# SPEC-003: Security Monitoring Implementation

**Status**: Ready for Implementation  
**Priority**: High  
**Estimated Effort**: 10-14 hours  
**Dependencies**: SPEC-001 (HTTP Middleware)  

## Overview

Implement comprehensive security monitoring for the Fortress API, focusing on authentication events, authorization failures, and suspicious activity detection. This specification integrates with the existing JWT + API Key authentication system and permission-based authorization middleware.

## Requirements

### Functional Requirements
- **FR-1**: Monitor authentication attempts (success/failure) for both JWT and API Key methods
- **FR-2**: Track authorization failures and permission violations
- **FR-3**: Detect suspicious activity patterns (brute force, unusual access patterns)
- **FR-4**: Monitor rate limiting violations and potential abuse
- **FR-5**: Provide security event logging for audit and compliance
- **FR-6**: Support security alerting based on threat patterns

### Non-Functional Requirements
- **NFR-1**: Security monitoring overhead <1% authentication performance impact  
- **NFR-2**: Real-time threat detection with <30 second detection latency
- **NFR-3**: Secure storage and handling of security event data
- **NFR-4**: Privacy-conscious logging (no sensitive data exposure)

## Technical Specification

### 1. Security Metrics Definition

```go
// pkg/metrics/security.go - Security-focused metrics
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Authentication metrics
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
    
    ActiveSessions = promauto.NewGauge(
        prometheus.GaugeOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "active_sessions_total",
            Help:      "Number of active authenticated sessions",
        },
    )
    
    // Authorization metrics
    PermissionChecks = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "permission_checks_total", 
            Help:      "Total permission checks by permission and result",
        },
        []string{"permission", "result"},
    )
    
    AuthorizationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "authorization_duration_seconds",
            Help:      "Authorization check processing duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1},
        },
        []string{"permission"},
    )
    
    // Security threat metrics
    SuspiciousActivity = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "security",
            Name:      "suspicious_activity_total",
            Help:      "Suspicious security events detected",
        },
        []string{"event_type", "severity"},
    )
    
    RateLimitViolations = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress", 
            Subsystem: "security",
            Name:      "rate_limit_violations_total",
            Help:      "Rate limiting violations by endpoint and client",
        },
        []string{"endpoint", "violation_type"},
    )
    
    SecurityEvents = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "security",
            Name:      "events_total",
            Help:      "Security events by type and severity",
        },
        []string{"event_type", "severity", "source"},
    )
    
    // API Key specific metrics
    APIKeyUsage = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth",
            Name:      "api_key_usage_total",
            Help:      "API key usage by client and result",
        },
        []string{"client_id", "result"},
    )
    
    APIKeyValidationErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "fortress",
            Subsystem: "auth", 
            Name:      "api_key_validation_errors_total",
            Help:      "API key validation errors by type",
        },
        []string{"error_type"},
    )
)
```

### 2. Enhanced Authentication Middleware

```go
// pkg/middleware/security_monitoring.go - Security monitoring middleware
package middleware

import (
    "strconv"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/metrics"
)

type SecurityMonitoringMiddleware struct {
    config *SecurityMonitoringConfig
    logger logger.Logger
    threatDetector *ThreatDetector
}

type SecurityMonitoringConfig struct {
    Enabled                    bool          `mapstructure:"enabled"`
    ThreatDetectionEnabled     bool          `mapstructure:"threat_detection_enabled"`
    BruteForceThreshold       int           `mapstructure:"brute_force_threshold"`
    BruteForceWindow          time.Duration `mapstructure:"brute_force_window"`
    SuspiciousPatternEnabled  bool          `mapstructure:"suspicious_pattern_enabled"`
    LogSecurityEvents         bool          `mapstructure:"log_security_events"`
}

func NewSecurityMonitoringMiddleware(cfg *SecurityMonitoringConfig, logger logger.Logger) *SecurityMonitoringMiddleware {
    if cfg == nil {
        cfg = &SecurityMonitoringConfig{
            Enabled:                   true,
            ThreatDetectionEnabled:    true,
            BruteForceThreshold:      10,
            BruteForceWindow:         time.Minute * 5,
            SuspiciousPatternEnabled: true,
            LogSecurityEvents:        true,
        }
    }
    
    return &SecurityMonitoringMiddleware{
        config:         cfg,
        logger:         logger,
        threatDetector: NewThreatDetector(cfg),
    }
}

func (smw *SecurityMonitoringMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !smw.config.Enabled {
            c.Next()
            return
        }
        
        start := time.Now()
        clientIP := c.ClientIP()
        userAgent := c.GetHeader("User-Agent")
        
        // Process request
        c.Next()
        
        // Record security metrics
        smw.recordAuthenticationMetrics(c, start)
        smw.recordPermissionMetrics(c, start)
        smw.detectSuspiciousActivity(c, clientIP, userAgent)
        smw.logSecurityEvent(c, clientIP, userAgent, start)
    }
}

func (smw *SecurityMonitoringMiddleware) recordAuthenticationMetrics(c *gin.Context, start time.Time) {
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        return
    }
    
    method := smw.getAuthMethod(authHeader)
    duration := time.Since(start).Seconds()
    
    // Record authentication duration
    metrics.AuthenticationDuration.WithLabelValues(method).Observe(duration)
    
    status := c.Writer.Status()
    result := "success"
    reason := ""
    
    if status == 401 {
        result = "failure"
        reason = smw.getFailureReason(c, method)
        
        // Record authentication failure
        metrics.AuthenticationAttempts.WithLabelValues(method, result, reason).Inc()
        
        // Check for brute force patterns
        if smw.config.ThreatDetectionEnabled {
            smw.threatDetector.RecordAuthFailure(c.ClientIP(), method)
        }
    } else if status < 400 {
        // Successful authentication
        metrics.AuthenticationAttempts.WithLabelValues(method, result, reason).Inc()
        
        // For API keys, record usage by client ID
        if method == "api_key" {
            clientID := smw.extractClientID(authHeader)
            if clientID != "" {
                metrics.APIKeyUsage.WithLabelValues(clientID, result).Inc()
            }
        }
    }
}

func (smw *SecurityMonitoringMiddleware) recordPermissionMetrics(c *gin.Context, start time.Time) {
    status := c.Writer.Status()
    
    if status == 403 {
        // Authorization failure
        permission := smw.getRequiredPermission(c)
        
        metrics.PermissionChecks.WithLabelValues(permission, "denied").Inc()
        
        // Log permission violation for security analysis
        if smw.config.LogSecurityEvents {
            smw.logger.Warn("Permission denied").
                AddField("permission", permission).
                AddField("client_ip", c.ClientIP()).
                AddField("user_agent", c.GetHeader("User-Agent")).
                AddField("endpoint", c.Request.URL.Path)
        }
    }
}

func (smw *SecurityMonitoringMiddleware) detectSuspiciousActivity(c *gin.Context, clientIP, userAgent string) {
    if !smw.config.ThreatDetectionEnabled {
        return
    }
    
    patterns := smw.threatDetector.AnalyzeRequest(c, clientIP, userAgent)
    
    for _, pattern := range patterns {
        severity := "medium"
        if pattern.IsCritical {
            severity = "high"
        }
        
        metrics.SuspiciousActivity.WithLabelValues(
            pattern.Type, severity,
        ).Inc()
        
        if smw.config.LogSecurityEvents {
            smw.logger.Warn("Suspicious activity detected").
                AddField("pattern_type", pattern.Type).
                AddField("severity", severity).
                AddField("client_ip", clientIP).
                AddField("details", pattern.Details)
        }
    }
}

func (smw *SecurityMonitoringMiddleware) getAuthMethod(authHeader string) string {
    parts := strings.Split(authHeader, " ")
    if len(parts) < 1 {
        return "unknown"
    }
    
    switch strings.ToLower(parts[0]) {
    case "bearer":
        return "jwt"
    case "apikey":
        return "api_key"
    default:
        return "unknown"
    }
}

func (smw *SecurityMonitoringMiddleware) getFailureReason(c *gin.Context, method string) string {
    // Extract failure reason from response or context
    if reason, exists := c.Get("auth_failure_reason"); exists {
        if reasonStr, ok := reason.(string); ok {
            return reasonStr
        }
    }
    
    // Default reasons based on auth method
    switch method {
    case "jwt":
        return "invalid_token"
    case "api_key":
        return "invalid_key"
    default:
        return "unknown"
    }
}

func (smw *SecurityMonitoringMiddleware) getRequiredPermission(c *gin.Context) string {
    if perm, exists := c.Get("required_permission"); exists {
        if permStr, ok := perm.(string); ok {
            return permStr
        }
    }
    return "unknown"
}

func (smw *SecurityMonitoringMiddleware) extractClientID(authHeader string) string {
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 {
        return ""
    }
    
    // Extract client ID from API key (implementation depends on key format)
    // This is a simplified example
    keyParts := strings.Split(parts[1], ".")
    if len(keyParts) >= 2 {
        return keyParts[0] // Assuming client ID is first part
    }
    
    return ""
}
```

### 3. Enhanced Authentication Middleware Integration

```go
// pkg/mw/mw.go - MODIFICATION to existing auth middleware
// ADD security monitoring to existing AuthMiddleware

func (amw *AuthMiddleware) WithAuth(c *gin.Context) {
    if !authRequired(c) {
        c.Next()
        return
    }

    start := time.Now()
    
    err := amw.authenticate(c)
    if err != nil {
        // NEW: Record authentication failure with reason
        c.Set("auth_failure_reason", categorizeAuthError(err))
        
        c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
        return
    }

    // NEW: Record successful authentication
    duration := time.Since(start).Seconds()
    method := getAuthMethodFromHeader(c.GetHeader("Authorization"))
    metrics.AuthenticationDuration.WithLabelValues(method).Observe(duration)
    metrics.AuthenticationAttempts.WithLabelValues(method, "success", "").Inc()

    c.Next()
}

func (amw *AuthMiddleware) validateAPIKey(apiKey string) error {
    start := time.Now()
    
    clientID, key, err := authutils.ExtractAPIKey(apiKey)
    if err != nil {
        // NEW: Record API key validation error
        metrics.APIKeyValidationErrors.WithLabelValues("invalid_format").Inc()
        return ErrInvalidAPIKey
    }

    rec, err := amw.store.APIKey.GetByClientID(amw.repo.DB(), clientID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // NEW: Record API key validation error
            metrics.APIKeyValidationErrors.WithLabelValues("key_not_found").Inc()
            return ErrInvalidAPIKey
        }
        // NEW: Record API key validation error
        metrics.APIKeyValidationErrors.WithLabelValues("database_error").Inc()
        return err
    }
    
    if rec.Status != model.ApikeyStatusValid {
        // NEW: Record API key validation error
        metrics.APIKeyValidationErrors.WithLabelValues("key_disabled").Inc()
        return ErrInvalidAPIKey
    }

    err = authutils.ValidateHashedKey(rec.SecretKey, key)
    if err != nil {
        // NEW: Record API key validation error
        metrics.APIKeyValidationErrors.WithLabelValues("key_mismatch").Inc()
        return err
    }
    
    // NEW: Record successful API key usage
    metrics.APIKeyUsage.WithLabelValues(clientID, "success").Inc()
    
    return nil
}

// NEW: Helper functions for security monitoring
func categorizeAuthError(err error) string {
    switch {
    case errors.Is(err, ErrUnexpectedAuthorizationHeader):
        return "invalid_header_format"
    case errors.Is(err, ErrAuthenticationTypeHeaderInvalid):
        return "invalid_auth_type"
    case errors.Is(err, ErrInvalidAPIKey):
        return "invalid_api_key"
    case strings.Contains(err.Error(), "token is expired"):
        return "token_expired"
    case strings.Contains(err.Error(), "signature is invalid"):
        return "invalid_signature"
    default:
        return "unknown_error"
    }
}

func getAuthMethodFromHeader(authHeader string) string {
    if authHeader == "" {
        return "none"
    }
    
    parts := strings.Split(authHeader, " ")
    if len(parts) < 1 {
        return "unknown"
    }
    
    switch strings.ToLower(parts[0]) {
    case "bearer":
        return "jwt"
    case "apikey":
        return "api_key"
    default:
        return "unknown"
    }
}
```

### 4. Enhanced Permission Middleware

```go
// pkg/mw/mw.go - MODIFICATION to existing PermMiddleware
// ADD security monitoring to existing permission checks

func (m *PermMiddleware) WithPerm(perm model.PermissionCode) func(c *gin.Context) {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Set required permission for monitoring
        c.Set("required_permission", perm.String())
        
        accessToken, err := authutils.GetTokenFromRequest(c)
        if err != nil {
            metrics.PermissionChecks.WithLabelValues(perm.String(), "error").Inc()
            c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
            return
        }
        
        tokenType := model.TokenTypeJWT
        if authutils.IsAPIKey(c) {
            tokenType = model.TokenTypeAPIKey
        }

        err = m.ensurePerm(m.store, m.repo.DB(), accessToken, perm.String(), tokenType.String())
        if err != nil {
            // NEW: Record permission denial
            duration := time.Since(start).Seconds()
            metrics.PermissionChecks.WithLabelValues(perm.String(), "denied").Inc()
            metrics.AuthorizationDuration.WithLabelValues(perm.String()).Observe(duration)
            
            c.AbortWithStatusJSON(401, map[string]string{"message": err.Error()})
            return
        }

        // NEW: Record successful permission check
        duration := time.Since(start).Seconds()
        metrics.PermissionChecks.WithLabelValues(perm.String(), "allowed").Inc()
        metrics.AuthorizationDuration.WithLabelValues(perm.String()).Observe(duration)

        c.Next()
    }
}
```

### 5. Threat Detection Engine

```go
// pkg/security/threat_detector.go - Threat detection logic
package security

import (
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/dwarvesf/fortress-api/pkg/middleware"
)

type ThreatDetector struct {
    config          *middleware.SecurityMonitoringConfig
    failureTracker  map[string]*AuthFailureTracker
    patternDetector *PatternDetector
    mu              sync.RWMutex
}

type AuthFailureTracker struct {
    Count      int
    FirstSeen  time.Time
    LastSeen   time.Time
    Methods    map[string]int
}

type SuspiciousPattern struct {
    Type       string
    Details    string
    IsCritical bool
    ClientIP   string
    Timestamp  time.Time
}

type PatternDetector struct {
    requestPatterns map[string]*RequestPattern
    mu              sync.RWMutex
}

type RequestPattern struct {
    Count       int
    LastSeen    time.Time
    Endpoints   map[string]int
    UserAgents  []string
}

func NewThreatDetector(config *middleware.SecurityMonitoringConfig) *ThreatDetector {
    td := &ThreatDetector{
        config:          config,
        failureTracker:  make(map[string]*AuthFailureTracker),
        patternDetector: &PatternDetector{
            requestPatterns: make(map[string]*RequestPattern),
        },
    }
    
    // Start cleanup routine
    go td.cleanupRoutine()
    
    return td
}

func (td *ThreatDetector) RecordAuthFailure(clientIP, method string) {
    td.mu.Lock()
    defer td.mu.Unlock()
    
    tracker, exists := td.failureTracker[clientIP]
    if !exists {
        tracker = &AuthFailureTracker{
            FirstSeen: time.Now(),
            Methods:   make(map[string]int),
        }
        td.failureTracker[clientIP] = tracker
    }
    
    tracker.Count++
    tracker.LastSeen = time.Now()
    tracker.Methods[method]++
    
    // Check if brute force threshold is exceeded
    if tracker.Count >= td.config.BruteForceThreshold {
        window := time.Since(tracker.FirstSeen)
        if window <= td.config.BruteForceWindow {
            // Brute force attack detected
            metrics.SuspiciousActivity.WithLabelValues(
                "brute_force", "high",
            ).Inc()
        }
    }
}

func (td *ThreatDetector) AnalyzeRequest(c *gin.Context, clientIP, userAgent string) []SuspiciousPattern {
    var patterns []SuspiciousPattern
    
    // Analyze request patterns
    if td.config.SuspiciousPatternEnabled {
        patterns = append(patterns, td.analyzeRequestPatterns(c, clientIP, userAgent)...)
    }
    
    return patterns
}

func (td *ThreatDetector) analyzeRequestPatterns(c *gin.Context, clientIP, userAgent string) []SuspiciousPattern {
    var patterns []SuspiciousPattern
    
    td.patternDetector.mu.Lock()
    defer td.patternDetector.mu.Unlock()
    
    pattern, exists := td.patternDetector.requestPatterns[clientIP]
    if !exists {
        pattern = &RequestPattern{
            LastSeen:   time.Now(),
            Endpoints:  make(map[string]int),
            UserAgents: []string{},
        }
        td.patternDetector.requestPatterns[clientIP] = pattern
    }
    
    pattern.Count++
    pattern.LastSeen = time.Now()
    pattern.Endpoints[c.Request.URL.Path]++
    
    // Check for rapid requests (potential bot activity)
    if pattern.Count > 100 && time.Since(pattern.LastSeen) < time.Minute {
        patterns = append(patterns, SuspiciousPattern{
            Type:       "rapid_requests",
            Details:    "High request frequency detected",
            IsCritical: false,
            ClientIP:   clientIP,
            Timestamp:  time.Now(),
        })
    }
    
    // Check for unusual user agent patterns
    if userAgent != "" {
        isNewUserAgent := true
        for _, ua := range pattern.UserAgents {
            if ua == userAgent {
                isNewUserAgent = false
                break
            }
        }
        
        if isNewUserAgent {
            pattern.UserAgents = append(pattern.UserAgents, userAgent)
            
            // If too many different user agents from same IP
            if len(pattern.UserAgents) > 10 {
                patterns = append(patterns, SuspiciousPattern{
                    Type:       "user_agent_variation",
                    Details:    "Multiple user agents from same IP",
                    IsCritical: false,
                    ClientIP:   clientIP,
                    Timestamp:  time.Now(),
                })
            }
        }
    }
    
    return patterns
}

func (td *ThreatDetector) cleanupRoutine() {
    ticker := time.NewTicker(time.Hour)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            td.cleanup()
        }
    }
}

func (td *ThreatDetector) cleanup() {
    td.mu.Lock()
    defer td.mu.Unlock()
    
    cutoff := time.Now().Add(-24 * time.Hour) // Keep 24 hours of data
    
    // Clean up auth failure tracker
    for ip, tracker := range td.failureTracker {
        if tracker.LastSeen.Before(cutoff) {
            delete(td.failureTracker, ip)
        }
    }
    
    // Clean up pattern detector
    td.patternDetector.mu.Lock()
    for ip, pattern := range td.patternDetector.requestPatterns {
        if pattern.LastSeen.Before(cutoff) {
            delete(td.patternDetector.requestPatterns, ip)
        }
    }
    td.patternDetector.mu.Unlock()
}
```

### 6. Configuration Integration

```go
// pkg/config/config.go - Security monitoring configuration
// ADDITION to existing MonitoringConfig

type MonitoringConfig struct {
    // ... existing fields ...
    
    Security SecurityMonitoringConfig `mapstructure:"security"`
}

type SecurityMonitoringConfig struct {
    Enabled                   bool          `mapstructure:"enabled" default:"true"`
    ThreatDetectionEnabled    bool          `mapstructure:"threat_detection_enabled" default:"true"`
    BruteForceThreshold       int           `mapstructure:"brute_force_threshold" default:"10"`
    BruteForceWindow          time.Duration `mapstructure:"brute_force_window" default:"5m"`
    SuspiciousPatternEnabled  bool          `mapstructure:"suspicious_pattern_enabled" default:"true"`
    LogSecurityEvents         bool          `mapstructure:"log_security_events" default:"true"`
    RateLimitMonitoring       bool          `mapstructure:"rate_limit_monitoring" default:"true"`
}
```

## Testing Strategy

### 1. Unit Tests

```go
// pkg/middleware/security_monitoring_test.go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/stretchr/testify/assert"
)

func TestAuthenticationMetrics(t *testing.T) {
    // Reset metrics for testing
    prometheus.DefaultRegisterer = prometheus.NewRegistry()
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    middleware := NewSecurityMonitoringMiddleware(nil, mockLogger)
    r.Use(middleware.Handler())
    
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Test successful JWT authentication
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer valid-jwt-token")
    r.ServeHTTP(w, req)
    
    // Verify metrics were recorded
    // Note: Would use prometheus testutil to verify actual values
    assert.Equal(t, 200, w.Code)
}

func TestAuthenticationFailureDetection(t *testing.T) {
    middleware := NewSecurityMonitoringMiddleware(nil, mockLogger)
    
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(middleware.Handler())
    
    r.GET("/test", func(c *gin.Context) {
        c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
    })
    
    // Test authentication failure
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    r.ServeHTTP(w, req)
    
    assert.Equal(t, 401, w.Code)
    // Verify failure metrics were recorded
}

func TestBruteForceDetection(t *testing.T) {
    config := &SecurityMonitoringConfig{
        BruteForceThreshold: 3,
        BruteForceWindow:    time.Minute,
        ThreatDetectionEnabled: true,
    }
    
    threatDetector := NewThreatDetector(config)
    
    clientIP := "192.168.1.100"
    
    // Record failures
    for i := 0; i < 5; i++ {
        threatDetector.RecordAuthFailure(clientIP, "jwt")
    }
    
    // Verify brute force detection triggered
    // Would check suspicious activity metrics
}

func TestSuspiciousPatternDetection(t *testing.T) {
    config := &SecurityMonitoringConfig{
        SuspiciousPatternEnabled: true,
    }
    
    threatDetector := NewThreatDetector(config)
    
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(httptest.NewRecorder())
    c.Request, _ = http.NewRequest("GET", "/api/test", nil)
    
    clientIP := "192.168.1.200"
    userAgent := "TestBot/1.0"
    
    patterns := threatDetector.AnalyzeRequest(c, clientIP, userAgent)
    
    // Should not detect patterns on first request
    assert.Empty(t, patterns)
    
    // Simulate rapid requests
    for i := 0; i < 150; i++ {
        patterns = threatDetector.AnalyzeRequest(c, clientIP, userAgent)
    }
    
    // Should detect rapid request pattern
    assert.NotEmpty(t, patterns)
    assert.Equal(t, "rapid_requests", patterns[0].Type)
}
```

### 2. Integration Tests

```go
// tests/integration/security_monitoring_test.go
package integration

import (
    "net/http"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/dwarvesf/fortress-api/pkg/testhelper"
)

func TestSecurityMetricsEndToEnd(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo *store.PostgresRepo) {
        router := setupTestRouterWithSecurity(repo, true)
        
        // Test various security scenarios
        scenarios := []struct {
            name           string
            method         string
            path           string
            headers        map[string]string
            expectedStatus int
            expectedMetric string
        }{
            {
                name:   "Valid JWT authentication",
                method: "GET",
                path:   "/api/v1/metadata/stacks",
                headers: map[string]string{
                    "Authorization": "Bearer " + generateValidJWT(),
                },
                expectedStatus: 200,
                expectedMetric: "fortress_auth_attempts_total{method=\"jwt\",result=\"success\"}",
            },
            {
                name:   "Invalid JWT authentication",
                method: "GET", 
                path:   "/api/v1/metadata/stacks",
                headers: map[string]string{
                    "Authorization": "Bearer invalid-token",
                },
                expectedStatus: 401,
                expectedMetric: "fortress_auth_attempts_total{method=\"jwt\",result=\"failure\"}",
            },
            {
                name:   "Valid API key authentication",
                method: "GET",
                path:   "/api/v1/metadata/stacks",
                headers: map[string]string{
                    "Authorization": "ApiKey " + generateValidAPIKey(),
                },
                expectedStatus: 200,
                expectedMetric: "fortress_auth_attempts_total{method=\"api_key\",result=\"success\"}",
            },
        }
        
        for _, scenario := range scenarios {
            t.Run(scenario.name, func(t *testing.T) {
                w := httptest.NewRecorder()
                req, _ := http.NewRequest(scenario.method, scenario.path, nil)
                
                for key, value := range scenario.headers {
                    req.Header.Set(key, value)
                }
                
                router.ServeHTTP(w, req)
                assert.Equal(t, scenario.expectedStatus, w.Code)
                
                // Verify metrics endpoint contains expected metric
                metricsReq, _ := http.NewRequest("GET", "/metrics", nil)
                metricsW := httptest.NewRecorder()
                router.ServeHTTP(metricsW, metricsReq)
                
                assert.Contains(t, metricsW.Body.String(), scenario.expectedMetric)
            })
        }
    })
}

func TestThreatDetectionIntegration(t *testing.T) {
    testhelper.TestWithTxDB(t, func(repo *store.PostgresRepo) {
        router := setupTestRouterWithSecurity(repo, true)
        
        clientIP := "192.168.1.100"
        
        // Simulate brute force attack
        for i := 0; i < 15; i++ {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("POST", "/api/v1/auth", nil)
            req.Header.Set("Authorization", "Bearer invalid-token")
            req.RemoteAddr = clientIP + ":12345"
            
            router.ServeHTTP(w, req)
            assert.Equal(t, 401, w.Code)
        }
        
        // Check that suspicious activity was detected
        metricsReq, _ := http.NewRequest("GET", "/metrics", nil)
        metricsW := httptest.NewRecorder()
        router.ServeHTTP(metricsW, metricsReq)
        
        assert.Contains(t, metricsW.Body.String(), 
            "fortress_security_suspicious_activity_total")
    })
}
```

### 3. Security Tests

```go
// tests/security/security_monitoring_test.go
package security

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
)

func TestPrivacyCompliance(t *testing.T) {
    // Ensure no sensitive data is logged in metrics
    middleware := NewSecurityMonitoringMiddleware(nil, mockLogger)
    
    // Test that sensitive information is not included in metrics
    // This would involve checking metric labels and values
}

func TestMetricCardinality(t *testing.T) {
    // Test that security metrics don't create unbounded cardinality
    // Especially important for client IP and user agent tracking
}

func TestSecurityEventIntegrity(t *testing.T) {
    // Ensure security events are accurately recorded
    // and timestamps are correct
}
```

## Deployment Strategy

### 1. Configuration Files

```yaml
# config/security-monitoring.yaml
monitoring:
  enabled: true
  security:
    enabled: true
    threat_detection_enabled: true
    brute_force_threshold: 10
    brute_force_window: 5m
    suspicious_pattern_enabled: true
    log_security_events: true
    rate_limit_monitoring: true
```

### 2. Alerting Rules

```yaml
# k8s/monitoring/security-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: fortress-security-alerts
spec:
  groups:
  - name: security
    rules:
    - alert: HighAuthenticationFailureRate
      expr: |
        rate(fortress_auth_attempts_total{result="failure"}[5m]) > 5
      for: 2m
      labels:
        severity: warning
        team: security
      annotations:
        summary: "High authentication failure rate detected"
        
    - alert: BruteForceAttackDetected
      expr: |
        rate(fortress_security_suspicious_activity_total{event_type="brute_force"}[5m]) > 0
      for: 1m
      labels:
        severity: critical
        team: security
      annotations:
        summary: "Brute force attack detected"
        
    - alert: HighPermissionViolationRate
      expr: |
        rate(fortress_auth_permission_checks_total{result="denied"}[5m]) > 2
      for: 5m
      labels:
        severity: warning
        team: security
      annotations:
        summary: "High rate of permission violations"
```

### 3. Staged Rollout Plan

**Phase 1: Basic Security Metrics (Week 1)**
- Deploy authentication and authorization metrics
- Monitor performance impact
- Validate metric accuracy

**Phase 2: Threat Detection (Week 2)**
- Enable threat detection engine
- Configure brute force detection
- Set up basic security alerting

**Phase 3: Advanced Monitoring (Week 3)**
- Enable suspicious pattern detection
- Fine-tune threat detection parameters
- Create security-focused dashboards

**Phase 4: Full Security Operations (Week 4)**
- Enable all security logging
- Train security team on new metrics
- Establish security incident response procedures

## Success Criteria

### Security Metrics
- **Threat Detection Accuracy**: >95% accuracy in identifying actual security threats
- **False Positive Rate**: <5% false positive rate for brute force detection
- **Detection Latency**: Security events detected within 30 seconds

### Performance Metrics
- **Authentication Overhead**: <1% performance impact on auth operations
- **Memory Usage**: <25MB additional memory for security monitoring
- **Monitoring Accuracy**: 99.9% accurate event recording

### Operational Metrics
- **Security Visibility**: 100% coverage of authentication and authorization events
- **Incident Response**: Security incidents identified proactively before manual detection
- **Compliance**: Security event logging meets audit requirements

## Risk Mitigation

### Performance Impact Risk
- **Mitigation**: Lightweight threat detection algorithms and efficient data structures
- **Monitoring**: Continuous performance benchmarking

### Privacy Risk  
- **Mitigation**: Careful data handling, no sensitive information in metrics
- **Monitoring**: Regular privacy compliance audits

### False Positive Risk
- **Mitigation**: Tunable thresholds and pattern recognition improvements
- **Monitoring**: Alert accuracy tracking and feedback loops

---

**Assignee**: Backend Engineering Team + Security Team  
**Reviewer**: Security Team Lead, Senior Backend Engineer  
**Implementation Timeline**: 2 weeks  
**Testing Timeline**: 1 week  
**Deployment Timeline**: 2 weeks (staged rollout)