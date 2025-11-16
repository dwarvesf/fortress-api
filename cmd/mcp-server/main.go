package main

import (
	"os"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/mcp/server"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig(config.DefaultConfigLoaders())

	// Initialize logger
	l := logger.NewLogrusLogger(cfg.LogLevel)

	v, err := vault.New(cfg)
	if err != nil {
		l.Error(err, "failed to init vault")
	}

	if v != nil {
		cfg = config.Generate(v)
	}

	// Initialize store with all sub-stores
	s := store.New()

	// Initialize database repository
	repo := store.NewPostgresStore(cfg)

	// Initialize services
	services, err := service.New(cfg, s, repo)
	if err != nil {
		l.Error(err, "Failed to initialize services")
		os.Exit(1)
	}

	// Create MCP server
	mcpServer := server.New(cfg, repo.DB(), l, s, repo, services)

	// Start server
	l.Info("Starting Fortress MCP Server...")

	if err := mcpServer.Serve(); err != nil {
		l.Error(err, "MCP server failed")
		os.Exit(1)
	}
}
