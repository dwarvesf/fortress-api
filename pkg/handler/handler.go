package handler

import (
	employee_handler "github.com/dwarvesf/fortress-api/pkg/handler/employee"
	healthz_handler "github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Handler struct {
	Healthcheck healthz_handler.IHandler
	Employee    employee_handler.IHandler
}

func New(store *store.Store, service *service.Service, logger logger.Logger) (*Handler, error) {
	return &Handler{
		Healthcheck: healthz_handler.New(),
		Employee:    employee_handler.New(store, service, logger),
	}, nil
}
