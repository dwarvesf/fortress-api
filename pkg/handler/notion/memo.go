// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

// ListMemos godoc
// @Summary Get list memos from DF Memos
// @Description Get list memos from DF Memos
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Router /notion/memos [get]
func (h *handler) ListMemos(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Memo, nil, []notion.DatabaseQuerySort{
		{
			Property:  "Created",
			Direction: notion.SortDirDesc,
		},
	}, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get memos from notion"))
		return
	}

	var memos []model.NotionMemo

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		author := ""
		if len(props["Author"].People) > 0 {
			author = props["Author"].People[0].Name
		}

		var tags []string
		if len(props["Tags"].MultiSelect) > 0 {
			for _, t := range props["Tags"].MultiSelect {
				tags = append(tags, t.Name)
			}
		}

		memos = append(memos, model.NotionMemo{
			ID:        r.ID,
			Name:      name,
			CreatedAt: props["Created"].Date.Start.Time,
			Tags:      tags,
			Author:    author,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](memos, nil, nil, nil, "get list memos successfully"))
}
