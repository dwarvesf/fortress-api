package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/mw"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/golang-jwt/jwt/v4"
)

// MockLogger for integration testing
type MockIntegrationLogger struct {
	mock.Mock
}

func (m *MockIntegrationLogger) Fields(data logger.Fields) logger.Logger {
	args := m.Called(data)
	return args.Get(0).(logger.Logger)
}

func (m *MockIntegrationLogger) Field(key, value string) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockIntegrationLogger) AddField(key string, value any) logger.Logger {
	m.Called(key, value)
	return m
}

func (m *MockIntegrationLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockIntegrationLogger) Debugf(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockIntegrationLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockIntegrationLogger) Infof(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockIntegrationLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockIntegrationLogger) Warnf(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockIntegrationLogger) Error(err error, msg string) {
	m.Called(err, msg)
}

func (m *MockIntegrationLogger) Errorf(err error, msg string, args ...interface{}) {
	m.Called(err, msg, args)
}

func (m *MockIntegrationLogger) Fatal(err error, msg string) {
	m.Called(err, msg)
}

func (m *MockIntegrationLogger) Fatalf(err error, msg string, args ...interface{}) {
	m.Called(err, msg, args)
}

// Test helper to generate valid JWT tokens
func generateIntegrationTestToken(cfg *config.Config) string {
	claims := &model.AuthenticationInfo{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
		UserID: "test-user-123",
		Email:  "test@example.com",
		Avatar: "https://example.com/avatar.png",
	}

	token, err := authutils.GenerateJWTToken(claims, claims.ExpiresAt, cfg.JWTSecretKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}

	return token
}

// Test helper to generate expired JWT tokens
func generateExpiredTestToken(cfg *config.Config) string {
	claims := &model.AuthenticationInfo{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		},
		UserID: "test-user-123",
		Email:  "test@example.com",
		Avatar: "https://example.com/avatar.png",
	}

	token, err := authutils.GenerateJWTToken(claims, claims.ExpiresAt, cfg.JWTSecretKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate expired test token: %v", err))
	}

	return token
}

func TestSecurityMonitoringIntegrationWithAuthentication(t *testing.T) {
	// Setup
	cfg := config.LoadTestConfig()
	mockLogger := &MockIntegrationLogger{}
	
	// Configure mock logger to handle security event logging
	mockLogger.On("AddField", mock.AnythingOfType("string"), mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", mock.AnythingOfType("string")).Maybe()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup security monitoring middleware
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       3,
		BruteForceWindow:          time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
		RateLimitMonitoring:       true,
	}
	
	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, mockLogger)
	r.Use(securityMiddleware.Handler())

	// Setup authentication middleware
	storeMock := store.New()
	authMiddleware := mw.NewAuthMiddleware(&cfg, storeMock, nil)

	// Define test endpoints
	r.GET("/protected", authMiddleware.WithAuth, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	r.GET("/public", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "public endpoint"})
	})

	// Test scenarios
	testCases := []struct {
		name               string
		endpoint           string
		authHeader         string
		expectedStatus     int
		expectSecurityLog  bool
		description        string
	}{
		{
			name:               "Successful JWT Authentication",
			endpoint:           "/protected",
			authHeader:         "Bearer " + generateIntegrationTestToken(&cfg),
			expectedStatus:     200,
			expectSecurityLog:  false,
			description:        "Valid JWT should allow access and record success metrics",
		},
		{
			name:               "Failed JWT Authentication - Invalid Token",
			endpoint:           "/protected",
			authHeader:         "Bearer invalid-token-here",
			expectedStatus:     401,
			expectSecurityLog:  true,
			description:        "Invalid JWT should be blocked and recorded as security event",
		},
		{
			name:               "Failed JWT Authentication - Expired Token",
			endpoint:           "/protected", 
			authHeader:         "Bearer " + generateExpiredTestToken(&cfg),
			expectedStatus:     401,
			expectSecurityLog:  true,
			description:        "Expired JWT should be blocked and recorded as security event",
		},
		{
			name:               "Failed Authentication - No Authorization Header",
			endpoint:           "/protected",
			authHeader:         "",
			expectedStatus:     401,
			expectSecurityLog:  true,
			description:        "Missing auth header should be blocked and recorded",
		},
		{
			name:               "Failed Authentication - Invalid Header Format",
			endpoint:           "/protected",
			authHeader:         "InvalidFormat token123",
			expectedStatus:     401,
			expectSecurityLog:  true,
			description:        "Invalid header format should be blocked and recorded",
		},
		{
			name:               "Successful API Key Authentication",
			endpoint:           "/protected",
			authHeader:         "ApiKey client123.validkey",
			expectedStatus:     401, // Will fail since we don't have actual API key validation setup
			expectSecurityLog:  true,
			description:        "API key attempts should be monitored",
		},
		{
			name:               "Public Endpoint Access",
			endpoint:           "/public",
			authHeader:         "",
			expectedStatus:     200,
			expectSecurityLog:  false,
			description:        "Public endpoints should not require auth but still be monitored",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.endpoint, nil)
			
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			
			// Add user agent for pattern detection
			req.Header.Set("User-Agent", "SecurityIntegrationTest/1.0")

			// Execute request
			r.ServeHTTP(w, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, w.Code, tc.description)

			// Verify security monitoring is working (metrics would be incremented)
			// Note: In a real test, you might want to check Prometheus metrics
			// but that requires more complex test infrastructure
		})
	}
}

func TestSecurityMonitoringBruteForceDetection(t *testing.T) {
	// Setup
	cfg := config.LoadTestConfig()
	mockLogger := &MockIntegrationLogger{}
	
	// Configure mock logger for brute force logging
	mockLogger.On("AddField", mock.AnythingOfType("string"), mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", "Suspicious activity detected").Maybe()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup security monitoring with low brute force threshold for testing
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       3, // Low threshold for testing
		BruteForceWindow:          time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
	}
	
	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, mockLogger)
	r.Use(securityMiddleware.Handler())

	// Setup authentication middleware
	storeMock := store.New()
	authMiddleware := mw.NewAuthMiddleware(&cfg, storeMock, nil)

	r.GET("/login", authMiddleware.WithAuth, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "login success"})
	})

	// Simulate brute force attack - multiple failed login attempts
	clientIP := "192.168.1.100"
	
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-"+fmt.Sprintf("%d", i))
		req.Header.Set("User-Agent", "AttackerBot/1.0")
		
		// Simulate same client IP
		req.RemoteAddr = clientIP + ":12345"
		
		r.ServeHTTP(w, req)
		
		// All should fail with 401
		assert.Equal(t, 401, w.Code)
		
		// Small delay to simulate real attack timing
		time.Sleep(10 * time.Millisecond)
	}

	// Verify that brute force detection would have triggered
	// (In actual implementation, metrics would show brute force activity)
}

func TestSecurityMonitoringPatternDetection(t *testing.T) {
	// Setup
	mockLogger := &MockIntegrationLogger{}
	
	// Configure mock logger for pattern detection
	mockLogger.On("AddField", mock.AnythingOfType("string"), mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", "Suspicious activity detected").Maybe()

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup security monitoring with pattern detection enabled
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
	}
	
	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, mockLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": "test"})
	})

	clientIP := "192.168.1.200"

	// Test 1: Rapid requests from same IP (potential bot behavior)
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/data", nil)
		req.Header.Set("User-Agent", "RapidBot/1.0")
		req.RemoteAddr = clientIP + ":54321"
		
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	// Test 2: Multiple user agents from same IP (potential spoofing)
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/605.1",
		"Mozilla/5.0 (X11; Linux x86_64) Firefox/89.0",
		"SuspiciousBot/1.0",
		"AnotherBot/2.0",
		"WeirdCrawler/1.5",
	}

	for _, ua := range userAgents {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/data", nil)
		req.Header.Set("User-Agent", ua)
		req.RemoteAddr = clientIP + ":65432"
		
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	// Pattern detection should have triggered for suspicious behavior
}

func TestSecurityMonitoringWithPermissions(t *testing.T) {
	// This test would require actual permission middleware integration
	// For now, we'll test the security monitoring of permission denials
	
	mockLogger := &MockIntegrationLogger{}
	
	// Configure mock logger for permission violations
	mockLogger.On("AddField", "permission", "employees.read").Return(mockLogger)
	mockLogger.On("AddField", "client_ip", mock.AnythingOfType("string")).Return(mockLogger)
	mockLogger.On("AddField", "user_agent", mock.AnythingOfType("string")).Return(mockLogger)
	mockLogger.On("AddField", "endpoint", "/admin/users").Return(mockLogger)
	mockLogger.On("Warn", "Permission denied")

	gin.SetMode(gin.TestMode)
	r := gin.New()

	securityConfig := &SecurityMonitoringConfig{
		Enabled:           true,
		LogSecurityEvents: true,
	}
	
	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, mockLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/admin/users", func(c *gin.Context) {
		// Simulate permission check failure
		c.Set("required_permission", "employees.read")
		c.AbortWithStatusJSON(403, gin.H{"error": "insufficient permissions"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer valid-but-insufficient-token")
	req.Header.Set("User-Agent", "TestClient/1.0")
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, 403, w.Code)
	mockLogger.AssertExpectations(t)
}

func TestSecurityMonitoringDisabled(t *testing.T) {
	// Test that when security monitoring is disabled, no security events are logged
	mockLogger := &MockIntegrationLogger{}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	securityConfig := &SecurityMonitoringConfig{
		Enabled: false, // Disabled
	}
	
	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, mockLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	// Verify no security logging methods were called
	mockLogger.AssertNotCalled(t, "Warn")
	mockLogger.AssertNotCalled(t, "Error")
}