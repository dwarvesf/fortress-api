# ADR-002: Metrics Selection and Instrumentation Strategy

**Status**: Proposed  
**Date**: 2025-08-08  
**Deciders**: Planning Team

## Context

Based on analysis of the Fortress API codebase, we need to define specific metrics to collect that provide maximum operational value while minimizing complexity and performance overhead. The system handles critical business functions including:

- Employee management and authentication  
- Project management and resource allocation
- Financial operations (invoices, payroll, accounting)
- Client relationship management
- Integration with external services (Discord, SendGrid, GitHub, Notion)

### Codebase Analysis Findings
- **HTTP Endpoints**: 100+ API endpoints across multiple domains
- **Database Operations**: Heavy GORM usage with complex queries  
- **Authentication**: JWT + API Key dual authentication system
- **External Dependencies**: 10+ external service integrations
- **Business Logic**: Complex permission system with role-based access

## Decision

We will implement a **focused metrics strategy** prioritizing the Golden Signals (Latency, Traffic, Errors, Saturation) with business-specific extensions:

### 1. HTTP Request Metrics (Highest Priority)

```go
// Core HTTP metrics
http_requests_total{method, endpoint, status} (counter)
http_request_duration_seconds{method, endpoint} (histogram)  
http_requests_in_flight (gauge)
http_request_size_bytes{method, endpoint} (histogram)
http_response_size_bytes{method, endpoint} (histogram)
```

**Endpoint Normalization Strategy**:
- Use Gin route templates (e.g., `/api/v1/employees/:id` not `/api/v1/employees/123`)
- Group low-traffic endpoints to prevent cardinality explosion
- Maximum 50 unique endpoint labels

### 2. Authentication & Security Metrics

```go
// Authentication events  
auth_attempts_total{method, result, reason} (counter)
jwt_validation_duration_seconds (histogram)
api_key_usage_total{client_id} (counter)

// Security events
permission_checks_total{permission, result} (counter)
rate_limit_violations_total{endpoint, client} (counter)
suspicious_activity_total{event_type, source_ip} (counter)
```

**Rationale**: Security is critical for RBAC system with sensitive financial data

### 3. Database Performance Metrics

```go
// GORM Plugin provides automatically:
gorm_dbstats_max_open_connections (gauge)
gorm_dbstats_open_connections (gauge) 
gorm_dbstats_in_use (gauge)
gorm_dbstats_idle (gauge)
gorm_dbstats_wait_count (counter)
gorm_dbstats_wait_duration_seconds (histogram)

// Custom business metrics:
database_queries_total{operation, table, result} (counter)
database_query_duration_seconds{operation, table} (histogram)
database_slow_queries_total{table} (counter)
```

**Integration**: Use GORM callbacks for custom query tracking

### 4. Business Operation Metrics

```go  
// Core business processes
invoice_operations_total{operation, status} (counter)
payroll_calculations_total{month, status} (counter)  
employee_operations_total{operation, result} (counter)
project_operations_total{operation, status} (counter)

// External service health
external_api_calls_total{service, operation, status} (counter)
external_api_duration_seconds{service, operation} (histogram)
external_api_failures_total{service, error_type} (counter)
```

**Services Tracked**: Discord, SendGrid, GitHub, Notion, Google APIs

### 5. Go Runtime Metrics

```go
// Automatically provided by Prometheus Go client:
go_memstats_heap_alloc_bytes (gauge)
go_memstats_heap_sys_bytes (gauge)
go_memstats_gc_duration_seconds (summary)
go_goroutines (gauge)
process_cpu_seconds_total (counter)
process_resident_memory_bytes (gauge)
process_max_fds (gauge)
process_open_fds (gauge)
```

**Rationale**: Essential for capacity planning and performance optimization

### 6. Application Health Metrics

```go
// Service status
service_health_status{component} (gauge) // 1=healthy, 0=unhealthy
service_startup_duration_seconds (histogram)
service_uptime_seconds (counter)

// Worker queue metrics (existing worker.go pattern)
worker_queue_size (gauge)
worker_jobs_total{job_type, status} (counter) 
worker_job_duration_seconds{job_type} (histogram)
```

## Label Strategy

### Cardinality Control
- **Maximum Labels per Metric**: 4-5 labels maximum
- **Dynamic Label Limits**: 
  - endpoint: max 50 values
  - user_id: exclude from high-frequency metrics
  - client_ip: hash or exclude for privacy
- **Label Normalization**: Consistent naming (snake_case)

### High-Value Labels
- `method`: HTTP method (GET, POST, PUT, DELETE)
- `endpoint`: Normalized route template
- `status`: HTTP status code or operation result
- `component`: Service component (auth, database, external_api)
- `operation`: Business operation type

## Implementation Patterns

### 1. Gin Middleware Integration
```go
// pkg/middleware/prometheus.go
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Record in-flight requests
        httpRequestsInFlight.Inc()
        defer httpRequestsInFlight.Dec()
        
        c.Next()
        
        // Record completed request metrics
        duration := time.Since(start).Seconds()
        endpoint := normalizeEndpoint(c.FullPath())
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(
            c.Request.Method, endpoint, status,
        ).Inc()
        
        httpDuration.WithLabelValues(
            c.Request.Method, endpoint,
        ).Observe(duration)
    }
}
```

### 2. Business Metrics in Handlers
```go
// Example in authentication handler
func (h *authHandler) authenticate(c *gin.Context) {
    timer := prometheus.NewTimer(jwtValidationDuration)
    defer timer.ObserveDuration()
    
    // Authentication logic...
    
    if err != nil {
        authAttempts.WithLabelValues(
            "jwt", "failure", categorizeError(err),
        ).Inc()
        return
    }
    
    authAttempts.WithLabelValues("jwt", "success", "").Inc()
}
```

### 3. Database Query Instrumentation
```go
// GORM callback for query metrics  
func (db *gorm.DB) MetricsCallback() {
    db.Callback().Query().After("gorm:query").Register("prometheus:query", func(db *gorm.DB) {
        start := db.Statement.Context.Value("query_start").(time.Time)
        duration := time.Since(start).Seconds()
        
        databaseQueryDuration.WithLabelValues(
            "select", db.Statement.Table,
        ).Observe(duration)
        
        if duration > 1.0 {
            databaseSlowQueries.WithLabelValues(db.Statement.Table).Inc()
        }
    })
}
```

## Performance Considerations

### Overhead Targets  
- **CPU Impact**: <2% under normal load
- **Memory Impact**: <50MB for metrics storage
- **Network Impact**: <100KB/minute metrics export  
- **Storage Impact**: 30-day retention ~1GB per replica

### Optimization Strategies
- **Sampling**: High-frequency operations may use sampling (10% for detailed timing)
- **Aggregation**: Pre-aggregate business metrics where possible
- **Batch Recording**: Group multiple metric updates
- **Async Recording**: Background goroutines for expensive calculations

## Quality Assurance

### Testing Strategy
- **Unit Tests**: Metric increment verification  
- **Integration Tests**: End-to-end request flow
- **Load Tests**: Performance overhead validation
- **Canary Deployment**: Gradual rollout with monitoring

### Alerting Integration  
All metrics designed to support alerting rules with clear thresholds:
- Error rates: >5% over 5 minutes
- Latency: 95th percentile >500ms over 10 minutes  
- Database connections: >80% utilization over 5 minutes
- Authentication failures: >10/second over 2 minutes

## Consequences

### Positive
- **Comprehensive Coverage**: All critical application paths monitored
- **Performance Optimization**: Identification of bottlenecks and slow queries
- **Security Insights**: Real-time detection of authentication attacks
- **Business Intelligence**: Understanding of feature usage and user patterns
- **Capacity Planning**: Data-driven scaling decisions

### Negative
- **Development Overhead**: ~40 hours to implement all metrics
- **Runtime Cost**: 1-2% performance impact
- **Complexity**: More moving parts to maintain
- **Storage Cost**: Additional metrics storage requirements

### Risks & Mitigations
- **High Cardinality**: Controlled through endpoint normalization and label limits
- **Memory Leaks**: Careful metric lifecycle management
- **Performance Degradation**: Load testing and gradual rollout
- **Alert Noise**: Thoughtful threshold setting and business impact focus

## Migration Strategy

### Phase 1: Core HTTP Metrics (Days 1-3)
- Implement basic Gin middleware  
- Add `/metrics` endpoint
- Deploy to staging environment

### Phase 2: Authentication & Database (Days 4-7)
- Add auth event tracking
- Configure GORM plugin
- Implement security metrics

### Phase 3: Business Logic (Days 8-12)
- Add business operation metrics  
- External service monitoring
- Worker queue instrumentation

### Phase 4: Production Deployment (Days 13-15)
- Performance validation
- Alerting rule configuration
- Production rollout

## Success Criteria

### Technical Metrics
- All endpoints instrumented with <5ms overhead
- Database connection pool visibility achieved
- Security events captured in real-time  
- Business KPIs trackable through metrics

### Operational Metrics  
- MTTD for critical issues <5 minutes
- Performance bottlenecks identified within 1 hour
- Security incidents detected automatically
- 95% of operational questions answerable through dashboards

---

**Approved By**: Engineering Team  
**Implementation Owner**: Backend Engineering Team  
**Review Date**: 2025-09-08