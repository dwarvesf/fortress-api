package main

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/job"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func main() {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger()
	log.Infof("Server starting")

	s := store.New()
	repo := store.NewPostgresStore(cfg)
	svc := service.New(cfg, s, repo)
	ctrl := controller.New(s, repo, svc, nil, log, cfg)

	if err := job.NewRevertIcy(ctrl, s, repo, svc, log, cfg).Run(); err != nil {
		log.Fatal(err, "failed to run job")
		return
	}

	log.Infof("done")
}
