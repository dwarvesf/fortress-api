# Integration Test Plan: Fortress API Monitoring

**Target**: Complete monitoring system integration  
**Test Count**: 35 integration tests  
**Coverage Target**: End-to-end monitoring workflows  
**Performance Target**: Each test <5s  

## Test Categories

### 1. HTTP Middleware Integration Tests (8 tests)

#### Test: TestHTTPMonitoring_EndToEnd_Success
```go
func TestHTTPMonitoring_EndToEnd_Success(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        // Setup router with monitoring enabled
        router := setupTestRouterWithMonitoring(repo, true)
        
        // Make authenticated request
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req.Header.Set("Authorization", validJWTToken)
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, 200, w.Code)
        
        // Verify metrics were recorded
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/metadata/stacks\",status=\"200\"} 1")
        assert.Contains(t, metricsResp, "fortress_http_request_duration_seconds")
    })
}
```

#### Test: TestHTTPMonitoring_MultipleRequests_Aggregation  
```go
func TestHTTPMonitoring_MultipleRequests_Aggregation(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithMonitoring(repo, true)
        
        // Make multiple requests to same endpoint
        requests := []struct {
            method string
            path   string
            status int
        }{
            {"GET", "/api/v1/metadata/stacks", 200},
            {"GET", "/api/v1/metadata/stacks", 200},
            {"GET", "/api/v1/metadata/stacks", 200},
            {"POST", "/api/v1/auth", 401},
            {"POST", "/api/v1/auth", 401},
        }
        
        for _, req := range requests {
            w := httptest.NewRecorder()
            httpReq := httptest.NewRequest(req.method, req.path, nil)
            if req.status == 200 {
                httpReq.Header.Set("Authorization", validJWTToken)
            }
            router.ServeHTTP(w, httpReq)
            assert.Equal(t, req.status, w.Code)
        }
        
        // Verify aggregated metrics
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/metadata/stacks\",status=\"200\"} 3")
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"POST\",endpoint=\"/api/v1/auth\",status=\"401\"} 2")
    })
}
```

#### Test: TestHTTPMonitoring_PathNormalization_Integration
```go
func TestHTTPMonitoring_PathNormalization_Integration(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithMonitoring(repo, true)
        
        // Create test employee first
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/create_employee.sql")
        
        // Make requests to parameterized endpoints
        employeeIDs := []string{"2655832e-f009-4b73-a535-64c3a22e558f", "3655832e-f009-4b73-a535-64c3a22e558f"}
        
        for _, id := range employeeIDs {
            w := httptest.NewRecorder()
            req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/employees/%s", id), nil)
            req.Header.Set("Authorization", validJWTToken)
            router.ServeHTTP(w, req)
            assert.Equal(t, 200, w.Code)
        }
        
        // Verify paths were normalized
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/employees/:id\",status=\"200\"} 2")
        
        // Should not contain individual employee IDs in metrics
        assert.NotContains(t, metricsResp, "2655832e-f009-4b73-a535-64c3a22e558f")
        assert.NotContains(t, metricsResp, "3655832e-f009-4b73-a535-64c3a22e558f")
    })
}
```

### 2. Database Monitoring Integration Tests (12 tests)

#### Test: TestDatabaseMonitoring_GORM_Operations
```go
func TestDatabaseMonitoring_GORM_Operations(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        // Enable database monitoring
        setupDatabaseMonitoringForTest(repo.DB())
        
        // Perform various database operations
        employee := &model.Employee{
            FullName: "Test Employee",
            Username: "testuser",
            Email:    "test@example.com",
        }
        
        // Create operation
        err := repo.DB().Create(employee).Error
        require.NoError(t, err)
        
        // Read operation
        var retrievedEmployee model.Employee
        err = repo.DB().First(&retrievedEmployee, employee.ID).Error
        require.NoError(t, err)
        
        // Update operation
        err = repo.DB().Model(&retrievedEmployee).Update("full_name", "Updated Name").Error
        require.NoError(t, err)
        
        // Delete operation
        err = repo.DB().Delete(&retrievedEmployee).Error
        require.NoError(t, err)
        
        // Verify all operations were tracked
        metricsResp := getMetricsEndpoint(setupTestRouterWithMonitoring(repo, true))
        assert.Contains(t, metricsResp, "fortress_database_operations_total{operation=\"create\",table=\"employees\",result=\"success\"} 1")
        assert.Contains(t, metricsResp, "fortress_database_operations_total{operation=\"select\",table=\"employees\",result=\"success\"} 1")
        assert.Contains(t, metricsResp, "fortress_database_operations_total{operation=\"update\",table=\"employees\",result=\"success\"} 1")
        assert.Contains(t, metricsResp, "fortress_database_operations_total{operation=\"delete\",table=\"employees\",result=\"success\"} 1")
    })
}
```

#### Test: TestDatabaseMonitoring_BusinessMetrics_PayrollFlow
```go
func TestDatabaseMonitoring_BusinessMetrics_PayrollFlow(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        setupDatabaseMonitoringForTest(repo.DB())
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/payroll_test_data.sql")
        
        // Simulate payroll calculation workflow
        var employees []model.Employee
        err := repo.DB().Where("working_status = ?", "active").Find(&employees).Error
        require.NoError(t, err)
        
        // Create payroll records
        for _, employee := range employees {
            payroll := &model.Payroll{
                EmployeeID: employee.ID,
                Month:      "2024-01",
                BaseSalary: 1000,
                Status:     "draft",
            }
            err = repo.DB().Create(payroll).Error
            require.NoError(t, err)
        }
        
        // Verify business domain metrics were recorded
        metricsResp := getMetricsEndpoint(setupTestRouterWithMonitoring(repo, true))
        assert.Contains(t, metricsResp, "fortress_database_business_operations_total{domain=\"hr\",operation=\"select\",result=\"success\"}")
        assert.Contains(t, metricsResp, "fortress_database_business_operations_total{domain=\"finance\",operation=\"create\",result=\"success\"}")
    })
}
```

#### Test: TestDatabaseMonitoring_SlowQueryDetection
```go
func TestDatabaseMonitoring_SlowQueryDetection(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        setupDatabaseMonitoringForTest(repo.DB())
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/large_dataset.sql")
        
        // Execute query that should be slow (intentionally inefficient)
        var count int64
        err := repo.DB().Raw(`
            SELECT COUNT(*) FROM employees e1 
            CROSS JOIN employees e2 
            WHERE e1.id != e2.id
        `).Scan(&count).Error
        require.NoError(t, err)
        
        // Verify slow query was detected
        metricsResp := getMetricsEndpoint(setupTestRouterWithMonitoring(repo, true))
        assert.Contains(t, metricsResp, "fortress_database_slow_queries_total")
    })
}
```

#### Test: TestDatabaseMonitoring_TransactionTracking
```go
func TestDatabaseMonitoring_TransactionTracking(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        setupDatabaseMonitoringForTest(repo.DB())
        
        // Successful transaction
        tx := repo.DB().Begin()
        employee := &model.Employee{
            FullName: "Transaction Test",
            Username: "txtest",
            Email:    "tx@example.com",
        }
        err := tx.Create(employee).Error
        require.NoError(t, err)
        tx.Commit()
        
        // Failed transaction (rollback)
        tx2 := repo.DB().Begin()
        invalidEmployee := &model.Employee{
            FullName: "Invalid",
            // Missing required username
        }
        tx2.Create(invalidEmployee) // This should fail
        tx2.Rollback()
        
        // Verify transaction metrics
        metricsResp := getMetricsEndpoint(setupTestRouterWithMonitoring(repo, true))
        assert.Contains(t, metricsResp, "fortress_database_transactions_total{result=\"commit\"} 1")
        assert.Contains(t, metricsResp, "fortress_database_transactions_total{result=\"rollback\"} 1")
    })
}
```

### 3. Security Monitoring Integration Tests (15 tests)

#### Test: TestSecurityMonitoring_Authentication_JWT_Flow
```go
func TestSecurityMonitoring_Authentication_JWT_Flow(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithSecurityMonitoring(repo, true)
        
        // Test successful JWT authentication
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w, req)
        assert.Equal(t, 200, w.Code)
        
        // Test failed JWT authentication
        w2 := httptest.NewRecorder()
        req2 := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req2.Header.Set("Authorization", "Bearer invalid-token")
        router.ServeHTTP(w2, req2)
        assert.Equal(t, 401, w2.Code)
        
        // Verify authentication metrics
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"jwt\",result=\"success\",reason=\"\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"jwt\",result=\"failure\",reason=\"invalid_token\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_duration_seconds")
    })
}
```

#### Test: TestSecurityMonitoring_Authorization_PermissionFlow
```go
func TestSecurityMonitoring_Authorization_PermissionFlow(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithSecurityMonitoring(repo, true)
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/permissions_test_data.sql")
        
        // Test successful authorization (user with employee.read permission)
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/employees", nil)
        req.Header.Set("Authorization", "Bearer "+employeeReadJWTToken)
        router.ServeHTTP(w, req)
        assert.Equal(t, 200, w.Code)
        
        // Test failed authorization (user without employee.delete permission)
        w2 := httptest.NewRecorder()
        req2 := httptest.NewRequest("DELETE", "/api/v1/employees/123", nil)
        req2.Header.Set("Authorization", "Bearer "+employeeReadJWTToken)
        router.ServeHTTP(w2, req2)
        assert.Equal(t, 403, w2.Code)
        
        // Verify authorization metrics
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_auth_permission_checks_total{permission=\"employees.read\",result=\"allowed\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_permission_checks_total{permission=\"employees.delete\",result=\"denied\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_authorization_duration_seconds")
    })
}
```

#### Test: TestSecurityMonitoring_APIKey_Authentication
```go
func TestSecurityMonitoring_APIKey_Authentication(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithSecurityMonitoring(repo, true)
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/api_keys_test_data.sql")
        
        // Test valid API key
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req.Header.Set("Authorization", "ApiKey "+validAPIKey)
        router.ServeHTTP(w, req)
        assert.Equal(t, 200, w.Code)
        
        // Test invalid API key
        w2 := httptest.NewRecorder()
        req2 := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req2.Header.Set("Authorization", "ApiKey invalid-key")
        router.ServeHTTP(w2, req2)
        assert.Equal(t, 401, w2.Code)
        
        // Verify API key metrics
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"api_key\",result=\"success\",reason=\"\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"api_key\",result=\"failure\",reason=\"invalid_key\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_api_key_usage_total{client_id=\"test_client\",result=\"success\"} 1")
        assert.Contains(t, metricsResp, "fortress_auth_api_key_validation_errors_total{error_type=\"key_not_found\"} 1")
    })
}
```

#### Test: TestSecurityMonitoring_BruteForce_Detection
```go
func TestSecurityMonitoring_BruteForce_Detection(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithSecurityMonitoring(repo, true)
        
        clientIP := "192.168.1.100"
        
        // Simulate brute force attack (10+ failed attempts within 5 minutes)
        for i := 0; i < 12; i++ {
            w := httptest.NewRecorder()
            req := httptest.NewRequest("POST", "/api/v1/auth", nil)
            req.Header.Set("Authorization", "Bearer invalid-token-"+strconv.Itoa(i))
            req.RemoteAddr = clientIP + ":12345"
            router.ServeHTTP(w, req)
            assert.Equal(t, 401, w.Code)
        }
        
        // Verify brute force detection
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_security_suspicious_activity_total{event_type=\"brute_force\",severity=\"high\"}")
    })
}
```

#### Test: TestSecurityMonitoring_SuspiciousPattern_Detection
```go
func TestSecurityMonitoring_SuspiciousPattern_Detection(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupTestRouterWithSecurityMonitoring(repo, true)
        
        clientIP := "192.168.1.200"
        
        // Simulate rapid requests pattern
        for i := 0; i < 150; i++ {
            w := httptest.NewRecorder()
            req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
            req.Header.Set("Authorization", "Bearer "+validJWTToken)
            req.Header.Set("User-Agent", "Bot"+strconv.Itoa(i%10))
            req.RemoteAddr = clientIP + ":12345"
            router.ServeHTTP(w, req)
            assert.Equal(t, 200, w.Code)
        }
        
        // Verify suspicious patterns were detected
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_security_suspicious_activity_total{event_type=\"rapid_requests\",severity=\"medium\"}")
        assert.Contains(t, metricsResp, "fortress_security_suspicious_activity_total{event_type=\"user_agent_variation\",severity=\"medium\"}")
    })
}
```

### 4. End-to-End Workflow Tests (10 tests)

#### Test: TestMonitoring_CompleteUserJourney
```go
func TestMonitoring_CompleteUserJourney(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupCompleteMonitoring(repo, true)
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/complete_test_data.sql")
        
        // Complete user journey: Login → View employees → Create project → Generate invoice
        
        // 1. Authentication
        w1 := httptest.NewRecorder()
        req1 := httptest.NewRequest("POST", "/api/v1/auth", nil)
        req1.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w1, req1)
        assert.Equal(t, 200, w1.Code)
        
        // 2. View employees (with database query)
        w2 := httptest.NewRecorder()
        req2 := httptest.NewRequest("GET", "/api/v1/employees", nil)
        req2.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w2, req2)
        assert.Equal(t, 200, w2.Code)
        
        // 3. Create project (database write)
        projectData := `{"name": "Test Project", "status": "active"}`
        w3 := httptest.NewRecorder()
        req3 := httptest.NewRequest("POST", "/api/v1/projects", strings.NewReader(projectData))
        req3.Header.Set("Authorization", "Bearer "+validJWTToken)
        req3.Header.Set("Content-Type", "application/json")
        router.ServeHTTP(w3, req3)
        assert.Equal(t, 201, w3.Code)
        
        // 4. Generate invoice (business operation)
        invoiceData := `{"client_id": "test-client", "amount": 5000}`
        w4 := httptest.NewRecorder()
        req4 := httptest.NewRequest("POST", "/api/v1/invoices", strings.NewReader(invoiceData))
        req4.Header.Set("Authorization", "Bearer "+validJWTToken)
        req4.Header.Set("Content-Type", "application/json")
        router.ServeHTTP(w4, req4)
        assert.Equal(t, 201, w4.Code)
        
        // Verify comprehensive monitoring captured all aspects
        metricsResp := getMetricsEndpoint(router)
        
        // HTTP monitoring
        assert.Contains(t, metricsResp, "fortress_http_requests_total")
        assert.Contains(t, metricsResp, "fortress_http_request_duration_seconds")
        
        // Security monitoring
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"jwt\",result=\"success\"}")
        assert.Contains(t, metricsResp, "fortress_auth_permission_checks_total")
        
        // Database monitoring
        assert.Contains(t, metricsResp, "fortress_database_operations_total")
        assert.Contains(t, metricsResp, "fortress_database_business_operations_total{domain=\"hr\"}")
        assert.Contains(t, metricsResp, "fortress_database_business_operations_total{domain=\"project_management\"}")
        assert.Contains(t, metricsResp, "fortress_database_business_operations_total{domain=\"finance\"}")
    })
}
```

#### Test: TestMonitoring_ErrorPropagation_Workflow
```go
func TestMonitoring_ErrorPropagation_Workflow(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupCompleteMonitoring(repo, true)
        
        // Test error scenarios across all monitoring components
        
        // 1. Authentication failure
        w1 := httptest.NewRecorder()
        req1 := httptest.NewRequest("GET", "/api/v1/employees", nil)
        req1.Header.Set("Authorization", "Bearer invalid-token")
        router.ServeHTTP(w1, req1)
        assert.Equal(t, 401, w1.Code)
        
        // 2. Authorization failure (valid auth, insufficient permissions)
        w2 := httptest.NewRecorder()
        req2 := httptest.NewRequest("DELETE", "/api/v1/employees/123", nil)
        req2.Header.Set("Authorization", "Bearer "+limitedPermissionJWTToken)
        router.ServeHTTP(w2, req2)
        assert.Equal(t, 403, w2.Code)
        
        // 3. Database error (non-existent resource)
        w3 := httptest.NewRecorder()
        req3 := httptest.NewRequest("GET", "/api/v1/employees/99999999-9999-9999-9999-999999999999", nil)
        req3.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w3, req3)
        assert.Equal(t, 404, w3.Code)
        
        // Verify error metrics were recorded properly
        metricsResp := getMetricsEndpoint(router)
        
        // Authentication errors
        assert.Contains(t, metricsResp, "fortress_auth_attempts_total{method=\"jwt\",result=\"failure\"}")
        
        // Authorization errors
        assert.Contains(t, metricsResp, "fortress_auth_permission_checks_total{permission=\"employees.delete\",result=\"denied\"}")
        
        // HTTP errors
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/employees\",status=\"401\"}")
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"DELETE\",endpoint=\"/api/v1/employees/:id\",status=\"403\"}")
        assert.Contains(t, metricsResp, "fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/employees/:id\",status=\"404\"}")
        
        // Database errors
        assert.Contains(t, metricsResp, "fortress_database_operations_total{operation=\"select\",table=\"employees\",result=\"error\"}")
    })
}
```

### 5. Performance Integration Tests (6 tests)

#### Test: TestMonitoring_PerformanceImpact_HTTPOverhead
```go
func TestMonitoring_PerformanceImpact_HTTPOverhead(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        // Test with monitoring disabled
        routerWithoutMonitoring := setupTestRouterWithMonitoring(repo, false)
        baselineDuration := measureHTTPPerformance(t, routerWithoutMonitoring, "/api/v1/metadata/stacks", 100)
        
        // Test with monitoring enabled
        routerWithMonitoring := setupTestRouterWithMonitoring(repo, true)
        monitoringDuration := measureHTTPPerformance(t, routerWithMonitoring, "/api/v1/metadata/stacks", 100)
        
        // Calculate overhead percentage
        overhead := (monitoringDuration - baselineDuration) / baselineDuration * 100
        
        // Verify overhead is less than 2%
        assert.LessOrEqual(t, overhead, 2.0, "HTTP monitoring overhead exceeds 2%: %.2f%%", overhead)
        
        t.Logf("HTTP Performance - Baseline: %v, With Monitoring: %v, Overhead: %.2f%%", 
            baselineDuration, monitoringDuration, overhead)
    })
}
```

#### Test: TestMonitoring_PerformanceImpact_DatabaseOverhead
```go
func TestMonitoring_PerformanceImpact_DatabaseOverhead(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        testhelper.LoadTestSQLFile(t, repo, "testdata/integration/performance_test_data.sql")
        
        // Measure database performance without monitoring
        baselineDuration := measureDatabasePerformance(t, repo, false, 50)
        
        // Measure database performance with monitoring
        setupDatabaseMonitoringForTest(repo.DB())
        monitoringDuration := measureDatabasePerformance(t, repo, true, 50)
        
        // Calculate overhead percentage
        overhead := (monitoringDuration - baselineDuration) / baselineDuration * 100
        
        // Verify overhead is less than 5%
        assert.LessOrEqual(t, overhead, 5.0, "Database monitoring overhead exceeds 5%: %.2f%%", overhead)
        
        t.Logf("Database Performance - Baseline: %v, With Monitoring: %v, Overhead: %.2f%%", 
            baselineDuration, monitoringDuration, overhead)
    })
}
```

#### Test: TestMonitoring_MemoryUsage_Integration
```go
func TestMonitoring_MemoryUsage_Integration(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        // Measure memory usage before monitoring setup
        var baselineMemStats runtime.MemStats
        runtime.GC()
        runtime.ReadMemStats(&baselineMemStats)
        
        // Setup complete monitoring
        router := setupCompleteMonitoring(repo, true)
        
        // Generate load to populate metrics
        generateMonitoringLoad(t, router, 1000)
        
        // Measure memory usage after monitoring
        var monitoringMemStats runtime.MemStats
        runtime.GC()
        runtime.ReadMemStats(&monitoringMemStats)
        
        // Calculate additional memory usage
        additionalMemory := monitoringMemStats.Alloc - baselineMemStats.Alloc
        
        // Verify memory usage increase is reasonable (<50MB)
        assert.LessOrEqual(t, additionalMemory, uint64(50*1024*1024), 
            "Monitoring memory usage exceeds 50MB: %d bytes", additionalMemory)
        
        t.Logf("Memory Usage - Baseline: %d bytes, With Monitoring: %d bytes, Additional: %d bytes", 
            baselineMemStats.Alloc, monitoringMemStats.Alloc, additionalMemory)
    })
}
```

### 6. External Service Integration Tests (4 tests)

#### Test: TestPrometheusIntegration_MetricsExport
```go
func TestPrometheusIntegration_MetricsExport(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        router := setupCompleteMonitoring(repo, true)
        
        // Generate various metrics
        generateMonitoringLoad(t, router, 50)
        
        // Test Prometheus metrics endpoint
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/metrics", nil)
        router.ServeHTTP(w, req)
        
        assert.Equal(t, 200, w.Code)
        assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"))
        
        metricsBody := w.Body.String()
        
        // Verify Prometheus format
        assert.Contains(t, metricsBody, "# HELP")
        assert.Contains(t, metricsBody, "# TYPE")
        
        // Verify all metric types are present
        assert.Contains(t, metricsBody, "fortress_http_requests_total")
        assert.Contains(t, metricsBody, "fortress_http_request_duration_seconds")
        assert.Contains(t, metricsBody, "fortress_database_operations_total")
        assert.Contains(t, metricsBody, "fortress_auth_attempts_total")
        assert.Contains(t, metricsBody, "fortress_security_suspicious_activity_total")
        
        // Verify metric values are reasonable
        lines := strings.Split(metricsBody, "\n")
        metricValues := extractMetricValues(lines)
        assert.True(t, len(metricValues) > 0, "Should have metric values")
    })
}
```

## Test Utilities

### Helper Functions

```go
func setupTestRouterWithMonitoring(repo store.DBRepo, enabled bool) *gin.Engine {
    cfg := &config.Config{
        Monitoring: config.MonitoringConfig{
            Enabled: enabled,
        },
    }
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    if enabled {
        prometheusMiddleware := middleware.NewPrometheusMiddleware(&middleware.PrometheusConfig{
            Enabled: true,
        })
        r.Use(prometheusMiddleware.Handler())
        r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }
    
    // Add test routes
    r.GET("/api/v1/metadata/stacks", func(c *gin.Context) {
        c.JSON(200, gin.H{"stacks": []string{"go", "typescript"}})
    })
    
    r.GET("/api/v1/employees", func(c *gin.Context) {
        var employees []model.Employee
        repo.DB().Find(&employees)
        c.JSON(200, gin.H{"employees": employees})
    })
    
    r.GET("/api/v1/employees/:id", func(c *gin.Context) {
        var employee model.Employee
        id := c.Param("id")
        err := repo.DB().First(&employee, "id = ?", id).Error
        if err != nil {
            c.JSON(404, gin.H{"error": "not found"})
            return
        }
        c.JSON(200, employee)
    })
    
    return r
}

func setupCompleteMonitoring(repo store.DBRepo, enabled bool) *gin.Engine {
    router := setupTestRouterWithMonitoring(repo, enabled)
    
    if enabled {
        setupDatabaseMonitoringForTest(repo.DB())
        setupSecurityMonitoringForTest(router)
    }
    
    return router
}

func getMetricsEndpoint(router *gin.Engine) string {
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/metrics", nil)
    router.ServeHTTP(w, req)
    return w.Body.String()
}

func measureHTTPPerformance(t *testing.T, router *gin.Engine, path string, iterations int) time.Duration {
    start := time.Now()
    
    for i := 0; i < iterations; i++ {
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", path, nil)
        req.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w, req)
        assert.Equal(t, 200, w.Code)
    }
    
    return time.Since(start) / time.Duration(iterations)
}

func measureDatabasePerformance(t *testing.T, repo store.DBRepo, withMonitoring bool, iterations int) time.Duration {
    start := time.Now()
    
    for i := 0; i < iterations; i++ {
        var employees []model.Employee
        err := repo.DB().Where("working_status = ?", "active").Find(&employees).Error
        assert.NoError(t, err)
    }
    
    return time.Since(start) / time.Duration(iterations)
}

func generateMonitoringLoad(t *testing.T, router *gin.Engine, requests int) {
    endpoints := []string{
        "/api/v1/metadata/stacks",
        "/api/v1/employees",
        "/api/v1/projects",
    }
    
    for i := 0; i < requests; i++ {
        endpoint := endpoints[i%len(endpoints)]
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", endpoint, nil)
        req.Header.Set("Authorization", "Bearer "+validJWTToken)
        router.ServeHTTP(w, req)
    }
}

func extractMetricValues(lines []string) map[string]float64 {
    values := make(map[string]float64)
    for _, line := range lines {
        if strings.HasPrefix(line, "fortress_") && !strings.HasPrefix(line, "#") {
            parts := strings.Split(line, " ")
            if len(parts) == 2 {
                value := 0.0
                fmt.Sscanf(parts[1], "%f", &value)
                values[parts[0]] = value
            }
        }
    }
    return values
}

// Test data constants
const (
    validJWTToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk1ODMzMzA5NDUsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIn0.test"
    employeeReadJWTToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.employee_read_token.test"
    limitedPermissionJWTToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.limited_permissions.test"
    validAPIKey = "test_client.valid_secret_key"
)
```

## Performance Criteria

- **Test Execution**: Each integration test must complete in <5 seconds
- **HTTP Overhead**: <2% performance impact validated
- **Database Overhead**: <5% query performance impact validated  
- **Memory Usage**: <50MB additional memory for complete monitoring stack
- **End-to-End Latency**: Complete monitoring adds <10ms to request processing

## Success Criteria

- **Workflow Coverage**: 100% of critical business workflows monitored end-to-end
- **Integration Accuracy**: 99.9% metric accuracy across all monitoring components
- **Performance Validation**: All performance targets met under load
- **Error Handling**: Graceful degradation tested across all integration points
- **External Services**: Successful integration with Prometheus metrics format

---

**Test Implementation Priority**: High  
**Estimated Implementation Time**: 32-40 hours  
**Dependencies**: Complete unit test implementation, test data fixtures  
**Review Requirements**: Full engineering team approval + performance validation