package discord

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/discord/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/discordevent"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/onleaverequest"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

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

const (
	discordReadingChannel           = "1225085624260759622"
	discordRandomChannel            = "788084358991970337"
	discordPlayGroundReadingChannel = "1119171172198797393"
)

func (h *handler) SyncDiscordInfo(c *gin.Context) {
	guildMembers, err := h.service.Discord.GetMembers()
	if err != nil {
		h.logger.Error(err, "failed to get guild members")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	for _, member := range guildMembers {
		if member.User.Bot {
			continue
		}

		communityProfile := model.DiscordAccount{
			DiscordID:       member.User.ID,
			PersonalEmail:   member.User.Email,
			DiscordUsername: member.User.Username,
			Roles:           member.Roles,         // currently an array of Discord role_id(s)
			MemoUsername:    member.User.Username, // default memo username is discord username
		}

		mochiPrf, err := h.service.MochiProfile.GetProfileByDiscordID(member.User.ID)
		if err == nil {
			for _, account := range mochiPrf.AssociatedAccounts {
				if account.Platform == mochiprofile.ProfilePlatformGithub {
					communityProfile.GithubUsername = fmt.Sprintf("%v", account.PlatformMetadata["username"])
				}
			}
		}

		_, err = h.store.DiscordAccount.Upsert(h.repo.DB(), &communityProfile)
		if err != nil {
			h.logger.Error(err, "failed to upsert discord account")
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// BirthdayDailyMessage check if today is birthday of any employee in the system
// if yes, send birthday message to employee through discord
func (h *handler) BirthdayDailyMessage(c *gin.Context) {
	// check app run mode
	projectID := consts.OperationID
	todoListID := consts.BirthdayToDoListID

	if h.config.Env != "prod" {
		projectID = consts.PlaygroundID
		todoListID = consts.PlaygroundBirthdayTodoID
	}

	//random message pool
	pool := []string{
		`Dear %s, we wish you courage and persistence in reaching all your greatest goals. Have a great birthday!`,
		`Happy Birthday to %s. No one knows your real age, except God, Human Resources and you yourself. Enjoy the blast!`,
		`Happy Birthday, %s! Thank you for being such a great team player and for giving us a perfect excuse to party on a weekday! Let's go grab a drink!`,
		`Just so you know you'd look much younger if not for working in this field :) Happy Birthday, %s`,
		`Congratulation on a great day! Here's to another year closer of retiring! Happy Birthday, %s!`,
		`%s, thank you for being a part of making this company more lively and cheerful. Wish you all the best in this special day.`,
		`Dear %s, we wish you a great birthday and a memorable year. From all the Dwarves brothers.`,
		`I can’t believe you are still single – lol. I hope you have a super day and get everything you want like a companion to share it with. Happy birthday to %s!`,
		`Here’s to another year of version controlling, bug reports, and comments about the documentation looking like code. Happy birthday, mate %s!`,
		`Hope your birthday loops run smoothly and that you don’t break out of the for loop too soon. Cheers, %s!`,
		`Happy birthday, %s. May your code works perfectly the first time you ran it.`,
		`I wish you could have a programming language that does not need compiling, installing, or debugging to run perfectly on the first run. Have a happy birthday, %s`,
	}

	// query active user
	filter := employee.EmployeeFilter{
		WorkingStatuses: []string{"full-time"},
	}

	employees, _, err := h.store.Employee.All(h.repo.DB(), filter, model.Pagination{
		Page: 0,
		Size: 1000,
	})
	if err != nil {
		h.logger.Error(err, "failed to get employees")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// format message if there is user's birthday
	var names string
	var birthDateNames []string
	todayDate := time.Now().Format("01/02")
	for _, e := range employees {
		now := time.Now()
		if now.Day() == e.DateOfBirth.Day() && now.Month() == e.DateOfBirth.Month() {
			if e.DiscordAccount != nil && e.DiscordAccount.DiscordID != "" {
				discordID := e.DiscordAccount.DiscordID
				names += fmt.Sprintf("<@%s>, ", discordID)
				birthDateNames = append(birthDateNames, e.FullName)
			}
		}
	}

	if len(birthDateNames) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, "no birthday today"))
		return
	}

	rand.New(rand.NewSource(time.Now().Unix()))
	msg := fmt.Sprintf(pool[rand.Intn(len(pool))], strings.TrimSuffix(names, ", "))

	//send message to Discord channel
	var discordMsg model.DiscordMessage
	discordMsg, err = h.service.Discord.PostBirthdayMsg(msg)
	if err != nil {
		h.logger.Error(err, "failed to post Discord message")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
		return
	}

	//Make Basecamp todos
	for _, birthDateName := range birthDateNames {
		birthDayTodo := bcModel.Todo{
			Title:   fmt.Sprintf("Prepare gift for %s, %s", birthDateName, todayDate),
			Content: fmt.Sprintf("Prepare gift for %s, %s", birthDateName, todayDate),
		}
		_, err := h.service.Basecamp.Todo.Create(projectID, todoListID, birthDayTodo)
		if err != nil {
			h.logger.Error(err, "failed create Basecamp todo")
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, birthDateName, "k"))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// OnLeaveMessage check if today is birthday of any employee in the system
// if yes, send birthday message to employee thru discord
func (h *handler) OnLeaveMessage(c *gin.Context) {
	todayDate := time.Now().Format("2006-01-02")
	onLeaveData, err := h.store.OnLeaveRequest.All(h.repo.DB(), onleaverequest.GetOnLeaveInput{Date: todayDate})
	if err != nil {
		h.logger.Error(err, "failed to get employees")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(onLeaveData) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, "there is no one on leave today"))
		return
	}

	var names string
	for _, e := range onLeaveData {
		names += fmt.Sprintf("<@%s>, ", e.Creator.DiscordAccount.DiscordID)
	}

	msg := fmt.Sprintf("Please be notified that %s will be absent today", strings.TrimSuffix(names, ", "))

	discordMsg, err := h.service.Discord.SendMessage(model.DiscordMessage{
		Content: msg,
	}, h.config.Discord.Webhooks.AuditLog)
	if err != nil {
		h.logger.Error(err, "failed to post Discord message")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
		return
	}

	h.logger.Infof("Discord message sent: %s", msg)
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// ReportBraineryMetrics reports brainery metrics to a channel
func (h *handler) ReportBraineryMetrics(c *gin.Context) {
	body := request.BraineryReportInput{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}
	if err := body.Validate(); err != nil {
		h.logger.Errorf(err, "failed to validate data", "body", body)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}
	now := time.Now()
	if body.View == "monthly" {
		now = now.Add(-24 * time.Hour)
	}

	latestPosts, logs, ncids, err := h.controller.BraineryLog.GetMetrics(now, body.View)
	if err != nil {
		h.logger.Error(err, "failed to get brainery metrics")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	metrics := view.ToBraineryMetric(latestPosts, logs, ncids, body.View)

	//send message to Discord channel
	var discordMsg *discordgo.Message
	discordMsg, err = h.service.Discord.ReportBraineryMetrics(body.View, &metrics, body.ChannelID)
	if err != nil {
		h.logger.Error(err, "failed to report brainery metrics discord message")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) DeliveryMetricsReport(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "discord",
			"method":  "DeliveryMetricsReport",
		},
	)

	in := request.DeliveryMetricReportInput{}
	if err := c.ShouldBindJSON(&in); err != nil {
		l.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	if err := in.Validate(); err != nil {
		l.Errorf(err, "failed to validate data", "body", in)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	if in.Sync {
		err := h.controller.DeliveryMetric.Sync()
		if err != nil {
			l.Errorf(err, "failed sync latest data", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}
	}

	if in.View == "weekly" {
		report, err := h.controller.DeliveryMetric.GetWeeklyReport()
		if err != nil {
			l.Errorf(err, "failed to get delivery metric weekly report", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}

		leaderBoard, err := h.controller.DeliveryMetric.GetWeeklyLeaderBoard()
		if err != nil {
			l.Errorf(err, "failed to get delivery metric weekly report", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}

		reportView := view.ToDeliveryMetricWeeklyReport(report)
		leaderBoardView := view.ToDeliveryMetricLeaderBoard(leaderBoard)

		discordMsg, err := h.service.Discord.DeliveryMetricWeeklyReport(reportView, leaderBoardView, in.ChannelID)
		if err != nil {
			h.logger.Error(err, "failed to post Discord message")
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
			return
		}
	}

	if in.View == "monthly" {
		report, err := h.controller.DeliveryMetric.GetMonthlyReport()
		if err != nil {
			l.Errorf(err, "failed to get delivery metric weekly report", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}

		currentMonthReport := report.Reports[0]
		previousMonthReport := report.Reports[1]

		if in.OnlyCompletedMonth {
			currentMonthReport = report.Reports[1]
			previousMonthReport = report.Reports[2]
		}

		leaderBoard, err := h.controller.DeliveryMetric.GetMonthlyLeaderBoard(currentMonthReport.Month)
		if err != nil {
			l.Errorf(err, "failed to get delivery metric weekly report", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}

		reportView := view.ToDeliveryMetricMonthlyReport(currentMonthReport, previousMonthReport)
		leaderBoardView := view.ToDeliveryMetricLeaderBoard(leaderBoard)

		discordMsg, err := h.service.Discord.DeliveryMetricMonthlyReport(reportView, leaderBoardView, in.ChannelID)
		if err != nil {
			h.logger.Error(err, "failed to post Discord message")
			c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, discordMsg, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// SyncMemo syncs memologs from the source memo.d.foundation
func (h *handler) SyncMemo(c *gin.Context) {
	targetChannelID := discordPlayGroundReadingChannel
	if h.config.Env == "prod" {
		targetChannelID = discordRandomChannel
	}

	memos, err := h.controller.MemoLog.Sync()
	if err != nil {
		h.logger.Error(err, "failed to sync memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(memos) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no new memo"))
		return
	}

	_, err = h.service.Discord.SendNewMemoMessage(h.config.Discord.IDs.DwarvesGuild, memos, targetChannelID)
	if err != nil {
		h.logger.Error(err, "failed to send new memo message")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) NotifyWeeklyMemos(c *gin.Context) {
	// get last 7 days
	end := time.Now()
	start := end.AddDate(0, 0, -7)

	var weekRangeStr string
	startDay := start.Day()
	endDay := end.Day()
	startMonth := strings.ToUpper(start.Month().String())
	endMonth := strings.ToUpper(end.Month().String())
	if startMonth == endMonth {
		weekRangeStr = fmt.Sprintf("%v - %v %v", startDay, endDay, startMonth)
	} else {
		weekRangeStr = fmt.Sprintf("%v %v - %v %v", startDay, startMonth, endDay, endMonth)
	}

	memos, err := h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
	if err != nil {
		h.logger.Error(err, "failed to retrieve weekly memos")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(memos) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no new memos in this week"))
		return
	}

	targetChannelID := discordPlayGroundReadingChannel
	if h.config.Env == "prod" {
		targetChannelID = discordRandomChannel
	}

	_, err = h.service.Discord.SendWeeklyMemosMessage(h.config.Discord.IDs.DwarvesGuild, memos, weekRangeStr, targetChannelID)
	if err != nil {
		h.logger.Error(err, "failed to send weekly memos report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// CreateScheduledEvent create new DF guild discord event
func (h *handler) CreateScheduledEvent(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "discord",
			"method":  "CreateScheduledEvent",
		},
	)

	in := request.DiscordEventInput{}
	if err := c.ShouldBindJSON(&in); err != nil {
		l.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	if err := in.Validate(); err != nil {
		l.Errorf(err, "failed to validate data", "body", in)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	// check if event already exists
	if _, err := h.store.DiscordEvent.
		One(h.repo.DB(),
			&discordevent.Query{DiscordEventID: in.ID}); !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "cannot find discord event", "body", in)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	evtType, err := in.EventType()
	if err != nil {
		l.Errorf(err, "failed to set event type", "body", in)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	// create event
	_, err = h.store.DiscordEvent.Create(h.repo.DB(), &model.Event{
		DiscordEventID:   in.ID,
		DiscordChannelID: in.DiscordChannelID,
		DiscordCreatorID: in.DiscordCreatorID,
		Name:             in.Name,
		Description:      in.Description,
		Date:             in.Date,
		EventType:        evtType,
	})
	if err != nil {
		h.logger.Error(err, "failed to create event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// ListScheduledEvent returns list of scheduled events
func (h *handler) ListScheduledEvent(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "discord",
			"method":  "ListScheduledEvent",
		},
	)

	var err error

	// get scheduled events from discord
	discordScheduledEvents, err := h.service.Discord.ListEvents()
	if err != nil {
		l.Error(err, "failed to get scheduled events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	eventIDs := make([]string, 0)
	for _, e := range discordScheduledEvents {
		eventIDs = append(eventIDs, e.ID)
	}

	// Get future events
	now := time.Now()
	events, err := h.store.DiscordEvent.All(h.repo.DB(), &discordevent.Query{
		DiscordEventIDs: eventIDs,
	}, false)
	if err != nil {
		l.Error(err, "failed to get events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	mapUpcomingEvents := make(map[string]bool)
	for _, e := range events {
		mapUpcomingEvents[e.DiscordEventID] = true
	}

	// Get completed events in the last 7 days, sometimes we need to update these events
	// If the event has date in the future, but cannot be found in discord, it means it has been done earlier
	after := now.AddDate(0, 0, -7)
	completedEvents, err := h.store.DiscordEvent.All(h.repo.DB(), &discordevent.Query{
		After: &after,
	}, false)
	if err != nil {
		l.Error(err, "failed to get completed events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	for i := range completedEvents {
		event := completedEvents[i]
		if _, ok := mapUpcomingEvents[event.DiscordEventID]; ok {
			continue
		}

		event.IsOver = true
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.After(events[j].Date)
	})

	c.JSON(http.StatusOK, view.CreateResponse[any](events, nil, nil, nil, "ok"))
}

// SetScheduledEventSpeakers sets speakers for a scheduled event
func (h *handler) SetScheduledEventSpeakers(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "discord",
			"method":  "SetScheduledEventSpeakers",
		},
	)

	in := []request.DiscordEventSpeakerInput{}
	if err := c.ShouldBindJSON(&in); err != nil {
		l.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	for _, i := range in {
		if err := i.Validate(); err != nil {
			l.Errorf(err, "failed to validate data", "body", in)
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}
	}

	// get event
	discordEventID := c.Param("id")
	event, err := h.store.DiscordEvent.One(h.repo.DB(), &discordevent.Query{DiscordEventID: discordEventID})
	if err != nil {
		l.Errorf(err, "failed to get event", "body", in)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	// delete all speakers
	if err = h.store.EventSpeaker.DeleteAllByEventID(h.repo.DB(), event.ID.String()); err != nil {
		l.Errorf(err, "failed to delete all speakers", "body", in)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	event.EventSpeakers = make([]model.EventSpeaker, 0)

	// get speakers
	for _, i := range in {
		speaker, err := h.store.DiscordAccount.OneByDiscordID(h.repo.DB(), i.ID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			discordUser, err := h.service.Discord.GetMember(i.ID)
			if err != nil {
				l.Errorf(err, "failed to get discord user", "body", in)
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
				return
			}

			speaker = &model.DiscordAccount{
				DiscordID:       i.ID,
				DiscordUsername: discordUser.User.Username,
			}

			_, err = h.store.DiscordAccount.Upsert(h.repo.DB(), speaker)
			if err != nil {
				l.Errorf(err, "failed to upsert speaker", "body", in)
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
				return
			}
		} else if err != nil {
			l.Errorf(err, "failed to get speaker", "body", in)
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
			return
		}

		event.EventSpeakers = append(event.EventSpeakers, model.EventSpeaker{
			EventID:          event.ID,
			DiscordAccountID: speaker.ID,
			Topic:            i.Topic,
		})
	}

	// set speakers
	err = h.store.DiscordEvent.SetSpeakers(h.repo.DB(), event)
	if err != nil {
		l.Errorf(err, "failed to set speakers", "body", in)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToDiscordEvent(*event), nil, nil, nil, ""))
}

// PostGolangNews fetches Golang news from social platform and post to golang channel
func (h *handler) PostGolangNews(c *gin.Context) {
	if err := h.controller.Reddit.SyncGolangNews(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// ListDiscordResearchTopics godoc
// @Summary Get list of research topics on discord
// @Description Get list of research topics on discord
// @id ListDiscordResearchTopics
// @Tags Discord
// @Accept  json
// @Produce  json
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} ListResearchTopicResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /discords/research-topics [get]
func (h *handler) ListDiscordResearchTopics(c *gin.Context) {
	var query model.Pagination
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}
	query.Standardize()

	limit, offset := query.ToLimitOffset()

	topics, total, err := h.controller.Discord.ListDiscordResearchTopics(context.Background(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToListResearchTopicResponse(topics),
		&view.PaginationResponse{
			Pagination: view.Pagination{
				Page: query.Page,
				Size: query.Size,
			},
			Total: total,
		}, nil, nil, ""))
}
