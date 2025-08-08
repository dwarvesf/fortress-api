package store

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

// SetupDatabaseMonitoring configures database monitoring plugins and callbacks
func SetupDatabaseMonitoring(db *gorm.DB, cfg *monitoring.DatabaseMonitoringConfig, env string) error {
	if !cfg.Enabled {
		logger.L.Info("Database monitoring disabled")
		return nil
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid database monitoring configuration: %w", err)
	}

	logger.L.Info("Setting up database monitoring")

	// Setup official GORM Prometheus plugin
	if err := setupGORMPrometheusPlugin(db, cfg, env); err != nil {
		logger.L.Warnf("Failed to setup GORM Prometheus plugin: %v", err)
		// Don't fail startup - continue with custom metrics only
	}

	// Setup custom callbacks for business metrics
	if cfg.CustomMetrics {
		registerCustomDatabaseCallbacks(db, cfg)
		logger.L.Info("Database custom metrics callbacks registered")
	}

	logger.L.Info("Database monitoring setup completed")
	return nil
}

func setupGORMPrometheusPlugin(db *gorm.DB, cfg *monitoring.DatabaseMonitoringConfig, env string) error {
	// Configure GORM Prometheus plugin
	pluginConfig := prometheus.Config{
		DBName:          "fortress",
		RefreshInterval: uint32(cfg.RefreshInterval.Seconds()),
		Labels: map[string]string{
			"service":     "fortress-api",
			"environment": env,
			"database":    "fortress",
		},
		// Use PostgreSQL-specific metrics collector
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.Postgres{},
		},
	}

	// Register the plugin with GORM
	err := db.Use(prometheus.New(pluginConfig))
	if err != nil {
		return fmt.Errorf("failed to register GORM prometheus plugin: %w", err)
	}

	logger.L.Info("GORM Prometheus plugin registered successfully")
	return nil
}

// EnableDatabaseMonitoring is a helper function to easily enable monitoring on existing DB
func EnableDatabaseMonitoring(db *gorm.DB, env string) error {
	config := monitoring.DefaultDatabaseConfig()
	
	// Disable monitoring in test environment to avoid conflicts
	if env == "test" {
		config.Enabled = false
		return nil
	}

	return SetupDatabaseMonitoring(db, config, env)
}