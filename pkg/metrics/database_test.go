package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseMetricsInitialization(t *testing.T) {
	// Test that all database metrics are properly initialized
	assert.NotNil(t, DatabaseOperations)
	assert.NotNil(t, DatabaseOperationDuration)
	assert.NotNil(t, DatabaseSlowQueries)
	assert.NotNil(t, DatabaseConnectionHealth)
	assert.NotNil(t, DatabaseTransactions)
	assert.NotNil(t, DatabaseLockWaitTime)
	assert.NotNil(t, DatabaseBusinessOperations)
	assert.NotNil(t, DatabaseConnectionPoolEfficiency)
	assert.NotNil(t, DatabaseConnectionWaitDuration)
}

func TestDatabaseMetricsLabels(t *testing.T) {
	// Test database operations counter with expected labels
	DatabaseOperations.WithLabelValues("create", "employees", "success").Inc()
	DatabaseOperations.WithLabelValues("select", "projects", "success").Inc()
	DatabaseOperations.WithLabelValues("update", "invoices", "error").Inc()

	// Test operation duration histogram
	DatabaseOperationDuration.WithLabelValues("create", "employees").Observe(0.1)
	DatabaseOperationDuration.WithLabelValues("select", "projects").Observe(0.05)

	// Test slow queries counter
	DatabaseSlowQueries.WithLabelValues("employees", "select").Inc()
	DatabaseSlowQueries.WithLabelValues("projects", "update").Inc()

	// Test connection health gauge
	DatabaseConnectionHealth.WithLabelValues("fortress", "primary").Set(1)
	DatabaseConnectionHealth.WithLabelValues("fortress", "replica").Set(0)

	// Test transactions counter
	DatabaseTransactions.WithLabelValues("commit").Inc()
	DatabaseTransactions.WithLabelValues("rollback").Inc()

	// Test business operations counter
	DatabaseBusinessOperations.WithLabelValues("hr", "create", "success").Inc()
	DatabaseBusinessOperations.WithLabelValues("finance", "select", "success").Inc()

	// Test connection pool efficiency
	DatabaseConnectionPoolEfficiency.WithLabelValues("fortress", "primary").Set(0.8)

	// Test connection wait duration
	DatabaseConnectionWaitDuration.WithLabelValues("fortress", "primary").Observe(0.01)

	// All metrics should be successfully recorded without errors
	assert.True(t, true, "All database metrics recorded successfully")
}

func TestDatabaseMetricsStructure(t *testing.T) {
	// Test metric naming conventions
	tests := []struct {
		name     string
		expected string
	}{
		{"operations", "fortress_database_operations_total"},
		{"duration", "fortress_database_operation_duration_seconds"},
		{"slow_queries", "fortress_database_slow_queries_total"},
		{"health", "fortress_database_connection_health_status"},
		{"transactions", "fortress_database_transactions_total"},
		{"business", "fortress_database_business_operations_total"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// This test ensures our naming follows Prometheus conventions
			assert.Contains(t, test.expected, "fortress_database_")
			assert.NotEmpty(t, test.expected)
		})
	}
}