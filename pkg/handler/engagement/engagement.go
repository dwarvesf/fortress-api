package engagement

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config

	// isIndexingMessages is used to make sure that there cannot be
	// a second IndexMessages invocation if the first one is not done
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
// @Security BearerAuth
// @Param Body body request.UpsertRollupRequest true "Body"
// @Success 200 {object} MessageResponse
// @Success 400 {object} ErrorResponse
// @Success 500 {object} ErrorResponse
// @Router /engagements/rollup [post]
func (h *handler) UpsertRollup(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "engagement",
			"method":  "UpsertRollup",
		},
	)
	body := request.UpsertRollupRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		l.Error(err, "error decoding body")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	l = l.AddField("body", body)
	if err := body.Validate(); err != nil {
		l.Error(err, "error validating data")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}

	discordUserID, err := strconv.ParseInt(body.DiscordUserID, 10, 64)
	if err != nil {
		l.Error(err, "unable to parse discordUserID to int64")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	lastMessageID, err := strconv.ParseInt(body.LastMessageID, 10, 64)
	if err != nil {
		l.Error(err, "unable to parse lastMessageID to int64")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	channelID, err := strconv.ParseInt(body.ChannelID, 10, 64)
	if err != nil {
		l.Error(err, "unable to parse channelID to int64")
		c.JSON(
			http.StatusBadRequest,
			view.CreateResponse[any](nil, nil, err, body, ""),
		)
		return
	}
	categoryID := int64(0)
	if body.CategoryID == "" {
		categoryID = -1
	} else {
		categoryID, err = strconv.ParseInt(body.CategoryID, 10, 64)
		if err != nil {
			l.Error(err, "unable to parse categoryID to int64")
			c.JSON(
				http.StatusBadRequest,
				view.CreateResponse[any](nil, nil, err, body, ""),
			)
			return
		}
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
	_, err = h.store.EngagementsRollup.Upsert(tx.DB(), rollup)
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
// @Security BearerAuth
// @Param channel-id path string true "Discord Channel ID"
// @Success 200 {object} MessageResponse
// @Success 400 {object} ErrorResponse
// @Success 500 {object} ErrorResponse
// @Router /engagements/channel/:channel-id/last-message-id [get]
func (h *handler) GetLastMessageID(c *gin.Context) {
	channelID := c.Param("channel-id")
	l := h.logger.Fields(logger.Fields{
		"handler":   "engagement",
		"method":    "UpsertRollupRecord",
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

func AggregateMessages(
	l logger.Logger,
	messages []*discordgo.Message,
	channelIDToCategoryID map[string]string,
) []*model.EngagementsRollup {
	userIDMessageIDToRecord := make(map[string]*model.EngagementsRollup)

	for _, message := range messages {
		l := l.AddField("messageID", message.ID)
		if message.Author == nil {
			l.Warn("missing author")
			continue
		}

		userID, err := strconv.ParseInt(message.Author.ID, 10, 64)
		if err != nil {
			l := l.AddField("userID", message.Author.ID)
			l.Error(err, "unable to parse user ID to int64")
			continue
		}
		messageID, err := strconv.ParseInt(message.ID, 10, 64)
		if err != nil {
			l := l.AddField("messageID", message.ID)
			l.Error(err, "unable to parse message ID to int64")
			continue
		}
		channelID, err := strconv.ParseInt(message.ChannelID, 10, 64)
		if err != nil {
			l := l.AddField("channelID", message.ChannelID)
			l.Error(err, "unable to parse channel ID to int64")
			continue
		}
		categoryID := int64(0)
		categoryIDStr := channelIDToCategoryID[message.ChannelID]
		if categoryIDStr == "" {
			categoryID = -1
		} else {
			categoryID, err = strconv.ParseInt(categoryIDStr, 10, 64)
			if err != nil {
				l := l.AddField("categoryIDStr", categoryIDStr)
				l.Error(err, "unable to parse category ID to int64")
				continue
			}
		}

		key := fmt.Sprintf("%d_%d", userID, channelID)
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
				CategoryID:      categoryID,
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
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Success 400 {object} ErrorResponse
// @Success 500 {object} ErrorResponse
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
			http.StatusAccepted,
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
	l.Debugf("fetched channels: %v", channels)

	tx, done := h.repo.NewTransaction()
	allMessages := make([]*discordgo.Message, 0)
	// TODO: parallelize the code as each channel can be processed singly
	for _, channel := range channels {
		l := l.AddField("channelID", channel.ID)
		if channel.LastMessageID == "" {
			l.Debugf("channel has no message", channel.ID)
			continue
		}

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
			l.Warnf("get messages after cursor error: %s", err.Error())
			continue
		}
		l.Debugf("fetched %d message(s)", len(messages))
		allMessages = append(allMessages, messages...)
	}

	channelIDToCategoryID := make(map[string]string, len(channels))
	for _, channel := range channels {
		channelIDToCategoryID[channel.ID] = channel.ParentID
	}

	records := AggregateMessages(l, allMessages, channelIDToCategoryID)
	l.Debugf("aggregated %d message(s) to %d records", len(allMessages), len(records))
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
