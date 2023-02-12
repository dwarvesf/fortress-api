package controller

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller/auth"
	"github.com/dwarvesf/fortress-api/pkg/controller/client"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Controller struct {
	Auth   auth.IController
	Client client.IController
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) *Controller {
	return &Controller{
		Auth:   auth.New(store, repo, service, logger, cfg),
		Client: client.New(store, repo, service, logger, cfg),
	}
}
