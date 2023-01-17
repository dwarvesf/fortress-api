// please edit this file only with approval from hnh
package earn

import (
	"net/http"
	"time"

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
// @Summary Get list items from DF earn
// @Description Get list items from DF earn
// @Tags earn
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.Earn
// @Failure 400 {object} view.ErrorResponse
func (h *handler) List(c *gin.Context) {
	resp, err := h.service.Notion.GetDatabase(h.config.Notion.EarnDBID, nil, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get items earn from notion"))
		return
	}

	var earns = []model.Earn{}
	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)
		if (props["Status"].Status == nil || props["Status"].Status.Name == "Done") || (props["Reward 🧊"].Number == nil || *props["Reward 🧊"].Number == 0) {
			continue
		}

		tags := []string{}
		for _, tag := range props["Tags"].MultiSelect {
			tags = append(tags, tag.Name)
		}
		functions := []string{}
		for _, f := range props["Function"].MultiSelect {
			functions = append(functions, f.Name)
		}
		employees := []model.Employee{}
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

		earns = append(earns, model.Earn{
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
