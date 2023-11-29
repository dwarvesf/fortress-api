// Package payroll please edit this file only with approval from hnh
package payroll

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

type IHandler interface {
	GetPayrollsByMonth(c *gin.Context)
	GetPayrollsBHXH(c *gin.Context)
	CommitPayroll(c *gin.Context)
	MarkPayrollAsPaid(c *gin.Context)
}

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	worker     *worker.Worker
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, worker *worker.Worker, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		worker:     worker,
		logger:     logger,
		config:     cfg,
	}
}
