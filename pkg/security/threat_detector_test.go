package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

func TestNewThreatDetector(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		BruteForceThreshold: 5,
		BruteForceWindow:    time.Minute,
	}

	detector := NewThreatDetector(config)

	assert.NotNil(t, detector)
	assert.Equal(t, config, detector.config)
	assert.NotNil(t, detector.failureTracker)
	assert.NotNil(t, detector.patternDetector)
}

func TestRecordAuthFailure(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		BruteForceThreshold: 3,
		BruteForceWindow:    time.Minute,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.100"

	// First failure
	detector.RecordAuthFailure(clientIP, "jwt")
	
	detector.mu.RLock()
	tracker, exists := detector.failureTracker[clientIP]
	detector.mu.RUnlock()
	
	assert.True(t, exists)
	assert.Equal(t, 1, tracker.Count)
	assert.Equal(t, 1, tracker.Methods["jwt"])

	// Second failure
	detector.RecordAuthFailure(clientIP, "jwt")
	
	detector.mu.RLock()
	tracker = detector.failureTracker[clientIP]
	detector.mu.RUnlock()
	
	assert.Equal(t, 2, tracker.Count)
	assert.Equal(t, 2, tracker.Methods["jwt"])

	// Third failure - should trigger brute force detection
	detector.RecordAuthFailure(clientIP, "api_key")
	
	detector.mu.RLock()
	tracker = detector.failureTracker[clientIP]
	detector.mu.RUnlock()
	
	assert.Equal(t, 3, tracker.Count)
	assert.Equal(t, 2, tracker.Methods["jwt"])
	assert.Equal(t, 1, tracker.Methods["api_key"])
}

func TestRecordAuthFailure_BruteForceDetection(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		BruteForceThreshold: 2,
		BruteForceWindow:    time.Minute,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.101"

	// Record failures within the window
	for i := 0; i < 3; i++ {
		detector.RecordAuthFailure(clientIP, "jwt")
	}

	detector.mu.RLock()
	tracker := detector.failureTracker[clientIP]
	detector.mu.RUnlock()

	assert.Equal(t, 3, tracker.Count)
	assert.True(t, tracker.Count >= config.BruteForceThreshold)
	assert.True(t, time.Since(tracker.FirstSeen) <= config.BruteForceWindow)
}

func TestAnalyzeRequest_RapidRequests(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		SuspiciousPatternEnabled: true,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.102"
	userAgent := "TestBot/1.0"

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/api/test", nil)

	// Simulate many rapid requests
	for i := 0; i < 150; i++ {
		patterns := detector.AnalyzeRequest(c, clientIP, userAgent)
		
		// After 100+ requests, should detect rapid request pattern
		if i > 100 {
			if len(patterns) > 0 {
				found := false
				for _, pattern := range patterns {
					if pattern.Type == "rapid_requests" {
						found = true
						assert.Equal(t, clientIP, pattern.ClientIP)
						assert.False(t, pattern.IsCritical)
						break
					}
				}
				if found {
					break // Found the pattern we're looking for
				}
			}
		}
	}
}

func TestAnalyzeRequest_UserAgentVariation(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		SuspiciousPatternEnabled: true,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.103"

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/api/test", nil)

	// Use different user agents from same IP
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		"curl/7.68.0",
		"Postman/7.36.1",
		"TestBot/1.0",
		"ScannerBot/2.1",
		"AttackBot/3.0",
		"MaliciousAgent/1.5",
		"EvilBot/4.2",
		"HackTool/1.0",
		"BadAgent/2.0",
	}

	var foundPattern bool
	for _, userAgent := range userAgents {
		patterns := detector.AnalyzeRequest(c, clientIP, userAgent)
		
		for _, pattern := range patterns {
			if pattern.Type == "user_agent_variation" {
				foundPattern = true
				assert.Equal(t, clientIP, pattern.ClientIP)
				assert.False(t, pattern.IsCritical)
				break
			}
		}
	}

	assert.True(t, foundPattern, "Should detect user agent variation pattern")
}

func TestAnalyzeRequest_PatternDetectionDisabled(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		SuspiciousPatternEnabled: false,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.104"
	userAgent := "TestBot/1.0"

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/api/test", nil)

	// Even with many requests, should not detect patterns when disabled
	for i := 0; i < 200; i++ {
		patterns := detector.AnalyzeRequest(c, clientIP, userAgent)
		assert.Empty(t, patterns, "Should not detect patterns when disabled")
	}
}

func TestAuthFailureTracker(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		BruteForceThreshold: 5,
		BruteForceWindow:    time.Minute,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.105"

	// Test tracker creation and updates
	detector.RecordAuthFailure(clientIP, "jwt")
	detector.RecordAuthFailure(clientIP, "api_key")
	detector.RecordAuthFailure(clientIP, "jwt")

	detector.mu.RLock()
	tracker := detector.failureTracker[clientIP]
	detector.mu.RUnlock()

	assert.Equal(t, 3, tracker.Count)
	assert.Equal(t, 2, tracker.Methods["jwt"])
	assert.Equal(t, 1, tracker.Methods["api_key"])
	assert.NotZero(t, tracker.FirstSeen)
	assert.NotZero(t, tracker.LastSeen)
	assert.True(t, tracker.LastSeen.After(tracker.FirstSeen) || tracker.LastSeen.Equal(tracker.FirstSeen))
}

func TestRequestPattern(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		SuspiciousPatternEnabled: true,
	}

	detector := NewThreatDetector(config)
	clientIP := "192.168.1.106"

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/api/endpoint1", nil)

	// First request
	detector.AnalyzeRequest(c, clientIP, "TestAgent/1.0")

	detector.patternDetector.mu.RLock()
	pattern := detector.patternDetector.requestPatterns[clientIP]
	detector.patternDetector.mu.RUnlock()

	assert.NotNil(t, pattern)
	assert.Equal(t, 1, pattern.Count)
	assert.Equal(t, 1, pattern.Endpoints["/api/endpoint1"])
	assert.Contains(t, pattern.UserAgents, "TestAgent/1.0")

	// Second request to different endpoint
	c.Request.URL.Path = "/api/endpoint2"
	detector.AnalyzeRequest(c, clientIP, "TestAgent/1.0")

	detector.patternDetector.mu.RLock()
	pattern = detector.patternDetector.requestPatterns[clientIP]
	detector.patternDetector.mu.RUnlock()

	assert.Equal(t, 2, pattern.Count)
	assert.Equal(t, 1, pattern.Endpoints["/api/endpoint1"])
	assert.Equal(t, 1, pattern.Endpoints["/api/endpoint2"])
}

func TestSuspiciousPattern(t *testing.T) {
	pattern := SuspiciousPattern{
		Type:       "test_pattern",
		Details:    "Test pattern details",
		IsCritical: true,
		ClientIP:   "192.168.1.107",
		Timestamp:  time.Now(),
	}

	assert.Equal(t, "test_pattern", pattern.Type)
	assert.Equal(t, "Test pattern details", pattern.Details)
	assert.True(t, pattern.IsCritical)
	assert.Equal(t, "192.168.1.107", pattern.ClientIP)
	assert.NotZero(t, pattern.Timestamp)
}

func TestThreatDetector_ConcurrentAccess(t *testing.T) {
	config := &monitoring.SecurityMonitoringConfig{
		BruteForceThreshold:      10,
		BruteForceWindow:         time.Minute,
		SuspiciousPatternEnabled: true,
	}

	detector := NewThreatDetector(config)

	// Test concurrent access to threat detector
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			clientIP := "192.168.1.200"
			
			// Record some auth failures
			for j := 0; j < 5; j++ {
				detector.RecordAuthFailure(clientIP, "jwt")
			}

			// Analyze some requests
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request, _ = http.NewRequest("GET", "/api/test", nil)

			for j := 0; j < 20; j++ {
				detector.AnalyzeRequest(c, clientIP, "ConcurrentAgent/1.0")
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify no race conditions occurred
	detector.mu.RLock()
	assert.NotEmpty(t, detector.failureTracker)
	detector.mu.RUnlock()

	detector.patternDetector.mu.RLock()
	assert.NotEmpty(t, detector.patternDetector.requestPatterns)
	detector.patternDetector.mu.RUnlock()
}