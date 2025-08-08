# Unit Test Plan: Security Monitoring (SPEC-003)

**Target**: pkg/middleware/security_monitoring.go, pkg/security/threat_detector.go, pkg/metrics/security.go  
**Test Count**: 30 unit tests  
**Coverage Target**: 95%+  
**Performance Target**: Each test <100ms  

## Test Categories

### 1. Security Middleware Initialization Tests

#### Test: TestNewSecurityMonitoringMiddleware_DefaultConfig
```go
func TestNewSecurityMonitoringMiddleware_DefaultConfig(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(nil, logger)
    
    assert.NotNil(t, middleware)
    assert.True(t, middleware.config.Enabled)
    assert.True(t, middleware.config.ThreatDetectionEnabled)
    assert.Equal(t, 10, middleware.config.BruteForceThreshold)
    assert.Equal(t, 5*time.Minute, middleware.config.BruteForceWindow)
    assert.True(t, middleware.config.SuspiciousPatternEnabled)
    assert.True(t, middleware.config.LogSecurityEvents)
    assert.NotNil(t, middleware.threatDetector)
}
```

#### Test: TestNewSecurityMonitoringMiddleware_CustomConfig
```go
func TestNewSecurityMonitoringMiddleware_CustomConfig(t *testing.T) {
    logger := logger.NewLogrusLogger()
    config := &SecurityMonitoringConfig{
        Enabled:                  false,
        ThreatDetectionEnabled:   false,
        BruteForceThreshold:     20,
        BruteForceWindow:        10 * time.Minute,
        SuspiciousPatternEnabled: false,
        LogSecurityEvents:       false,
    }
    
    middleware := NewSecurityMonitoringMiddleware(config, logger)
    
    assert.NotNil(t, middleware)
    assert.False(t, middleware.config.Enabled)
    assert.False(t, middleware.config.ThreatDetectionEnabled)
    assert.Equal(t, 20, middleware.config.BruteForceThreshold)
    assert.Equal(t, 10*time.Minute, middleware.config.BruteForceWindow)
    assert.False(t, middleware.config.SuspiciousPatternEnabled)
    assert.False(t, middleware.config.LogSecurityEvents)
}
```

### 2. Authentication Monitoring Tests

#### Test: TestRecordAuthenticationMetrics_JWT_Success
```go
func TestRecordAuthenticationMetrics_JWT_Success(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    c.Request.Header.Set("Authorization", "Bearer valid-jwt-token")
    c.Writer.WriteHeader(200)
    
    start := time.Now().Add(-50 * time.Millisecond)
    middleware.recordAuthenticationMetrics(c, start)
    
    // Verify authentication success metrics
    authCount := testutil.ToFloat64(AuthenticationAttempts.WithLabelValues("jwt", "success", ""))
    assert.Equal(t, 1.0, authCount)
    
    durationCount := testutil.ToFloat64(AuthenticationDuration.WithLabelValues("jwt"))
    assert.Equal(t, 1.0, durationCount)
}
```

#### Test: TestRecordAuthenticationMetrics_JWT_Failure
```go
func TestRecordAuthenticationMetrics_JWT_Failure(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{
        Enabled:                true,
        ThreatDetectionEnabled: true,
    }, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    c.Request.Header.Set("Authorization", "Bearer invalid-token")
    c.Set("auth_failure_reason", "invalid_token")
    c.Writer.WriteHeader(401)
    c.Set("client_ip", "192.168.1.100")
    
    start := time.Now().Add(-30 * time.Millisecond)
    middleware.recordAuthenticationMetrics(c, start)
    
    // Verify authentication failure metrics
    authCount := testutil.ToFloat64(AuthenticationAttempts.WithLabelValues("jwt", "failure", "invalid_token"))
    assert.Equal(t, 1.0, authCount)
    
    durationCount := testutil.ToFloat64(AuthenticationDuration.WithLabelValues("jwt"))
    assert.Equal(t, 1.0, durationCount)
}
```

#### Test: TestRecordAuthenticationMetrics_APIKey_Success
```go
func TestRecordAuthenticationMetrics_APIKey_Success(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    c.Request.Header.Set("Authorization", "ApiKey client123.secret456")
    c.Writer.WriteHeader(200)
    
    start := time.Now().Add(-25 * time.Millisecond)
    middleware.recordAuthenticationMetrics(c, start)
    
    // Verify API key success metrics
    authCount := testutil.ToFloat64(AuthenticationAttempts.WithLabelValues("api_key", "success", ""))
    assert.Equal(t, 1.0, authCount)
    
    apiKeyUsageCount := testutil.ToFloat64(APIKeyUsage.WithLabelValues("client123", "success"))
    assert.Equal(t, 1.0, apiKeyUsageCount)
}
```

#### Test: TestGetAuthMethod_VariousHeaders
```go
func TestGetAuthMethod_VariousHeaders(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    tests := []struct {
        authHeader     string
        expectedMethod string
    }{
        {"Bearer jwt-token-here", "jwt"},
        {"bearer jwt-token-here", "jwt"}, // case insensitive
        {"ApiKey client.secret", "api_key"},
        {"apikey client.secret", "api_key"}, // case insensitive
        {"Basic base64creds", "unknown"},
        {"", "unknown"},
        {"InvalidFormat", "unknown"},
    }
    
    for _, test := range tests {
        result := middleware.getAuthMethod(test.authHeader)
        assert.Equal(t, test.expectedMethod, result, "Auth header: %s", test.authHeader)
    }
}
```

### 3. Authorization Monitoring Tests

#### Test: TestRecordPermissionMetrics_Success
```go
func TestRecordPermissionMetrics_Success(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/employees", nil)
    c.Set("required_permission", "employees.read")
    c.Writer.WriteHeader(200)
    
    start := time.Now().Add(-15 * time.Millisecond)
    middleware.recordPermissionMetrics(c, start)
    
    // Should not record permission check for successful request (only failures)
    permissionCount := testutil.ToFloat64(PermissionChecks.WithLabelValues("employees.read", "denied"))
    assert.Equal(t, 0.0, permissionCount)
}
```

#### Test: TestRecordPermissionMetrics_Denied
```go
func TestRecordPermissionMetrics_Denied(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{
        Enabled:           true,
        LogSecurityEvents: true,
    }, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("DELETE", "/api/v1/employees/123", nil)
    c.Set("required_permission", "employees.delete")
    c.Writer.WriteHeader(403)
    
    start := time.Now().Add(-10 * time.Millisecond)
    middleware.recordPermissionMetrics(c, start)
    
    // Verify permission denied metrics
    permissionCount := testutil.ToFloat64(PermissionChecks.WithLabelValues("employees.delete", "denied"))
    assert.Equal(t, 1.0, permissionCount)
}
```

#### Test: TestGetRequiredPermission
```go
func TestGetRequiredPermission(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    tests := []struct {
        name               string
        setupContext       func(*gin.Context)
        expectedPermission string
    }{
        {
            name: "permission_set",
            setupContext: func(c *gin.Context) {
                c.Set("required_permission", "projects.write")
            },
            expectedPermission: "projects.write",
        },
        {
            name: "permission_not_set",
            setupContext: func(c *gin.Context) {
                // No permission set
            },
            expectedPermission: "unknown",
        },
        {
            name: "permission_wrong_type",
            setupContext: func(c *gin.Context) {
                c.Set("required_permission", 123) // Wrong type
            },
            expectedPermission: "unknown",
        },
    }
    
    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            gin.SetMode(gin.TestMode)
            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            c.Request = httptest.NewRequest("GET", "/test", nil)
            
            test.setupContext(c)
            
            permission := middleware.getRequiredPermission(c)
            assert.Equal(t, test.expectedPermission, permission)
        })
    }
}
```

### 4. Threat Detection Tests

#### Test: TestNewThreatDetector
```go
func TestNewThreatDetector(t *testing.T) {
    config := &SecurityMonitoringConfig{
        BruteForceThreshold: 5,
        BruteForceWindow:   2 * time.Minute,
    }
    
    detector := NewThreatDetector(config)
    
    assert.NotNil(t, detector)
    assert.Equal(t, config, detector.config)
    assert.NotNil(t, detector.failureTracker)
    assert.NotNil(t, detector.patternDetector)
    assert.NotNil(t, detector.patternDetector.requestPatterns)
}
```

#### Test: TestRecordAuthFailure_BruteForceDetection
```go
func TestRecordAuthFailure_BruteForceDetection(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    config := &SecurityMonitoringConfig{
        BruteForceThreshold: 3,
        BruteForceWindow:   1 * time.Minute,
    }
    detector := NewThreatDetector(config)
    
    clientIP := "192.168.1.100"
    
    // Record failures below threshold
    detector.RecordAuthFailure(clientIP, "jwt")
    detector.RecordAuthFailure(clientIP, "jwt")
    
    // No brute force detected yet
    bruteForceCount := testutil.ToFloat64(SuspiciousActivity.WithLabelValues("brute_force", "high"))
    assert.Equal(t, 0.0, bruteForceCount)
    
    // Record failure that exceeds threshold
    detector.RecordAuthFailure(clientIP, "jwt")
    
    // Brute force should be detected
    bruteForceCount = testutil.ToFloat64(SuspiciousActivity.WithLabelValues("brute_force", "high"))
    assert.Equal(t, 1.0, bruteForceCount)
}
```

#### Test: TestRecordAuthFailure_WindowExpiry
```go
func TestRecordAuthFailure_WindowExpiry(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    config := &SecurityMonitoringConfig{
        BruteForceThreshold: 3,
        BruteForceWindow:   100 * time.Millisecond, // Short window for testing
    }
    detector := NewThreatDetector(config)
    
    clientIP := "192.168.1.200"
    
    // Record failures
    detector.RecordAuthFailure(clientIP, "jwt")
    detector.RecordAuthFailure(clientIP, "jwt")
    
    // Wait for window to expire
    time.Sleep(150 * time.Millisecond)
    
    // Record another failure - should not trigger brute force (window expired)
    detector.RecordAuthFailure(clientIP, "jwt")
    
    bruteForceCount := testutil.ToFloat64(SuspiciousActivity.WithLabelValues("brute_force", "high"))
    assert.Equal(t, 0.0, bruteForceCount)
}
```

### 5. Pattern Detection Tests

#### Test: TestAnalyzeRequestPatterns_RapidRequests
```go
func TestAnalyzeRequestPatterns_RapidRequests(t *testing.T) {
    config := &SecurityMonitoringConfig{
        SuspiciousPatternEnabled: true,
    }
    detector := NewThreatDetector(config)
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    
    clientIP := "192.168.1.150"
    userAgent := "TestBot/1.0"
    
    // Make many rapid requests
    for i := 0; i < 150; i++ {
        patterns := detector.AnalyzeRequest(c, clientIP, userAgent)
        
        if i >= 100 { // Should detect rapid requests after 100
            assert.NotEmpty(t, patterns)
            found := false
            for _, pattern := range patterns {
                if pattern.Type == "rapid_requests" {
                    found = true
                    assert.False(t, pattern.IsCritical)
                    assert.Equal(t, clientIP, pattern.ClientIP)
                    break
                }
            }
            assert.True(t, found, "Rapid requests pattern should be detected")
        }
    }
}
```

#### Test: TestAnalyzeRequestPatterns_UserAgentVariation
```go
func TestAnalyzeRequestPatterns_UserAgentVariation(t *testing.T) {
    config := &SecurityMonitoringConfig{
        SuspiciousPatternEnabled: true,
    }
    detector := NewThreatDetector(config)
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    
    clientIP := "192.168.1.175"
    
    // Use many different user agents from same IP
    userAgents := []string{
        "Mozilla/5.0 (Chrome)", "Mozilla/5.0 (Firefox)", "Mozilla/5.0 (Safari)",
        "Googlebot", "Bingbot", "curl/7.68.0", "PostmanRuntime/7.26.8",
        "Python-requests/2.25.1", "Go-http-client/1.1", "wget/1.20.3",
        "Apache-HttpClient/4.5.12", // This is the 11th user agent, should trigger
    }
    
    var finalPatterns []SuspiciousPattern
    for _, ua := range userAgents {
        patterns := detector.AnalyzeRequest(c, clientIP, ua)
        if len(patterns) > 0 {
            finalPatterns = patterns
        }
    }
    
    // Should detect user agent variation
    assert.NotEmpty(t, finalPatterns)
    found := false
    for _, pattern := range finalPatterns {
        if pattern.Type == "user_agent_variation" {
            found = true
            assert.False(t, pattern.IsCritical)
            assert.Equal(t, clientIP, pattern.ClientIP)
            break
        }
    }
    assert.True(t, found, "User agent variation should be detected")
}
```

### 6. Client ID Extraction Tests

#### Test: TestExtractClientID_ValidFormat
```go
func TestExtractClientID_ValidFormat(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    tests := []struct {
        authHeader       string
        expectedClientID string
    }{
        {"ApiKey client123.secret456", "client123"},
        {"apikey dwarves.foundation.abcd1234", "dwarves"},
        {"ApiKey single_part", ""},              // Invalid format
        {"ApiKey ", ""},                         // Empty key
        {"Bearer jwt-token", ""},                // Wrong auth type
    }
    
    for _, test := range tests {
        clientID := middleware.extractClientID(test.authHeader)
        assert.Equal(t, test.expectedClientID, clientID, "Auth header: %s", test.authHeader)
    }
}
```

### 7. Failure Reason Tests

#### Test: TestGetFailureReason_ContextReason
```go
func TestGetFailureReason_ContextReason(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Set("auth_failure_reason", "token_expired")
    
    reason := middleware.getFailureReason(c, "jwt")
    assert.Equal(t, "token_expired", reason)
}
```

#### Test: TestGetFailureReason_DefaultReasons
```go
func TestGetFailureReason_DefaultReasons(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    tests := []struct {
        method         string
        expectedReason string
    }{
        {"jwt", "invalid_token"},
        {"api_key", "invalid_key"},
        {"unknown", "unknown"},
    }
    
    for _, test := range tests {
        reason := middleware.getFailureReason(c, test.method)
        assert.Equal(t, test.expectedReason, reason, "Method: %s", test.method)
    }
}
```

### 8. Threat Detector Cleanup Tests

#### Test: TestThreatDetector_Cleanup
```go
func TestThreatDetector_Cleanup(t *testing.T) {
    config := &SecurityMonitoringConfig{
        BruteForceThreshold: 5,
        BruteForceWindow:   1 * time.Hour,
    }
    detector := NewThreatDetector(config)
    
    clientIP := "192.168.1.250"
    
    // Add some failure tracking data
    detector.RecordAuthFailure(clientIP, "jwt")
    
    // Manually set old timestamp to simulate expired data
    if tracker, exists := detector.failureTracker[clientIP]; exists {
        tracker.LastSeen = time.Now().Add(-25 * time.Hour) // 25 hours ago
    }
    
    // Add pattern data
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("GET", "/test", nil)
    detector.AnalyzeRequest(c, clientIP, "TestAgent")
    
    // Manually set old timestamp for pattern data
    if pattern, exists := detector.patternDetector.requestPatterns[clientIP]; exists {
        pattern.LastSeen = time.Now().Add(-25 * time.Hour)
    }
    
    // Run cleanup
    detector.cleanup()
    
    // Verify old data was cleaned up
    _, failureExists := detector.failureTracker[clientIP]
    assert.False(t, failureExists, "Old failure tracking data should be cleaned up")
    
    _, patternExists := detector.patternDetector.requestPatterns[clientIP]
    assert.False(t, patternExists, "Old pattern data should be cleaned up")
}
```

### 9. Suspicious Activity Detection Tests

#### Test: TestDetectSuspiciousActivity_Disabled
```go
func TestDetectSuspiciousActivity_Disabled(t *testing.T) {
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{
        Enabled:                true,
        ThreatDetectionEnabled: false,
    }, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    
    clientIP := "192.168.1.300"
    userAgent := "SuspiciousBot/1.0"
    
    // Should not panic and should not detect anything when disabled
    assert.NotPanics(t, func() {
        middleware.detectSuspiciousActivity(c, clientIP, userAgent)
    })
}
```

#### Test: TestDetectSuspiciousActivity_MultiplePatterns
```go
func TestDetectSuspiciousActivity_MultiplePatterns(t *testing.T) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{
        Enabled:                  true,
        ThreatDetectionEnabled:   true,
        SuspiciousPatternEnabled: true,
        LogSecurityEvents:        true,
    }, logger)
    
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/v1/test", nil)
    
    clientIP := "192.168.1.350"
    userAgent := "AttackBot/2.0"
    
    // Simulate multiple patterns being detected
    middleware.detectSuspiciousActivity(c, clientIP, userAgent)
    
    // Should handle multiple patterns gracefully
    assert.True(t, true) // Test passes if no panic
}
```

### 10. Performance Tests

#### Test: BenchmarkSecurityMiddleware_Enabled
```go
func BenchmarkSecurityMiddleware_Enabled(b *testing.B) {
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    setupTestSecurityMetrics()
    
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
    r.Use(middleware.Handler())
    r.GET("/bench", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/bench", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        w.Body.Reset()
        r.ServeHTTP(w, req)
    }
}
```

#### Test: BenchmarkThreatDetection
```go
func BenchmarkThreatDetection(b *testing.B) {
    config := &SecurityMonitoringConfig{
        ThreatDetectionEnabled:   true,
        SuspiciousPatternEnabled: true,
    }
    detector := NewThreatDetector(config)
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("GET", "/api/test", nil)
    
    clientIP := "192.168.1.400"
    userAgent := "BenchmarkBot/1.0"
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        detector.AnalyzeRequest(c, clientIP, userAgent)
    }
}
```

### 11. Error Handling Tests

#### Test: TestSecurityMonitoring_GracefulDegradation
```go
func TestSecurityMonitoring_GracefulDegradation(t *testing.T) {
    // Test with invalid Prometheus registry to simulate metric errors
    logger := logger.NewLogrusLogger()
    middleware := NewSecurityMonitoringMiddleware(&SecurityMonitoringConfig{Enabled: true}, logger)
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(middleware.Handler())
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    
    // Should not panic even with metric collection errors
    assert.NotPanics(t, func() {
        r.ServeHTTP(w, req)
    })
    
    assert.Equal(t, 200, w.Code)
}
```

### 12. Request Pattern Tests

#### Test: TestRequestPattern_ConcurrentAccess
```go
func TestRequestPattern_ConcurrentAccess(t *testing.T) {
    config := &SecurityMonitoringConfig{
        SuspiciousPatternEnabled: true,
    }
    detector := NewThreatDetector(config)
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request = httptest.NewRequest("GET", "/api/test", nil)
    
    clientIP := "192.168.1.500"
    userAgent := "ConcurrentBot/1.0"
    
    // Test concurrent access to pattern detection
    var wg sync.WaitGroup
    concurrentRequests := 10
    
    wg.Add(concurrentRequests)
    for i := 0; i < concurrentRequests; i++ {
        go func() {
            defer wg.Done()
            detector.AnalyzeRequest(c, clientIP, userAgent)
        }()
    }
    
    wg.Wait()
    
    // Should handle concurrent access without panicking
    assert.True(t, true)
}
```

## Test Utilities

### Setup Helper Functions

```go
func setupTestSecurityMetrics() {
    AuthenticationAttempts = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "attempts_total",
            Help:      "Total authentication attempts by method and result",
        },
        []string{"method", "result", "reason"},
    )
    
    AuthenticationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "duration_seconds",
            Help:      "Authentication processing duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"method"},
    )
    
    PermissionChecks = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "permission_checks_total", 
            Help:      "Total permission checks by permission and result",
        },
        []string{"permission", "result"},
    )
    
    SuspiciousActivity = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "security",
            Name:      "suspicious_activity_total",
            Help:      "Suspicious security events detected",
        },
        []string{"event_type", "severity"},
    )
    
    APIKeyUsage = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "api_key_usage_total",
            Help:      "API key usage by client and result",
        },
        []string{"client_id", "result"},
    )
}

func mockLogger() logger.Logger {
    return logger.NewLogrusLogger()
}
```

## Performance Criteria

- **Test Execution**: Each unit test must complete in <100ms
- **Security Overhead**: Security monitoring should add <1% overhead to authentication
- **Memory Usage**: No memory leaks in threat detection components
- **Concurrent Safety**: All security components must be safe for parallel execution

## Success Criteria

- **Code Coverage**: 95%+ line coverage for security monitoring packages
- **Test Reliability**: 0% flaky tests, consistent threat detection accuracy
- **Performance**: Security monitoring overhead <1% validated through benchmarks
- **Privacy Compliance**: All tests verify no sensitive data exposure
- **Threat Detection**: Accurate threat pattern recognition in all test scenarios

---

**Test Implementation Priority**: Critical (Security-focused)  
**Estimated Implementation Time**: 24-28 hours  
**Dependencies**: Prometheus client, testify framework, security threat patterns  
**Review Requirements**: Security team + senior backend engineer approval required