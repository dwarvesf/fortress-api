# Performance Test Plan: Fortress API Monitoring

**Target**: Complete monitoring system performance validation  
**Test Count**: 15 performance tests  
**Coverage Target**: All monitoring components under load  
**Performance Targets**: <2% HTTP overhead, <5% database overhead, <50MB memory  

## Performance Test Categories

### 1. Baseline Performance Tests (3 tests)

#### Test: BenchmarkHTTP_WithoutMonitoring
```go
func BenchmarkHTTP_WithoutMonitoring(b *testing.B) {
    testhelper.TestWithTxDB(b, func(repo store.DBRepo) {
        router := setupTestRouterWithoutMonitoring(repo)
        
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req.Header.Set("Authorization", validJWTToken)
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            w.Body.Reset()
            router.ServeHTTP(w, req)
            if w.Code != 200 {
                b.Fatalf("Expected 200, got %d", w.Code)
            }
        }
    })
}
```

#### Test: BenchmarkDatabase_WithoutMonitoring
```go
func BenchmarkDatabase_WithoutMonitoring(b *testing.B) {
    testhelper.TestWithTxDB(b, func(repo store.DBRepo) {
        // Seed test data
        setupPerformanceTestData(repo)
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            var employees []model.Employee
            err := repo.DB().Where("working_status = ?", "active").Find(&employees).Error
            if err != nil {
                b.Fatalf("Database query failed: %v", err)
            }
        }
    })
}
```

#### Test: BenchmarkSecurity_WithoutMonitoring
```go
func BenchmarkSecurity_WithoutMonitoring(b *testing.B) {
    testhelper.TestWithTxDB(b, func(repo store.DBRepo) {
        router := setupTestRouterWithoutSecurity(repo)
        
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/employees", nil)
        req.Header.Set("Authorization", validJWTToken)
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            w.Body.Reset()
            router.ServeHTTP(w, req)
            if w.Code != 200 {
                b.Fatalf("Expected 200, got %d", w.Code)
            }
        }
    })
}
```

### 2. Monitoring Overhead Tests (4 tests)

#### Test: BenchmarkHTTP_WithMonitoring
```go
func BenchmarkHTTP_WithMonitoring(b *testing.B) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestMetrics()
    
    testhelper.TestWithTxDB(b, func(repo store.DBRepo) {
        router := setupTestRouterWithMonitoring(repo, true)
        
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
        req.Header.Set("Authorization", validJWTToken)
        
        b.ResetTimer()
        b.ReportAllocs()
        
        for i := 0; i < b.N; i++ {
            w.Body.Reset()
            router.ServeHTTP(w, req)
            if w.Code != 200 {
                b.Fatalf("Expected 200, got %d", w.Code)
            }
        }
    })
}
```

#### Test: TestHTTPMonitoring_OverheadValidation
```go
func TestHTTPMonitoring_OverheadValidation(t *testing.T) {
    iterations := 10000
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        // Measure baseline performance
        routerBaseline := setupTestRouterWithoutMonitoring(repo)
        baselineTime := measureRequestTime(routerBaseline, "/api/v1/metadata/stacks", iterations)
        
        // Measure with monitoring
        prometheus.DefaultRegisterer = prometheus.NewRegistry()
        setupTestMetrics()
        routerWithMonitoring := setupTestRouterWithMonitoring(repo, true)
        monitoringTime := measureRequestTime(routerWithMonitoring, "/api/v1/metadata/stacks", iterations)
        
        // Calculate overhead
        overhead := (monitoringTime - baselineTime) / baselineTime * 100
        
        t.Logf("HTTP Monitoring Overhead: Baseline=%v, WithMonitoring=%v, Overhead=%.2f%%", 
            baselineTime, monitoringTime, overhead)
        
        // Verify overhead is within acceptable limits
        assert.LessOrEqual(t, overhead, 2.0, "HTTP monitoring overhead exceeds 2%%: %.2f%%", overhead)
    })
}
```

#### Test: TestDatabaseMonitoring_OverheadValidation  
```go
func TestDatabaseMonitoring_OverheadValidation(t *testing.T) {
    iterations := 1000
    
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        setupPerformanceTestData(repo)
        
        // Measure baseline database performance
        baselineTime := measureDatabaseQueryTime(repo, false, iterations)
        
        // Enable monitoring and measure again
        setupDatabaseMonitoringForTest(repo.DB())
        monitoringTime := measureDatabaseQueryTime(repo, true, iterations)
        
        // Calculate overhead
        overhead := (monitoringTime - baselineTime) / baselineTime * 100
        
        t.Logf("Database Monitoring Overhead: Baseline=%v, WithMonitoring=%v, Overhead=%.2f%%", 
            baselineTime, monitoringTime, overhead)
        
        // Verify overhead is within acceptable limits
        assert.LessOrEqual(t, overhead, 5.0, "Database monitoring overhead exceeds 5%%: %.2f%%", overhead)
    })
}
```

### 3. Load Testing (4 tests)

#### Test: TestHTTPLoad_HighThroughput
```go
func TestHTTPLoad_HighThroughput(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        prometheus.DefaultRegisterer = prometheus.NewRegistry()
        setupTestMetrics()
        router := setupCompleteMonitoring(repo, true)
        
        // Test configuration
        concurrentUsers := 50
        requestsPerUser := 100
        totalRequests := concurrentUsers * requestsPerUser
        
        // Rate limiting and timeout
        rateLimiter := time.NewTicker(time.Millisecond) // 1000 RPS max
        defer rateLimiter.Stop()
        
        timeout := time.After(60 * time.Second)
        done := make(chan bool, concurrentUsers)
        errors := make(chan error, totalRequests)
        
        start := time.Now()
        
        // Launch concurrent users
        for i := 0; i < concurrentUsers; i++ {
            go func(userID int) {
                defer func() { done <- true }()
                
                for j := 0; j < requestsPerUser; j++ {
                    select {
                    case <-timeout:
                        errors <- fmt.Errorf("timeout reached")
                        return
                    case <-rateLimiter.C:
                        // Execute request
                        w := httptest.NewRecorder()
                        req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
                        req.Header.Set("Authorization", validJWTToken)
                        req.Header.Set("User-Agent", fmt.Sprintf("LoadTest-User-%d", userID))
                        
                        router.ServeHTTP(w, req)
                        
                        if w.Code != 200 {
                            errors <- fmt.Errorf("user %d request %d failed: %d", userID, j, w.Code)
                        }
                    }
                }
            }(i)
        }
        
        // Wait for completion
        completedUsers := 0
        for completedUsers < concurrentUsers {
            select {
            case <-done:
                completedUsers++
            case <-timeout:
                t.Fatalf("Load test timed out after 60 seconds")
            }
        }
        
        duration := time.Since(start)
        rps := float64(totalRequests) / duration.Seconds()
        
        t.Logf("Load Test Results: %d requests in %v (%.2f RPS)", totalRequests, duration, rps)
        
        // Check for errors
        close(errors)
        errorCount := 0
        for err := range errors {
            t.Logf("Error: %v", err)
            errorCount++
        }
        
        errorRate := float64(errorCount) / float64(totalRequests) * 100
        assert.LessOrEqual(t, errorRate, 1.0, "Error rate exceeds 1%%: %.2f%%", errorRate)
        
        // Verify monitoring continued to work under load
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_http_requests_total")
        assert.Contains(t, metricsResp, fmt.Sprintf("fortress_http_requests_total{method=\"GET\",endpoint=\"/api/v1/metadata/stacks\",status=\"200\"} %d", totalRequests))
    })
}
```

#### Test: TestDatabaseLoad_HighQueryVolume
```go
func TestDatabaseLoad_HighQueryVolume(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        setupPerformanceTestData(repo)
        setupDatabaseMonitoringForTest(repo.DB())
        
        // Test configuration
        concurrentQueries := 20
        queriesPerGoRoutine := 100
        totalQueries := concurrentQueries * queriesPerGoRoutine
        
        start := time.Now()
        done := make(chan bool, concurrentQueries)
        errors := make(chan error, totalQueries)
        
        queryTypes := []func(store.DBRepo) error{
            func(r store.DBRepo) error {
                var employees []model.Employee
                return r.DB().Where("working_status = ?", "active").Find(&employees).Error
            },
            func(r store.DBRepo) error {
                var projects []model.Project
                return r.DB().Where("status = ?", "active").Find(&projects).Error
            },
            func(r store.DBRepo) error {
                var count int64
                return r.DB().Model(&model.Employee{}).Count(&count).Error
            },
        }
        
        // Launch concurrent database operations
        for i := 0; i < concurrentQueries; i++ {
            go func(goroutineID int) {
                defer func() { done <- true }()
                
                for j := 0; j < queriesPerGoRoutine; j++ {
                    queryType := queryTypes[j%len(queryTypes)]
                    err := queryType(repo)
                    if err != nil {
                        errors <- fmt.Errorf("goroutine %d query %d failed: %v", goroutineID, j, err)
                    }
                }
            }(i)
        }
        
        // Wait for completion
        for i := 0; i < concurrentQueries; i++ {
            <-done
        }
        
        duration := time.Since(start)
        qps := float64(totalQueries) / duration.Seconds()
        
        t.Logf("Database Load Test Results: %d queries in %v (%.2f QPS)", totalQueries, duration, qps)
        
        // Check for errors
        close(errors)
        errorCount := 0
        for err := range errors {
            t.Logf("Database Error: %v", err)
            errorCount++
        }
        
        errorRate := float64(errorCount) / float64(totalQueries) * 100
        assert.LessOrEqual(t, errorRate, 1.0, "Database error rate exceeds 1%%: %.2f%%", errorRate)
        
        // Verify database monitoring metrics were recorded
        router := setupTestRouterWithMonitoring(repo, true)
        metricsResp := getMetricsEndpoint(router)
        assert.Contains(t, metricsResp, "fortress_database_operations_total")
        assert.Contains(t, metricsResp, "fortress_database_operation_duration_seconds")
    })
}
```

### 4. Memory Performance Tests (2 tests)

#### Test: TestMonitoring_MemoryUsage_LongRunning
```go
func TestMonitoring_MemoryUsage_LongRunning(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        prometheus.DefaultRegisterer = prometheus.NewRegistry()
        setupTestMetrics()
        router := setupCompleteMonitoring(repo, true)
        
        // Measure initial memory
        var initialMem runtime.MemStats
        runtime.GC()
        runtime.ReadMemStats(&initialMem)
        
        // Generate sustained load for memory testing
        duration := 2 * time.Minute
        requestInterval := 10 * time.Millisecond
        start := time.Now()
        requestCount := 0
        
        for time.Since(start) < duration {
            // HTTP requests
            w := httptest.NewRecorder()
            req := httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
            req.Header.Set("Authorization", validJWTToken)
            router.ServeHTTP(w, req)
            
            // Database operations
            var employees []model.Employee
            repo.DB().Where("working_status = ?", "active").Find(&employees)
            
            requestCount++
            time.Sleep(requestInterval)
            
            // Periodic memory checks
            if requestCount%1000 == 0 {
                var currentMem runtime.MemStats
                runtime.ReadMemStats(&currentMem)
                memoryIncrease := currentMem.Alloc - initialMem.Alloc
                
                t.Logf("After %d requests: Memory increase = %d bytes", requestCount, memoryIncrease)
                
                // Early failure if memory grows too much
                if memoryIncrease > 100*1024*1024 { // 100MB
                    t.Fatalf("Memory usage exceeded 100MB after %d requests", requestCount)
                }
            }
        }
        
        // Final memory check
        runtime.GC()
        var finalMem runtime.MemStats
        runtime.ReadMemStats(&finalMem)
        
        totalMemoryIncrease := finalMem.Alloc - initialMem.Alloc
        avgMemoryPerRequest := float64(totalMemoryIncrease) / float64(requestCount)
        
        t.Logf("Memory Test Results: %d requests, %d bytes total increase, %.2f bytes per request", 
            requestCount, totalMemoryIncrease, avgMemoryPerRequest)
        
        // Verify memory usage is reasonable
        assert.LessOrEqual(t, totalMemoryIncrease, uint64(50*1024*1024), 
            "Total memory increase exceeds 50MB: %d bytes", totalMemoryIncrease)
        assert.LessOrEqual(t, avgMemoryPerRequest, 1024.0, 
            "Average memory per request exceeds 1KB: %.2f bytes", avgMemoryPerRequest)
    })
}
```

### 5. Scalability Tests (2 tests)

#### Test: TestMonitoring_MetricCardinality_Control
```go
func TestMonitoring_MetricCardinality_Control(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo store.DBRepo) {
        prometheus.DefaultRegisterer = prometheus.NewRegistry()
        setupTestMetrics()
        router := setupCompleteMonitoring(repo, true)
        
        // Generate diverse requests to test cardinality
        endpoints := []string{
            "/api/v1/metadata/stacks",
            "/api/v1/employees",
            "/api/v1/projects", 
            "/api/v1/invoices",
            "/api/v1/clients",
        }
        
        methods := []string{"GET", "POST", "PUT", "DELETE"}
        statusCodes := []int{200, 201, 400, 401, 403, 404, 500}
        
        requestCount := 0
        for _, endpoint := range endpoints {
            for _, method := range methods {
                for _, status := range statusCodes {
                    // Skip invalid combinations
                    if (method == "POST" && status == 200) || (method == "GET" && status == 201) {
                        continue
                    }
                    
                    w := httptest.NewRecorder()
                    req := httptest.NewRequest(method, endpoint, nil)
                    
                    // Manipulate response status for testing
                    if status >= 400 {
                        req.Header.Set("Authorization", "Bearer invalid-token")
                    } else {
                        req.Header.Set("Authorization", validJWTToken)
                    }
                    
                    router.ServeHTTP(w, req)
                    requestCount++
                }
            }
        }
        
        // Analyze metric cardinality
        metricsResp := getMetricsEndpoint(router)
        metricLines := strings.Split(metricsResp, "\n")
        
        uniqueMetrics := make(map[string]bool)
        for _, line := range metricLines {
            if strings.HasPrefix(line, "fortress_") && !strings.HasPrefix(line, "#") {
                metricParts := strings.Split(line, " ")
                if len(metricParts) >= 2 {
                    uniqueMetrics[metricParts[0]] = true
                }
            }
        }
        
        cardinalityCount := len(uniqueMetrics)
        t.Logf("Metric Cardinality: %d unique metric series from %d requests", cardinalityCount, requestCount)
        
        // Verify cardinality is controlled (should not explode)
        expectedMaxCardinality := len(endpoints) * len(methods) * len(statusCodes) * 3 // HTTP, DB, Security metrics
        assert.LessOrEqual(t, cardinalityCount, expectedMaxCardinality, 
            "Metric cardinality too high: %d (expected max: %d)", cardinalityCount, expectedMaxCardinality)
        assert.LessOrEqual(t, cardinalityCount, 1000, 
            "Metric cardinality exceeds 1000: %d", cardinalityCount)
    })
}
```

## Performance Test Utilities

### Helper Functions

```go
func measureRequestTime(router *gin.Engine, endpoint string, iterations int) time.Duration {
    start := time.Now()
    
    for i := 0; i < iterations; i++ {
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", endpoint, nil)
        req.Header.Set("Authorization", validJWTToken)
        router.ServeHTTP(w, req)
        if w.Code != 200 {
            panic(fmt.Sprintf("Request failed: %d", w.Code))
        }
    }
    
    return time.Since(start) / time.Duration(iterations)
}

func measureDatabaseQueryTime(repo store.DBRepo, withMonitoring bool, iterations int) time.Duration {
    start := time.Now()
    
    for i := 0; i < iterations; i++ {
        var employees []model.Employee
        err := repo.DB().Where("working_status = ?", "active").Find(&employees).Error
        if err != nil {
            panic(fmt.Sprintf("Database query failed: %v", err))
        }
    }
    
    return time.Since(start) / time.Duration(iterations)
}

func setupPerformanceTestData(repo store.DBRepo) {
    // Create test employees for performance testing
    employees := make([]model.Employee, 100)
    for i := 0; i < 100; i++ {
        employees[i] = model.Employee{
            FullName:      fmt.Sprintf("Employee %d", i),
            Username:      fmt.Sprintf("emp%d", i),
            Email:         fmt.Sprintf("emp%d@test.com", i),
            WorkingStatus: model.WorkingStatusActive,
        }
    }
    
    err := repo.DB().CreateInBatches(employees, 50).Error
    if err != nil {
        panic(fmt.Sprintf("Failed to create test data: %v", err))
    }
    
    // Create test projects
    projects := make([]model.Project, 50)
    for i := 0; i < 50; i++ {
        projects[i] = model.Project{
            Name:   fmt.Sprintf("Project %d", i),
            Status: model.ProjectStatusActive,
        }
    }
    
    err = repo.DB().CreateInBatches(projects, 25).Error
    if err != nil {
        panic(fmt.Sprintf("Failed to create test projects: %v", err))
    }
}

func setupTestRouterWithoutMonitoring(repo store.DBRepo) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    // Add test routes without any monitoring middleware
    r.GET("/api/v1/metadata/stacks", func(c *gin.Context) {
        c.JSON(200, gin.H{"stacks": []string{"go", "typescript", "react", "nodejs"}})
    })
    
    r.GET("/api/v1/employees", func(c *gin.Context) {
        var employees []model.Employee
        repo.DB().Where("working_status = ?", "active").Find(&employees)
        c.JSON(200, gin.H{"employees": employees})
    })
    
    return r
}

func getMemoryStats() runtime.MemStats {
    var mem runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&mem)
    return mem
}

func formatBytes(bytes uint64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## Performance Success Criteria

### Response Time Targets
- **HTTP Middleware**: <2% overhead on request processing time
- **Database Monitoring**: <5% overhead on query execution time
- **Security Monitoring**: <1% overhead on authentication operations
- **End-to-End Monitoring**: <10ms additional latency for complete request flow

### Resource Usage Targets  
- **Memory Usage**: <50MB additional memory for complete monitoring stack
- **CPU Usage**: <2% additional CPU usage under normal load
- **Storage**: <100MB per day for metric storage at 1000 RPS
- **Network**: <1KB per request for metric transmission

### Scalability Targets
- **Request Throughput**: Handle 1000+ RPS with monitoring enabled
- **Concurrent Users**: Support 100+ concurrent users with full monitoring
- **Metric Cardinality**: Maintain <10,000 unique metric series
- **Long-running Stability**: No memory leaks over 24+ hour runs

### Load Testing Targets
- **High Volume**: Process 100,000+ requests with <1% error rate
- **Sustained Load**: Maintain performance over 2+ hour duration
- **Burst Handling**: Handle 5x traffic spikes without degradation
- **Recovery**: Return to baseline performance within 60 seconds

## Performance Monitoring

### Continuous Monitoring
- **CI/CD Integration**: Performance regression detection in pipeline
- **Baseline Tracking**: Automated baseline performance measurement
- **Threshold Alerts**: Automated alerts when performance degrades
- **Trend Analysis**: Long-term performance trend tracking

### Performance Reports
- **Benchmark Results**: Detailed benchmark comparison reports
- **Resource Usage**: Memory and CPU usage analysis
- **Scalability Analysis**: Performance under various load scenarios
- **Regression Detection**: Automated detection of performance regressions

---

**Test Implementation Priority**: Critical  
**Estimated Implementation Time**: 24-30 hours  
**Dependencies**: Complete integration test implementation  
**Review Requirements**: Performance engineering team + senior backend approval