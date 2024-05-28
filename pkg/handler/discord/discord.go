package discord

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/onleaverequest"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
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
	start := timeutil.GetStartDayOfWeek(time.Now())
	end := timeutil.GetEndDayOfWeek(time.Now())
	var weekRangeStr string

	// parse week range to string format
	// eg. 13 - 17 APR or 27 APR - 2 MAY
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
