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

	vault, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
	}

	if vault != nil {
		cfg = config.Generate(vault)
	}

	svc := service.New(cfg)
	st := store.New()
	repo := store.NewPostgresStore(cfg)

	cronjob.NewSyncAuditJob(st, repo, svc, log, cfg).Run()
}
