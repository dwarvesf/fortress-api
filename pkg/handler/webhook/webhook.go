package webhook

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type handler struct {
	store      *store.Store
	service    *service.Service
	controller *controller.Controller
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
	worker     *worker.Worker
}

// New returns a handler
func New(ctrl *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config, worker *worker.Worker) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
		worker:     worker,
		controller: ctrl,
	}
}
