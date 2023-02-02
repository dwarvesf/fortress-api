// please edit this file only with approval from hnh
package audit

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/cronjob/audit"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"

	"github.com/gin-gonic/gin"
)

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

// Sync godoc
// @Summary Sync audit info from Notion to database
// @Description Sync audit info from Notion to database
// @Tags Audit
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.MessageResponse
// @Router /audits [put]
func (h *handler) Sync(c *gin.Context) {
	go func() {
		auditSync := audit.New(
			h.store, h.repo, h.service, h.logger, h.config,
		)

		auditSync.SyncAuditCycle()
	}()

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "sync audit from notion successfully"))
}
