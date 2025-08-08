# ADR-003: Implementation Architecture and Integration Strategy

**Status**: Proposed  
**Date**: 2025-08-08  
**Deciders**: Planning Team

## Context

The Fortress API has a well-established architecture that monitoring must integrate seamlessly with:

- **Layered Architecture**: Routes → Controllers → Services → Stores → Database
- **Gin Framework**: Middleware-based request processing
- **Domain-Driven Structure**: Clear package separation by business domain  
- **Existing Patterns**: Consistent interface/implementation patterns
- **Docker + Kubernetes**: Containerized deployment with existing infrastructure

We need to design monitoring integration that:
1. Respects existing architectural boundaries
2. Follows established code organization patterns  
3. Minimizes disruption to current development workflow
4. Provides clear upgrade path for future enhancements

## Decision

We will implement a **non-invasive monitoring architecture** that integrates at the middleware level with optional business-specific instrumentation:

### 1. Package Structure

```
pkg/
├── metrics/
│   ├── registry.go          # Metric definitions and registration
│   ├── http.go             # HTTP request metrics  
│   ├── database.go         # Database performance metrics
│   ├── business.go         # Business logic metrics
│   ├── security.go         # Authentication/authorization metrics
│   └── collector.go        # Custom metrics collection logic
├── middleware/
│   ├── prometheus.go       # HTTP metrics middleware
│   ├── logging.go          # Structured logging middleware  
│   └── security.go         # Security event logging
└── monitoring/
    ├── health.go           # Enhanced health checks
    ├── profiler.go         # Performance profiling endpoints
    └── diagnostics.go      # System diagnostics
```

**Rationale**: Follows existing domain separation while keeping monitoring concerns isolated

### 2. Middleware Integration Strategy

```go
// Middleware chain integration in pkg/routes/routes.go
func NewRoutes(cfg *config.Config, svc *service.Service, s *store.Store, repo store.DBRepo, worker *worker.Worker, logger logger.Logger) *gin.Engine {
    r := gin.New()
    
    // Core middleware (existing)
    r.Use(gin.Recovery())
    
    // NEW: Monitoring middleware (non-intrusive)
    r.Use(middleware.PrometheusMiddleware())
    r.Use(middleware.StructuredLogging(logger))
    
    // Existing middleware continues unchanged
    r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"))
    
    // NEW: Metrics endpoint
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    r.GET("/health", h.Health.DetailedHealth) // Enhanced health check
    
    // Existing routes unchanged
    loadV1Routes(r, h, repo, s, cfg)
    
    return r
}
```

**Design Principles**:
- **Non-Breaking**: Existing routes and handlers require no changes
- **Opt-In Enhancement**: Business metrics added incrementally  
- **Clear Boundaries**: Monitoring logic isolated from business logic
- **Performance First**: Minimal overhead design

### 3. Metrics Collection Architecture

#### Core Metrics Registry
```go
// pkg/metrics/registry.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP Metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "http",
            Name:      "requests_total",
            Help:      "Total HTTP requests processed",
        },
        []string{"method", "endpoint", "status"},
    )
    
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "http", 
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration distribution",
            Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "endpoint"},
    )
    
    // Database Metrics (enhanced)
    DatabaseOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "operations_total",
            Help:      "Total database operations",
        },
        []string{"operation", "table", "result"},
    )
    
    // Business Metrics
    BusinessOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress", 
            Subsystem: "business",
            Name:      "operations_total",
            Help:      "Business operations processed",
        },
        []string{"domain", "operation", "result"},
    )
)

// Initialize sets up all metric collectors
func Initialize(cfg *config.Config) {
    // Register custom collectors
    prometheus.MustRegister(&healthCollector{})
    prometheus.MustRegister(&databaseStatsCollector{repo: repo})
}
```

#### Non-Intrusive Business Metrics
```go
// pkg/metrics/business.go
package metrics

// Business metric helpers that can be called from any layer
func RecordInvoiceOperation(operation string, success bool) {
    result := "success"
    if !success {
        result = "failure"
    }
    
    BusinessOperations.WithLabelValues(
        "invoice", operation, result,
    ).Inc()
}

func RecordPayrollCalculation(month string, success bool) {
    result := "success"
    if !success {
        result = "failure"
    }
    
    BusinessOperations.WithLabelValues(
        "payroll", "calculation", result,
    ).Inc()
}

// Usage in existing handlers (optional enhancement):
// func (h *invoiceHandler) UpdateStatus(c *gin.Context) {
//     // existing logic...
//     err := h.controller.UpdateStatus(...)
//     
//     // NEW: optional metrics recording
//     metrics.RecordInvoiceOperation("update_status", err == nil)
//     
//     // existing response logic...
// }
```

### 4. Database Integration Strategy

#### GORM Plugin Integration
```go
// Database initialization enhancement in store/store.go
func NewPostgresStore(cfg *config.Config) DBRepo {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // NEW: Add monitoring plugins (non-breaking)
    if cfg.Monitoring.Enabled {
        // Official GORM Prometheus plugin
        db.Use(prometheus.New(prometheus.Config{
            DBName:          "fortress_db",
            RefreshInterval: 15,
            Labels: map[string]string{
                "service": "fortress-api",
                "env":     cfg.Env,
            },
        }))
        
        // Custom callback for business query tracking
        registerMetricsCallbacks(db)
    }
    
    return &postgresRepo{db: db}
}

func registerMetricsCallbacks(db *gorm.DB) {
    // Optional query performance tracking
    db.Callback().Query().Before("gorm:query").Register("metrics:query_start", 
        func(db *gorm.DB) {
            db.Set("query_start_time", time.Now())
        })
    
    db.Callback().Query().After("gorm:query").Register("metrics:query_end",
        func(db *gorm.DB) {
            if start, ok := db.Get("query_start_time"); ok {
                duration := time.Since(start.(time.Time)).Seconds()
                
                metrics.DatabaseOperations.WithLabelValues(
                    "select", db.Statement.Table, "success",
                ).Inc()
                
                // Track slow queries
                if duration > 1.0 {
                    metrics.DatabaseSlowQueries.WithLabelValues(
                        db.Statement.Table,
                    ).Inc()
                }
            }
        })
}
```

### 5. Security Metrics Integration

#### Enhanced Authentication Middleware
```go
// pkg/middleware/security.go
package middleware

func SecurityMetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        // Record authentication events
        if authMethod := getAuthMethod(c); authMethod != "" {
            result := "success"
            reason := ""
            
            if c.Writer.Status() == 401 {
                result = "failure"  
                reason = getFailureReason(c)
            }
            
            metrics.AuthAttempts.WithLabelValues(
                authMethod, result, reason,
            ).Inc()
            
            metrics.AuthDuration.WithLabelValues(authMethod).Observe(
                time.Since(start).Seconds(),
            )
        }
        
        // Record permission checks
        if perm := getRequiredPermission(c); perm != "" {
            result := "allowed"
            if c.Writer.Status() == 403 {
                result = "denied"
            }
            
            metrics.PermissionChecks.WithLabelValues(perm, result).Inc()
        }
    }
}

// Enhanced authentication middleware wraps existing logic
func EnhanceAuthMiddleware(originalAuth gin.HandlerFunc) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Call original authentication logic
        originalAuth(c)
        
        // Add metrics recording (non-intrusive)
        recordAuthMetrics(c, time.Since(start))
    }
}
```

### 6. Configuration Integration

#### Enhanced Config Structure
```go
// pkg/config/config.go - Add monitoring section
type Config struct {
    // Existing config fields...
    
    // NEW: Monitoring configuration
    Monitoring MonitoringConfig `mapstructure:"monitoring"`
}

type MonitoringConfig struct {
    Enabled           bool          `mapstructure:"enabled" default:"true"`
    MetricsPath       string        `mapstructure:"metrics_path" default:"/metrics"`
    HealthPath        string        `mapstructure:"health_path" default:"/health"`
    CollectBusiness   bool          `mapstructure:"collect_business" default:"true"`
    CollectSecurity   bool          `mapstructure:"collect_security" default:"true"`
    SampleRate        float64       `mapstructure:"sample_rate" default:"1.0"`
    SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold" default:"1s"`
}
```

### 7. Kubernetes Integration

#### ServiceMonitor Configuration
```yaml
# k8s/monitoring/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: fortress-api
  namespace: default
  labels:
    app: fortress-api
    release: prometheus
spec:
  selector:
    matchLabels:
      app: fortress-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
    honorLabels: true
```

#### Enhanced Deployment
```yaml
# k8s/deployment.yaml - Add monitoring annotations
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fortress-api
spec:
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
        # Enable structured logging
        fluentbit.io/parser: "json"
    spec:
      containers:
      - name: api
        env:
        - name: MONITORING_ENABLED
          value: "true"
        - name: LOG_FORMAT
          value: "json"
        # Health check configuration
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
```

## Implementation Strategy

### Phase 1: Infrastructure Setup
1. Add Prometheus client dependencies
2. Create metrics package structure
3. Implement basic HTTP middleware
4. Deploy `/metrics` endpoint

### Phase 2: Core Instrumentation  
1. Add GORM Prometheus plugin
2. Implement authentication metrics
3. Configure Kubernetes ServiceMonitor
4. Basic Grafana dashboard

### Phase 3: Enhanced Monitoring
1. Structured logging implementation
2. Business metrics integration
3. Security event tracking
4. Advanced dashboard creation

### Phase 4: Production Readiness
1. Performance optimization
2. Alert rule configuration
3. Documentation and training
4. Production deployment

## Testing Strategy

### Unit Testing
```go
// pkg/middleware/prometheus_test.go
func TestPrometheusMiddleware(t *testing.T) {
    // Test metric collection accuracy
    registry := prometheus.NewRegistry()
    // ... test implementation
}
```

### Integration Testing
```go
// Test end-to-end metric collection
func TestMetricsEndToEnd(t *testing.T) {
    // Make HTTP requests
    // Verify metrics are recorded
    // Check metric accuracy
}
```

### Load Testing
```bash
# Validate performance overhead
hey -n 10000 -c 50 http://localhost:8080/api/v1/health
```

## Consequences

### Positive
- **Non-Disruptive**: Existing code requires minimal changes
- **Incremental**: Can be deployed and enhanced gradually  
- **Maintainable**: Clear separation of monitoring concerns
- **Performant**: Designed for minimal overhead
- **Scalable**: Architecture supports future monitoring needs

### Negative
- **Complexity**: Additional packages and middleware to maintain
- **Learning Curve**: Team needs to understand Prometheus patterns
- **Resource Usage**: Additional memory and CPU overhead
- **Deployment Complexity**: More Kubernetes resources to manage

### Migration Risks
- **Performance Regression**: Mitigated by load testing and gradual rollout
- **High Cardinality**: Controlled through label design and limits
- **Breaking Changes**: Architecture designed to be additive only
- **Operational Overhead**: Balanced by improved observability benefits

## Success Criteria

### Technical Requirements
- Zero breaking changes to existing functionality
- <2% performance overhead under normal load
- All critical paths instrumented
- Kubernetes integration working seamlessly

### Operational Requirements  
- Mean Time to Detection (MTTD) <5 minutes
- 95% of operational questions answerable through metrics
- Security events detected in real-time
- Business KPIs trackable through dashboards

---

**Architecture Review**: Engineering Team  
**Implementation Timeline**: 3-4 weeks  
**Go-Live Target**: 2025-09-01