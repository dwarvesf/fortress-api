package handler

import (
	"github.com/dwarvesf/fortress-api/pkg/handler/auth"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee"
	"github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata"
	"github.com/dwarvesf/fortress-api/pkg/handler/project"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Handler struct {
	Healthcheck healthz.IHandler
	Employee    employee.IHandler
	Metadata    metadata.IHandler
	Auth        auth.IHandler
	Project     project.IHandler
}

func New(store *store.Store, service *service.Service, logger logger.Logger) *Handler {
	return &Handler{
		Healthcheck: healthz.New(),
		Employee:    employee.New(store, service, logger),
		Metadata:    metadata.New(store, service, logger),
		Auth:        auth.New(store, service, logger),
		Project:     project.New(store, service, logger),
	}
}
