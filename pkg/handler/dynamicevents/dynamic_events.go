package dynamicevents

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type handler struct {
	store      *store.Store
	controller *controller.Controller
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
	service    *service.Service
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, controller *controller.Controller, logger logger.Logger, cfg *config.Config, service *service.Service) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		controller: controller,
		logger:     logger,
		config:     cfg,
		service:    service,
	}
}
