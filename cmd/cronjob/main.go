package main

import (
	_ "github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/cronjob"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func main() {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger()
	log.Infof("Cronjob starting")

	vault, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
	}

	if vault != nil {
		cfg = vault.LoadConfig()
	}

	svc := service.New(cfg)
	st := store.New()
	repo := store.NewPostgresStore(cfg)

	job := cronjob.New(st, repo, svc, log, cfg)
	if err := job.Run(); err != nil {
		log.Fatal(err, "error when running cronjob")
	}
}
