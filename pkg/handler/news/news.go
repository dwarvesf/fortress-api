package news

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type handler struct {
	store      *store.Store
	controller *controller.Controller
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, controller *controller.Controller, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		controller: controller,
		logger:     logger,
		config:     cfg,
	}
}
