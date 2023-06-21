package brainerylogs

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/brainerylogs/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
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

// Create godoc
// @Summary Create brainery logs
// @Description Create brainery logs
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @PSuccess 200 {object} view.Create
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /brainery-logs [post]
func (h *handler) Create(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "brainerylogs",
			"method":  "Create",
		},
	)

	body := request.CreateBraineryLogRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}
	if err := body.Validate(); err != nil {
		l.Errorf(err, "failed to validate data", "body", body)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	publishedAt, _ := time.Parse(time.RFC3339Nano, body.PublishedAt)

	b := model.BraineryLog{
		Title:       body.Title,
		URL:         body.URL,
		GithubID:    body.GithubID,
		DiscordID:   body.DiscordID,
		Tags:        body.Tags,
		PublishedAt: &publishedAt,
		Reward:      body.Reward,
	}

	emp, err := h.store.Employee.GetByDiscordID(h.repo.DB(), body.DiscordID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get employee by discordID", "discordID", body.DiscordID)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		b.EmployeeID = emp.ID
	}

	_, err = h.store.BraineryLog.Create(h.repo.DB(), &b)
	if err != nil {
		l.Errorf(err, "failed to create brainery logs", "braineryLog", b)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any]("success", nil, nil, body, ""))
}
