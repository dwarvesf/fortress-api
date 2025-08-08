package security

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

type ThreatDetector struct {
	config          *monitoring.SecurityMonitoringConfig
	failureTracker  map[string]*AuthFailureTracker
	patternDetector *PatternDetector
	mu              sync.RWMutex
}

type AuthFailureTracker struct {
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
	Methods   map[string]int
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
	Count      int
	LastSeen   time.Time
	Endpoints  map[string]int
	UserAgents []string
}

func NewThreatDetector(config *monitoring.SecurityMonitoringConfig) *ThreatDetector {
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