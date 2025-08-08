# Implementation Guide: Go Web API Monitoring with Prometheus/Grafana/Loki

**Reference**: monitoring-best-practices-research.md  
**Target**: Production Go web API with Gin, GORM, PostgreSQL  
**Scope**: Minimal complexity, maximum value monitoring implementation

## Quick Start Implementation

### Phase 1: Basic Prometheus Integration (Day 1)

#### 1.1 Dependencies
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp
go get github.com/zsais/go-gin-prometheus  # For Gin middleware
```

#### 1.2 Basic HTTP Metrics Setup
```go
// pkg/metrics/http.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP Request metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    HTTPRequestsInFlight = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Number of HTTP requests currently being processed",
        },
    )
)
```

#### 1.3 Gin Middleware Integration
```go
// pkg/middleware/prometheus.go
package middleware

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "your-project/pkg/metrics"
)

func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Increment in-flight requests
        metrics.HTTPRequestsInFlight.Inc()
        defer metrics.HTTPRequestsInFlight.Dec()
        
        // Process request
        c.Next()
        
        // Record metrics
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        // Normalize endpoint to prevent cardinality explosion
        endpoint := normalizeEndpoint(c.FullPath())
        
        metrics.HTTPRequestsTotal.WithLabelValues(
            c.Request.Method,
            endpoint,
            status,
        ).Inc()
        
        metrics.HTTPRequestDuration.WithLabelValues(
            c.Request.Method,
            endpoint,
        ).Observe(duration)
    }
}

// Normalize endpoint paths to reduce cardinality
func normalizeEndpoint(path string) string {
    if path == "" {
        return "unknown"
    }
    return path
}
```

#### 1.4 Router Setup
```go
// cmd/server/main.go or router setup
func setupRouter() *gin.Engine {
    r := gin.New()
    
    // Add Prometheus middleware
    r.Use(middleware.PrometheusMiddleware())
    
    // Standard middleware
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    
    // Expose metrics endpoint
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // Your API routes
    v1 := r.Group("/api/v1")
    {
        v1.GET("/health", healthCheck)
        // ... other routes
    }
    
    return r
}
```

### Phase 2: Database Monitoring (Day 2)

#### 2.1 GORM Prometheus Plugin
```go
// In your database initialization
import "gorm.io/plugin/prometheus"

func InitializeDatabase() *gorm.DB {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Add Prometheus plugin
    db.Use(prometheus.New(prometheus.Config{
        DBName:          "fortress_db",
        RefreshInterval: 15, // seconds
        Labels: map[string]string{
            "service": "fortress-api",
            "env":     os.Getenv("ENVIRONMENT"),
        },
    }))
    
    return db
}
```

#### 2.2 Custom Database Metrics
```go
// pkg/metrics/database.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Database operation metrics
    DatabaseOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "database_operations_total",
            Help: "Total database operations",
        },
        []string{"operation", "table", "result"},
    )
    
    DatabaseOperationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_operation_duration_seconds",
            Help: "Duration of database operations",
        },
        []string{"operation", "table"},
    )
)
```

### Phase 3: Business Logic Metrics (Day 3-5)

#### 3.1 Authentication Metrics
```go
// pkg/metrics/auth.go
package metrics

var (
    AuthenticationAttempts = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_attempts_total",
            Help: "Total authentication attempts",
        },
        []string{"method", "result", "reason"},
    )
    
    JWTValidationDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name: "jwt_validation_duration_seconds",
            Help: "JWT validation duration",
        },
    )
    
    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_sessions_total",
            Help: "Number of active user sessions",
        },
    )
)
```

#### 3.2 Integration in Auth Middleware
```go
// pkg/middleware/auth.go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        token := extractToken(c)
        if token == "" {
            metrics.AuthenticationAttempts.WithLabelValues(
                "jwt", "failure", "missing_token",
            ).Inc()
            c.AbortWithStatusJSON(401, gin.H{"error": "Missing token"})
            return
        }
        
        // Validate token
        claims, err := validateJWT(token)
        if err != nil {
            metrics.AuthenticationAttempts.WithLabelValues(
                "jwt", "failure", "invalid_token",
            ).Inc()
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        // Record successful validation
        metrics.AuthenticationAttempts.WithLabelValues(
            "jwt", "success", "",
        ).Inc()
        
        metrics.JWTValidationDuration.Observe(
            time.Since(start).Seconds(),
        )
        
        c.Set("user", claims)
        c.Next()
    }
}
```

### Phase 4: Kubernetes Deployment (Day 6-7)

#### 4.1 ServiceMonitor for Prometheus Discovery
```yaml
# k8s/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: fortress-api
  namespace: default
  labels:
    app: fortress-api
spec:
  selector:
    matchLabels:
      app: fortress-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
```

#### 4.2 Service Configuration
```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: fortress-api
  labels:
    app: fortress-api
spec:
  selector:
    app: fortress-api
  ports:
  - name: http
    port: 8080
    targetPort: 8080
```

#### 4.3 Deployment with Prometheus Annotations
```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fortress-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fortress-api
  template:
    metadata:
      labels:
        app: fortress-api
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: api
        image: fortress-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
```

### Phase 5: Alerting Rules (Day 8-10)

#### 5.1 Critical Alerting Rules
```yaml
# k8s/alerting-rules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: fortress-api-alerts
  namespace: default
spec:
  groups:
  - name: fortress-api
    rules:
    - alert: HighErrorRate
      expr: |
        (
          rate(http_requests_total{app="fortress-api",status=~"5.."}[5m])
          /
          rate(http_requests_total{app="fortress-api"}[5m])
        ) > 0.05
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High error rate on Fortress API"
        description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"
    
    - alert: HighLatency
      expr: |
        histogram_quantile(0.95,
          sum(rate(http_request_duration_seconds_bucket{app="fortress-api"}[5m])) by (le)
        ) > 0.5
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "High latency on Fortress API"
        description: "95th percentile latency is {{ $value }}s"
    
    - alert: DatabaseConnectionsHigh
      expr: |
        (
          gorm_dbstats_in_use{app="fortress-api"}
          /
          gorm_dbstats_max_open_connections{app="fortress-api"}
        ) > 0.8
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Database connections high"
        description: "Database connection usage is {{ $value | humanizePercentage }}"
    
    - alert: AuthenticationFailureSpike
      expr: |
        rate(auth_attempts_total{app="fortress-api",result="failure"}[5m]) > 10
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High authentication failure rate"
        description: "Authentication failures at {{ $value }} per second"
```

### Phase 6: Grafana Dashboard (Day 11-12)

#### 6.1 Essential Dashboard Panels

**HTTP Overview Panel**
```json
{
  "title": "HTTP Request Rate",
  "type": "stat",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total{app=\"fortress-api\"}[5m]))",
      "legendFormat": "Requests/sec"
    }
  ]
}
```

**Error Rate Panel**
```json
{
  "title": "HTTP Error Rate",
  "type": "stat",
  "targets": [
    {
      "expr": "sum(rate(http_requests_total{app=\"fortress-api\",status=~\"4..|5..\"}[5m])) / sum(rate(http_requests_total{app=\"fortress-api\"}[5m]))",
      "legendFormat": "Error Rate"
    }
  ]
}
```

**Latency Panel**
```json
{
  "title": "Response Time Percentiles",
  "type": "graph",
  "targets": [
    {
      "expr": "histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{app=\"fortress-api\"}[5m])) by (le))",
      "legendFormat": "50th percentile"
    },
    {
      "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{app=\"fortress-api\"}[5m])) by (le))",
      "legendFormat": "95th percentile"
    },
    {
      "expr": "histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{app=\"fortress-api\"}[5m])) by (le))",
      "legendFormat": "99th percentile"
    }
  ]
}
```

### Phase 7: Structured Logging with Loki (Day 13-15)

#### 7.1 Structured Logging Setup
```go
// pkg/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

var Logger *slog.Logger

func InitLogger() {
    // JSON for production, text for development
    var handler slog.Handler
    if os.Getenv("ENVIRONMENT") == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }
    
    Logger = slog.New(handler)
}

// Structured logging helpers
func LogHTTPRequest(method, path string, status int, duration float64, userID string) {
    Logger.Info("HTTP request",
        slog.String("method", method),
        slog.String("path", path),
        slog.Int("status", status),
        slog.Float64("duration_ms", duration*1000),
        slog.String("user_id", userID),
        slog.String("component", "http"),
    )
}

func LogAuthEvent(event, userID, reason string) {
    Logger.Info("Authentication event",
        slog.String("event", event),
        slog.String("user_id", userID),
        slog.String("reason", reason),
        slog.String("component", "auth"),
    )
}
```

#### 7.2 Logging Middleware
```go
// pkg/middleware/logging.go
func StructuredLogging() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        userID := getUserID(c) // Extract from JWT claims
        
        logger.LogHTTPRequest(
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            duration.Seconds(),
            userID,
        )
    }
}
```

## Testing and Validation

### Metrics Validation
```bash
# Test metrics endpoint
curl http://localhost:8080/metrics | grep fortress

# Validate specific metrics exist
curl -s http://localhost:8080/metrics | grep -E "(http_requests_total|http_request_duration)"
```

### Load Testing with Metrics
```bash
# Install hey for load testing
go install github.com/rakyll/hey@latest

# Generate load
hey -n 1000 -c 10 http://localhost:8080/api/v1/health

# Check metrics during load
watch -n 1 'curl -s http://localhost:8080/metrics | grep http_requests_in_flight'
```

### Alert Testing
```bash
# Simulate high error rate
for i in {1..100}; do
  curl -s http://localhost:8080/api/v1/nonexistent || true
done
```

## Production Checklist

- [ ] Prometheus endpoint secured/access controlled
- [ ] Metrics cardinality within limits (<10k series per service)
- [ ] Alert routing configured (Slack/PagerDuty)
- [ ] Grafana dashboards accessible to team
- [ ] Log retention policies configured
- [ ] Backup procedures for metrics data
- [ ] Runbook documentation created
- [ ] Team training on dashboard usage
- [ ] SLO targets defined and agreed upon
- [ ] Incident response procedures updated

## Next Steps

1. **Week 1**: Implement Phases 1-3 (basic metrics)
2. **Week 2**: Deploy to staging with K8s monitoring (Phases 4-5)  
3. **Week 3**: Add Grafana dashboards and structured logging (Phases 6-7)
4. **Week 4**: Production rollout with full alerting

This implementation guide provides a practical, step-by-step approach to achieve comprehensive monitoring with minimal complexity and maximum operational value.