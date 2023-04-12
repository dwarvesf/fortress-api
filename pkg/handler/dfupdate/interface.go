package dfupdate

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type From struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type ProjectChangelog struct {
	ProjectPageID string `json:"project_page_id,omitempty"`
	IsPreview     bool   `json:"is_preview"`
	From          From   `json:"from,omitempty"`
}

// IHandler functions of dfupdate handler interface
type IHandler interface {
	Send(c *gin.Context)
}

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}
