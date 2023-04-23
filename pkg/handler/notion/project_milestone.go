package notion

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

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

	filter.And = append(filter.And, notion.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Select: &notion.SelectDatabaseQueryFilter{
				Equals: "Active",
			},
		},
	})

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.Project, filter, nil, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "failed to get projects from notion"))
		return
	}

	var projects []struct {
		Name       string                         `json:"name"`
		Milestones []model.NotionProjectMilestone `json:"milestones"`
	}

	var wg sync.WaitGroup
	var pmu sync.Mutex

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		if len(props["Project"].Title) == 0 || len(props["Parent item"].Relation) > 0 {
			continue
		}

		if c.Query("project_name") != "" {
			matched, err := regexp.MatchString(".*"+strings.ToLower(c.Query("project_name"))+".*", strings.ToLower(props["Project"].Title[0].Text.Content))
			if err != nil || !matched {
				continue
			}
		}

		var p = struct {
			Name       string                         `json:"name"`
			Milestones []model.NotionProjectMilestone `json:"milestones"`
		}{}
		var milestones []model.NotionProjectMilestone

		p.Name = props["Project"].Title[0].Text.Content

		wg.Add(1)
		go func() {
			defer wg.Done()

			workers := make(chan struct{}, 10) // limit to 10 workers

			var mmilestones = make(map[string]model.NotionProjectMilestone)
			for _, p := range props["Sub-item"].Relation {
				workers <- struct{}{}
				go func(p notion.Relation) {
					defer func() { <-workers }()
					if x, found := h.service.Cache.Get(p.ID); found {
						mmilestones[p.ID] = x.(model.NotionProjectMilestone)
						return
					}
					resp, err := h.service.Notion.GetPage(p.ID)
					if err != nil {
						return
					}
					props := resp.Properties.(notion.DatabasePageProperties)
					name := ""
					if len(props["Project"].Title) > 0 {
						name = props["Project"].Title[0].Text.Content
					}
					m := model.NotionProjectMilestone{
						ID:            resp.ID,
						Name:          name,
						SubMilestones: []*model.NotionProjectMilestone{},
					}
					if props["Milestone Date"].Date != nil {
						m.StartDate = props["Milestone Date"].Date.Start.Time
						m.EndDate = props["Milestone Date"].Date.End.Time
					}
					mmilestones[p.ID] = m
					h.service.Cache.Set(p.ID, m, 0)
				}(p)
			}
			for i := 0; i < cap(workers); i++ {
				workers <- struct{}{}
			}
			for _, m := range mmilestones {
				milestones = append(milestones, m)
			}
			sort.Slice(milestones[:], func(i, j int) bool {
				return milestones[i].StartDate.Before(milestones[j].StartDate)
			})

			p.Milestones = milestones

			pmu.Lock()
			projects = append(projects, p)
			pmu.Unlock()
		}()
	}
	wg.Wait()
	sort.Slice(projects[:], func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	c.JSON(http.StatusOK, view.CreateResponse[any](projects, nil, nil, nil, "get list milestones successfully"))
}

// func (h *handler) getMilestones(item *model.ProjectMilestone, subItems []*model.ProjectMilestone) []*model.ProjectMilestone {
// 	resp, err := h.service.Notion.GetPage(item.ID)
// 	if err != nil {
// 		return subItems
// 	}
// 	props := resp.Properties.(notion.DatabasePageProperties)
// 	for _, p := range props["Sub-item"].Relation {
// 		resp, err := h.service.Notion.GetPage(p.ID)
// 		if err != nil {
// 			continue
// 		}
// 		props := resp.Properties.(notion.DatabasePageProperties)
// 		name := ""
// 		if len(props["Project"].Title) > 0 {
// 			name = props["Project"].Title[0].Text.Content
// 		}
// 		m := &model.ProjectMilestone{
// 			ID:            resp.ID,
// 			Name:          name,
// 			SubMilestones: []*model.ProjectMilestone{},
// 		}
// 		if props["Milestone Date"].Date != nil {
// 			m.StartDate = props["Milestone Date"].Date.Start.Time
// 			m.EndDate = props["Milestone Date"].Date.End.Time
// 		}
// 		subItems = append(subItems, m)
// 	}

// 	return subItems
// }
