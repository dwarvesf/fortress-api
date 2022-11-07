package handler

import (
	auth_handler "github.com/dwarvesf/fortress-api/pkg/handler/auth"
	employee_handler "github.com/dwarvesf/fortress-api/pkg/handler/employee"
	healthz_handler "github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	metadata_handler "github.com/dwarvesf/fortress-api/pkg/handler/metadata"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Handler struct {
	Healthcheck healthz_handler.IHandler
	Employee    employee_handler.IHandler
	Metadata    metadata_handler.IHandler
	Auth        auth_handler.IHandler
}

func New(store *store.Store, service *service.Service, logger logger.Logger) *Handler {
	return &Handler{
		Healthcheck: healthz_handler.New(),
		Employee:    employee_handler.New(store, service, logger),
		Metadata:    metadata_handler.New(store, service, logger),
		Auth:        auth_handler.New(store, service, logger),
	}
}
