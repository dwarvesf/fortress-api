package notion

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

// ListProjectMilestones godoc
// @Summary Get list  project milestones
// @Description Get list  project milestones
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/projects/milestones [get]
func (h *handler) ListProjectMilestones(c *gin.Context) {
	filter := &notion.DatabaseQueryFilter{}

	filter.And = append(filter.And, []notion.DatabaseQueryFilter{{
		Property: "Type",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: "Project",
			},
		},
	},
		{
			Property: "Status",
			DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
				Select: &notion.SelectDatabaseQueryFilter{
					DoesNotEqual: "Done",
				},
			},
		},
	}...)

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Project, filter, nil, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "failed to get projects from notion"))
		return
	}

	var projects []struct {
		Name       string                         `json:"name"`
		Milestones []model.NotionProjectMilestone `json:"milestones"`
	}
	var milestones []model.NotionProjectMilestone

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		if props["Project"].Select == nil || (props["Project"].Select != nil && props["Project"].Select.Name == "") {
			continue
		}

		if c.Query("project_name") != "" {
			matched, err := regexp.MatchString(".*"+strings.ToLower(c.Query("project_name"))+".*", strings.ToLower(props["Project"].Select.Name))
			if err != nil || !matched {
				continue
			}
		}

		m := model.NotionProjectMilestone{
			ID:      r.ID,
			Name:    props["Scope"].Title[0].PlainText,
			Project: props["Project"].Select.Name,
		}
		if props["Date"].Date != nil {
			m.StartDate = props["Date"].Date.Start.Time
			if props["Date"].Date.End != nil {
				m.EndDate = props["Date"].Date.End.Time
			}
		}
		milestones = append(milestones, m)
	}

	// group milestones by project name
	for _, m := range milestones {
		found := false
		for i := range projects {
			if projects[i].Name == m.Project {
				projects[i].Milestones = append(projects[i].Milestones, m)
				found = true
				break
			}
		}
		if !found {
			p := struct {
				Name       string                         `json:"name"`
				Milestones []model.NotionProjectMilestone `json:"milestones"`
			}{
				Name:       m.Project,
				Milestones: []model.NotionProjectMilestone{m},
			}
			projects = append(projects, p)
		}
	}

	for i := range projects {
		sort.Slice(projects[i].Milestones[:], func(j, k int) bool {
			return projects[i].Milestones[j].StartDate.Before(projects[i].Milestones[k].StartDate)
		})
	}

	// pp.Println(milestones)
	sort.Slice(projects[:], func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	c.JSON(http.StatusOK, view.CreateResponse[any](projects, nil, nil, nil, "get list milestones successfully"))
}
