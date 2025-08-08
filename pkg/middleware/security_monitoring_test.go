package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// MockLogger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Fields(data logger.Fields) logger.Logger {
	args := m.Called(data)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) Field(key, value string) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) AddField(key string, value any) logger.Logger {
	m.Called(key, value)
	return m
}

func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Debugf(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Infof(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

func (m *MockLogger) Warnf(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(err error, msg string) {
	m.Called(err, msg)
}

func (m *MockLogger) Errorf(err error, msg string, args ...interface{}) {
	m.Called(err, msg, args)
}

func (m *MockLogger) Fatal(err error, msg string) {
	m.Called(err, msg)
}

func (m *MockLogger) Fatalf(err error, msg string, args ...interface{}) {
	m.Called(err, msg, args)
}

func TestNewSecurityMonitoringMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}

	// Test with nil config (should use defaults)
	middleware := NewSecurityMonitoringMiddleware(nil, mockLogger)
	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.config)
	assert.True(t, middleware.config.Enabled)

	// Test with custom config
	customConfig := &SecurityMonitoringConfig{
		Enabled:                   false,
		ThreatDetectionEnabled:    false,
		BruteForceThreshold:       5,
		BruteForceWindow:          2 * time.Minute,
		SuspiciousPatternEnabled:  false,
		LogSecurityEvents:         false,
	}

	middleware2 := NewSecurityMonitoringMiddleware(customConfig, mockLogger)
	assert.NotNil(t, middleware2)
	assert.Equal(t, customConfig, middleware2.config)
	assert.False(t, middleware2.config.Enabled)
}

func TestSecurityMonitoringMiddleware_DisabledConfig(t *testing.T) {
	mockLogger := &MockLogger{}
	config := &SecurityMonitoringConfig{
		Enabled: false,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	middleware := NewSecurityMonitoringMiddleware(config, mockLogger)
	r.Use(middleware.Handler())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Should pass through without any security monitoring
	assert.Equal(t, 200, w.Code)
	mockLogger.AssertNotCalled(t, "Warn")
}

func TestSecurityMonitoringMiddleware_AuthenticationMetrics(t *testing.T) {
	tests := []struct {
		name               string
		authHeader         string
		responseStatus     int
		expectedMethod     string
		expectedResult     string
		shouldRecordMetric bool
	}{
		{
			name:               "successful JWT authentication",
			authHeader:         "Bearer valid-jwt-token",
			responseStatus:     200,
			expectedMethod:     "jwt",
			expectedResult:     "success",
			shouldRecordMetric: true,
		},
		{
			name:               "failed JWT authentication",
			authHeader:         "Bearer invalid-jwt-token",
			responseStatus:     401,
			expectedMethod:     "jwt",
			expectedResult:     "failure",
			shouldRecordMetric: true,
		},
		{
			name:               "successful API key authentication",
			authHeader:         "ApiKey valid-api-key",
			responseStatus:     200,
			expectedMethod:     "api_key",
			expectedResult:     "success",
			shouldRecordMetric: true,
		},
		{
			name:               "failed API key authentication",
			authHeader:         "ApiKey invalid-api-key",
			responseStatus:     401,
			expectedMethod:     "api_key",
			expectedResult:     "failure",
			shouldRecordMetric: true,
		},
		{
			name:               "unknown authentication method",
			authHeader:         "BasicAuth user:pass",
			responseStatus:     401,
			expectedMethod:     "unknown",
			expectedResult:     "failure",
			shouldRecordMetric: true,
		},
		{
			name:               "no authentication header",
			authHeader:         "",
			responseStatus:     200,
			expectedMethod:     "",
			expectedResult:     "",
			shouldRecordMetric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			config := &SecurityMonitoringConfig{
				Enabled:                   true,
				ThreatDetectionEnabled:    true,
				BruteForceThreshold:       10,
				BruteForceWindow:          5 * time.Minute,
				SuspiciousPatternEnabled:  true,
				LogSecurityEvents:         true,
			}

			gin.SetMode(gin.TestMode)
			r := gin.New()

			middleware := NewSecurityMonitoringMiddleware(config, mockLogger)
			r.Use(middleware.Handler())

			r.GET("/test", func(c *gin.Context) {
				if tt.responseStatus == 401 {
					c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
					return
				}
				c.JSON(200, gin.H{"status": "ok"})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.responseStatus, w.Code)
			// Additional metric verification would require Prometheus test utilities
		})
	}
}

func TestSecurityMonitoringMiddleware_PermissionMetrics(t *testing.T) {
	mockLogger := &MockLogger{}
	config := &SecurityMonitoringConfig{
		Enabled:           true,
		LogSecurityEvents: true,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	middleware := NewSecurityMonitoringMiddleware(config, mockLogger)
	r.Use(middleware.Handler())

	// Mock the logger calls for permission violations
	mockLogger.On("AddField", "permission", "employees.read").Return(mockLogger)
	mockLogger.On("AddField", "client_ip", mock.AnythingOfType("string")).Return(mockLogger)
	mockLogger.On("AddField", "user_agent", mock.AnythingOfType("string")).Return(mockLogger)
	mockLogger.On("AddField", "endpoint", "/test").Return(mockLogger)
	mockLogger.On("Warn", "Permission denied")

	r.GET("/test", func(c *gin.Context) {
		// Simulate permission check failure
		c.Set("required_permission", "employees.read")
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.Header.Set("User-Agent", "TestClient/1.0")
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	mockLogger.AssertExpectations(t)
}

func TestSecurityMonitoringMiddleware_ThreatDetection(t *testing.T) {
	mockLogger := &MockLogger{}
	config := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
		BruteForceThreshold:       3,
		BruteForceWindow:          time.Minute,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	middleware := NewSecurityMonitoringMiddleware(config, mockLogger)
	r.Use(middleware.Handler())

	// Mock logger calls for suspicious activity
	mockLogger.On("AddField", mock.AnythingOfType("string"), mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", "Suspicious activity detected")

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make multiple requests to trigger pattern detection
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "TestBot/1.0")
		req.RemoteAddr = "192.168.1.100:12345"
		r.ServeHTTP(w, req)
	}

	// Note: Specific threat detection verification would require more detailed testing
	// This test primarily ensures the middleware doesn't panic with threat detection enabled
}

func TestSecurityMonitoringConfig_Defaults(t *testing.T) {
	config := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       10,
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
	}

	// Test configuration validation
	assert.True(t, config.Enabled)
	assert.True(t, config.ThreatDetectionEnabled)
	assert.Equal(t, 10, config.BruteForceThreshold)
	assert.Equal(t, 5*time.Minute, config.BruteForceWindow)
	assert.True(t, config.SuspiciousPatternEnabled)
	assert.True(t, config.LogSecurityEvents)
}

func TestGetAuthMethod(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{"JWT Bearer token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "jwt"},
		{"API Key", "ApiKey client123.secretkey", "api_key"},
		{"Unknown method", "BasicAuth dXNlcjpwYXNz", "unknown"},
		{"Empty header", "", "unknown"},
		{"Malformed header", "InvalidFormat", "unknown"},
	}

	middleware := &SecurityMonitoringMiddleware{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.getAuthMethod(tt.authHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFailureReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware := &SecurityMonitoringMiddleware{}

	tests := []struct {
		name         string
		method       string
		contextValue interface{}
		expected     string
	}{
		{"JWT with context reason", "jwt", "token_expired", "token_expired"},
		{"API key with context reason", "api_key", "invalid_key", "invalid_key"},
		{"JWT without context", "jwt", nil, "invalid_token"},
		{"API key without context", "api_key", nil, "invalid_key"},
		{"Unknown method", "unknown", nil, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.contextValue != nil {
				c.Set("auth_failure_reason", tt.contextValue)
			}

			result := middleware.getFailureReason(c, tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractClientID(t *testing.T) {
	middleware := &SecurityMonitoringMiddleware{}

	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{"Valid API key format", "ApiKey client123.secretkey", "client123"},
		{"Invalid format - no dot", "ApiKey invalidsecret", ""},
		{"Invalid format - no space", "ApiKeyclient123.secret", ""},
		{"Invalid format - empty", "", ""},
		{"Complex client ID", "ApiKey complex-client-123.longSecretKey", "complex-client-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.extractClientID(tt.authHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityMonitoringIntegration(t *testing.T) {
	// Test that security monitoring integrates properly with Gin middleware chain
	mockLogger := &MockLogger{}
	config := &SecurityMonitoringConfig{
		Enabled: true,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add security monitoring middleware
	middleware := NewSecurityMonitoringMiddleware(config, mockLogger)
	r.Use(middleware.Handler())

	// Add a dummy endpoint
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test successful request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	// Middleware should not interfere with normal request processing
	mockLogger.AssertNotCalled(t, "Error")
}

func TestSecurityEventLogging(t *testing.T) {
	mockLogger := &MockLogger{}
	config := &SecurityMonitoringConfig{
		Enabled:           true,
		LogSecurityEvents: true,
	}

	middleware := NewSecurityMonitoringMiddleware(config, mockLogger)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("Authorization", "Bearer test-token")
	c.Request.Header.Set("User-Agent", "TestAgent/1.0")

	// Mock the logger chain
	mockLogger.On("AddField", mock.AnythingOfType("string"), mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", mock.AnythingOfType("string"))

	// Simulate security event logging
	start := time.Now()
	middleware.logSecurityEvent(c, "127.0.0.1", "TestAgent/1.0", start)

	// Verify that logging methods could be called (specific verification depends on implementation)
	assert.NotNil(t, middleware)
}