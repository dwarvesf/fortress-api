// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

// ListEarns godoc
// @Summary Get list items from DF earn
// @Description Get list items from DF earn
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/earns [get]
func (h *handler) ListEarns(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Earn, nil, nil, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get items earn from notion"))
		return
	}

	var earns []model.NotionEarn
	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)
		if (props["Status"].Status == nil || props["Status"].Status.Name == "Done") || (props["Reward 🧊"].Number == nil || *props["Reward 🧊"].Number == 0) {
			continue
		}

		var tags []string
		for _, tag := range props["Tags"].MultiSelect {
			tags = append(tags, tag.Name)
		}
		var functions []string
		for _, f := range props["Function"].MultiSelect {
			functions = append(functions, f.Name)
		}
		var employees []model.Employee
		for _, e := range props["PICs"].People {
			employees = append(employees, model.Employee{
				FullName:      e.Name,
				PersonalEmail: e.Person.Email,
				Avatar:        e.AvatarURL,
			})
		}
		var dueData *time.Time
		if props["Due Date"].Date != nil {
			dueData = &props["Due Date"].Date.Start.Time
		}

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		earns = append(earns, model.NotionEarn{
			ID:       r.ID,
			Name:     name,
			Reward:   int(*props["Reward 🧊"].Number),
			Progress: int(*props["Progress"].Number * 100),
			Tags:     tags,
			PICs:     employees,
			Status:   props["Status"].Status.Name,
			Function: functions,
			DueDate:  dueData,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](earns, nil, nil, nil, "get list earn items successfully"))
}
