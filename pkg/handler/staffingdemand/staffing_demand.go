// please edit this file only with approval from hnh
package staffingdemand

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
// @Summary Get list  staffing demands from DF Staffing Demand
// @Description Get list  staffing demands from DF Staffing Demand
// @Tags staffing-demands
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.StaffingDemand
// @Failure 400 {object} view.ErrorResponse
func (h *handler) List(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.StaffingDemand, nil, nil, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get staffing demands from notion"))
		return
	}

	var staffingDemands = []model.StaffingDemand{}

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		name := props["Project Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		request := ""
		if len(props["Request"].RichText) > 0 {
			request = props["Request"].RichText[0].Text.Content
		}

		staffingDemands = append(staffingDemands, model.StaffingDemand{
			ID:      r.ID,
			Name:    name,
			Request: request,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](staffingDemands, nil, nil, nil, "get list staffing demands successfully"))
}
