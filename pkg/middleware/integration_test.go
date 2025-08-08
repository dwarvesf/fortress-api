package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/handler/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

func TestIntegration_PrometheusMiddleware_WithMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup middleware with monitoring enabled
	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: true,
		ExcludePaths:   []string{"/healthz"},
	}
	middleware := NewPrometheusMiddleware(config)

	// Setup routes similar to the actual application
	r := gin.New()
	r.Use(middleware.Handler())

	// Add metrics endpoint
	metricsHandler := metrics.New()
	r.GET("/metrics", metricsHandler.Metrics)

	// Add test API endpoints
	r.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test response"})
	})
	r.POST("/api/v1/users", func(c *gin.Context) {
		c.JSON(201, gin.H{"id": 123, "status": "created"})
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Make some API requests to generate metrics
	testCases := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/api/v1/test", 200},
		{"POST", "/api/v1/users", 201},
		{"GET", "/api/v1/test", 200}, // Duplicate request
		{"GET", "/healthz", 200},     // Should be excluded
	}

	for _, tc := range testCases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, tc.status, w.Code, "Path: %s", tc.path)
	}

	// Now fetch metrics and verify they contain expected data
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	metricsOutput := w.Body.String()

	// Verify that metrics are in Prometheus format
	assert.Contains(t, metricsOutput, "fortress_http_requests_total")
	assert.Contains(t, metricsOutput, "fortress_http_request_duration_seconds")

	// Verify that API requests were tracked (but healthz was excluded)
	assert.Contains(t, metricsOutput, `method="GET"`)
	assert.Contains(t, metricsOutput, `method="POST"`)
	assert.Contains(t, metricsOutput, `endpoint="/api/v1/test"`)
	assert.Contains(t, metricsOutput, `endpoint="/api/v1/users"`)

	// Verify that excluded paths are not present
	assert.NotContains(t, metricsOutput, `endpoint="/healthz"`)

	// Verify that we have the expected counter values
	// We made 2 GET requests to /api/v1/test
	assert.Contains(t, metricsOutput, `fortress_http_requests_total{endpoint="/api/v1/test",method="GET",status="200"} 2`)

	// We made 1 POST request to /api/v1/users  
	assert.Contains(t, metricsOutput, `fortress_http_requests_total{endpoint="/api/v1/users",method="POST",status="201"} 1`)

	t.Logf("Metrics output sample:\n%s", metricsOutput[:500]) // Log first 500 chars for debugging
}

func TestIntegration_PrometheusMiddleware_PathNormalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &monitoring.PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: true,
	}
	middleware := NewPrometheusMiddleware(config)

	r := gin.New()
	r.Use(middleware.Handler())

	metricsHandler := metrics.New()
	r.GET("/metrics", metricsHandler.Metrics)

	// Simulate requests to dynamic endpoints
	r.GET("/api/v1/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	// Make requests with different IDs
	testPaths := []string{
		"/api/v1/users/123",
		"/api/v1/users/456",
		"/api/v1/users/789",
	}

	for _, path := range testPaths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "Path: %s", path)
	}

	// Fetch metrics
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	metricsOutput := w.Body.String()

	// Verify that paths were normalized to a single endpoint
	normalizedCount := strings.Count(metricsOutput, `endpoint="/api/v1/users/:id"`)
	assert.True(t, normalizedCount > 0, "Should have normalized endpoint")

	// Verify that we don't have individual numbered paths
	assert.NotContains(t, metricsOutput, `endpoint="/api/v1/users/123"`)
	assert.NotContains(t, metricsOutput, `endpoint="/api/v1/users/456"`)
	assert.NotContains(t, metricsOutput, `endpoint="/api/v1/users/789"`)

	// Verify total count for normalized endpoint is 3
	assert.Contains(t, metricsOutput, `fortress_http_requests_total{endpoint="/api/v1/users/:id",method="GET",status="200"} 3`)
}