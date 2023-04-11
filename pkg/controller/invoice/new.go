package invoice

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type controller struct {
	store   *store.Store
	service *service.Service
	worker  *worker.Worker
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
		worker:  worker,
	}
}

type IController interface {
	UpdateStatus(in UpdateStatusInput) (*model.Invoice, error)
	MarkInvoiceAsError(invoice *model.Invoice) (*model.Invoice, error)
	MarkInvoiceAsPaid(invoice *model.Invoice, sendThankYouEmail bool) (*model.Invoice, error)
}
