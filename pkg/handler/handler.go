package handler

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/auth"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee"
	"github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile"
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
	Profile     profile.IHandler
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) *Handler {
	return &Handler{
		Healthcheck: healthz.New(),
		Employee:    employee.New(store, repo, service, logger, cfg),
		Metadata:    metadata.New(store, repo, service, logger),
		Auth:        auth.New(store, repo, service, logger),
		Project:     project.New(store, repo, service, logger),
		Profile:     profile.New(store, repo, service, logger),
	}
}
