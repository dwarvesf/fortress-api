// Package notion please edit this file only with approval from hnh
package notion

import (
	"net/http"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListIssues godoc
// @Summary Get list issues from DF Issues & Resolution Log
// @Description Get list issues from DF Issues & Resolution Log
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/issues [get]
func (h *handler) ListIssues(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Issue, nil, []notion.DatabaseQuerySort{
		{
			Property:  "Incident Date",
			Direction: notion.SortDirDesc,
		},
	}, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get issues from notion"))
		return
	}

	var issues []model.NotionIssue
	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)
		if props["Status"].Status == nil || props["Status"].Status.Name == "Done" {
			continue
		}

		source := ""
		if len(props["Source"].MultiSelect) > 0 {
			source = props["Source"].MultiSelect[0].Name
		}
		serverity := ""
		if props["Severity"].Select != nil {
			serverity = props["Severity"].Select.Name
		}
		scope := ""
		if len(props["Scope"].MultiSelect) > 0 {
			scope = props["Scope"].MultiSelect[0].Name
		}
		var projects []string
		if len(props["Project"].Relation) > 0 {
			for _, p := range props["Project"].Relation {
				projects = append(projects, p.ID)
			}
		}
		pic := ""
		if len(props["PIC"].People) > 0 {
			pic = props["PIC"].People[0].Name
		}
		priority := ""
		if props["Priority"].Select != nil {
			priority = props["Priority"].Select.Name
		}
		profile := ""
		if len(props["Profile"].Relation) > 0 {
			profile = props["Profile"].Relation[0].ID
		}
		resolution := ""
		if len(props["Resolution"].RichText) > 0 {
			resolution = props["Resolution"].RichText[0].Text.Content
		}
		var incidentDate time.Time
		if props["Incident Date"].Date != nil {
			incidentDate = props["Incident Date"].Date.Start.Time
		}
		var solvedDate time.Time
		if props["Solved Date"].Date != nil {
			solvedDate = props["Solved Date"].Date.Start.Time
		}
		rootCause := ""
		if len(props["Rootcause"].RichText) > 0 {
			rootCause = props["Rootcause"].RichText[0].Text.Content
		}

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		issues = append(issues, model.NotionIssue{
			ID:           r.ID,
			Name:         name,
			Status:       props["Status"].Status.Name,
			Source:       source,
			RootCause:    rootCause,
			IncidentDate: &incidentDate,
			SolvedDate:   &solvedDate,
			Severity:     serverity,
			Scope:        scope,
			Projects:     projects,
			PIC:          pic,
			Priority:     priority,
			Profile:      profile,
			Resolution:   resolution,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](issues, nil, nil, nil, "get list issues successfully"))
}
