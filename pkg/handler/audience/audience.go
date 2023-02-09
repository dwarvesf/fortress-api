// please edit this file only with approval from hnh
package audience

import (
	"net/http"

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
// @Summary Get list audiences from DF Audience
// @Description Get list audiences from DF Audience
// @Tags Audience
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.Audience
// @Failure 400 {object} view.ErrorResponse
func (h *handler) List(c *gin.Context) {
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

	var audiences = []model.Audience{}
	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		fullname := ""
		if len(props["Full Name"].Title) > 0 {
			fullname = props["Full Name"].Title[0].Text.Content
		}
		sources := []string{}
		for _, c := range props["Source"].MultiSelect {
			sources = append(sources, c.Name)
		}
		email := ""
		if props["Email"].Email != nil {
			email = *props["Email"].Email
		}

		audiences = append(audiences, model.Audience{
			ID:        r.ID,
			FullName:  fullname,
			Email:     email,
			CreatedAt: r.CreatedTime,
			Sources:   sources,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](audiences, nil, nil, nil, "get list audiences successfully"))
}
