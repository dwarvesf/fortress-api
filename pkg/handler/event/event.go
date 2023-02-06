// please edit this file only with approval from hnh
package event

import (
	"net/http"
	"strconv"
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
// @Summary Get list events from DF Dwarves Community Events
// @Description Get list events from DF Dwarves Community Events
// @Tags events
// @Accept  json
// @Produce  json
// @Success 200 {object} []model.Event
// @Failure 400 {object} view.ErrorResponse
func (h *handler) List(c *gin.Context) {
	filter := &notion.DatabaseQueryFilter{}

	nextDays := 7
	if c.Query("d") != "" {
		d, ok := c.GetQuery("d")
		if !ok {
			d = "7"
		}
		var err error
		nextDays, err = strconv.Atoi(d)
		if err != nil {
			nextDays = 7
		}
	}

	from := time.Now()
	to := from.Add(24 * time.Hour * time.Duration(nextDays))
	filter.And = append(filter.And, notion.DatabaseQueryFilter{
		Property: "Date",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Date: &notion.DatePropertyFilter{
				OnOrAfter: &from,
			},
		},
	})
	filter.And = append(filter.And, notion.DatabaseQueryFilter{
		Property: "Date",
		DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
			Date: &notion.DatePropertyFilter{
				OnOrBefore: &to,
			},
		},
	})

	resp, err := h.service.Notion.GetDatabase(h.config.Notion.EventDBID, filter, []notion.DatabaseQuerySort{
		{
			Property:  "Date",
			Direction: notion.SortDirAsc,
		},
	}, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "can't get events from notion"))
		return
	}

	var events = []model.Event{}

	for _, r := range resp.Results {
		props := r.Properties.(notion.DatabasePageProperties)

		name := props["Name"].Title[0].Text.Content
		if r.Icon != nil && r.Icon.Emoji != nil {
			name = *r.Icon.Emoji + " " + props["Name"].Title[0].Text.Content
		}

		activityType := ""
		if props["Activity Type"].Select != nil {
			activityType = props["Activity Type"].Select.Name
		}

		var date model.DateTime
		if props["Date"].Date != nil {
			date.Time = props["Date"].Date.Start.Time
			date.HasTime = props["Date"].Date.Start.HasTime()
		}

		events = append(events, model.Event{
			ID:           r.ID,
			Name:         name,
			ActivityType: activityType,
			Date:         date,
			CreatedAt:    r.CreatedTime,
		})
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](events, nil, nil, nil, "get list events successfully"))
}
