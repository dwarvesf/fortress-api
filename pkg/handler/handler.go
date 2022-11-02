package handler

import (
	healthz_handler "github.com/dwarvesf/fortress-api/pkg/handler/healthz"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Handler struct {
	Healthcheck healthz_handler.IHandler
}

func New(store *store.Store, service *service.Service) (*Handler, error) {
	return &Handler{
		Healthcheck: healthz_handler.New(),
	}, nil
}
