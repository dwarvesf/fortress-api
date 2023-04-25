// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListDigests godoc
// @Summary Get list digests from DF Internal Digest
// @Description Get list digests from DF Internal Digest
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/digests [get]
func (h *handler) ListDigests(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Digest, nil, []notion.DatabaseQuerySort{
		{
			Property:  "Created at",
			Direction: notion.SortDirDesc,
		},
	}, 5)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get digests from notion"))
		return
	}

	var digests []model.NotionDigest

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		digests = append(digests, model.NotionDigest{
			ID:        r.ID,
			Name:      name,
			CreatedAt: props["Created at"].Date.Start.Time,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](digests, nil, nil, nil, "get list digests successfully"))
}
