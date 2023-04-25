// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListStaffingDemands godoc
// @Summary Get list  staffing demands from DF Staffing Demand
// @Description Get list  staffing demands from DF Staffing Demand
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/staffing-demands [get]
func (h *handler) ListStaffingDemands(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.StaffingDemand, nil, nil, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "failed to get staffing demands from notion"))
		return
	}

	var staffingDemands []model.NotionStaffingDemand

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

		staffingDemands = append(staffingDemands, model.NotionStaffingDemand{
			ID:      r.ID,
			Name:    name,
			Request: request,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](staffingDemands, nil, nil, nil, "get list staffing demands successfully"))
}
