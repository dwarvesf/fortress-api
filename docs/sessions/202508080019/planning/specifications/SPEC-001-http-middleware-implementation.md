# SPEC-001: HTTP Middleware Implementation

**Status**: Ready for Implementation  
**Priority**: High  
**Estimated Effort**: 8-12 hours  
**Dependencies**: None  

## Overview

Implement comprehensive HTTP request monitoring middleware for the Fortress API using Prometheus metrics collection. This middleware will integrate seamlessly with the existing Gin framework and provide the foundation for all HTTP-level monitoring.

## Requirements

### Functional Requirements
- **FR-1**: Collect HTTP request metrics (count, duration, size) for all endpoints
- **FR-2**: Support endpoint normalization to prevent metric cardinality explosion
- **FR-3**: Integrate with existing middleware chain without breaking changes
- **FR-4**: Provide configurable sampling and filtering options
- **FR-5**: Support both development and production environments

### Non-Functional Requirements  
- **NFR-1**: Middleware overhead <5ms per request under normal load
- **NFR-2**: Memory usage <50MB for metric storage
- **NFR-3**: Graceful degradation if Prometheus is unavailable
- **NFR-4**: Zero breaking changes to existing request/response flow

## Technical Specification

### 1. Package Structure

```go
// pkg/metrics/http.go - HTTP metric definitions
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
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
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 10MB
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
)
```

### 2. Middleware Implementation

```go
// pkg/middleware/prometheus.go - Main middleware implementation
package middleware

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/metrics"
)

type PrometheusMiddleware struct {
    config *PrometheusConfig
}

type PrometheusConfig struct {
    Enabled           bool     `mapstructure:"enabled"`
    ExcludePaths      []string `mapstructure:"exclude_paths"`
    SampleRate        float64  `mapstructure:"sample_rate"`
    MaxEndpoints      int      `mapstructure:"max_endpoints"`
    NormalizePaths    bool     `mapstructure:"normalize_paths"`
}

func NewPrometheusMiddleware(cfg *PrometheusConfig) *PrometheusMiddleware {
    if cfg == nil {
        cfg = &PrometheusConfig{
            Enabled:        true,
            ExcludePaths:   []string{"/metrics", "/healthz", "/swagger"},
            SampleRate:     1.0,
            MaxEndpoints:   100,
            NormalizePaths: true,
        }
    }
    
    return &PrometheusMiddleware{config: cfg}
}

func (pm *PrometheusMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !pm.config.Enabled {
            c.Next()
            return
        }
        
        // Skip excluded paths
        if pm.shouldExclude(c.Request.URL.Path) {
            c.Next()
            return
        }
        
        // Sample requests if configured
        if pm.config.SampleRate < 1.0 && !pm.shouldSample() {
            c.Next()
            return
        }
        
        start := time.Now()
        
        // Record in-flight requests
        metrics.HTTPRequestsInFlight.Inc()
        defer metrics.HTTPRequestsInFlight.Dec()
        
        // Get request size
        requestSize := c.Request.ContentLength
        if requestSize < 0 {
            requestSize = 0
        }
        
        c.Next()
        
        // Record completed request metrics
        pm.recordMetrics(c, start, requestSize)
    }
}

func (pm *PrometheusMiddleware) recordMetrics(c *gin.Context, start time.Time, requestSize int64) {
    duration := time.Since(start).Seconds()
    method := c.Request.Method
    endpoint := pm.normalizeEndpoint(c.FullPath())
    status := strconv.Itoa(c.Writer.Status())
    
    // Record core metrics
    metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
    metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
    
    // Record size metrics
    if requestSize > 0 {
        metrics.HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
    }
    
    responseSize := float64(c.Writer.Size())
    if responseSize > 0 {
        metrics.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(responseSize)
    }
}

func (pm *PrometheusMiddleware) normalizeEndpoint(path string) string {
    if path == "" {
        return "unknown"
    }
    
    if !pm.config.NormalizePaths {
        return path
    }
    
    // Use Gin route template if available (e.g., "/api/v1/employees/:id")
    // This prevents cardinality explosion from dynamic path parameters
    return path
}

func (pm *PrometheusMiddleware) shouldExclude(path string) bool {
    for _, excluded := range pm.config.ExcludePaths {
        if strings.Contains(path, excluded) {
            return true
        }
    }
    return false
}

func (pm *PrometheusMiddleware) shouldSample() bool {
    return rand.Float64() < pm.config.SampleRate
}
```

### 3. Router Integration

```go
// pkg/routes/routes.go - Integration with existing router
// MODIFICATION to existing NewRoutes function

func NewRoutes(cfg *config.Config, svc *service.Service, s *store.Store, repo store.DBRepo, worker *worker.Worker, logger logger.Logger) *gin.Engine {
    // Existing setup
    docs.SwaggerInfo.Title = "Swagger API"
    docs.SwaggerInfo.Description = "This is a swagger for API."
    docs.SwaggerInfo.Version = "1.0"
    docs.SwaggerInfo.Schemes = []string{"https", "http"}
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
    pprof.Register(r)

    ctrl := controller.New(s, repo, svc, worker, logger, cfg)
    h := handler.New(s, repo, svc, ctrl, worker, logger, cfg)

    // MODIFICATION: Add Prometheus middleware before existing middleware
    if cfg.Monitoring.Enabled {
        prometheusMiddleware := middleware.NewPrometheusMiddleware(&middleware.PrometheusConfig{
            Enabled:        cfg.Monitoring.Enabled,
            ExcludePaths:   []string{"/metrics", "/healthz", "/swagger"},
            SampleRate:     cfg.Monitoring.SampleRate,
            NormalizePaths: true,
        })
        r.Use(prometheusMiddleware.Handler())
    }

    // Existing middleware continues unchanged
    r.Use(
        gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"),
        gin.Recovery(),
    )

    // config CORS (existing)
    setupCORS(r, cfg)

    r.GET("/healthz", h.Healthcheck.Healthz)
    
    // NEW: Add metrics endpoint
    if cfg.Monitoring.Enabled {
        r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }

    // use ginSwagger middleware to serve the API docs (existing)
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // load API here (existing)
    loadV1Routes(r, h, repo, s, cfg)

    return r
}
```

### 4. Configuration Integration

```go
// pkg/config/config.go - Add monitoring configuration
// MODIFICATION to existing Config struct

type Config struct {
    // ... existing fields ...
    
    // NEW: Monitoring configuration
    Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

type MonitoringConfig struct {
    Enabled    bool    `mapstructure:"enabled" default:"true"`
    SampleRate float64 `mapstructure:"sample_rate" default:"1.0"`
    
    // HTTP middleware specific
    ExcludePaths      []string `mapstructure:"exclude_paths"`
    NormalizePaths    bool     `mapstructure:"normalize_paths" default:"true"`
    MaxEndpoints      int      `mapstructure:"max_endpoints" default:"100"`
    
    // Endpoints
    MetricsPath string `mapstructure:"metrics_path" default:"/metrics"`
    HealthPath  string `mapstructure:"health_path" default:"/health"`
}
```

### 5. Dependencies

```go
// go.mod additions
require (
    github.com/prometheus/client_golang v1.17.0
)
```

## Testing Strategy

### 1. Unit Tests

```go
// pkg/middleware/prometheus_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/testutil"
    "github.com/stretchr/testify/assert"
)

func TestPrometheusMiddleware(t *testing.T) {
    // Reset metrics for testing
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    middleware := NewPrometheusMiddleware(nil)
    r.Use(middleware.Handler())
    
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Make test request
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    r.ServeHTTP(w, req)
    
    // Verify response
    assert.Equal(t, 200, w.Code)
    
    // Verify metrics were recorded
    // Note: In actual implementation, would use testutil.ToFloat64()
    // to verify metric values
}

func TestEndpointNormalization(t *testing.T) {
    middleware := NewPrometheusMiddleware(nil)
    
    tests := []struct {
        input    string
        expected string
    }{
        {"/api/v1/employees/123", "/api/v1/employees/:id"},
        {"/api/v1/projects/456/members", "/api/v1/projects/:id/members"},
        {"", "unknown"},
    }
    
    for _, test := range tests {
        result := middleware.normalizeEndpoint(test.input)
        assert.Equal(t, test.expected, result)
    }
}

func TestExcludedPaths(t *testing.T) {
    config := &PrometheusConfig{
        ExcludePaths: []string{"/metrics", "/healthz"},
    }
    middleware := NewPrometheusMiddleware(config)
    
    assert.True(t, middleware.shouldExclude("/metrics"))
    assert.True(t, middleware.shouldExclude("/healthz"))
    assert.False(t, middleware.shouldExclude("/api/v1/test"))
}

func TestSampling(t *testing.T) {
    config := &PrometheusConfig{
        SampleRate: 0.5,
    }
    middleware := NewPrometheusMiddleware(config)
    
    // Test sampling over multiple iterations
    samples := 0
    iterations := 1000
    
    for i := 0; i < iterations; i++ {
        if middleware.shouldSample() {
            samples++
        }
    }
    
    // Should be approximately 50% (allowing for randomness)
    assert.InDelta(t, 500, samples, 50)
}
```

### 2. Integration Tests

```go
// tests/integration/monitoring_test.go
package integration

import (
    "net/http"
    "testing"
    "time"
    
    "github.com/dwarvesf/fortress-api/pkg/testhelper"
    "github.com/stretchr/testify/assert"
)

func TestMetricsEndpoint(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo *store.PostgresRepo) {
        // Set up test server with monitoring enabled
        router := setupTestRouter(repo, true)
        
        // Make some test requests
        testRequests := []struct {
            method string
            path   string
            status int
        }{
            {"GET", "/api/v1/metadata/stacks", 200},
            {"GET", "/api/v1/nonexistent", 404},
            {"POST", "/api/v1/auth", 401},
        }
        
        for _, req := range testRequests {
            w := httptest.NewRecorder()
            httpReq, _ := http.NewRequest(req.method, req.path, nil)
            router.ServeHTTP(w, httpReq)
        }
        
        // Check metrics endpoint
        w := httptest.NewRecorder()
        req, _ := http.NewRequest("GET", "/metrics", nil)
        router.ServeHTTP(w, req)
        
        assert.Equal(t, 200, w.Code)
        
        // Verify metrics are present
        body := w.Body.String()
        assert.Contains(t, body, "fortress_http_requests_total")
        assert.Contains(t, body, "fortress_http_request_duration_seconds")
        assert.Contains(t, body, "fortress_http_requests_in_flight")
    })
}

func TestPerformanceOverhead(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo *store.PostgresRepo) {
        // Test with monitoring disabled
        routerWithoutMetrics := setupTestRouter(repo, false)
        baselineTime := measureRequestTime(routerWithoutMetrics, "/api/v1/metadata/stacks")
        
        // Test with monitoring enabled
        routerWithMetrics := setupTestRouter(repo, true)
        metricsTime := measureRequestTime(routerWithMetrics, "/api/v1/metadata/stacks")
        
        // Overhead should be minimal (< 5ms)
        overhead := metricsTime - baselineTime
        assert.Less(t, overhead, 5*time.Millisecond)
    })
}

func measureRequestTime(router *gin.Engine, path string) time.Duration {
    iterations := 100
    start := time.Now()
    
    for i := 0; i < iterations; i++ {
        w := httptest.NewRecorder()
        req, _ := http.NewRequest("GET", path, nil)
        router.ServeHTTP(w, req)
    }
    
    return time.Since(start) / time.Duration(iterations)
}
```

### 3. Load Testing

```bash
# scripts/test-monitoring-performance.sh
#!/bin/bash

echo "Testing monitoring middleware performance..."

# Start test server
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Run load test without monitoring
echo "Baseline performance (no monitoring):"
curl -X POST localhost:8080/disable-monitoring
hey -n 10000 -c 50 http://localhost:8080/api/v1/metadata/stacks

# Run load test with monitoring
echo "Performance with monitoring enabled:"
curl -X POST localhost:8080/enable-monitoring  
hey -n 10000 -c 50 http://localhost:8080/api/v1/metadata/stacks

# Clean up
kill $SERVER_PID
```

## Deployment Strategy

### 1. Feature Flag Implementation

```go
// Deploy with feature flag to enable gradual rollout
func (pm *PrometheusMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check feature flag or configuration
        if !pm.config.Enabled || !pm.isEnabledForRequest(c) {
            c.Next()
            return
        }
        
        // ... existing middleware logic
    }
}

func (pm *PrometheusMiddleware) isEnabledForRequest(c *gin.Context) bool {
    // Could implement percentage-based rollout
    // or user-specific enablement
    return pm.config.Enabled
}
```

### 2. Staging Deployment

```yaml
# k8s/staging/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fortress-api-config
data:
  config.yaml: |
    monitoring:
      enabled: true
      sample_rate: 1.0
      exclude_paths:
        - "/metrics"
        - "/healthz" 
        - "/swagger"
      normalize_paths: true
```

### 3. Production Rollout Plan

**Phase 1: Canary Deployment (Week 1)**
- Deploy to 10% of production traffic
- Monitor performance impact
- Validate metric collection accuracy

**Phase 2: Gradual Rollout (Week 2)**  
- Increase to 50% of traffic
- Fine-tune configuration based on usage patterns
- Create initial dashboards

**Phase 3: Full Deployment (Week 3)**
- Enable for 100% of traffic
- Set up alerting rules
- Train team on metrics usage

## Success Criteria

### Performance Metrics
- **Middleware Overhead**: <2ms average per request
- **Memory Usage**: <50MB for metrics storage
- **CPU Overhead**: <1% additional CPU usage

### Functional Metrics
- **Metric Accuracy**: 99.9% accurate request counting
- **Cardinality Control**: <10,000 unique metric series
- **Reliability**: 99.9% uptime for metrics collection

### Operational Metrics
- **Dashboard Integration**: Metrics visible in Grafana within 30 seconds
- **Alert Integration**: Compatible with defined alerting rules
- **Team Adoption**: 90% of engineering team using metrics for debugging within 2 weeks

## Risk Mitigation

### High Cardinality Risk
- **Mitigation**: Implement endpoint normalization and path limits
- **Monitoring**: Alert if metric cardinality exceeds 50,000 series

### Performance Impact Risk
- **Mitigation**: Comprehensive load testing and gradual rollout
- **Monitoring**: Continuous performance benchmarking

### Memory Leak Risk
- **Mitigation**: Proper metric lifecycle management and testing
- **Monitoring**: Memory usage alerts and regular profiling

---

**Assignee**: Backend Engineering Team  
**Reviewer**: Senior Backend Engineer, DevOps Lead  
**Implementation Timeline**: 1 week  
**Testing Timeline**: 3-5 days  
**Deployment Timeline**: 2 weeks (staged rollout)