package memologs

import (
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/memologs/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

func (h *handler) Create(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Create",
		},
	)

	body := request.CreateMemoLogsRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "[memologs.Create] failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	memologs := make([]model.MemoLog, 0)
	for _, b := range body {
		publishedAt, _ := time.Parse(time.RFC3339Nano, b.PublishedAt)

		// TODO: enrich author info and add to `community_members` if not exist

		b := model.MemoLog{
			Title: b.Title,
			URL:   b.URL,
			// Authors:     authors, // TODO: change to new authors with community member
			Tags:        b.Tags,
			PublishedAt: &publishedAt,
			Description: b.Description,
			Reward:      b.Reward,
		}
		memologs = append(memologs, b)
	}

	logs, err := h.store.MemoLog.Create(h.repo.DB(), memologs)
	if err != nil {
		l.Errorf(err, "[memologs.Create] failed to create new memo logs", "memologs", memologs)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, memologs, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(logs), nil, nil, body, ""))
}

func (h *handler) List(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "List",
		},
	)

	memoLogs, err := h.store.MemoLog.List(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(memoLogs), nil, nil, nil, ""))
}

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "memologs",
			"method":  "Sync",
		},
	)

	results, err := h.controller.MemoLog.Sync()
	if err != nil {
		l.Error(err, "failed to sync memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToMemoLog(results), nil, nil, nil, "ok"))
}
