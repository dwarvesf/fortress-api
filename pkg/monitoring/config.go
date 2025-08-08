package monitoring

import (
	"strings"
	"time"
)

// PrometheusConfig configures the Prometheus monitoring middleware
type PrometheusConfig struct {
	// Enabled enables or disables Prometheus metrics collection
	Enabled bool

	// SampleRate controls the percentage of requests to monitor (0.0 to 1.0)
	SampleRate float64

	// NormalizePaths enables path parameter normalization (e.g., /users/:id)
	NormalizePaths bool

	// MaxEndpoints limits the number of unique endpoints to track
	MaxEndpoints int

	// ExcludePaths defines paths to exclude from monitoring
	ExcludePaths []string

	// RequestTimeout defines the timeout for recording metrics
	RequestTimeout time.Duration
}

// DatabaseMonitoringConfig configures database monitoring
type DatabaseMonitoringConfig struct {
	// Enabled enables or disables database monitoring
	Enabled bool

	// RefreshInterval for GORM Prometheus plugin refresh
	RefreshInterval time.Duration

	// CustomMetrics enables custom GORM callbacks
	CustomMetrics bool

	// SlowQueryThreshold defines the threshold for slow query detection
	SlowQueryThreshold time.Duration

	// HealthCheckInterval for database connection health checks
	HealthCheckInterval time.Duration

	// BusinessMetrics enables business-specific database metrics
	BusinessMetrics bool

	// MaxTableCardinality limits unique table names tracked
	MaxTableCardinality int
}

// SecurityMonitoringConfig configures security monitoring
type SecurityMonitoringConfig struct {
	// Enabled enables or disables security monitoring
	Enabled bool

	// ThreatDetectionEnabled enables threat detection engine
	ThreatDetectionEnabled bool

	// BruteForceThreshold number of auth failures before brute force detection
	BruteForceThreshold int

	// BruteForceWindow time window for brute force detection
	BruteForceWindow time.Duration

	// SuspiciousPatternEnabled enables suspicious pattern detection
	SuspiciousPatternEnabled bool

	// LogSecurityEvents enables security event logging
	LogSecurityEvents bool

	// RateLimitMonitoring enables rate limit violation monitoring
	RateLimitMonitoring bool
}

// DefaultConfig returns a PrometheusConfig with sensible defaults
func DefaultConfig() *PrometheusConfig {
	return &PrometheusConfig{
		Enabled:        true,
		SampleRate:     1.0,
		NormalizePaths: true,
		MaxEndpoints:   100,
		ExcludePaths: []string{
			"/metrics",
			"/healthz",
			"/health",
			"/ping",
			"/favicon.ico",
		},
		RequestTimeout: 30 * time.Second,
	}
}

// Validate validates the configuration and returns error if invalid
func (c *PrometheusConfig) Validate() error {
	if c.SampleRate < 0.0 || c.SampleRate > 1.0 {
		c.SampleRate = 1.0
	}
	
	if c.MaxEndpoints <= 0 {
		c.MaxEndpoints = 100
	}

	if c.RequestTimeout <= 0 {
		c.RequestTimeout = 30 * time.Second
	}

	return nil
}

// ShouldExclude returns true if the given path should be excluded from monitoring
func (c *PrometheusConfig) ShouldExclude(path string) bool {
	for _, excludePath := range c.ExcludePaths {
		if path == excludePath || strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	return false
}

// DefaultDatabaseConfig returns a DatabaseMonitoringConfig with sensible defaults
func DefaultDatabaseConfig() *DatabaseMonitoringConfig {
	return &DatabaseMonitoringConfig{
		Enabled:               true,
		RefreshInterval:       15 * time.Second,
		CustomMetrics:         true,
		SlowQueryThreshold:    1 * time.Second,
		HealthCheckInterval:   30 * time.Second,
		BusinessMetrics:       true,
		MaxTableCardinality:   100,
	}
}

// Validate validates the database monitoring configuration
func (c *DatabaseMonitoringConfig) Validate() error {
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = 15 * time.Second
	}

	if c.SlowQueryThreshold <= 0 {
		c.SlowQueryThreshold = 1 * time.Second
	}

	if c.HealthCheckInterval <= 0 {
		c.HealthCheckInterval = 30 * time.Second
	}

	if c.MaxTableCardinality <= 0 {
		c.MaxTableCardinality = 100
	}

	return nil
}

// DefaultSecurityConfig returns a SecurityMonitoringConfig with sensible defaults
func DefaultSecurityConfig() *SecurityMonitoringConfig {
	return &SecurityMonitoringConfig{
		Enabled:                   true,
		ThreatDetectionEnabled:    true,
		BruteForceThreshold:       10,
		BruteForceWindow:          5 * time.Minute,
		SuspiciousPatternEnabled:  true,
		LogSecurityEvents:         true,
		RateLimitMonitoring:       true,
	}
}

// Validate validates the security monitoring configuration
func (c *SecurityMonitoringConfig) Validate() error {
	if c.BruteForceThreshold <= 0 {
		c.BruteForceThreshold = 10
	}

	if c.BruteForceWindow <= 0 {
		c.BruteForceWindow = 5 * time.Minute
	}

	return nil
}