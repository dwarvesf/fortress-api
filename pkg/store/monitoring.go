package store

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

type QueryContext struct {
	StartTime      time.Time
	Operation      string
	TableName      string
	BusinessDomain string
}

var (
	// Cache for business domain lookups to avoid repeated string operations
	domainCache = make(map[string]string)
	domainMutex = sync.RWMutex{}
)

// registerCustomDatabaseCallbacks registers GORM callbacks for database monitoring
func registerCustomDatabaseCallbacks(db *gorm.DB, cfg *monitoring.DatabaseMonitoringConfig) {
	if !cfg.Enabled || !cfg.CustomMetrics {
		return
	}

	// Before callbacks - capture start time and context
	db.Callback().Create().Before("gorm:create").Register("metrics:before_create", beforeCallback("create"))
	db.Callback().Query().Before("gorm:query").Register("metrics:before_query", beforeCallback("select"))
	db.Callback().Update().Before("gorm:update").Register("metrics:before_update", beforeCallback("update"))
	db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", beforeCallback("delete"))

	// After callbacks - record metrics
	db.Callback().Create().After("gorm:create").Register("metrics:after_create", afterCallback("create", cfg))
	db.Callback().Query().After("gorm:query").Register("metrics:after_query", afterCallback("select", cfg))
	db.Callback().Update().After("gorm:update").Register("metrics:after_update", afterCallback("update", cfg))
	db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", afterCallback("delete", cfg))

	// Transaction tracking is handled in the repository layer via FinallyFunc

	// Start connection health monitoring if enabled
	if cfg.HealthCheckInterval > 0 {
		go startConnectionHealthMonitor(db, cfg)
	}
}

func beforeCallback(operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		// Minimal context - defer expensive operations to after callback
		ctx := &QueryContext{
			StartTime: time.Now(),
			Operation: operation,
		}
		db.Set("metrics:context", ctx)
	}
}

func afterCallback(operation string, cfg *monitoring.DatabaseMonitoringConfig) func(*gorm.DB) {
	return func(db *gorm.DB) {
		ctxValue, exists := db.Get("metrics:context")
		if !exists {
			return
		}

		ctx, ok := ctxValue.(*QueryContext)
		if !ok {
			return
		}

		duration := time.Since(ctx.StartTime)

		// Get table name only once when needed
		tableName := getTableName(db)
		
		// Determine result once
		result := "success"
		if db.Error != nil {
			result = "error"
		}

		// Record basic metrics (most critical) - use simplified labels
		metrics.DatabaseOperations.WithLabelValues(
			ctx.Operation, tableName, result,
		).Inc()

		metrics.DatabaseOperationDuration.WithLabelValues(
			ctx.Operation, tableName,
		).Observe(duration.Seconds())

		// Only check slow queries if enabled and if duration exceeds threshold
		if cfg.SlowQueryThreshold > 0 && duration > cfg.SlowQueryThreshold {
			metrics.DatabaseSlowQueries.WithLabelValues(
				tableName, ctx.Operation,
			).Inc()
		}

		// Business-specific metrics - only if explicitly enabled
		// Defer domain inference until needed and cache result
		if cfg.BusinessMetrics {
			domain := inferBusinessDomainCached(tableName)
			if domain != "" {
				metrics.DatabaseBusinessOperations.WithLabelValues(
					domain, ctx.Operation, result,
				).Inc()
			}
		}
	}
}


func getTableName(db *gorm.DB) string {
	if db.Statement != nil && db.Statement.Table != "" {
		return db.Statement.Table
	}

	if db.Statement != nil && db.Statement.Schema != nil {
		return db.Statement.Schema.Table
	}

	return "unknown"
}

// Static domain map for O(1) lookups (moved to package level for reuse)
var staticDomainMap = map[string]string{
	"employees":             "hr",
	"employee_roles":        "hr",
	"employee_positions":    "hr",
	"employee_chapters":     "hr",
	"employee_commissions":  "hr",
	"employee_invitations":  "hr",
	"employee_organizations": "hr",
	"employee_stacks":       "hr",
	"projects":              "project_management",
	"project_members":       "project_management",
	"project_heads":         "project_management",
	"project_stacks":        "project_management",
	"project_slots":         "project_management",
	"invoices":              "finance",
	"invoice_numbers":       "finance",
	"payrolls":              "finance",
	"cached_payrolls":       "finance",
	"base_salaries":         "finance",
	"salary_advances":       "finance",
	"employee_bonuses":      "finance",
	"accounting":            "finance",
	"clients":               "client_management",
	"client_contacts":       "client_management",
	"audits":                "compliance",
	"audit_cycles":          "compliance",
	"audit_items":           "compliance",
	"audit_participants":    "compliance",
	"permissions":           "security",
	"api_keys":              "security",
	"roles":                 "security",
	"banks":                 "finance",
	"bank_accounts":         "finance",
}

func inferBusinessDomain(db *gorm.DB) string {
	tableName := getTableName(db)
	return inferBusinessDomainCached(tableName)
}

// Optimized version that works with table name directly and uses cache
func inferBusinessDomainCached(tableName string) string {
	// Check cache first (read lock)
	domainMutex.RLock()
	if domain, exists := domainCache[tableName]; exists {
		domainMutex.RUnlock()
		return domain
	}
	domainMutex.RUnlock()

	// Direct map lookup (fastest path)
	if domain, exists := staticDomainMap[tableName]; exists {
		// Cache the result
		domainMutex.Lock()
		domainCache[tableName] = domain
		domainMutex.Unlock()
		return domain
	}

	// Optimized prefix matching using byte-level comparison
	var domain string
	if len(tableName) >= 8 && tableName[:8] == "employee" {
		domain = "hr"
	} else if len(tableName) >= 7 && tableName[:7] == "project" {
		domain = "project_management"
	} else if len(tableName) >= 7 && (tableName[:7] == "invoice" || tableName[:7] == "payroll") {
		domain = "finance"
	} else if len(tableName) >= 4 && tableName[:4] == "bank" {
		domain = "finance"
	} else if len(tableName) >= 10 && tableName[:10] == "accounting" {
		domain = "finance"
	} else if len(tableName) >= 5 && tableName[:5] == "audit" {
		domain = "compliance"
	} else if len(tableName) >= 6 && tableName[:6] == "client" {
		domain = "client_management"
	} else {
		domain = "" // No business domain mapping
	}

	// Cache the result
	domainMutex.Lock()
	domainCache[tableName] = domain
	domainMutex.Unlock()
	
	return domain
}

// startConnectionHealthMonitor monitors database connection health
func startConnectionHealthMonitor(db *gorm.DB, cfg *monitoring.DatabaseMonitoringConfig) {
	ticker := time.NewTicker(cfg.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkDatabaseHealth(db, "fortress", "primary")
		}
	}
}

// checkDatabaseHealth checks database connectivity and records health metrics
func checkDatabaseHealth(db *gorm.DB, dbName, instance string) {
	sqlDB, err := db.DB()
	if err != nil {
		metrics.DatabaseConnectionHealth.WithLabelValues(dbName, instance).Set(0)
		return
	}

	// Test basic connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sqlDB.PingContext(ctx)
	if err != nil {
		metrics.DatabaseConnectionHealth.WithLabelValues(dbName, instance).Set(0)
		return
	}

	// Connection is healthy
	metrics.DatabaseConnectionHealth.WithLabelValues(dbName, instance).Set(1)

	// Record connection pool metrics
	stats := sqlDB.Stats()
	recordConnectionPoolStats(stats, dbName, instance)
}

// recordConnectionPoolStats records connection pool efficiency metrics
func recordConnectionPoolStats(stats sql.DBStats, dbName, instance string) {
	// Connection pool efficiency
	if stats.MaxOpenConnections > 0 {
		efficiency := float64(stats.InUse) / float64(stats.MaxOpenConnections)
		metrics.DatabaseConnectionPoolEfficiency.WithLabelValues(dbName, instance).Set(efficiency)
	}

	// Average connection wait time
	if stats.WaitCount > 0 {
		avgWaitDuration := stats.WaitDuration / time.Duration(stats.WaitCount)
		metrics.DatabaseConnectionWaitDuration.WithLabelValues(dbName, instance).Observe(avgWaitDuration.Seconds())
	}
}