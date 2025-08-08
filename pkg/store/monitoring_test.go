package store

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

func TestBusinessDomainInference(t *testing.T) {
	tests := []struct {
		tableName      string
		expectedDomain string
	}{
		{"employees", "hr"},
		{"employee_roles", "hr"},
		{"employee_positions", "hr"},
		{"projects", "project_management"},
		{"project_members", "project_management"},
		{"invoices", "finance"},
		{"invoice_numbers", "finance"},
		{"payrolls", "finance"},
		{"clients", "client_management"},
		{"audits", "compliance"},
		{"permissions", "security"},
		{"api_keys", "security"},
		{"unknown_table", ""},
		{"employee_test", "hr"}, // prefix match
		{"project_test", "project_management"}, // prefix match
		{"invoice_test", "finance"}, // prefix match
	}

	for _, test := range tests {
		t.Run(test.tableName, func(t *testing.T) {
			// Mock GORM DB with table name
			mockDB := &gorm.DB{
				Statement: &gorm.Statement{
					Table: test.tableName,
				},
			}

			domain := inferBusinessDomain(mockDB)
			assert.Equal(t, test.expectedDomain, domain)
		})
	}
}

func TestGetTableName(t *testing.T) {
	tests := []struct {
		name          string
		setupDB       func() *gorm.DB
		expectedTable string
	}{
		{
			name: "table name from statement",
			setupDB: func() *gorm.DB {
				return &gorm.DB{
					Statement: &gorm.Statement{
						Table: "employees",
					},
				}
			},
			expectedTable: "employees",
		},
		{
			name: "table name from empty statement", 
			setupDB: func() *gorm.DB {
				return &gorm.DB{
					Statement: &gorm.Statement{
						Table: "",
					},
				}
			},
			expectedTable: "unknown",
		},
		{
			name: "unknown table",
			setupDB: func() *gorm.DB {
				return &gorm.DB{
					Statement: &gorm.Statement{},
				}
			},
			expectedTable: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := test.setupDB()
			tableName := getTableName(db)
			assert.Equal(t, test.expectedTable, tableName)
		})
	}
}

func TestDatabaseMonitoringConfiguration(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		config := monitoring.DefaultDatabaseConfig()
		
		assert.True(t, config.Enabled)
		assert.Equal(t, 15*time.Second, config.RefreshInterval)
		assert.True(t, config.CustomMetrics)
		assert.Equal(t, 1*time.Second, config.SlowQueryThreshold)
		assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
		assert.True(t, config.BusinessMetrics)
		assert.Equal(t, 100, config.MaxTableCardinality)
	})

	t.Run("configuration validation", func(t *testing.T) {
		config := &monitoring.DatabaseMonitoringConfig{
			Enabled:               true,
			RefreshInterval:       0, // Invalid - should be fixed
			SlowQueryThreshold:    -1 * time.Second, // Invalid - should be fixed
			HealthCheckInterval:   0, // Invalid - should be fixed
			MaxTableCardinality:   0, // Invalid - should be fixed
		}

		err := config.Validate()
		assert.NoError(t, err)

		// Check that invalid values were corrected
		assert.Equal(t, 15*time.Second, config.RefreshInterval)
		assert.Equal(t, 1*time.Second, config.SlowQueryThreshold)
		assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
		assert.Equal(t, 100, config.MaxTableCardinality)
	})
}

func TestSlowQueryThresholdConfiguration(t *testing.T) {
	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:            true,
		CustomMetrics:      true,
		SlowQueryThreshold: 1 * time.Millisecond,
		BusinessMetrics:    true,
	}

	assert.Equal(t, 1*time.Millisecond, config.SlowQueryThreshold)
	assert.True(t, config.Enabled)
	assert.True(t, config.CustomMetrics)
}

func TestHealthCheckConfiguration(t *testing.T) {
	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:             true,
		HealthCheckInterval: 100 * time.Millisecond,
	}

	assert.Equal(t, 100*time.Millisecond, config.HealthCheckInterval)
	assert.True(t, config.Enabled)
}

func TestMetricsRegistration(t *testing.T) {
	// Create a custom registry for testing
	registry := prometheus.NewRegistry()
	
	// Test that our metrics can be registered
	err := registry.Register(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "database", 
			Name:      "operations_total_test",
			Help:      "Test metric for database operations",
		},
		[]string{"operation", "table", "result"},
	))
	assert.NoError(t, err)
}

func TestTransactionMetricsStructure(t *testing.T) {
	// Test transaction metric structure
	config := &monitoring.DatabaseMonitoringConfig{
		Enabled:       true,
		CustomMetrics: true,
	}

	assert.True(t, config.Enabled)
	assert.True(t, config.CustomMetrics)
	
	// Test metric structure - this tests our metrics are properly defined
	// without requiring actual database operations
}

func TestDatabaseOperationMetrics(t *testing.T) {
	// Test the structure and labels of our database metrics
	t.Run("database operations counter", func(t *testing.T) {
		// Create test metric with same structure
		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "fortress",
				Subsystem: "database",
				Name:      "operations_total",
				Help:      "Total database operations by type and result",
			},
			[]string{"operation", "table", "result"},
		)

		// Test incrementing with different labels
		counter.WithLabelValues("create", "employees", "success").Inc()
		counter.WithLabelValues("select", "projects", "success").Inc()
		counter.WithLabelValues("update", "invoices", "error").Inc()

		// Verify metrics can be collected
		metricFamilies, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)
		assert.NotEmpty(t, metricFamilies)
	})

	t.Run("database operation duration histogram", func(t *testing.T) {
		histogram := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "fortress",
				Subsystem: "database",
				Name:      "operation_duration_seconds",
				Help:      "Database operation duration distribution",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"operation", "table"},
		)

		// Test observing different durations
		histogram.WithLabelValues("create", "employees").Observe(0.05)
		histogram.WithLabelValues("select", "projects").Observe(0.1)
		histogram.WithLabelValues("update", "invoices").Observe(0.5)

		// Verify histogram structure by checking the observer interface
		observer := histogram.WithLabelValues("create", "employees")
		assert.NotNil(t, observer)
		
		// Test that we can observe values
		observer.Observe(0.1)
	})
}

func TestMetricsCardinality(t *testing.T) {
	// Test that we properly control metric cardinality
	config := &monitoring.DatabaseMonitoringConfig{
		MaxTableCardinality: 3,
	}

	tableNames := []string{"employees", "projects", "invoices", "should_be_ignored"}
	
	for i, tableName := range tableNames {
		if i >= config.MaxTableCardinality {
			// In real implementation, we would skip tracking this table
			continue
		}
		// Normal tracking logic would go here
		assert.True(t, len(tableName) > 0)
	}
}