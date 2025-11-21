package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/request"
	"github.com/dwarvesf/fortress-api/pkg/routes"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

// @title           FORTRESS API DOCUMENT
// @version         v0.1.39
// @description     This is api document for fortress project.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Nam Nguyen
// @contact.url    https://d.foundation
// @contact.email  benjamin@d.foundation

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger(cfg.LogLevel)
	log.Infof("Server starting with log level: %s", cfg.LogLevel)

	v, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
		// In production, Vault is required
		if cfg.Env != "local" {
			log.Fatal(err, "Vault initialization failed in production environment")
		}
	}

	if v != nil {
		cfg = config.Generate(v)
		// Re-initialize logger with Vault config
		log = logger.NewLogrusLogger(cfg.LogLevel)
		log.Debugf("Logger re-initialized with Vault config, log level: %s", cfg.LogLevel)
	}

	// Validate critical configuration
	if cfg.Invoice.TemplatePath == "" {
		log.Fatal(fmt.Errorf("INVOICE_TEMPLATE_PATH is empty"), "INVOICE_TEMPLATE_PATH is not configured - check Vault settings")
	}
	log.Infof("Config loaded - INVOICE_TEMPLATE_PATH: %s", cfg.Invoice.TemplatePath)

	s := store.New()
	repo := store.NewPostgresStore(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := service.New(cfg, s, repo)
	if err != nil {
		logger.L.Error(err, "failed to initialize service")
		os.Exit(1)
	}

	// Start parquet background sync service
	if svc.ParquetSync != nil {
		if err := svc.ParquetSync.StartBackgroundSync(ctx); err != nil {
			log.Warnf("Failed to start parquet sync service: %v", err)
		} else {
			log.Info("Parquet sync service started successfully")
		}
	}

	queue := make(chan model.WorkerMessage, 1000)
	w := worker.New(ctx, queue, svc, log)

	go func() {
		err := w.ProcessMessage()
		if err != nil {
			log.Error(err, "failed to process message")
		}
	}()

	router := routes.NewRoutes(cfg, svc, s, repo, w, log)
	request.RegisCustomValidators(router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ApiServer.Port),
		Handler: router,
	}

	// serve http server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err, "failed to listen and serve")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	cancel()

	shutdownServer(srv, log)
}

func shutdownServer(srv *http.Server, l logger.Logger) {
	l.Info("Server Shutting Down")
	if err := srv.Shutdown(context.Background()); err != nil {
		l.Error(err, "failed to shutdown server")
	}

	l.Info("Server Exit")
}
