// please edit this file only with approval from hnh
package hiring

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
	"github.com/thoas/go-funk"
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
// @Summary Get list hirings from DF Dwarves Hiring
// @Description Get list hirings from DF Dwarves Hiring
// @Tags hiring
// @Accept  json
// @Produce  json
// @Success 200 {object} view.HiringResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /hiring-positions [get]
func (h *handler) List(c *gin.Context) {
	filter := &notion.DatabaseQueryFilter{}

	status := "Active"
	if c.Query("status") != "" {
		if !funk.Contains([]string{"Active", "Inactive"}, c.Query("status")) {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, nil, nil, "status must be Active or Inactive"))
			return
		}
		status = c.Query("status")
	}
	filter.And = append(filter.And, notion.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: status,
			},
		},
	})

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.HiringDBID, nil, []notion.DatabaseQuerySort{
		{
			Direction: notion.SortDirDesc,
			Timestamp: notion.SortTimeStampCreatedTime,
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get hiring positions from notion"))
		return
	}

	var positions = []model.HiringPosition{}

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		if props["Status"].Select == nil || props["Status"].Select.Name != status {
			continue
		}

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		var projects []string
		if props["Project"].MultiSelect != nil {
			for _, p := range props["Project"].MultiSelect {
				projects = append(projects, p.Name)
			}
		}

		positions = append(positions, model.HiringPosition{
			ID:        r.ID,
			Name:      name,
			Status:    props["Status"].Select.Name,
			Projects:  projects,
			CreatedAt: r.CreatedTime,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](positions, nil, nil, nil, "get list hiring positions successfully"))
}
