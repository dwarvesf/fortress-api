package engagement

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

func New(
	controller *controller.Controller,
	store *store.Store,
	repo store.DBRepo,
	service *service.Service,
	logger logger.Logger,
	cfg *config.Config,
) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

func (h *handler) UpsertRollup(c *gin.Context) {

	body := UpsertRollupRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "engagement",
		"method":  "UpsertRollup",
		"body":    body,
	})

	discordUserID, err := decimal.NewFromString(body.DiscordUserID)
	if err != nil {
		l.Error(err, "unable to convert discordUserID to decimal")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	lastMessageID, err := decimal.NewFromString(body.LastMessageID)
	if err != nil {
		l.Error(err, "unable to convert lastMessageID to decimal")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	channelID, err := decimal.NewFromString(body.ChannelID)
	if err != nil {
		l.Error(err, "unable to convert channelID to decimal")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	categoryID, err := decimal.NewFromString(body.CategoryID)
	if err != nil {
		l.Error(err, "unable to convert categoryID to decimal")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}

	tx, done := h.repo.NewTransaction()
	rollup := &model.EngagementsRollup{
		DiscordUserID: discordUserID,
		LastMessageID: lastMessageID,
		ChannelID:     channelID,
		CategoryID:    categoryID,
		MessageCount:  body.MessageCount,
		ReactionCount: body.ReactionCount,
	}
	rollup, err = h.store.EngagementsRollup.Upsert(tx.DB(), rollup)
	if err != nil {
		l.Error(err, "unable to upsert engagements rollup")
		c.JSON(
			http.StatusInternalServerError,
			view.CreateResponse[any](nil, nil, done(err), body, ""),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		view.CreateResponse[any]("success", nil, done(nil), body, ""),
	)
}

func (h *handler) GetLastMessageID(c *gin.Context) {
	channelID := c.Param("channel-id")
	l := h.logger.Fields(logger.Fields{
		"handler":   "engagement",
		"method":    "GetLastMessageID",
		"channelID": channelID,
	})

	tx, done := h.repo.NewTransaction()
	lastMessageID, err := h.store.EngagementsRollup.GetLastMessageID(tx.DB(), channelID)
	if err != nil {
		l.Error(err, "unable to get last message ID")
		c.JSON(
			http.StatusInternalServerError,
			view.CreateResponse[any](nil, nil, done(err), channelID, ""),
		)
		return
	}

	c.JSON(
		http.StatusOK,
		view.CreateResponse[any](lastMessageID, nil, done(nil), nil, "success"),
	)
}
