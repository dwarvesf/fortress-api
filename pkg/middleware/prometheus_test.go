package middleware

import (
	"math/rand"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

func setupTestMetrics() {
	// Reset Prometheus registry for testing
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
}

func TestNewPrometheusMiddleware_DefaultConfig(t *testing.T) {
	middleware := NewPrometheusMiddleware(nil)
	
	assert.NotNil(t, middleware)
	assert.True(t, middleware.config.Enabled)
	assert.Equal(t, 1.0, middleware.config.SampleRate)
	assert.True(t, middleware.config.NormalizePaths)
	assert.Equal(t, 100, middleware.config.MaxEndpoints)
	assert.Contains(t, middleware.config.ExcludePaths, "/metrics")
	assert.Contains(t, middleware.config.ExcludePaths, "/healthz")
}

func TestNewPrometheusMiddleware_CustomConfig(t *testing.T) {
	config := &monitoring.PrometheusConfig{
		Enabled:        false,
		SampleRate:     0.5,
		NormalizePaths: false,
		MaxEndpoints:   50,
		ExcludePaths:   []string{"/custom"},
	}
	
	middleware := NewPrometheusMiddleware(config)
	
	assert.NotNil(t, middleware)
	assert.False(t, middleware.config.Enabled)
	assert.Equal(t, 0.5, middleware.config.SampleRate)
	assert.False(t, middleware.config.NormalizePaths)
	assert.Equal(t, 50, middleware.config.MaxEndpoints)
	assert.Contains(t, middleware.config.ExcludePaths, "/custom")
}

func TestPrometheusHandler_Enabled_Success(t *testing.T) {
	// Clean setup
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
	
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: false,
		ExcludePaths:   []string{},
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	// Verify metrics were recorded using the counter directly
	counter := metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200")
	metric := &dto.Metric{}
	err := counter.Write(metric)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, metric.Counter.GetValue())
	
	// Verify duration was recorded (check that histogram has samples)
	durationHistogram := metrics.HTTPRequestDuration.WithLabelValues("GET", "/test").(prometheus.Histogram)
	durationMetric := &dto.Metric{}
	err = durationHistogram.Write(durationMetric)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), durationMetric.GetHistogram().GetSampleCount())
}

func TestPrometheusHandler_Disabled_NoMetrics(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupTestMetrics()
	
	middleware := NewPrometheusMiddleware(&monitoring.PrometheusConfig{Enabled: false})
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	// Verify no metrics were recorded
	requestCount := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
	assert.Equal(t, 0.0, requestCount)
}

func TestPrometheusHandler_ExcludedPaths(t *testing.T) {
	// Clean setup
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
	
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: false,
		ExcludePaths:   []string{"/metrics", "/health", "/debug"},
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/metrics", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	r.GET("/debug/vars", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	r.GET("/api/test", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	
	tests := []struct {
		path           string
		shouldTrack    bool
		expectedCount  float64
	}{
		{"/metrics", false, 0.0},
		{"/health", false, 0.0}, 
		{"/debug/vars", false, 0.0},
		{"/api/test", true, 1.0},
	}
	
	for _, test := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", test.path, nil)
		r.ServeHTTP(w, req)
		
		assert.Equal(t, 200, w.Code)
		
		counter := metrics.HTTPRequestsTotal.WithLabelValues("GET", test.path, "200")
		metric := &dto.Metric{}
		err := counter.Write(metric)
		assert.NoError(t, err, "Path: %s", test.path)
		
		actualCount := 0.0
		if metric.Counter != nil {
			actualCount = metric.Counter.GetValue()
		}
		assert.Equal(t, test.expectedCount, actualCount, "Path: %s", test.path)
	}
}

func TestPrometheusHandler_SamplingRate(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupTestMetrics()
	
	// Set seed for deterministic testing
	rand.Seed(12345)
	
	config := &monitoring.PrometheusConfig{
		Enabled:    true,
		SampleRate: 0.5,
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// Make multiple requests
	totalRequests := 1000
	for i := 0; i < totalRequests; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}
	
	// Check that approximately 50% were sampled (allow for randomness)
	sampledCount := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
	expectedCount := float64(totalRequests) * 0.5
	tolerance := expectedCount * 0.1 // 10% tolerance
	
	assert.InDelta(t, expectedCount, sampledCount, tolerance)
}

func TestShouldSample_RateCalculation(t *testing.T) {
	tests := []struct {
		sampleRate    float64
		iterations    int
		tolerance     float64
	}{
		{0.0, 1000, 0.05},   // No sampling
		{0.25, 1000, 0.05},  // 25% sampling
		{0.5, 1000, 0.05},   // 50% sampling  
		{1.0, 1000, 0.01},   // 100% sampling
	}
	
	for _, test := range tests {
		config := &monitoring.PrometheusConfig{SampleRate: test.sampleRate}
		middleware := NewPrometheusMiddleware(config)
		
		sampledCount := 0
		for i := 0; i < test.iterations; i++ {
			if middleware.shouldSample() {
				sampledCount++
			}
		}
		
		expectedRate := test.sampleRate
		actualRate := float64(sampledCount) / float64(test.iterations)
		
		assert.InDelta(t, expectedRate, actualRate, test.tolerance,
			"Sample rate %.2f: expected %.2f, got %.2f", 
			test.sampleRate, expectedRate, actualRate)
	}
}

func TestNormalizeEndpoint_DefaultBehavior(t *testing.T) {
	config := &monitoring.PrometheusConfig{NormalizePaths: true}
	middleware := NewPrometheusMiddleware(config)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"", "unknown"},
		{"/api/v1/employees/123", "/api/v1/employees/:id"},
		{"/api/v1/projects/456/members", "/api/v1/projects/:id/members"},
		{"/api/v1/projects/456/members/789", "/api/v1/projects/:id/members/:id"},
		{"/static/js/app.js", "/static/js/app.js"}, // No normalization for static
		{"/api/v1/metadata/stacks", "/api/v1/metadata/stacks"}, // No params to normalize
	}
	
	for _, test := range tests {
		result := middleware.normalizeEndpoint(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestNormalizeEndpoint_Disabled(t *testing.T) {
	config := &monitoring.PrometheusConfig{NormalizePaths: false}
	middleware := NewPrometheusMiddleware(config)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"", "unknown"},
		{"/api/v1/employees/123", "/api/v1/employees/123"}, // No normalization
		{"/api/v1/projects/456/members", "/api/v1/projects/456/members"},
		{"/exact/path", "/exact/path"},
	}
	
	for _, test := range tests {
		result := middleware.normalizeEndpoint(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestRecordMetrics_RequestSuccess(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupTestMetrics()
	
	middleware := NewPrometheusMiddleware(&monitoring.PrometheusConfig{Enabled: true})
	
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/api/v1/test", strings.NewReader("test body"))
	c.Request.ContentLength = 9 // len("test body")
	
	start := time.Now().Add(-100 * time.Millisecond) // Simulate 100ms duration
	middleware.recordMetrics(c, start, 9, 0)
	
	// Verify request count metric
	requestCount := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/test", "200"))
	assert.Equal(t, 1.0, requestCount)
	
	// Verify duration was recorded (check that histogram has samples)
	durationHistogram := metrics.HTTPRequestDuration.WithLabelValues("POST", "/api/v1/test").(prometheus.Histogram)
	durationMetric := &dto.Metric{}
	err := durationHistogram.Write(durationMetric)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), durationMetric.GetHistogram().GetSampleCount())
	
	// Verify request size was recorded
	sizeHistogram := metrics.HTTPRequestSize.WithLabelValues("POST", "/api/v1/test").(prometheus.Histogram)
	sizeMetric := &dto.Metric{}
	err = sizeHistogram.Write(sizeMetric)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), sizeMetric.GetHistogram().GetSampleCount())
}

func TestRecordMetrics_RequestError(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupTestMetrics()
	
	middleware := NewPrometheusMiddleware(&monitoring.PrometheusConfig{Enabled: true})
	
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
	
	// Simulate 404 response
	c.AbortWithStatusJSON(404, gin.H{"error": "not found"})
	
	start := time.Now().Add(-50 * time.Millisecond)
	middleware.recordMetrics(c, start, 0, 0)
	
	// Verify error status was recorded
	requestCount := testutil.ToFloat64(metrics.HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/nonexistent", "404"))
	assert.Equal(t, 1.0, requestCount)
	
	// Verify duration was still recorded
	durationHistogram := metrics.HTTPRequestDuration.WithLabelValues("GET", "/api/v1/nonexistent").(prometheus.Histogram)
	durationMetric := &dto.Metric{}
	err := durationHistogram.Write(durationMetric)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), durationMetric.GetHistogram().GetSampleCount())
}

func TestInFlightRequests_Tracking(t *testing.T) {
	// Clean setup
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
	
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: false,
		ExcludePaths:   []string{},
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/slow", func(c *gin.Context) {
		// Check in-flight counter during request processing
		gauge := metrics.HTTPRequestsInFlight
		metric := &dto.Metric{}
		err := gauge.Write(metric)
		assert.NoError(t, err)
		assert.Equal(t, 1.0, metric.Gauge.GetValue())
		
		time.Sleep(10 * time.Millisecond)
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// Initial in-flight should be 0
	initialGauge := metrics.HTTPRequestsInFlight
	initialMetric := &dto.Metric{}
	err := initialGauge.Write(initialMetric)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, initialMetric.Gauge.GetValue())
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/slow", nil)
	r.ServeHTTP(w, req)
	
	// After request completion, in-flight should be 0 again
	finalGauge := metrics.HTTPRequestsInFlight
	finalMetric := &dto.Metric{}
	err = finalGauge.Write(finalMetric)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, finalMetric.Gauge.GetValue())
	
	assert.Equal(t, 200, w.Code)
}

func TestInFlightRequests_ConcurrentRequests(t *testing.T) {
	// Clean setup
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
	
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: false,
		ExcludePaths:   []string{},
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	
	var maxInFlight int64
	r.GET("/concurrent", func(c *gin.Context) {
		gauge := metrics.HTTPRequestsInFlight
		metric := &dto.Metric{}
		err := gauge.Write(metric)
		if err == nil && metric.Gauge != nil {
			current := int64(metric.Gauge.GetValue())
			for {
				max := atomic.LoadInt64(&maxInFlight)
				if current <= max || atomic.CompareAndSwapInt64(&maxInFlight, max, current) {
					break
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// Launch concurrent requests
	concurrentCount := 5
	var wg sync.WaitGroup
	wg.Add(concurrentCount)
	
	for i := 0; i < concurrentCount; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/concurrent", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, 200, w.Code)
		}()
	}
	
	wg.Wait()
	
	// Verify maximum concurrent requests were tracked
	assert.Equal(t, int64(concurrentCount), atomic.LoadInt64(&maxInFlight))
	
	// Final in-flight should be 0
	finalGauge := metrics.HTTPRequestsInFlight
	finalMetric := &dto.Metric{}
	err := finalGauge.Write(finalMetric)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, finalMetric.Gauge.GetValue())
}

func TestRecordMetrics_ResponseSize(t *testing.T) {
	// Clean setup
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	metrics.InitHTTPMetrics(registry)
	
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: false,
		ExcludePaths:   []string{},
	}
	middleware := NewPrometheusMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Handler())
	
	responseBody := `{"message": "test response", "data": ["item1", "item2", "item3"]}`
	r.GET("/response-test", func(c *gin.Context) {
		c.JSON(200, responseBody)
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/response-test", nil)
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 200, w.Code)
	
	// Verify response size was recorded
	responseHistogram := metrics.HTTPResponseSize.WithLabelValues("GET", "/response-test").(prometheus.Histogram)
	responseMetric := &dto.Metric{}
	err := responseHistogram.Write(responseMetric)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), responseMetric.GetHistogram().GetSampleCount())
	
	// Response should be non-zero size
	assert.True(t, w.Body.Len() > 0)
}

func TestPrometheusHandler_MetricRegistrationError(t *testing.T) {
	// This test verifies graceful handling when metrics can't be registered
	// In practice, this might happen if metrics with the same name already exist
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Create middleware without proper metric setup to simulate error
	middleware := &PrometheusMiddleware{
		config: &monitoring.PrometheusConfig{Enabled: true},
	}
	
	r.Use(middleware.Handler())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	
	// Should not panic even with metric errors
	assert.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})
	
	assert.Equal(t, 200, w.Code)
}

func TestPrometheusConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *monitoring.PrometheusConfig
		expectValid bool
	}{
		{
			name: "valid_config",
			config: &monitoring.PrometheusConfig{
				Enabled:        true,
				SampleRate:     0.5,
				NormalizePaths: true,
				MaxEndpoints:   100,
				ExcludePaths:   []string{"/metrics"},
			},
			expectValid: true,
		},
		{
			name: "invalid_sample_rate_negative",
			config: &monitoring.PrometheusConfig{
				Enabled:    true,
				SampleRate: -0.1,
			},
			expectValid: false,
		},
		{
			name: "invalid_sample_rate_over_one",
			config: &monitoring.PrometheusConfig{
				Enabled:    true,
				SampleRate: 1.5,
			},
			expectValid: false,
		},
		{
			name: "zero_max_endpoints",
			config: &monitoring.PrometheusConfig{
				Enabled:      true,
				MaxEndpoints: 0,
			},
			expectValid: false,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			middleware := NewPrometheusMiddleware(test.config)
			
			if test.expectValid {
				assert.NotNil(t, middleware)
				assert.NotNil(t, middleware.config)
			} else {
				// Should use defaults for invalid values
				assert.NotNil(t, middleware)
				assert.True(t, middleware.config.SampleRate >= 0.0)
				assert.True(t, middleware.config.SampleRate <= 1.0)
				assert.True(t, middleware.config.MaxEndpoints > 0)
			}
		})
	}
}

// Performance benchmarks
func BenchmarkPrometheusHandler_Enabled(b *testing.B) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	setupTestMetrics()
	
	middleware := NewPrometheusMiddleware(&monitoring.PrometheusConfig{Enabled: true})
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/bench", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/bench", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkPrometheusHandler_Disabled(b *testing.B) {
	middleware := NewPrometheusMiddleware(&monitoring.PrometheusConfig{Enabled: false})
	
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Handler())
	r.GET("/bench", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/bench", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w.Body.Reset()
		r.ServeHTTP(w, req)
	}
}