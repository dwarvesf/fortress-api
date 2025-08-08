package monitoring

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test PrometheusConfig Default Configuration
func TestPrometheusConfigDefaults(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 1.0, config.SampleRate)
	assert.True(t, config.NormalizePaths)
	assert.Equal(t, 100, config.MaxEndpoints)
	assert.Equal(t, 30*time.Second, config.RequestTimeout)

	// Check default exclude paths
	expectedPaths := []string{
		"/metrics",
		"/healthz",
		"/health",
		"/ping",
		"/favicon.ico",
	}
	assert.Equal(t, expectedPaths, config.ExcludePaths)
}

// Test PrometheusConfig Validation
func TestPrometheusConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		config         *PrometheusConfig
		expectedSample float64
		expectedMax    int
		expectedTimeout time.Duration
	}{
		{
			name: "valid configuration unchanged",
			config: &PrometheusConfig{
				Enabled:        true,
				SampleRate:     0.5,
				NormalizePaths: true,
				MaxEndpoints:   200,
				RequestTimeout: 45 * time.Second,
			},
			expectedSample: 0.5,
			expectedMax:    200,
			expectedTimeout: 45 * time.Second,
		},
		{
			name: "negative sample rate corrected",
			config: &PrometheusConfig{
				SampleRate: -0.1,
			},
			expectedSample: 1.0,
		},
		{
			name: "sample rate over 1.0 corrected",
			config: &PrometheusConfig{
				SampleRate: 1.5,
			},
			expectedSample: 1.0,
		},
		{
			name: "zero max endpoints corrected",
			config: &PrometheusConfig{
				MaxEndpoints: 0,
			},
			expectedMax: 100,
		},
		{
			name: "negative max endpoints corrected",
			config: &PrometheusConfig{
				MaxEndpoints: -10,
			},
			expectedMax: 100,
		},
		{
			name: "zero timeout corrected",
			config: &PrometheusConfig{
				RequestTimeout: 0,
			},
			expectedTimeout: 30 * time.Second,
		},
		{
			name: "negative timeout corrected",
			config: &PrometheusConfig{
				RequestTimeout: -5 * time.Second,
			},
			expectedTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err)
			
			if tt.expectedSample != 0 {
				assert.Equal(t, tt.expectedSample, tt.config.SampleRate)
			}
			if tt.expectedMax != 0 {
				assert.Equal(t, tt.expectedMax, tt.config.MaxEndpoints)
			}
			if tt.expectedTimeout != 0 {
				assert.Equal(t, tt.expectedTimeout, tt.config.RequestTimeout)
			}
		})
	}
}

// Test ShouldExclude functionality
func TestPrometheusConfigShouldExclude(t *testing.T) {
	config := &PrometheusConfig{
		ExcludePaths: []string{
			"/metrics",
			"/healthz",
			"/debug",
		},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/metrics", true},
		{"/healthz", true},
		{"/debug", true},
		{"/debug/pprof", true}, // prefix match
		{"/api/v1/users", false},
		{"/health", false}, // not exact match or prefix
		{"/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := config.ShouldExclude(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test DatabaseMonitoringConfig Default Configuration
func TestDatabaseMonitoringConfigDefaults(t *testing.T) {
	config := DefaultDatabaseConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 15*time.Second, config.RefreshInterval)
	assert.True(t, config.CustomMetrics)
	assert.Equal(t, 1*time.Second, config.SlowQueryThreshold)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.True(t, config.BusinessMetrics)
	assert.Equal(t, 100, config.MaxTableCardinality)
}

// Test DatabaseMonitoringConfig Validation
func TestDatabaseMonitoringConfigValidation(t *testing.T) {
	tests := []struct {
		name               string
		config             *DatabaseMonitoringConfig
		expectedRefresh    time.Duration
		expectedSlow       time.Duration
		expectedHealth     time.Duration
		expectedCardinality int
	}{
		{
			name: "valid configuration unchanged",
			config: &DatabaseMonitoringConfig{
				Enabled:               true,
				RefreshInterval:       30 * time.Second,
				CustomMetrics:         true,
				SlowQueryThreshold:    2 * time.Second,
				HealthCheckInterval:   60 * time.Second,
				BusinessMetrics:       true,
				MaxTableCardinality:   200,
			},
			expectedRefresh:     30 * time.Second,
			expectedSlow:        2 * time.Second,
			expectedHealth:      60 * time.Second,
			expectedCardinality: 200,
		},
		{
			name: "zero refresh interval corrected",
			config: &DatabaseMonitoringConfig{
				RefreshInterval: 0,
			},
			expectedRefresh: 15 * time.Second,
		},
		{
			name: "negative refresh interval corrected",
			config: &DatabaseMonitoringConfig{
				RefreshInterval: -10 * time.Second,
			},
			expectedRefresh: 15 * time.Second,
		},
		{
			name: "zero slow query threshold corrected",
			config: &DatabaseMonitoringConfig{
				SlowQueryThreshold: 0,
			},
			expectedSlow: 1 * time.Second,
		},
		{
			name: "negative slow query threshold corrected",
			config: &DatabaseMonitoringConfig{
				SlowQueryThreshold: -1 * time.Second,
			},
			expectedSlow: 1 * time.Second,
		},
		{
			name: "zero health check interval corrected",
			config: &DatabaseMonitoringConfig{
				HealthCheckInterval: 0,
			},
			expectedHealth: 30 * time.Second,
		},
		{
			name: "negative health check interval corrected",
			config: &DatabaseMonitoringConfig{
				HealthCheckInterval: -5 * time.Second,
			},
			expectedHealth: 30 * time.Second,
		},
		{
			name: "zero cardinality corrected",
			config: &DatabaseMonitoringConfig{
				MaxTableCardinality: 0,
			},
			expectedCardinality: 100,
		},
		{
			name: "negative cardinality corrected",
			config: &DatabaseMonitoringConfig{
				MaxTableCardinality: -50,
			},
			expectedCardinality: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err)
			
			if tt.expectedRefresh != 0 {
				assert.Equal(t, tt.expectedRefresh, tt.config.RefreshInterval)
			}
			if tt.expectedSlow != 0 {
				assert.Equal(t, tt.expectedSlow, tt.config.SlowQueryThreshold)
			}
			if tt.expectedHealth != 0 {
				assert.Equal(t, tt.expectedHealth, tt.config.HealthCheckInterval)
			}
			if tt.expectedCardinality != 0 {
				assert.Equal(t, tt.expectedCardinality, tt.config.MaxTableCardinality)
			}
		})
	}
}

// Test edge cases and disabled configurations
func TestConfigEdgeCases(t *testing.T) {
	t.Run("PrometheusConfig disabled", func(t *testing.T) {
		config := &PrometheusConfig{
			Enabled: false,
			// Set some invalid values that should be corrected
			SampleRate:     -1.0, // Invalid, should be corrected to 1.0
			MaxEndpoints:   0,    // Invalid, should be corrected to 100
			RequestTimeout: 0,    // Invalid, should be corrected to 30s
		}
		
		err := config.Validate()
		assert.NoError(t, err)
		assert.False(t, config.Enabled)
		// Disabled configs should still have valid defaults for invalid values
		assert.Equal(t, 1.0, config.SampleRate)
		assert.Equal(t, 100, config.MaxEndpoints)
		assert.Equal(t, 30*time.Second, config.RequestTimeout)
	})

	t.Run("DatabaseMonitoringConfig disabled", func(t *testing.T) {
		config := &DatabaseMonitoringConfig{
			Enabled: false,
			// Set some invalid values that should be corrected
			RefreshInterval:       0, // Invalid, should be corrected to 15s
			SlowQueryThreshold:    0, // Invalid, should be corrected to 1s
			HealthCheckInterval:   0, // Invalid, should be corrected to 30s
			MaxTableCardinality:   0, // Invalid, should be corrected to 100
		}
		
		err := config.Validate()
		assert.NoError(t, err)
		assert.False(t, config.Enabled)
		// Disabled configs should still have valid defaults for invalid values
		assert.Equal(t, 15*time.Second, config.RefreshInterval)
		assert.Equal(t, 1*time.Second, config.SlowQueryThreshold)
		assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
		assert.Equal(t, 100, config.MaxTableCardinality)
	})

	t.Run("PrometheusConfig with nil ExcludePaths", func(t *testing.T) {
		config := &PrometheusConfig{
			ExcludePaths: nil,
		}
		
		// Should not panic with nil slice
		assert.False(t, config.ShouldExclude("/any/path"))
	})

	t.Run("PrometheusConfig with empty ExcludePaths", func(t *testing.T) {
		config := &PrometheusConfig{
			ExcludePaths: []string{},
		}
		
		// Should not exclude anything
		assert.False(t, config.ShouldExclude("/metrics"))
		assert.False(t, config.ShouldExclude("/healthz"))
	})
}

// Test configuration feature toggles
func TestConfigFeatureToggles(t *testing.T) {
	t.Run("PrometheusConfig features", func(t *testing.T) {
		tests := []struct {
			name           string
			enabled        bool
			normalizePaths bool
			sampleRate     float64
		}{
			{"all enabled", true, true, 1.0},
			{"monitoring disabled", false, true, 1.0},
			{"path normalization disabled", true, false, 1.0},
			{"reduced sample rate", true, true, 0.1},
			{"all features minimal", false, false, 0.0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &PrometheusConfig{
					Enabled:        tt.enabled,
					NormalizePaths: tt.normalizePaths,
					SampleRate:     tt.sampleRate,
				}
				
				err := config.Validate()
				assert.NoError(t, err)
				assert.Equal(t, tt.enabled, config.Enabled)
				assert.Equal(t, tt.normalizePaths, config.NormalizePaths)
			})
		}
	})

	t.Run("DatabaseMonitoringConfig features", func(t *testing.T) {
		tests := []struct {
			name            string
			enabled         bool
			customMetrics   bool
			businessMetrics bool
		}{
			{"all enabled", true, true, true},
			{"monitoring disabled", false, true, true},
			{"custom metrics disabled", true, false, true},
			{"business metrics disabled", true, true, false},
			{"all metrics disabled", true, false, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				config := &DatabaseMonitoringConfig{
					Enabled:         tt.enabled,
					CustomMetrics:   tt.customMetrics,
					BusinessMetrics: tt.businessMetrics,
				}
				
				err := config.Validate()
				assert.NoError(t, err)
				assert.Equal(t, tt.enabled, config.Enabled)
				assert.Equal(t, tt.customMetrics, config.CustomMetrics)
				assert.Equal(t, tt.businessMetrics, config.BusinessMetrics)
			})
		}
	})
}

// Test configuration integration
func TestConfigIntegration(t *testing.T) {
	t.Run("all default configs are valid", func(t *testing.T) {
		promConfig := DefaultConfig()
		dbConfig := DefaultDatabaseConfig()
		secConfig := DefaultSecurityConfig()

		assert.NoError(t, promConfig.Validate())
		assert.NoError(t, dbConfig.Validate())
		assert.NoError(t, secConfig.Validate())
	})

	t.Run("configs work together", func(t *testing.T) {
		// Test that configs can be used together in a monitoring setup
		promConfig := DefaultConfig()
		dbConfig := DefaultDatabaseConfig()
		
		// Modify some settings
		promConfig.SampleRate = 0.5
		dbConfig.SlowQueryThreshold = 2 * time.Second
		
		// Both should still be valid
		assert.NoError(t, promConfig.Validate())
		assert.NoError(t, dbConfig.Validate())
		
		// Settings should be preserved
		assert.Equal(t, 0.5, promConfig.SampleRate)
		assert.Equal(t, 2*time.Second, dbConfig.SlowQueryThreshold)
	})
}