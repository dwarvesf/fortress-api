# Unit Test Plan: HTTP Middleware Monitoring (SPEC-001)

**Target**: pkg/middleware/prometheus.go and pkg/metrics/http.go  
**Test Count**: 25 unit tests  
**Coverage Target**: 95%+  
**Performance Target**: Each test <100ms  

## Test Categories

### 1. Middleware Initialization Tests

#### Test: TestNewPrometheusMiddleware_DefaultConfig
```go
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
```

#### Test: TestNewPrometheusMiddleware_CustomConfig
```go
func TestNewPrometheusMiddleware_CustomConfig(t *testing.T) {
    config := &PrometheusConfig{
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
```

### 2. Middleware Handler Tests

#### Test: TestPrometheusHandler_Enabled_Success
```go
func TestPrometheusHandler_Enabled_Success(t *testing.T) {
    // Reset Prometheus registry
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
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
    
    // Verify metrics were recorded
    requestCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
    assert.Equal(t, 1.0, requestCount)
    
    // Verify duration was recorded
    durationCount := testutil.ToFloat64(HTTPRequestDuration.WithLabelValues("GET", "/test"))
    assert.Equal(t, 1.0, durationCount)
}
```

#### Test: TestPrometheusHandler_Disabled_NoMetrics
```go
func TestPrometheusHandler_Disabled_NoMetrics(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: false})
    
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
    requestCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
    assert.Equal(t, 0.0, requestCount)
}
```

#### Test: TestPrometheusHandler_ExcludedPaths
```go
func TestPrometheusHandler_ExcludedPaths(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    config := &PrometheusConfig{
        Enabled:      true,
        ExcludePaths: []string{"/metrics", "/health", "/debug"},
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
        
        count := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("GET", test.path, "200"))
        assert.Equal(t, test.expectedCount, count, "Path: %s", test.path)
    }
}
```

### 3. Sampling Tests

#### Test: TestPrometheusHandler_SamplingRate
```go
func TestPrometheusHandler_SamplingRate(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    // Set seed for deterministic testing
    rand.Seed(12345)
    
    config := &PrometheusConfig{
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
    sampledCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("GET", "/test", "200"))
    expectedCount := float64(totalRequests) * 0.5
    tolerance := expectedCount * 0.1 // 10% tolerance
    
    assert.InDelta(t, expectedCount, sampledCount, tolerance)
}
```

#### Test: TestShouldSample_RateCalculation
```go
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
        config := &PrometheusConfig{SampleRate: test.sampleRate}
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
```

### 4. Path Normalization Tests

#### Test: TestNormalizeEndpoint_DefaultBehavior
```go
func TestNormalizeEndpoint_DefaultBehavior(t *testing.T) {
    config := &PrometheusConfig{NormalizePaths: true}
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
```

#### Test: TestNormalizeEndpoint_Disabled
```go
func TestNormalizeEndpoint_Disabled(t *testing.T) {
    config := &PrometheusConfig{NormalizePaths: false}
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
```

### 5. Metric Recording Tests

#### Test: TestRecordMetrics_RequestSuccess
```go
func TestRecordMetrics_RequestSuccess(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("POST", "/api/v1/test", strings.NewReader("test body"))
    c.Request.ContentLength = 9 // len("test body")
    
    start := time.Now().Add(-100 * time.Millisecond) // Simulate 100ms duration
    middleware.recordMetrics(c, start, 9)
    
    // Verify request count metric
    requestCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("POST", "/api/v1/test", "200"))
    assert.Equal(t, 1.0, requestCount)
    
    // Verify duration was recorded (should be > 0.1 seconds)
    durationCount := testutil.ToFloat64(HTTPRequestDuration.WithLabelValues("POST", "/api/v1/test"))
    assert.Equal(t, 1.0, durationCount)
    
    // Verify request size was recorded
    sizeCount := testutil.ToFloat64(HTTPRequestSize.WithLabelValues("POST", "/api/v1/test"))
    assert.Equal(t, 1.0, sizeCount)
}
```

#### Test: TestRecordMetrics_RequestError
```go
func TestRecordMetrics_RequestError(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/nonexistent", nil)
    
    // Simulate 404 response
    c.AbortWithStatusJSON(404, gin.H{"error": "not found"})
    
    start := time.Now().Add(-50 * time.Millisecond)
    middleware.recordMetrics(c, start, 0)
    
    // Verify error status was recorded
    requestCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("GET", "/api/v1/nonexistent", "404"))
    assert.Equal(t, 1.0, requestCount)
    
    // Verify duration was still recorded
    durationCount := testutil.ToFloat64(HTTPRequestDuration.WithLabelValues("GET", "/api/v1/nonexistent"))
    assert.Equal(t, 1.0, durationCount)
}
```

### 6. In-Flight Request Tests

#### Test: TestInFlightRequests_Tracking
```go
func TestInFlightRequests_Tracking(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(middleware.Handler())
    r.GET("/slow", func(c *gin.Context) {
        // Check in-flight counter during request processing
        inFlight := testutil.ToFloat64(HTTPRequestsInFlight)
        assert.Equal(t, 1.0, inFlight)
        
        time.Sleep(10 * time.Millisecond)
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Initial in-flight should be 0
    initialInFlight := testutil.ToFloat64(HTTPRequestsInFlight)
    assert.Equal(t, 0.0, initialInFlight)
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/slow", nil)
    r.ServeHTTP(w, req)
    
    // After request completion, in-flight should be 0 again
    finalInFlight := testutil.ToFloat64(HTTPRequestsInFlight)
    assert.Equal(t, 0.0, finalInFlight)
    
    assert.Equal(t, 200, w.Code)
}
```

#### Test: TestInFlightRequests_ConcurrentRequests
```go
func TestInFlightRequests_ConcurrentRequests(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(middleware.Handler())
    
    var maxInFlight int64
    r.GET("/concurrent", func(c *gin.Context) {
        current := testutil.ToFloat64(HTTPRequestsInFlight)
        if int64(current) > maxInFlight {
            maxInFlight = int64(current)
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
    assert.Equal(t, int64(concurrentCount), maxInFlight)
    
    // Final in-flight should be 0
    finalInFlight := testutil.ToFloat64(HTTPRequestsInFlight)
    assert.Equal(t, 0.0, finalInFlight)
}
```

### 7. Response Size Tests

#### Test: TestRecordMetrics_ResponseSize
```go
func TestRecordMetrics_ResponseSize(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
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
    sizeCount := testutil.ToFloat64(HTTPResponseSize.WithLabelValues("GET", "/response-test"))
    assert.Equal(t, 1.0, sizeCount)
    
    // Response should be non-zero size
    assert.True(t, w.Body.Len() > 0)
}
```

### 8. Error Handling Tests

#### Test: TestPrometheusHandler_MetricRegistrationError
```go
func TestPrometheusHandler_MetricRegistrationError(t *testing.T) {
    // This test verifies graceful handling when metrics can't be registered
    // In practice, this might happen if metrics with the same name already exist
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    // Create middleware without proper metric setup to simulate error
    middleware := &PrometheusMiddleware{
        config: &PrometheusConfig{Enabled: true},
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
```

### 9. Configuration Validation Tests

#### Test: TestPrometheusConfig_Validation
```go
func TestPrometheusConfig_Validation(t *testing.T) {
    tests := []struct {
        name        string
        config      *PrometheusConfig
        expectValid bool
    }{
        {
            name: "valid_config",
            config: &PrometheusConfig{
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
            config: &PrometheusConfig{
                Enabled:    true,
                SampleRate: -0.1,
            },
            expectValid: false,
        },
        {
            name: "invalid_sample_rate_over_one",
            config: &PrometheusConfig{
                Enabled:    true,
                SampleRate: 1.5,
            },
            expectValid: false,
        },
        {
            name: "zero_max_endpoints",
            config: &PrometheusConfig{
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
```

### 10. Performance Tests

#### Test: BenchmarkPrometheusHandler_Enabled
```go
func BenchmarkPrometheusHandler_Enabled(b *testing.B) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: true})
    
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
```

#### Test: BenchmarkPrometheusHandler_Disabled  
```go
func BenchmarkPrometheusHandler_Disabled(b *testing.B) {
    middleware := NewPrometheusMiddleware(&PrometheusConfig{Enabled: false})
    
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
```

## Test Utilities

### Setup Helper Functions

```go
func setupTestMetrics() {
    // Re-create metrics for testing
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "http",
            Name:      "requests_total",
            Help:      "Total number of HTTP requests processed",
        },
        []string{"method", "endpoint", "status"},
    )
    
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "http", 
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration distribution",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )
    
    HTTPRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "http",
            Name:      "request_size_bytes",
            Help:      "HTTP request size distribution",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "endpoint"},
    )
    
    HTTPResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "http",
            Name:      "response_size_bytes",
            Help:      "HTTP response size distribution", 
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "endpoint"},
    )
    
    HTTPRequestsInFlight = promauto.NewGauge(
        prometheus.GaugeOpts{
            Namespace: "fortress",
            Subsystem: "http",
            Name:      "requests_in_flight",
            Help:      "Number of HTTP requests currently being processed",
        },
    )
}
```

## Performance Criteria

- **Test Execution**: Each unit test must complete in <100ms
- **Memory Usage**: No memory leaks in middleware tests
- **CPU Overhead**: Monitoring should add <5% CPU overhead to test execution
- **Concurrent Safety**: All tests must be safe for parallel execution

## Success Criteria

- **Code Coverage**: 95%+ line coverage for middleware and metrics packages
- **Test Reliability**: 0% flaky tests, consistent results
- **Performance**: Monitoring overhead <2% validated through benchmarks
- **Error Handling**: Graceful degradation tested for all failure scenarios

---

**Test Implementation Priority**: High  
**Estimated Implementation Time**: 16-20 hours  
**Dependencies**: Prometheus client library, testify framework  
**Review Requirements**: Senior backend engineer approval required