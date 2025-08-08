package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Business operation metrics
	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "operations_total",
			Help:      "Total database operations by type and result",
		},
		[]string{"operation", "table", "result"},
	)

	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "operation_duration_seconds",
			Help:      "Database operation duration distribution",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	DatabaseSlowQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "slow_queries_total",
			Help:      "Total number of slow database queries",
		},
		[]string{"table", "operation"},
	)

	DatabaseConnectionHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "connection_health_status",
			Help:      "Database connection health status (1=healthy, 0=unhealthy)",
		},
		[]string{"database", "instance"},
	)

	DatabaseTransactions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "transactions_total",
			Help:      "Total database transactions by result",
		},
		[]string{"result"}, // commit, rollback
	)

	DatabaseLockWaitTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "lock_wait_duration_seconds",
			Help:      "Time spent waiting for database locks",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"table", "lock_type"},
	)

	// Business-specific metrics
	DatabaseBusinessOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "business_operations_total",
			Help:      "Business-critical database operations",
		},
		[]string{"domain", "operation", "result"},
	)

	// Connection pool metrics (complementing GORM plugin)
	DatabaseConnectionPoolEfficiency = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "connection_pool_efficiency",
			Help:      "Database connection pool efficiency ratio (InUse/MaxOpen)",
		},
		[]string{"database", "instance"},
	)

	DatabaseConnectionWaitDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "database",
			Name:      "connection_wait_duration_seconds",
			Help:      "Time spent waiting for database connections",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"database", "instance"},
	)
)