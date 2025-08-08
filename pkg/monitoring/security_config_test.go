package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecurityMonitoringConfigDefaults(t *testing.T) {
	config := DefaultSecurityConfig()

	assert.True(t, config.Enabled)
	assert.True(t, config.ThreatDetectionEnabled)
	assert.Equal(t, 10, config.BruteForceThreshold)
	assert.Equal(t, 5*time.Minute, config.BruteForceWindow)
	assert.True(t, config.SuspiciousPatternEnabled)
	assert.True(t, config.LogSecurityEvents)
	assert.True(t, config.RateLimitMonitoring)
}

func TestSecurityMonitoringConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityMonitoringConfig
		want   *SecurityMonitoringConfig
	}{
		{
			name: "valid configuration unchanged",
			config: &SecurityMonitoringConfig{
				Enabled:                   true,
				ThreatDetectionEnabled:    true,
				BruteForceThreshold:       15,
				BruteForceWindow:          10 * time.Minute,
				SuspiciousPatternEnabled:  true,
				LogSecurityEvents:         true,
				RateLimitMonitoring:       true,
			},
			want: &SecurityMonitoringConfig{
				Enabled:                   true,
				ThreatDetectionEnabled:    true,
				BruteForceThreshold:       15,
				BruteForceWindow:          10 * time.Minute,
				SuspiciousPatternEnabled:  true,
				LogSecurityEvents:         true,
				RateLimitMonitoring:       true,
			},
		},
		{
			name: "zero brute force threshold corrected",
			config: &SecurityMonitoringConfig{
				BruteForceThreshold: 0,
				BruteForceWindow:    5 * time.Minute,
			},
			want: &SecurityMonitoringConfig{
				BruteForceThreshold: 10,
				BruteForceWindow:    5 * time.Minute,
			},
		},
		{
			name: "negative brute force threshold corrected",
			config: &SecurityMonitoringConfig{
				BruteForceThreshold: -5,
				BruteForceWindow:    5 * time.Minute,
			},
			want: &SecurityMonitoringConfig{
				BruteForceThreshold: 10,
				BruteForceWindow:    5 * time.Minute,
			},
		},
		{
			name: "zero brute force window corrected",
			config: &SecurityMonitoringConfig{
				BruteForceThreshold: 15,
				BruteForceWindow:    0,
			},
			want: &SecurityMonitoringConfig{
				BruteForceThreshold: 15,
				BruteForceWindow:    5 * time.Minute,
			},
		},
		{
			name: "negative brute force window corrected",
			config: &SecurityMonitoringConfig{
				BruteForceThreshold: 15,
				BruteForceWindow:    -1 * time.Minute,
			},
			want: &SecurityMonitoringConfig{
				BruteForceThreshold: 15,
				BruteForceWindow:    5 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.BruteForceThreshold, tt.config.BruteForceThreshold)
			assert.Equal(t, tt.want.BruteForceWindow, tt.config.BruteForceWindow)
		})
	}
}

func TestSecurityMonitoringConfigDisabled(t *testing.T) {
	config := &SecurityMonitoringConfig{
		Enabled: false,
	}

	err := config.Validate()
	assert.NoError(t, err)
	assert.False(t, config.Enabled)

	// Even disabled configs should have valid defaults after validation
	assert.Equal(t, 10, config.BruteForceThreshold)
	assert.Equal(t, 5*time.Minute, config.BruteForceWindow)
}

func TestSecurityMonitoringConfigFeatureToggles(t *testing.T) {
	tests := []struct {
		name                     string
		threatDetectionEnabled   bool
		suspiciousPatternEnabled bool
		rateLimitMonitoring      bool
		logSecurityEvents        bool
	}{
		{"all features enabled", true, true, true, true},
		{"only threat detection", true, false, false, false},
		{"only pattern detection", false, true, false, false},
		{"only rate limit monitoring", false, false, true, false},
		{"only security logging", false, false, false, true},
		{"all features disabled", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SecurityMonitoringConfig{
				Enabled:                   true,
				ThreatDetectionEnabled:    tt.threatDetectionEnabled,
				SuspiciousPatternEnabled:  tt.suspiciousPatternEnabled,
				RateLimitMonitoring:       tt.rateLimitMonitoring,
				LogSecurityEvents:         tt.logSecurityEvents,
				BruteForceThreshold:       10,
				BruteForceWindow:          5 * time.Minute,
			}

			err := config.Validate()
			assert.NoError(t, err)
			assert.Equal(t, tt.threatDetectionEnabled, config.ThreatDetectionEnabled)
			assert.Equal(t, tt.suspiciousPatternEnabled, config.SuspiciousPatternEnabled)
			assert.Equal(t, tt.rateLimitMonitoring, config.RateLimitMonitoring)
			assert.Equal(t, tt.logSecurityEvents, config.LogSecurityEvents)
		})
	}
}

func TestSecurityMonitoringConfigEdgeCases(t *testing.T) {
	// Test with nil config - should not panic
	var config *SecurityMonitoringConfig
	assert.NotPanics(t, func() {
		if config != nil {
			config.Validate()
		}
	})

	// Test with very high thresholds (edge case but valid)
	config = &SecurityMonitoringConfig{
		BruteForceThreshold: 1000,
		BruteForceWindow:    24 * time.Hour,
	}

	err := config.Validate()
	assert.NoError(t, err)
	assert.Equal(t, 1000, config.BruteForceThreshold)
	assert.Equal(t, 24*time.Hour, config.BruteForceWindow)
}

func TestSecurityMonitoringConfigIntegration(t *testing.T) {
	// Test that security config integrates with the overall monitoring config structure
	config := DefaultSecurityConfig()

	// Should have all expected fields
	assert.NotNil(t, config)
	assert.IsType(t, &SecurityMonitoringConfig{}, config)

	// Should be valid by default
	err := config.Validate()
	assert.NoError(t, err)

	// Should support being disabled
	config.Enabled = false
	err = config.Validate()
	assert.NoError(t, err)
	assert.False(t, config.Enabled)
}