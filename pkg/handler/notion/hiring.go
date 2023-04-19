// Package notion Package hiring please edit this file only with approval from hnh
package notion

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListHiringPositions godoc
// @Summary Get list hiring from DF Dwarves Hiring
// @Description Get list hiring from DF Dwarves Hiring
// @Tags hiring
// @Accept  json
// @Produce  json
// @Success 200 {object} view.HiringResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /hiring-positions [get]
func (h *handler) ListHiringPositions(c *gin.Context) {
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

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Hiring, nil, []notion.DatabaseQuerySort{
		{
			Direction: notion.SortDirDesc,
			Timestamp: notion.SortTimeStampCreatedTime,
		},
	}, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get hiring positions from notion"))
		return
	}

	var positions []model.NotionHiringPosition

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

		positions = append(positions, model.NotionHiringPosition{
			ID:        r.ID,
			Name:      name,
			Status:    props["Status"].Select.Name,
			Projects:  projects,
			CreatedAt: r.CreatedTime,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](positions, nil, nil, nil, "get list hiring positions successfully"))
}
