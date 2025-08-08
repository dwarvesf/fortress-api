package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
	"github.com/dwarvesf/fortress-api/pkg/security"
)

type SecurityMonitoringMiddleware struct {
	config         *SecurityMonitoringConfig
	logger         logger.Logger
	threatDetector *security.ThreatDetector
}

type SecurityMonitoringConfig struct {
	Enabled                   bool          `mapstructure:"enabled"`
	ThreatDetectionEnabled    bool          `mapstructure:"threat_detection_enabled"`
	BruteForceThreshold       int           `mapstructure:"brute_force_threshold"`
	BruteForceWindow          time.Duration `mapstructure:"brute_force_window"`
	SuspiciousPatternEnabled  bool          `mapstructure:"suspicious_pattern_enabled"`
	LogSecurityEvents         bool          `mapstructure:"log_security_events"`
	RateLimitMonitoring       bool          `mapstructure:"rate_limit_monitoring"`
}

func NewSecurityMonitoringMiddleware(cfg *SecurityMonitoringConfig, logger logger.Logger) *SecurityMonitoringMiddleware {
	if cfg == nil {
		cfg = &SecurityMonitoringConfig{
			Enabled:                   true,
			ThreatDetectionEnabled:    true,
			BruteForceThreshold:       10,
			BruteForceWindow:          time.Minute * 5,
			SuspiciousPatternEnabled:  true,
			LogSecurityEvents:         true,
			RateLimitMonitoring:       true,
		}
	}

	// Convert to monitoring.SecurityMonitoringConfig for threat detector
	monitoringConfig := &monitoring.SecurityMonitoringConfig{
		Enabled:                   cfg.Enabled,
		ThreatDetectionEnabled:    cfg.ThreatDetectionEnabled,
		BruteForceThreshold:       cfg.BruteForceThreshold,
		BruteForceWindow:          cfg.BruteForceWindow,
		SuspiciousPatternEnabled:  cfg.SuspiciousPatternEnabled,
		LogSecurityEvents:         cfg.LogSecurityEvents,
		RateLimitMonitoring:       cfg.RateLimitMonitoring,
	}

	return &SecurityMonitoringMiddleware{
		config:         cfg,
		logger:         logger,
		threatDetector: security.NewThreatDetector(monitoringConfig),
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
			smw.logger.AddField("permission", permission).
				AddField("client_ip", c.ClientIP()).
				AddField("user_agent", c.GetHeader("User-Agent")).
				AddField("endpoint", c.Request.URL.Path).
				Warn("Permission denied")
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
			smw.logger.AddField("pattern_type", pattern.Type).
				AddField("severity", severity).
				AddField("client_ip", clientIP).
				AddField("details", pattern.Details).
				Warn("Suspicious activity detected")
		}
	}
}

func (smw *SecurityMonitoringMiddleware) logSecurityEvent(c *gin.Context, clientIP, userAgent string, start time.Time) {
	if !smw.config.LogSecurityEvents {
		return
	}

	// Log security events for audit trail
	// This is a placeholder for security event logging
	// In a real implementation, you might want to log to a separate security audit log
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