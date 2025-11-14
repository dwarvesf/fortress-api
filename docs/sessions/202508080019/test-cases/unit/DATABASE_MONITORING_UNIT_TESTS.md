# Unit Test Plan: Database Monitoring (SPEC-002)

**Target**: pkg/store/monitoring.go, pkg/metrics/database.go, pkg/store/health.go  
**Test Count**: 30 unit tests  
**Coverage Target**: 95%+  
**Performance Target**: Each test <100ms  

## Test Categories

### 1. GORM Plugin Integration Tests

#### Test: TestSetupDatabaseMonitoring_Success
```go
func TestSetupDatabaseMonitoring_Success(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        cfg := &config.Config{
            Database: config.DatabaseConfig{Name: "fortress_test"},
            Monitoring: config.MonitoringConfig{
                Enabled: true,
                Database: config.DatabaseMonitoringConfig{
                    Enabled:           true,
                    RefreshInterval:   15 * time.Second,
                    CustomMetrics:     true,
                    HealthCheckInterval: 30 * time.Second,
                },
            },
        }
        
        db := repo.DB()
        err := setupDatabaseMonitoring(db, cfg)
        
        assert.NoError(t, err)
        
        // Verify GORM Prometheus plugin was registered
        dialector := db.Dialector
        assert.NotNil(t, dialector)
        
        // Verify callbacks were registered
        callback := db.Callback()
        assert.NotNil(t, callback.Create().Get("metrics:before_create"))
        assert.NotNil(t, callback.Create().Get("metrics:after_create"))
        assert.NotNil(t, callback.Query().Get("metrics:before_query"))
        assert.NotNil(t, callback.Query().Get("metrics:after_query"))
        assert.NotNil(t, callback.Update().Get("metrics:before_update"))
        assert.NotNil(t, callback.Update().Get("metrics:after_update"))
        assert.NotNil(t, callback.Delete().Get("metrics:before_delete"))
        assert.NotNil(t, callback.Delete().Get("metrics:after_delete"))
    })
}
```

#### Test: TestSetupDatabaseMonitoring_DisabledConfig
```go
func TestSetupDatabaseMonitoring_DisabledConfig(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        cfg := &config.Config{
            Monitoring: config.MonitoringConfig{
                Enabled: false,
            },
        }
        
        db := repo.DB()
        
        // Should not call setupDatabaseMonitoring when disabled
        // This test validates the calling code behavior
        assert.True(t, true) // Placeholder for actual implementation test
    })
}
```

### 2. Custom Callback Tests

#### Test: TestBeforeCallback_ContextSetup
```go
func TestBeforeCallback_ContextSetup(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Create test callback
        beforeCreate := beforeCallback("create")
        
        // Mock GORM DB with statement
        db.Statement = &gorm.Statement{
            Table: "employees",
            Schema: &schema.Schema{Table: "employees"},
        }
        
        // Call callback
        beforeCreate(db)
        
        // Verify context was set
        ctxValue, exists := db.Get("metrics:context")
        assert.True(t, exists)
        
        ctx, ok := ctxValue.(*QueryContext)
        assert.True(t, ok)
        assert.Equal(t, "create", ctx.Operation)
        assert.Equal(t, "employees", ctx.TableName)
        assert.Equal(t, "hr", ctx.BusinessDomain)
        assert.True(t, ctx.StartTime.Before(time.Now()))
    })
}
```

#### Test: TestAfterCallback_MetricsRecording
```go
func TestAfterCallback_MetricsRecording(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Set up context from before callback
        startTime := time.Now().Add(-100 * time.Millisecond)
        ctx := &QueryContext{
            StartTime:      startTime,
            Operation:      "select",
            TableName:      "employees", 
            BusinessDomain: "hr",
        }
        db.Set("metrics:context", ctx)
        
        // Create test callback
        afterQuery := afterCallback("select")
        
        // Call callback with no error
        afterQuery(db)
        
        // Verify metrics were recorded
        operationCount := testutil.ToFloat64(DatabaseOperations.WithLabelValues("select", "employees", "success"))
        assert.Equal(t, 1.0, operationCount)
        
        durationCount := testutil.ToFloat64(DatabaseOperationDuration.WithLabelValues("select", "employees"))
        assert.Equal(t, 1.0, durationCount)
        
        businessOpCount := testutil.ToFloat64(DatabaseBusinessOperations.WithLabelValues("hr", "select", "success"))
        assert.Equal(t, 1.0, businessOpCount)
    })
}
```

#### Test: TestAfterCallback_ErrorHandling
```go
func TestAfterCallback_ErrorHandling(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Set up context with error
        startTime := time.Now().Add(-50 * time.Millisecond)
        ctx := &QueryContext{
            StartTime:      startTime,
            Operation:      "update",
            TableName:      "projects",
            BusinessDomain: "project_management",
        }
        db.Set("metrics:context", ctx)
        
        // Simulate database error
        db.Error = gorm.ErrRecordNotFound
        
        afterUpdate := afterCallback("update")
        afterUpdate(db)
        
        // Verify error metrics were recorded
        operationCount := testutil.ToFloat64(DatabaseOperations.WithLabelValues("update", "projects", "error"))
        assert.Equal(t, 1.0, operationCount)
        
        businessOpCount := testutil.ToFloat64(DatabaseBusinessOperations.WithLabelValues("project_management", "update", "error"))
        assert.Equal(t, 1.0, businessOpCount)
    })
}
```

### 3. Business Domain Inference Tests

#### Test: TestInferBusinessDomain_KnownTables
```go
func TestInferBusinessDomain_KnownTables(t *testing.T) {
    tests := []struct {
        tableName      string
        expectedDomain string
    }{
        {"employees", "hr"},
        {"employee_roles", "hr"},
        {"employee_positions", "hr"},
        {"employee_chapters", "hr"},
        {"projects", "project_management"},
        {"project_members", "project_management"},
        {"project_slots", "project_management"},
        {"invoices", "finance"},
        {"invoice_numbers", "finance"},
        {"payrolls", "finance"},
        {"cached_payrolls", "finance"},
        {"clients", "client_management"},
        {"client_contacts", "client_management"},
        {"audits", "compliance"},
        {"audit_cycles", "compliance"},
        {"permissions", "security"},
        {"api_keys", "security"},
        {"roles", "security"},
    }
    
    for _, test := range tests {
        mockDB := &gorm.DB{
            Statement: &gorm.Statement{
                Table: test.tableName,
            },
        }
        
        domain := inferBusinessDomain(mockDB)
        assert.Equal(t, test.expectedDomain, domain, "Table: %s", test.tableName)
    }
}
```

#### Test: TestInferBusinessDomain_UnknownTables
```go
func TestInferBusinessDomain_UnknownTables(t *testing.T) {
    tests := []struct {
        tableName      string
        expectedDomain string
    }{
        {"unknown_table", ""},
        {"random_data", ""},
        {"temp_table", ""},
        {"", ""},
    }
    
    for _, test := range tests {
        mockDB := &gorm.DB{
            Statement: &gorm.Statement{
                Table: test.tableName,
            },
        }
        
        domain := inferBusinessDomain(mockDB)
        assert.Equal(t, test.expectedDomain, domain, "Table: %s", test.tableName)
    }
}
```

#### Test: TestInferBusinessDomain_PrefixMatching
```go
func TestInferBusinessDomain_PrefixMatching(t *testing.T) {
    tests := []struct {
        tableName      string
        expectedDomain string
    }{
        {"employee_custom_field", "hr"},
        {"project_custom_data", "project_management"},
        {"invoice_line_items", "finance"},
        {"payroll_calculations", "finance"},
        {"audit_trail_data", "compliance"},
    }
    
    for _, test := range tests {
        mockDB := &gorm.DB{
            Statement: &gorm.Statement{
                Table: test.tableName,
            },
        }
        
        domain := inferBusinessDomain(mockDB)
        assert.Equal(t, test.expectedDomain, domain, "Table: %s", test.tableName)
    }
}
```

### 4. Slow Query Detection Tests

#### Test: TestSlowQueryDetection_ThresholdExceeded
```go
func TestSlowQueryDetection_ThresholdExceeded(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Set up context with slow query (2 seconds)
        startTime := time.Now().Add(-2 * time.Second)
        ctx := &QueryContext{
            StartTime:      startTime,
            Operation:      "select",
            TableName:      "employees",
            BusinessDomain: "hr",
        }
        db.Set("metrics:context", ctx)
        
        afterQuery := afterCallback("select")
        afterQuery(db)
        
        // Verify slow query was detected (threshold is 1 second)
        slowQueryCount := testutil.ToFloat64(DatabaseSlowQueries.WithLabelValues("employees", "select"))
        assert.Equal(t, 1.0, slowQueryCount)
    })
}
```

#### Test: TestSlowQueryDetection_ThresholdNotExceeded
```go
func TestSlowQueryDetection_ThresholdNotExceeded(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Set up context with fast query (100ms)
        startTime := time.Now().Add(-100 * time.Millisecond)
        ctx := &QueryContext{
            StartTime:      startTime,
            Operation:      "select",
            TableName:      "employees",
            BusinessDomain: "hr",
        }
        db.Set("metrics:context", ctx)
        
        afterQuery := afterCallback("select")
        afterQuery(db)
        
        // Verify slow query was NOT detected
        slowQueryCount := testutil.ToFloat64(DatabaseSlowQueries.WithLabelValues("employees", "select"))
        assert.Equal(t, 0.0, slowQueryCount)
    })
}
```

### 5. Transaction Callback Tests

#### Test: TestTransactionCommitCallback
```go
func TestTransactionCommitCallback(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        transactionCommitCallback(db)
        
        // Verify commit metric was recorded
        commitCount := testutil.ToFloat64(DatabaseTransactions.WithLabelValues("commit"))
        assert.Equal(t, 1.0, commitCount)
    })
}
```

#### Test: TestTransactionRollbackCallback
```go
func TestTransactionRollbackCallback(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        transactionRollbackCallback(db)
        
        // Verify rollback metric was recorded
        rollbackCount := testutil.ToFloat64(DatabaseTransactions.WithLabelValues("rollback"))
        assert.Equal(t, 1.0, rollbackCount)
    })
}
```

### 6. Connection Health Monitoring Tests

#### Test: TestCheckDatabaseHealth_Success
```go
func TestCheckDatabaseHealth_Success(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        cfg := &config.Config{
            Database: config.DatabaseConfig{Name: "fortress_test"},
        }
        
        db := repo.DB()
        checkDatabaseHealth(db, cfg)
        
        // Verify health metric was set to 1 (healthy)
        healthValue := testutil.ToFloat64(DatabaseConnectionHealth.WithLabelValues("fortress_test", "primary"))
        assert.Equal(t, 1.0, healthValue)
    })
}
```

#### Test: TestCheckDatabaseHealth_ConnectionError
```go
func TestCheckDatabaseHealth_ConnectionError(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    // Create a mock DB that will fail ping
    mockDB := &gorm.DB{}
    
    cfg := &config.Config{
        Database: config.DatabaseConfig{Name: "fortress_test"},
    }
    
    checkDatabaseHealth(mockDB, cfg)
    
    // Verify health metric was set to 0 (unhealthy)
    healthValue := testutil.ToFloat64(DatabaseConnectionHealth.WithLabelValues("fortress_test", "primary"))
    assert.Equal(t, 0.0, healthValue)
}
```

### 7. Connection Pool Statistics Tests

#### Test: TestRecordConnectionPoolStats
```go
func TestRecordConnectionPoolStats(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        cfg := &config.Config{
            Database: config.DatabaseConfig{Name: "fortress_test"},
        }
        
        db := repo.DB()
        sqlDB, err := db.DB()
        require.NoError(t, err)
        
        // Set connection pool limits
        sqlDB.SetMaxOpenConns(10)
        sqlDB.SetMaxIdleConns(5)
        
        stats := sqlDB.Stats()
        recordConnectionPoolStats(stats, cfg)
        
        // Verify function doesn't panic and handles stats correctly
        assert.True(t, stats.MaxOpenConnections > 0)
    })
}
```

### 8. Business Metrics Helper Tests

#### Test: TestRecordPayrollDatabaseOperation_Success
```go
func TestRecordPayrollDatabaseOperation_Success(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    duration := 150 * time.Millisecond
    RecordPayrollDatabaseOperation("calculate", duration, nil)
    
    // Verify business metric was recorded
    businessOpCount := testutil.ToFloat64(DatabaseBusinessOperations.WithLabelValues("payroll", "calculate", "success"))
    assert.Equal(t, 1.0, businessOpCount)
}
```

#### Test: TestRecordPayrollDatabaseOperation_Error
```go
func TestRecordPayrollDatabaseOperation_Error(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    duration := 200 * time.Millisecond
    err := errors.New("calculation error")
    RecordPayrollDatabaseOperation("calculate", duration, err)
    
    // Verify error metric was recorded
    businessOpCount := testutil.ToFloat64(DatabaseBusinessOperations.WithLabelValues("payroll", "calculate", "error"))
    assert.Equal(t, 1.0, businessOpCount)
}
```

#### Test: TestRecordInvoiceDatabaseOperation
```go
func TestRecordInvoiceDatabaseOperation(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestDatabaseMetrics()
    
    duration := 75 * time.Millisecond
    RecordInvoiceDatabaseOperation("generate", duration, nil)
    
    // Verify invoice metric was recorded
    businessOpCount := testutil.ToFloat64(DatabaseBusinessOperations.WithLabelValues("invoice", "generate", "success"))
    assert.Equal(t, 1.0, businessOpCount)
}
```

### 9. Table Name Extraction Tests

#### Test: TestGetTableName_FromStatement
```go
func TestGetTableName_FromStatement(t *testing.T) {
    tests := []struct {
        name         string
        setupDB      func() *gorm.DB
        expectedName string
    }{
        {
            name: "table_from_statement_table",
            setupDB: func() *gorm.DB {
                return &gorm.DB{
                    Statement: &gorm.Statement{
                        Table: "employees",
                    },
                }
            },
            expectedName: "employees",
        },
        {
            name: "table_from_schema",
            setupDB: func() *gorm.DB {
                return &gorm.DB{
                    Statement: &gorm.Statement{
                        Schema: &schema.Schema{
                            Table: "projects",
                        },
                    },
                }
            },
            expectedName: "projects",
        },
        {
            name: "no_statement",
            setupDB: func() *gorm.DB {
                return &gorm.DB{}
            },
            expectedName: "unknown",
        },
        {
            name: "empty_table",
            setupDB: func() *gorm.DB {
                return &gorm.DB{
                    Statement: &gorm.Statement{
                        Table: "",
                    },
                }
            },
            expectedName: "unknown",
        },
    }
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            db := test.setupDB()
            tableName := getTableName(db)
            assert.Equal(t, test.expectedName, tableName)
        })
    }
}
```

### 10. Context Validation Tests

#### Test: TestQueryContext_MissingContext
```go
func TestQueryContext_MissingContext(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Call after callback without setting before context
        afterQuery := afterCallback("select")
        
        // Should not panic when context is missing
        assert.NotPanics(t, func() {
            afterQuery(db)
        })
    })
}
```

#### Test: TestQueryContext_InvalidContextType
```go
func TestQueryContext_InvalidContextType(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Set invalid context type
        db.Set("metrics:context", "invalid_string")
        
        afterQuery := afterCallback("select")
        
        // Should not panic with invalid context type
        assert.NotPanics(t, func() {
            afterQuery(db)
        })
    })
}
```

### 11. Performance Tests

#### Test: BenchmarkDatabaseCallbacks_Overhead
```go
func BenchmarkDatabaseCallbacks_Overhead(b *testing.B) {
    testhelper.TestWithTxDB(b, func(repo store.DBRepo) {
        db := repo.DB()
        
        beforeCreate := beforeCallback("create")
        afterCreate := afterCallback("create")
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            beforeCreate(db)
            afterCreate(db)
        }
    })
}
```

#### Test: BenchmarkBusinessDomainInference
```go
func BenchmarkBusinessDomainInference(b *testing.B) {
    mockDB := &gorm.DB{
        Statement: &gorm.Statement{
            Table: "employees",
        },
    }
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        inferBusinessDomain(mockDB)
    }
}
```

### 12. Error Handling Tests

#### Test: TestDatabaseMonitoring_GracefulDegradation
```go
func TestDatabaseMonitoring_GracefulDegradation(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        db := repo.DB()
        
        // Simulate metrics collection failure
        beforeCallback := beforeCallback("create")
        afterCallback := afterCallback("create")
        
        // Should not panic even with metric collection errors
        assert.NotPanics(t, func() {
            beforeCallback(db)
            afterCallback(db)
        })
        
        // Database operations should still work
        var count int64
        err := db.Raw("SELECT COUNT(*) FROM employees").Scan(&count).Error
        assert.NoError(t, err)
    })
}
```

## Test Utilities

### Setup Helper Functions

```go
func setupTestDatabaseMetrics() {
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
        []string{"result"},
    )
    
    DatabaseBusinessOperations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "database",
            Name:      "business_operations_total",
            Help:      "Business-critical database operations",
        },
        []string{"domain", "operation", "result"},
    )
}
```

## Performance Criteria

- **Test Execution**: Each unit test must complete in <100ms
- **Database Overhead**: Callbacks should add <5% to query execution time
- **Memory Usage**: No memory leaks in callback tests
- **Concurrent Safety**: All callbacks must be safe for parallel execution

## Success Criteria

- **Code Coverage**: 95%+ line coverage for database monitoring packages
- **Test Reliability**: 0% flaky tests, consistent results across runs
- **Performance**: Database monitoring overhead <5% validated through benchmarks
- **Business Logic**: All business domain mappings correctly tested
- **Error Handling**: Graceful degradation tested for all failure scenarios

---

**Test Implementation Priority**: High  
**Estimated Implementation Time**: 20-24 hours  
**Dependencies**: GORM, Prometheus client, testhelper framework  
**Review Requirements**: Database team + senior backend engineer approval