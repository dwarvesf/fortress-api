package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/mw"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/golang-jwt/jwt/v4"
)

// BenchmarkLogger for performance testing (minimal overhead)
type BenchmarkLogger struct{}

func (b *BenchmarkLogger) Fields(data logger.Fields) logger.Logger       { return b }
func (b *BenchmarkLogger) Field(key, value string) logger.Logger         { return b }
func (b *BenchmarkLogger) AddField(key string, value any) logger.Logger  { return b }
func (b *BenchmarkLogger) Debug(msg string)                             {}
func (b *BenchmarkLogger) Debugf(msg string, args ...interface{})        {}
func (b *BenchmarkLogger) Info(msg string)                              {}
func (b *BenchmarkLogger) Infof(msg string, args ...interface{})         {}
func (b *BenchmarkLogger) Warn(msg string)                              {}
func (b *BenchmarkLogger) Warnf(msg string, args ...interface{})         {}
func (b *BenchmarkLogger) Error(err error, msg string)                   {}
func (b *BenchmarkLogger) Errorf(err error, msg string, args ...interface{}) {}
func (b *BenchmarkLogger) Fatal(err error, msg string)                   {}
func (b *BenchmarkLogger) Fatalf(err error, msg string, args ...interface{}) {}

// Test helper to generate valid JWT tokens for benchmarking
func generateBenchmarkToken(cfg *config.Config) string {
	claims := &model.AuthenticationInfo{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
		UserID: "benchmark-user-123",
		Email:  "benchmark@example.com",
		Avatar: "https://example.com/avatar.png",
	}

	token, err := authutils.GenerateJWTToken(claims, claims.ExpiresAt, cfg.JWTSecretKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate benchmark token: %v", err))
	}

	return token
}

// Baseline benchmark - HTTP request without any security monitoring
func BenchmarkBaselineHTTPRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "BenchmarkClient/1.0")
		r.ServeHTTP(w, req)
	}
}

// Benchmark HTTP request with security monitoring enabled (all features)
func BenchmarkSecurityMonitoringEnabled(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Full security monitoring configuration
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       10,
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", fmt.Sprintf("BenchmarkClient/1.0-%d", i%10))
		req.Header.Set("Authorization", "Bearer test-token-"+fmt.Sprintf("%d", i%5))
		r.ServeHTTP(w, req)
	}
}

// Benchmark HTTP request with security monitoring disabled
func BenchmarkSecurityMonitoringDisabled(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Security monitoring disabled
	securityConfig := &SecurityMonitoringConfig{
		Enabled: false,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "BenchmarkClient/1.0")
		req.Header.Set("Authorization", "Bearer test-token")
		r.ServeHTTP(w, req)
	}
}

// Benchmark authentication flow with security monitoring
func BenchmarkAuthenticationWithSecurityMonitoring(b *testing.B) {
	cfg := config.LoadTestConfig()
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup security monitoring
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       100, // High threshold to avoid triggering during bench
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         false, // Disable logging for performance
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	// Setup authentication middleware
	storeMock := store.New()
	authMiddleware := mw.NewAuthMiddleware(&cfg, storeMock, nil)

	// Generate valid token for benchmarking
	validToken := generateBenchmarkToken(&cfg)

	r.GET("/protected", authMiddleware.WithAuth, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "authenticated"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		req.Header.Set("User-Agent", fmt.Sprintf("BenchClient/%d", i%20))
		r.ServeHTTP(w, req)
	}
}

// Benchmark authentication failures with security monitoring
func BenchmarkAuthenticationFailuresWithSecurityMonitoring(b *testing.B) {
	cfg := config.LoadTestConfig()
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup security monitoring with lower thresholds for failure testing
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       1000, // High threshold to avoid excessive brute force alerts
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         false, // Disable logging for performance
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	// Setup authentication middleware
	storeMock := store.New()
	authMiddleware := mw.NewAuthMiddleware(&cfg, storeMock, nil)

	r.GET("/protected", authMiddleware.WithAuth, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "authenticated"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token-"+fmt.Sprintf("%d", i%100))
		req.Header.Set("User-Agent", fmt.Sprintf("TestClient/%d", i%50))
		r.ServeHTTP(w, req)
	}
}

// Benchmark threat detection specifically
func BenchmarkThreatDetection(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Focus on threat detection features
	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       50,
		BruteForceWindow:          2 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         false, // Disable logging for pure performance measurement
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/api/endpoint", func(c *gin.Context) {
		// Simulate some auth failure to trigger threat detection
		if c.GetHeader("Authorization") == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(200, gin.H{"data": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/endpoint", nil)
		
		// Alternate between authorized and unauthorized requests
		if i%3 == 0 {
			// Unauthorized request to trigger threat detection
			req.Header.Set("User-Agent", fmt.Sprintf("SuspiciousBot/%d", i%10))
		} else {
			// Authorized request
			req.Header.Set("Authorization", "Bearer valid-token")
			req.Header.Set("User-Agent", fmt.Sprintf("LegitClient/%d", i%5))
		}
		
		r.ServeHTTP(w, req)
	}
}

// Benchmark permission monitoring
func BenchmarkPermissionMonitoring(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	securityConfig := &SecurityMonitoringConfig{
		Enabled:           true,
		LogSecurityEvents: false, // Disable logging for performance
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/admin/action", func(c *gin.Context) {
		// Simulate permission check
		c.Set("required_permission", "admin.write")
		
		// Alternate between success and permission denied
		if c.GetHeader("X-Admin") == "true" {
			c.JSON(200, gin.H{"result": "success"})
		} else {
			c.AbortWithStatusJSON(403, gin.H{"error": "permission denied"})
		}
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/admin/action", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		
		// Alternate between admin and regular user
		if i%4 == 0 {
			req.Header.Set("X-Admin", "true")
		}
		
		r.ServeHTTP(w, req)
	}
}

// Benchmark concurrent security monitoring (simulates real-world load)
func BenchmarkSecurityMonitoringConcurrent(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       100,
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         false,
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/concurrent-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/concurrent-test", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer token-%d", i%10))
			req.Header.Set("User-Agent", fmt.Sprintf("Client/%d", i%20))
			r.ServeHTTP(w, req)
			i++
		}
	})
}

// Benchmark memory allocations for security monitoring
func BenchmarkSecurityMonitoringMemory(b *testing.B) {
	benchLogger := &BenchmarkLogger{}
	gin.SetMode(gin.TestMode)
	r := gin.New()

	securityConfig := &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       50,
		BruteForceWindow:          2 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         false,
		RateLimitMonitoring:       true,
	}

	securityMiddleware := NewSecurityMonitoringMiddleware(securityConfig, benchLogger)
	r.Use(securityMiddleware.Handler())

	r.GET("/memory-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"data": "test"})
	})

	b.ResetTimer()
	b.ReportAllocs()
	b.ReportMetric(0, "ns/op")

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/memory-test", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer token-%d", i))
		req.Header.Set("User-Agent", fmt.Sprintf("TestAgent/%d.0", i%5))
		r.ServeHTTP(w, req)
	}
}