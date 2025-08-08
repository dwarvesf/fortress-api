# SPEC-002: Database Monitoring Integration

**Status**: Ready for Implementation  
**Priority**: High  
**Estimated Effort**: 6-8 hours  
**Dependencies**: SPEC-001 (HTTP Middleware)  

## Overview

Integrate comprehensive database monitoring into the Fortress API using the official GORM Prometheus plugin plus custom business logic metrics. This will provide visibility into database performance, connection health, and query patterns critical for operational excellence.

## Requirements

### Functional Requirements
- **FR-1**: Monitor database connection pool metrics (connections, utilization, wait times)
- **FR-2**: Track query performance and identify slow queries  
- **FR-3**: Instrument business-critical database operations
- **FR-4**: Provide database health status for alerting
- **FR-5**: Support multiple database instances (primary, replica if applicable)

### Non-Functional Requirements
- **NFR-1**: Database monitoring overhead <2% query performance impact
- **NFR-2**: Graceful degradation if metrics collection fails
- **NFR-3**: Compatible with existing GORM setup and patterns
- **NFR-4**: Support for both development and production environments

## Technical Specification

### 1. GORM Prometheus Plugin Integration

```go
// pkg/store/store.go - Enhanced database setup
// MODIFICATION to existing NewPostgresStore function

package store

import (
    "time"
    
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "gorm.io/plugin/prometheus"
    
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/metrics"
)

func NewPostgresStore(cfg *config.Config) DBRepo {
    db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{
        // ... existing configuration
    })
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // NEW: Add monitoring if enabled
    if cfg.Monitoring.Enabled && cfg.Monitoring.Database.Enabled {
        if err := setupDatabaseMonitoring(db, cfg); err != nil {
            // Log warning but don't fail startup
            log.Warnf("Failed to setup database monitoring: %v", err)
        }
    }

    return &postgresRepo{db: db}
}

func setupDatabaseMonitoring(db *gorm.DB, cfg *config.Config) error {
    // Official GORM Prometheus plugin
    err := db.Use(prometheus.New(prometheus.Config{
        DBName:          cfg.Database.Name,
        RefreshInterval: uint32(cfg.Monitoring.Database.RefreshInterval),
        Labels: map[string]string{
            "service":     "fortress-api",
            "environment": cfg.Env,
            "database":    cfg.Database.Name,
        },
        MetricsCollector: []prometheus.MetricsCollector{
            &prometheus.MySQL{
                VariableNames: []string{
                    "Threads_running",
                    "Threads_connected", 
                    "Max_used_connections",
                    "Open_tables",
                },
            },
        },
    }))
    if err != nil {
        return fmt.Errorf("failed to register prometheus plugin: %w", err)
    }

    // Custom callbacks for business metrics
    if cfg.Monitoring.Database.CustomMetrics {
        registerCustomDatabaseCallbacks(db, cfg)
    }

    return nil
}
```

### 2. Custom Database Metrics

```go
// pkg/metrics/database.go - Custom database metrics
package metrics

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Business operation metrics
    DatabaseOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "operations_total",
            Help:      "Total database operations by type and result",
        },
        []string{"operation", "table", "result"},
    )
    
    DatabaseOperationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress", 
            Subsystem: "database",
            Name:      "operation_duration_seconds",
            Help:      "Database operation duration distribution",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"operation", "table"},
    )
    
    DatabaseSlowQueries = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database", 
            Name:      "slow_queries_total",
            Help:      "Total number of slow database queries",
        },
        []string{"table", "operation"},
    )
    
    DatabaseConnectionHealth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "connection_health_status",
            Help:      "Database connection health status (1=healthy, 0=unhealthy)",
        },
        []string{"database", "instance"},
    )
    
    DatabaseTransactions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "transactions_total", 
            Help:      "Total database transactions by result",
        },
        []string{"result"}, // commit, rollback
    )
    
    DatabaseLockWaitTime = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "lock_wait_duration_seconds",
            Help:      "Time spent waiting for database locks",
        },
        []string{"table", "lock_type"},
    )
)

// Business-specific metrics
var (
    DatabaseBusinessOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "business_operations_total",
            Help:      "Business-critical database operations",
        },
        []string{"domain", "operation", "result"},
    )
)
```

### 3. GORM Callback Implementation

```go
// pkg/store/monitoring.go - Custom GORM callbacks
package store

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"
    
    "gorm.io/gorm"
    "github.com/dwarvesf/fortress-api/pkg/metrics"
)

type QueryContext struct {
    StartTime     time.Time
    Operation     string
    TableName     string
    BusinessDomain string
}

func registerCustomDatabaseCallbacks(db *gorm.DB, cfg *config.Config) {
    // Before callbacks - capture start time and context
    db.Callback().Create().Before("gorm:create").Register("metrics:before_create", beforeCallback("create"))
    db.Callback().Query().Before("gorm:query").Register("metrics:before_query", beforeCallback("select"))
    db.Callback().Update().Before("gorm:update").Register("metrics:before_update", beforeCallback("update"))
    db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", beforeCallback("delete"))
    
    // After callbacks - record metrics
    db.Callback().Create().After("gorm:create").Register("metrics:after_create", afterCallback("create"))
    db.Callback().Query().After("gorm:query").Register("metrics:after_query", afterCallback("select"))
    db.Callback().Update().After("gorm:update").Register("metrics:after_update", afterCallback("update"))
    db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", afterCallback("delete"))
    
    // Transaction callbacks
    db.Callback().Begin().After("gorm:begin_transaction").Register("metrics:begin_transaction", transactionBeginCallback)
    db.Callback().Commit().After("gorm:commit_transaction").Register("metrics:commit_transaction", transactionCommitCallback)
    db.Callback().Rollback().After("gorm:rollback_transaction").Register("metrics:rollback_transaction", transactionRollbackCallback)
    
    // Start connection health monitoring  
    go startConnectionHealthMonitor(db, cfg)
}

func beforeCallback(operation string) func(*gorm.DB) {
    return func(db *gorm.DB) {
        ctx := &QueryContext{
            StartTime:  time.Now(),
            Operation:  operation,
            TableName:  getTableName(db),
            BusinessDomain: inferBusinessDomain(db),
        }
        db.Set("metrics:context", ctx)
    }
}

func afterCallback(operation string) func(*gorm.DB) {
    return func(db *gorm.DB) {
        ctxValue, exists := db.Get("metrics:context")
        if !exists {
            return
        }
        
        ctx, ok := ctxValue.(*QueryContext)
        if !ok {
            return
        }
        
        duration := time.Since(ctx.StartTime)
        
        // Determine result
        result := "success"
        if db.Error != nil {
            result = "error"
        }
        
        // Record metrics
        metrics.DatabaseOperations.WithLabelValues(
            ctx.Operation, ctx.TableName, result,
        ).Inc()
        
        metrics.DatabaseOperationDuration.WithLabelValues(
            ctx.Operation, ctx.TableName,
        ).Observe(duration.Seconds())
        
        // Track slow queries (configurable threshold)
        slowQueryThreshold := 1.0 // 1 second default
        if duration.Seconds() > slowQueryThreshold {
            metrics.DatabaseSlowQueries.WithLabelValues(
                ctx.TableName, ctx.Operation,
            ).Inc()
        }
        
        // Business-specific metrics
        if ctx.BusinessDomain != "" {
            metrics.DatabaseBusinessOperations.WithLabelValues(
                ctx.BusinessDomain, ctx.Operation, result,
            ).Inc()
        }
    }
}

func transactionBeginCallback(db *gorm.DB) {
    // Could add transaction tracking metrics if needed
}

func transactionCommitCallback(db *gorm.DB) {
    metrics.DatabaseTransactions.WithLabelValues("commit").Inc()
}

func transactionRollbackCallback(db *gorm.DB) {
    metrics.DatabaseTransactions.WithLabelValues("rollback").Inc()
}

func getTableName(db *gorm.DB) string {
    if db.Statement != nil && db.Statement.Table != "" {
        return db.Statement.Table
    }
    
    if db.Statement != nil && db.Statement.Model != nil {
        return db.Statement.Schema.Table
    }
    
    return "unknown"
}

func inferBusinessDomain(db *gorm.DB) string {
    tableName := getTableName(db)
    
    // Map table names to business domains
    domainMap := map[string]string{
        "employees":            "hr",
        "employee_roles":       "hr", 
        "employee_positions":   "hr",
        "projects":            "project_management",
        "project_members":     "project_management",
        "invoices":            "finance",
        "invoice_numbers":     "finance",
        "payrolls":            "finance", 
        "clients":             "client_management",
        "audits":              "compliance",
        "permissions":         "security",
        "api_keys":            "security",
    }
    
    if domain, exists := domainMap[tableName]; exists {
        return domain
    }
    
    // Try to infer from table prefix
    if strings.HasPrefix(tableName, "employee") {
        return "hr"
    } else if strings.HasPrefix(tableName, "project") {
        return "project_management"
    } else if strings.HasPrefix(tableName, "invoice") || strings.HasPrefix(tableName, "payroll") {
        return "finance"
    }
    
    return "" // No business domain mapping
}
```

### 4. Connection Health Monitoring

```go
// pkg/store/health.go - Database health monitoring
package store

import (
    "context"
    "database/sql"
    "time"
    
    "gorm.io/gorm"
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/metrics"
)

func startConnectionHealthMonitor(db *gorm.DB, cfg *config.Config) {
    ticker := time.NewTicker(time.Duration(cfg.Monitoring.Database.HealthCheckInterval) * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            checkDatabaseHealth(db, cfg)
        }
    }
}

func checkDatabaseHealth(db *gorm.DB, cfg *config.Config) {
    sqlDB, err := db.DB()
    if err != nil {
        metrics.DatabaseConnectionHealth.WithLabelValues(
            cfg.Database.Name, "primary",
        ).Set(0)
        return
    }
    
    // Test basic connectivity
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    err = sqlDB.PingContext(ctx)
    if err != nil {
        metrics.DatabaseConnectionHealth.WithLabelValues(
            cfg.Database.Name, "primary",
        ).Set(0)
        return
    }
    
    // Check connection pool stats
    stats := sqlDB.Stats()
    
    // Set health to 1 (healthy) if basic checks pass
    metrics.DatabaseConnectionHealth.WithLabelValues(
        cfg.Database.Name, "primary",
    ).Set(1)
    
    // Additional connection pool metrics (supplementing GORM plugin)
    recordConnectionPoolStats(stats, cfg)
}

func recordConnectionPoolStats(stats sql.DBStats, cfg *config.Config) {
    // Note: These metrics complement the GORM Prometheus plugin
    // Only add if we need additional insights not provided by the plugin
    
    // Connection pool efficiency
    poolEfficiency := float64(stats.InUse) / float64(stats.MaxOpenConnections)
    if poolEfficiency > 0.8 { // 80% threshold
        // Could increment a "high_pool_utilization" counter
    }
    
    // Connection wait analysis  
    if stats.WaitCount > 0 {
        avgWaitDuration := stats.WaitDuration / time.Duration(stats.WaitCount)
        // Record average wait time if needed
        _ = avgWaitDuration
    }
}
```

### 5. Business Logic Integration

```go
// pkg/metrics/business_database.go - Business-specific database metrics helpers
package metrics

import (
    "context"
    "time"
)

// Helper functions for business logic to use

func RecordPayrollDatabaseOperation(operation string, duration time.Duration, err error) {
    result := "success"
    if err != nil {
        result = "error"
    }
    
    DatabaseBusinessOperations.WithLabelValues(
        "payroll", operation, result,
    ).Inc()
    
    if duration > 0 {
        // Could add specific payroll operation timing if needed
    }
}

func RecordInvoiceDatabaseOperation(operation string, duration time.Duration, err error) {
    result := "success" 
    if err != nil {
        result = "error"
    }
    
    DatabaseBusinessOperations.WithLabelValues(
        "invoice", operation, result,
    ).Inc()
}

// Usage in business logic (optional enhancement):
// func (s *payrollStore) CalculatePayroll(ctx context.Context, month string) error {
//     start := time.Now()
//     
//     // existing business logic...
//     err := s.performCalculation(ctx, month)
//     
//     // record metrics
//     metrics.RecordPayrollDatabaseOperation("calculate", time.Since(start), err)
//     
//     return err
// }
```

### 6. Configuration Enhancement

```go
// pkg/config/config.go - Database monitoring configuration
// ADDITION to existing MonitoringConfig

type MonitoringConfig struct {
    // ... existing fields ...
    
    Database DatabaseMonitoringConfig `mapstructure:"database"`
}

type DatabaseMonitoringConfig struct {
    Enabled                bool          `mapstructure:"enabled" default:"true"`
    RefreshInterval        time.Duration `mapstructure:"refresh_interval" default:"15s"`
    CustomMetrics          bool          `mapstructure:"custom_metrics" default:"true"`
    SlowQueryThreshold     time.Duration `mapstructure:"slow_query_threshold" default:"1s"`
    HealthCheckInterval    time.Duration `mapstructure:"health_check_interval" default:"30s"`
    BusinessMetrics        bool          `mapstructure:"business_metrics" default:"true"`
}
```

## Testing Strategy

### 1. Unit Tests

```go
// pkg/store/monitoring_test.go
package store

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
    
    "github.com/dwarvesf/fortress-api/pkg/testhelper"
)

func TestDatabaseMetricsCallbacks(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        db := repo.DB()
        
        // Reset metrics for testing
        // Note: Would need to implement metric reset functionality
        
        // Test create operation
        employee := &model.Employee{
            FullName: "Test Employee",
            Username: "testuser",
        }
        
        err := db.Create(employee).Error
        assert.NoError(t, err)
        
        // Verify metrics were recorded
        // Note: Would use prometheus testutil to verify metrics
    })
}

func TestSlowQueryDetection(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        db := repo.DB()
        
        // Execute a query that should be flagged as slow
        // (would need to create a test scenario for this)
        
        // Verify slow query metric was incremented
    })
}

func TestBusinessDomainInference(t *testing.T) {
    tests := []struct {
        tableName      string
        expectedDomain string
    }{
        {"employees", "hr"},
        {"employee_roles", "hr"},
        {"projects", "project_management"},
        {"invoices", "finance"},
        {"payrolls", "finance"},
        {"unknown_table", ""},
    }
    
    for _, test := range tests {
        // Mock GORM DB with table name
        mockDB := &gorm.DB{
            Statement: &gorm.Statement{
                Table: test.tableName,
            },
        }
        
        domain := inferBusinessDomain(mockDB)
        assert.Equal(t, test.expectedDomain, domain)
    }
}

func TestConnectionHealthMonitoring(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        db := repo.DB()
        
        // Test health check function
        cfg := &config.Config{
            Database: config.DatabaseConfig{Name: "test_db"},
            Monitoring: config.MonitoringConfig{
                Database: config.DatabaseMonitoringConfig{
                    HealthCheckInterval: time.Second,
                },
            },
        }
        
        checkDatabaseHealth(db, cfg)
        
        // Verify health metric was set to 1 (healthy)
        // Note: Would use prometheus testutil to verify
    })
}
```

### 2. Integration Tests

```go
// tests/integration/database_monitoring_test.go
package integration

import (
    "testing"
    "time"
    
    "github.com/dwarvesf/fortress-api/pkg/testhelper"
    "github.com/stretchr/testify/assert"
)

func TestGORMPrometheusPlugin(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        // Perform various database operations
        testOperations := []func(){
            func() {
                var employees []model.Employee
                repo.DB().Find(&employees)
            },
            func() {
                employee := &model.Employee{
                    FullName: "Test User",
                    Username: "testuser",
                }
                repo.DB().Create(employee)
            },
            func() {
                repo.DB().Model(&model.Employee{}).Where("id = ?", 1).Update("full_name", "Updated Name")
            },
        }
        
        for _, op := range testOperations {
            op()
        }
        
        // Check that standard GORM metrics are available
        // Would need to query Prometheus metrics endpoint
    })
}

func TestDatabaseMetricsAccuracy(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        // Record initial metric values
        // Perform known number of operations
        // Verify metric increments match expected values
        
        initialCount := getCurrentMetricValue("fortress_database_operations_total")
        
        // Perform 10 operations
        for i := 0; i < 10; i++ {
            var employee model.Employee
            repo.DB().First(&employee, 1)
        }
        
        finalCount := getCurrentMetricValue("fortress_database_operations_total") 
        assert.Equal(t, initialCount+10, finalCount)
    })
}

func getCurrentMetricValue(metricName string) float64 {
    // Helper function to get current metric value
    // Would implement using prometheus testutil
    return 0
}
```

### 3. Performance Tests

```go
// tests/performance/database_monitoring_test.go
package performance

import (
    "testing"
    "time"
    
    "github.com/dwarvesf/fortress-api/pkg/testhelper"
)

func BenchmarkDatabaseOperationsWithMonitoring(b *testing.B) {
    testhelper.TestWithTxDB(b, func(repo DBRepo) {
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
            var employee model.Employee
            repo.DB().First(&employee, 1)
        }
    })
}

func BenchmarkDatabaseOperationsWithoutMonitoring(b *testing.B) {
    // Same test but with monitoring disabled
    testhelper.TestWithTxDB(b, func(repo DBRepo) {
        // Disable monitoring callbacks temporarily
        
        b.ResetTimer()
        
        for i := 0; i < b.N; i++ {
            var employee model.Employee  
            repo.DB().First(&employee, 1)
        }
    })
}

func TestMonitoringOverhead(t *testing.T) {
    // Measure the overhead of monitoring callbacks
    testhelper.TestWithTxDB(t, func(repo DBRepo) {
        operations := 1000
        
        // Test with monitoring
        startWith := time.Now()
        for i := 0; i < operations; i++ {
            var employee model.Employee
            repo.DB().First(&employee, 1)
        }
        durationWith := time.Since(startWith)
        
        // Disable monitoring and test again
        // (would need to implement dynamic disable)
        
        startWithout := time.Now()
        for i := 0; i < operations; i++ {
            var employee model.Employee
            repo.DB().First(&employee, 1)
        }
        durationWithout := time.Since(startWithout)
        
        overhead := durationWith - durationWithout
        overheadPercentage := float64(overhead) / float64(durationWithout) * 100
        
        // Overhead should be less than 5%
        assert.Less(t, overheadPercentage, 5.0, "Monitoring overhead exceeds 5%")
    })
}
```

## Deployment Strategy

### 1. Configuration Files

```yaml
# config/monitoring.yaml - Database monitoring configuration
monitoring:
  enabled: true
  database:
    enabled: true
    refresh_interval: 15s
    custom_metrics: true
    slow_query_threshold: 1s
    health_check_interval: 30s
    business_metrics: true
```

### 2. Kubernetes ServiceMonitor Update

```yaml
# k8s/monitoring/servicemonitor.yaml - Updated for database metrics
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor  
metadata:
  name: fortress-api
spec:
  selector:
    matchLabels:
      app: fortress-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
    # Database metrics may need longer scrape timeout
    scrapeTimeout: 15s
```

### 3. Staged Rollout Plan

**Phase 1: GORM Plugin Only (Week 1)**
- Deploy official GORM Prometheus plugin
- Monitor basic connection pool metrics
- Validate no performance impact

**Phase 2: Custom Callbacks (Week 2)**
- Add custom GORM callbacks for business metrics
- Monitor slow query detection
- Fine-tune thresholds

**Phase 3: Health Monitoring (Week 3)**
- Enable connection health monitoring
- Set up database alerting rules
- Create database-focused dashboards

**Phase 4: Business Integration (Week 4)**
- Add business logic integration points
- Create business-specific database dashboards
- Train teams on database metrics

## Success Criteria

### Performance Metrics
- **Query Overhead**: <5% performance impact on database operations
- **Memory Usage**: <25MB additional memory for database metrics
- **Monitoring Accuracy**: 99.9% accurate query counting and timing

### Operational Metrics
- **Slow Query Detection**: 100% of queries >1s threshold detected
- **Connection Health**: Real-time visibility into connection pool status
- **Business Insights**: Database performance correlated with business operations

### Technical Metrics
- **Plugin Stability**: Zero database connection issues due to monitoring
- **Metric Availability**: Database metrics available within 30 seconds
- **Cardinality Control**: Database metrics remain under 1,000 series per instance

## Risk Mitigation

### Database Performance Risk
- **Mitigation**: Comprehensive benchmarking and gradual rollout
- **Monitoring**: Continuous performance comparison with baseline

### Connection Pool Risk
- **Mitigation**: Thorough testing of GORM plugin with existing connection settings
- **Monitoring**: Close monitoring of connection pool metrics during deployment

### Callback Overhead Risk
- **Mitigation**: Minimal callback logic and async processing where possible
- **Monitoring**: Performance profiling and callback execution timing

---

**Assignee**: Backend Engineering Team  
**Reviewer**: Database Team, Senior Backend Engineer  
**Implementation Timeline**: 1 week  
**Testing Timeline**: 3-5 days  
**Deployment Timeline**: 2 weeks (staged rollout)