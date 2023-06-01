package engagement

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/engagement/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config

	// isIndexingMessages is used to make sure that there cannot be
	// a second AggregateMessages invocation if the first one is not done
	isIndexingMessages bool
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

// UpsertRollup godoc
// @Summary Upsert engagement rollup
// @Description Upsert engagement rollup
// @Tags Engagement
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body request.UpsertRollupRequest true "Body"
// @Success 200 {object} view.MessageResponse
// @Success 400 {object} view.ErrorResponse
// @Success 500 {object} view.ErrorResponse
// @Router /engagements/rollup [post]
func (h *handler) UpsertRollup(c *gin.Context) {

	body := request.UpsertRollupRequest{}
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

// GetLastMessageID godoc
// @Summary Get local last message ID of a channel
// @Description Get local last message ID of a channel
// @Tags Engagement
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param channel-id path string true "Discord Channel ID"
// @Success 200 {object} view.MessageResponse
// @Success 400 {object} view.ErrorResponse
// @Success 500 {object} view.ErrorResponse
// @Router /engagements/channel/:channel-id/last-message-id [get]
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

func AggregateMessages(l logger.Logger, messages []*discordgo.Message) []*model.EngagementsRollup {
	userIDMessageIDToRecord := make(map[string]*model.EngagementsRollup)

	for _, message := range messages {
		l := l.AddField("messageID", message.ID)
		if message.Author == nil {
			l.Warn("missing author")
			continue
		}

		userID, err := decimal.NewFromString(message.Author.ID)
		if err != nil {
			l := l.AddField("userID", message.Author.ID)
			l.Error(err, "unable to convert user ID to decimal")
			continue
		}
		messageID, err := decimal.NewFromString(message.ID)
		if err != nil {
			l := l.AddField("messageID", message.ID)
			l.Error(err, "unable to convert message ID to decimal")
			continue
		}
		channelID, err := decimal.NewFromString(message.ChannelID)
		if err != nil {
			l := l.AddField("channelID", message.ChannelID)
			l.Error(err, "unable to convert channel ID to decimal")
			continue
		}

		key := fmt.Sprintf("%s_%s", userID.String(), channelID.String())
		record, ok := userIDMessageIDToRecord[key]
		if ok {
			record.MessageCount += 1
			record.LastMessageID = messageID
		} else {
			userIDMessageIDToRecord[key] = &model.EngagementsRollup{
				DiscordUserID:   userID,
				LastMessageID:   messageID,
				DiscordUsername: fmt.Sprintf("%s#%s", message.Author.Username, message.Author.Discriminator),
				ChannelID:       channelID,
				MessageCount:    1,
			}
		}
	}

	records := make([]*model.EngagementsRollup, 0, len(userIDMessageIDToRecord))
	for _, record := range userIDMessageIDToRecord {
		records = append(records, record)
	}

	return records
}

// IndexMessages godoc
// @Summary Index messages of provided Discord server
// @Description Index messages of provided Discord server
// @Tags Engagement
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.MessageResponse
// @Success 400 {object} view.ErrorResponse
// @Success 500 {object} view.ErrorResponse
// @Router /cronjobs/index-engagement-messages [post]
func (h *handler) IndexMessages(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "engagement",
			"method":  "IndexMessages",
		},
	)

	if h.isIndexingMessages {
		l.Warn("handler is indexing messages")
		c.JSON(
			503,
			view.CreateResponse[any]("handler is indexing messages", nil, nil, nil, ""),
		)
		return
	}
	h.isIndexingMessages = true
	defer func() {
		h.isIndexingMessages = false
	}()

	l = h.logger.Fields(
		logger.Fields{
			"guildID": h.config.Discord.IDs.DwarvesGuild,
		},
	)
	channels, err := h.service.Discord.GetChannels()
	if err != nil {
		l.Error(err, "get channels error")
		c.JSON(
			http.StatusInternalServerError,
			view.CreateResponse[any](nil, nil, err, nil, ""),
		)
		return
	}

	tx, done := h.repo.NewTransaction()
	allMessages := make([]*discordgo.Message, 0)
	for _, channel := range channels {
		// TODO: parallelize the code as each channel can be processed singly
		if channel.LastMessageID == "" {
			continue
		}

		l := l.AddField("channelID", channel.ID)
		cursorMessageID, err := h.store.EngagementsRollup.GetLastMessageID(tx.DB(), channel.ID)
		if err != nil {
			l.Error(done(err), "get cursor message id error")
			c.JSON(
				http.StatusInternalServerError,
				view.CreateResponse[any](nil, nil, err, nil, ""),
			)
			return
		}
		messages, err := h.service.Discord.GetMessagesAfterCursor(
			channel.ID,
			cursorMessageID,
			channel.LastMessageID,
		)
		if err != nil {
			l := l.AddField("cursorMessageID", cursorMessageID)
			l.Error(err, "get messages after cursor error")
			c.JSON(
				http.StatusInternalServerError,
				view.CreateResponse[any](nil, nil, err, nil, ""),
			)
			return
		}
		allMessages = append(allMessages, messages...)
	}

	records := AggregateMessages(l, allMessages)
	for _, record := range records {
		_, err := h.store.EngagementsRollup.Upsert(tx.DB(), record)
		if err != nil {
			l.Error(done(err), "upsert record error")
			c.JSON(
				http.StatusInternalServerError,
				view.CreateResponse[any](nil, nil, err, nil, ""),
			)
			return
		}
		// wait 500ms after each insert to avoid overwhelming the database
		time.Sleep(500 * time.Millisecond)
	}

	c.JSON(
		http.StatusOK,
		view.CreateResponse[any]("success", nil, done(nil), nil, ""),
	)
}
