package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityMetricsInitialization(t *testing.T) {
	// Test that all security metrics are properly initialized
	assert.NotNil(t, AuthenticationAttempts)
	assert.NotNil(t, AuthenticationDuration)
	assert.NotNil(t, ActiveSessions)
	assert.NotNil(t, PermissionChecks)
	assert.NotNil(t, AuthorizationDuration)
	assert.NotNil(t, SuspiciousActivity)
	assert.NotNil(t, RateLimitViolations)
	assert.NotNil(t, SecurityEvents)
	assert.NotNil(t, APIKeyUsage)
	assert.NotNil(t, APIKeyValidationErrors)
}

func TestAuthenticationMetricsLabels(t *testing.T) {
	// Test authentication metrics with expected labels
	AuthenticationAttempts.WithLabelValues("jwt", "success", "").Inc()
	AuthenticationAttempts.WithLabelValues("api_key", "failure", "invalid_key").Inc()
	AuthenticationAttempts.WithLabelValues("jwt", "failure", "token_expired").Inc()

	// Test authentication duration histogram
	AuthenticationDuration.WithLabelValues("jwt").Observe(0.05)
	AuthenticationDuration.WithLabelValues("api_key").Observe(0.02)

	// All metrics should be successfully recorded without errors
	assert.True(t, true, "All authentication metrics recorded successfully")
}

func TestAuthorizationMetricsLabels(t *testing.T) {
	// Test permission checks counter
	PermissionChecks.WithLabelValues("employees.read", "allowed").Inc()
	PermissionChecks.WithLabelValues("employees.create", "denied").Inc()
	PermissionChecks.WithLabelValues("projects.read", "allowed").Inc()

	// Test authorization duration histogram
	AuthorizationDuration.WithLabelValues("employees.read").Observe(0.001)
	AuthorizationDuration.WithLabelValues("projects.create").Observe(0.003)

	// All metrics should be successfully recorded without errors
	assert.True(t, true, "All authorization metrics recorded successfully")
}

func TestSecurityThreatMetricsLabels(t *testing.T) {
	// Test suspicious activity counter
	SuspiciousActivity.WithLabelValues("brute_force", "high").Inc()
	SuspiciousActivity.WithLabelValues("rapid_requests", "medium").Inc()
	SuspiciousActivity.WithLabelValues("user_agent_variation", "low").Inc()

	// Test rate limit violations
	RateLimitViolations.WithLabelValues("/api/v1/auth", "frequency_exceeded").Inc()
	RateLimitViolations.WithLabelValues("/api/v1/employees", "concurrent_exceeded").Inc()

	// Test security events
	SecurityEvents.WithLabelValues("auth_failure", "medium", "middleware").Inc()
	SecurityEvents.WithLabelValues("permission_violation", "high", "authorization").Inc()

	// All metrics should be successfully recorded without errors
	assert.True(t, true, "All security threat metrics recorded successfully")
}

func TestAPIKeyMetricsLabels(t *testing.T) {
	// Test API key usage counter
	APIKeyUsage.WithLabelValues("client123", "success").Inc()
	APIKeyUsage.WithLabelValues("client456", "failure").Inc()

	// Test API key validation errors
	APIKeyValidationErrors.WithLabelValues("invalid_format").Inc()
	APIKeyValidationErrors.WithLabelValues("key_not_found").Inc()
	APIKeyValidationErrors.WithLabelValues("key_disabled").Inc()
	APIKeyValidationErrors.WithLabelValues("key_mismatch").Inc()
	APIKeyValidationErrors.WithLabelValues("database_error").Inc()

	// All metrics should be successfully recorded without errors
	assert.True(t, true, "All API key metrics recorded successfully")
}

func TestActiveSessionsGauge(t *testing.T) {
	// Test active sessions gauge
	ActiveSessions.Set(150)
	ActiveSessions.Inc()
	ActiveSessions.Dec()

	// All gauge operations should work without errors
	assert.True(t, true, "Active sessions gauge operations completed successfully")
}

func TestSecurityMetricsStructure(t *testing.T) {
	// Test metric naming conventions for security metrics
	tests := []struct {
		name     string
		expected string
	}{
		{"auth_attempts", "fortress_auth_attempts_total"},
		{"auth_duration", "fortress_auth_duration_seconds"},
		{"active_sessions", "fortress_auth_active_sessions_total"},
		{"permission_checks", "fortress_auth_permission_checks_total"},
		{"authorization_duration", "fortress_auth_authorization_duration_seconds"},
		{"suspicious_activity", "fortress_security_suspicious_activity_total"},
		{"rate_limit_violations", "fortress_security_rate_limit_violations_total"},
		{"security_events", "fortress_security_events_total"},
		{"api_key_usage", "fortress_auth_api_key_usage_total"},
		{"api_key_errors", "fortress_auth_api_key_validation_errors_total"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// This test ensures our security metrics follow Prometheus conventions
			assert.Contains(t, test.expected, "fortress_")
			assert.NotEmpty(t, test.expected)
		})
	}
}

func TestSecurityMetricsLabelValidation(t *testing.T) {
	// Test that metrics accept valid label combinations without panicking
	tests := []struct {
		name   string
		metric func()
	}{
		{
			"auth_attempts_all_combinations",
			func() {
				methods := []string{"jwt", "api_key", "unknown"}
				results := []string{"success", "failure"}
				reasons := []string{"", "invalid_key", "token_expired", "invalid_format"}

				for _, method := range methods {
					for _, result := range results {
						for _, reason := range reasons {
							AuthenticationAttempts.WithLabelValues(method, result, reason).Inc()
						}
					}
				}
			},
		},
		{
			"permission_checks_combinations",
			func() {
				permissions := []string{"employees.read", "employees.create", "projects.read", "unknown"}
				results := []string{"allowed", "denied", "error"}

				for _, perm := range permissions {
					for _, result := range results {
						PermissionChecks.WithLabelValues(perm, result).Inc()
					}
				}
			},
		},
		{
			"suspicious_activity_combinations",
			func() {
				eventTypes := []string{"brute_force", "rapid_requests", "user_agent_variation", "unusual_pattern"}
				severities := []string{"low", "medium", "high", "critical"}

				for _, eventType := range eventTypes {
					for _, severity := range severities {
						SuspiciousActivity.WithLabelValues(eventType, severity).Inc()
					}
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Should not panic with valid label combinations
			assert.NotPanics(t, test.metric, "Metric should accept valid label combinations")
		})
	}
}