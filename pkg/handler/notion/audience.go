// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListAudiences godoc
// @Summary Get list audiences from DF Audience
// @Description Get list audiences from DF Audience
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/audiences [get]
func (h *handler) ListAudiences(c *gin.Context) {
	filter := &notion.DatabaseQueryFilter{}

	filterNewSubscriber := true
	if c.Query("new_subscriber") == "false" {
		filterNewSubscriber = false
	}

	filter.And = append(filter.And, notion.DatabaseQueryFilter{
		Property: "New Subscriber",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Checkbox: &notion.CheckboxDatabaseQueryFilter{
				Equals: &filterNewSubscriber,
			},
		},
	})

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Audience, filter, []notion.DatabaseQuerySort{
		{
			Direction: notion.SortDirDesc,
			Timestamp: notion.SortTimeStampCreatedTime,
		},
	}, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get list audiences from notion"))
		return
	}

	var audiences []model.NotionAudience
	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		fullName := ""
		if len(props["Full Name"].Title) > 0 {
			fullName = props["Full Name"].Title[0].Text.Content
		}
		var sources []string
		for _, c := range props["Source"].MultiSelect {
			sources = append(sources, c.Name)
		}
		email := ""
		if props["Email"].Email != nil {
			email = *props["Email"].Email
		}

		audiences = append(audiences, model.NotionAudience{
			ID:        r.ID,
			FullName:  fullName,
			Email:     email,
			CreatedAt: r.CreatedTime,
			Sources:   sources,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](audiences, nil, nil, nil, "get list audiences successfully"))
}
