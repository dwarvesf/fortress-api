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
	log := logger.NewLogrusLogger()
	log.Infof("Server starting")

	v, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
	}

	if v != nil {
		cfg = config.Generate(v)
	}

	s := store.New()
	repo := store.NewPostgresStore(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	svc, err := service.New(cfg, s, repo)
	if err != nil {
		logger.L.Error(err, "failed to initialize service")
		os.Exit(1)
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
