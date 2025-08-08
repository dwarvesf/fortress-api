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