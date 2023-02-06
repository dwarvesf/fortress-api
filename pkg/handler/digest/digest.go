// please edit this file only with approval from hnh
package digest

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
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

// List godoc
// @Summary Get list digests from DF Internal Digest
// @Description Get list digests from DF Internal Digest
// @Tags digests
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.Digest
// @Failure 400 {object} view.ErrorResponse
func (h *handler) List(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.DigestDBID, nil, []notion.DatabaseQuerySort{
		{
			Property:  "Created at",
			Direction: notion.SortDirDesc,
		},
	}, 5)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get digests from notion"))
		return
	}

	var digests = []model.Digest{}

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		digests = append(digests, model.Digest{
			ID:        r.ID,
			Name:      name,
			CreatedAt: props["Created at"].Date.Start.Time,
		})

	}

	c.JSON(http.StatusOK, view.CreateResponse[any](digests, nil, nil, nil, "get list digests successfully"))
}
