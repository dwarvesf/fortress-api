package discord

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
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
	"github.com/dwarvesf/fortress-api/pkg/service/discord/helpers"
	"github.com/dwarvesf/fortress-api/pkg/service/duckdb"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/discordevent"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/onleaverequest"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller      *controller.Controller
	store           *store.Store
	service         *service.Service
	logger          logger.Logger
	repo            store.DBRepo
	config          *config.Config
	dataTransformer helpers.DataTransformer
}

func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	// Initialize data transformer for parquet data processing
	dataTransformer := helpers.NewDataTransformer(helpers.DataTransformationConfig{
		DateFormats: []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
		},
		DefaultReward:           "25",
		DefaultCategory:         []string{"others"},
		EnableValidation:        false, // Skip validation for performance
		SkipInvalidRecords:      true,  // Skip invalid records during batch processing
		MaxContentLength:        10000,
		AuthorResolutionRetries: 3,
	})

	return &handler{
		controller:      controller,
		store:           store,
		repo:            repo,
		service:         service,
		logger:          logger,
		config:          cfg,
		dataTransformer: dataTransformer,
	}
}

const (
	discordReadingChannel           = "1225085624260759622"
	discordRandomChannel            = "788084358991970337"
	discordPlayGroundReadingChannel = "1064460652720160808" // quang's channel
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

	_, err = h.service.Discord.SendNewMemoMessage(
		h.config.Discord.IDs.DwarvesGuild,
		memos,
		targetChannelID,
		func(discordAccountID string) (*model.DiscordAccount, error) {
			return h.store.DiscordAccount.One(h.repo.DB(), discordAccountID)
		},
	)
	if err != nil {
		h.logger.Error(err, "failed to send new memo message")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// SweepMemo sweeps memologs
func (h *handler) SweepMemo(c *gin.Context) {
	err := h.controller.MemoLog.Sweep()
	if err != nil {
		h.logger.Error(err, "failed to sweep memologs")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
func (h *handler) NotifyTopMemoAuthors(c *gin.Context) {
	in := request.TopMemoAuthorsInput{}
	if err := c.ShouldBindJSON(&in); err != nil {
		h.logger.Error(err, "failed to decode body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	if err := in.Validate(); err != nil {
		h.logger.Error(err, "failed to validate input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, in, ""))
		return
	}

	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -in.Days+1)

	topAuthors, err := h.store.MemoLog.GetTopAuthors(h.repo.DB(), in.Limit, &start, &end)
	if err != nil {
		h.logger.Error(err, "failed to retrieve top memo authors")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if len(topAuthors) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no memo authors found"))
		return
	}

	var topAuthorsStr string
	for i, author := range topAuthors {
		topAuthorsStr += fmt.Sprintf("%d. <@%s> (%d memos)\n", i+1, author.DiscordID, author.TotalMemos)
	}

	targetChannelID := discordPlayGroundReadingChannel
	if h.config.Env == "prod" {
		targetChannelID = discordRandomChannel
	}

	title := fmt.Sprintf("Top %d Memo Authors (Last %d Days)", in.Limit, in.Days)
	msg := &discordgo.MessageEmbed{
		Title:       title,
		Description: topAuthorsStr,
	}

	_, err = h.service.Discord.SendEmbeddedMessageWithChannel(nil, msg, targetChannelID)
	if err != nil {
		h.logger.Error(err, "failed to send top memo authors message")
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

	// Try DuckDB first, fall back to database on failure
	memos, err := h.getMemosFromParquet(start, end, "weekly")
	if err != nil {
		h.logger.Warnf("Failed to get memos from parquet, falling back to database: %v", err)
		// Fallback to database query
		memos, err = h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
		if err != nil {
			h.logger.Error(err, "failed to retrieve weekly memos from database")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	if len(memos) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no new memos in this week"))
		return
	}

	targetChannelID := discordPlayGroundReadingChannel
	if h.config.Env == "prod" {
		targetChannelID = discordRandomChannel
	}

	// Detect new authors by comparing with historical data
	newAuthors, err := h.getNewAuthorsFromParquet(memos, "weekly")
	if err != nil {
		h.logger.Warnf("Failed to detect new authors: %v", err)
		newAuthors = []string{} // Continue with empty list if detection fails
	}

	// Create a simple author resolver that handles invalid IDs gracefully
	authorResolver := func(discordAccountID string) (*model.DiscordAccount, error) {
		// If the ID contains invalid characters like commas, create a cleaned username
		if strings.Contains(discordAccountID, ",") || strings.Contains(discordAccountID, " ") {
			cleanedUsername := strings.ReplaceAll(strings.ReplaceAll(discordAccountID, ",", "_"), " ", "_")
			return &model.DiscordAccount{
				DiscordID:       "",
				DiscordUsername: cleanedUsername,
				MemoUsername:    cleanedUsername,
			}, nil
		}
		
		// First try to find employee by Discord/GitHub username, then get their Discord account
		employee, err := h.store.Employee.GetByDiscordUsername(h.repo.DB(), discordAccountID)
		if err == nil && employee.DiscordAccount != nil {
			return employee.DiscordAccount, nil
		}
		
		// Fallback: if it's a valid UUID, try direct lookup
		if model.IsUUIDFromString(discordAccountID) {
			return h.store.DiscordAccount.One(h.repo.DB(), discordAccountID)
		}
		
		// If all else fails, return a fallback account with the username
		return &model.DiscordAccount{
			DiscordID:       "",
			DiscordUsername: discordAccountID,
			MemoUsername:    discordAccountID,
		}, nil
	}

	_, err = h.service.Discord.SendWeeklyMemosMessage(
		h.config.Discord.IDs.DwarvesGuild,
		memos,
		weekRangeStr,
		targetChannelID,
		authorResolver,
		newAuthors,
		h.resolveAuthorsFromParquetByTitle,
		func(username string) (string, error) {
			employee, err := h.store.Employee.GetByDiscordUsername(h.repo.DB(), username)
			if err != nil {
				return "", err
			}
			if employee.DiscordAccount == nil {
				return "", fmt.Errorf("no discord account found for user %s", username)
			}
			return employee.DiscordAccount.DiscordID, nil
		},
	)
	if err != nil {
		h.logger.Error(err, "failed to send weekly memos report")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) NotifyMonthlyMemos(c *gin.Context) {
	// Calculate monthly time range (first day of month to now)
	end := time.Now()
	start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())
	
	// Try DuckDB first, fall back to database on failure
	memos, err := h.getMemosFromParquet(start, end, "monthly")
	if err != nil {
		h.logger.Warnf("Failed to get memos from parquet, falling back to database: %v", err)
		// Fallback to database query
		memos, err = h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
		if err != nil {
			h.logger.Error(err, "failed to retrieve monthly memos from database")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	if len(memos) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "no new memos this month"))
		return
	}

	// Format month range string (e.g., "AUGUST 2025")
	monthRangeStr := fmt.Sprintf("%s %d", strings.ToUpper(end.Month().String()), end.Year())
	
	// Determine target channel
	targetChannelID := discordPlayGroundReadingChannel
	if h.config.Env == "prod" {
		targetChannelID = discordRandomChannel
	}

	// Call new service method
	_, err = h.service.Discord.SendMonthlyMemosMessage(
		h.config.Discord.IDs.DwarvesGuild,
		memos,
		monthRangeStr,
		targetChannelID,
		func(discordAccountID string) (*model.DiscordAccount, error) {
			return h.store.DiscordAccount.One(h.repo.DB(), discordAccountID)
		},
		func(username string) (string, error) {
			employee, err := h.store.Employee.GetByDiscordUsername(h.repo.DB(), username)
			if err != nil {
				return "", err
			}
			if employee.DiscordAccount == nil {
				return "", fmt.Errorf("no discord account found for user %s", username)
			}
			return employee.DiscordAccount.DiscordID, nil
		},
	)

	if err != nil {
		h.logger.Error(err, "failed to send monthly memo message")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "monthly memo notification sent successfully"))
}

// getMemosFromParquet retrieves memos from parquet file with fallback to database
func (h *handler) getMemosFromParquet(start, end time.Time, period string) ([]model.MemoLog, error) {
	// Check if parquet querying is disabled via environment variable
	if os.Getenv("DISABLE_PARQUET_QUERY") == "true" {
		h.logger.Info("Parquet querying disabled via environment variable, falling back to database")
		return h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
	}
	
	// Check if DuckDB service is available
	if h.service.DuckDB == nil {
		h.logger.Warn("DuckDB service not available, falling back to database")
		return h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
	}

	// Use local parquet file if caching is enabled and file is ready
	var parquetURL string
	if h.service.ParquetSync.IsLocalFileReady() {
		parquetURL = h.service.ParquetSync.GetLocalFilePath()
		h.logger.Debugf("Using local parquet file: %s", parquetURL)
	} else {
		// Fallback to remote URL if local file not ready
		parquetURL = h.service.ParquetSync.GetRemoteURL()
		h.logger.Debugf("Using remote parquet file: %s", parquetURL)
	}
	
	// Build DuckDB query with time range filters and efficient pagination
	options := duckdb.QueryOptions{
		Filters: []duckdb.QueryFilter{
			{Column: "date", Operator: ">=", Value: start.Format("2006-01-02")},
			{Column: "date", Operator: "<=", Value: end.Format("2006-01-02")},
		},
		OrderBy: []string{"date DESC"},
		Limit:   100,  // Reduced limit for faster queries
		Offset:  0,    // Start from most recent
	}
	
	h.logger.Infof("Attempting to query parquet data for %s period from %s to %s", period, start.Format("2006-01-02"), end.Format("2006-01-02"))
	
	// Query parquet file using DuckDB with reduced timeout for faster fallback
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	parquetData, err := h.service.DuckDB.QueryParquetWithFilters(ctx, parquetURL, options)
	if err != nil {
		h.logger.Warnf("Failed to query parquet data, falling back to database: %v", err)
		return h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
	}
	
	h.logger.Infof("Successfully retrieved %d records from parquet file", len(parquetData))
	
	// Transform parquet data to ParquetMemoRecord format
	var parquetRecords []helpers.ParquetMemoRecord
	for _, row := range parquetData {
		record := helpers.ParquetMemoRecord{}
		
		// Extract date - handle both string and time.Time types
		if dateVal, ok := row["date"]; ok {
			switch date := dateVal.(type) {
			case string:
				record.Date = date
			case time.Time:
				record.Date = date.Format("2006-01-02")
			default:
				// Try to convert to string as fallback
				record.Date = fmt.Sprintf("%v", dateVal)
			}
		}
		
		// Extract title
		if titleVal, ok := row["title"]; ok {
			if titleStr, ok := titleVal.(string); ok {
				record.Title = titleStr
			}
		}
		
		// Extract authors (handle array format)
		if authorsVal, ok := row["authors"]; ok {
			switch authors := authorsVal.(type) {
			case []interface{}:
				for _, author := range authors {
					if authorStr, ok := author.(string); ok {
						record.Authors = append(record.Authors, authorStr)
					}
				}
			case string:
				// Handle single author as string
				record.Authors = []string{authors}
			}
		}
		
		// Extract tags (handle array format)
		if tagsVal, ok := row["tags"]; ok {
			switch tags := tagsVal.(type) {
			case []interface{}:
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						record.Tags = append(record.Tags, tagStr)
					}
				}
			case string:
				// Handle single tag as string
				record.Tags = []string{tags}
			}
		}
		
		// Extract URL
		if urlVal, ok := row["url"]; ok {
			if urlStr, ok := urlVal.(string); ok {
				record.URL = urlStr
			}
		}
		
		// Extract content
		if contentVal, ok := row["content"]; ok {
			if contentStr, ok := contentVal.(string); ok {
				record.Content = contentStr
			}
		}
		
		parquetRecords = append(parquetRecords, record)
	}
	
	// Transform parquet records to MemoLog models using helper
	memoLogs, err := h.dataTransformer.TransformParquetRecords(parquetRecords)
	if err != nil {
		h.logger.Warnf("Failed to transform parquet data, falling back to database: %v", err)
		return h.store.MemoLog.GetLimitByTimeRange(h.repo.DB(), &start, &end, 1000)
	}
	
	h.logger.Infof("Successfully transformed %d parquet records to MemoLog models", len(memoLogs))
	
	// Filter results by date range again to ensure accuracy (parquet data might have timezone issues)
	var filteredMemos []model.MemoLog
	for _, memo := range memoLogs {
		var memoDate time.Time
		if memo.PublishedAt != nil {
			memoDate = *memo.PublishedAt
		} else {
			memoDate = memo.CreatedAt
		}
		
		if (memoDate.Equal(start) || memoDate.After(start)) && 
		   (memoDate.Equal(end) || memoDate.Before(end)) {
			filteredMemos = append(filteredMemos, memo)
		}
	}
	
	h.logger.Infof("Returning %d filtered memos from parquet data", len(filteredMemos))
	return filteredMemos, nil
}

// getNewAuthorsFromParquet detects new authors by comparing current period with historical data
func (h *handler) getNewAuthorsFromParquet(currentMemos []model.MemoLog, period string) ([]string, error) {
	// Check if parquet querying is disabled
	if os.Getenv("DISABLE_PARQUET_QUERY") == "true" {
		h.logger.Info("Parquet querying disabled via environment variable, skipping new author detection")
		return []string{}, nil
	}
	
	// Check if DuckDB service is available
	if h.service.DuckDB == nil {
		h.logger.Warn("DuckDB service not available, skipping new author detection")
		return []string{}, nil
	}
	
	// Calculate historical period for comparison
	var historicalStart, historicalEnd time.Time
	now := time.Now()
	
	if period == "weekly" {
		// Look back 4 weeks for historical comparison
		historicalEnd = now.AddDate(0, 0, -7)  // End 1 week ago
		historicalStart = historicalEnd.AddDate(0, 0, -28) // Start 4 weeks before that
	} else { // monthly
		// Look back 3 months for historical comparison
		historicalEnd = now.AddDate(0, -1, 0)   // End 1 month ago
		historicalStart = historicalEnd.AddDate(0, -3, 0) // Start 3 months before that
	}
	
	h.logger.Infof("Querying historical authors from %s to %s for new author detection", 
		historicalStart.Format("2006-01-02"), historicalEnd.Format("2006-01-02"))
	
	// Query historical authors from parquet with optimized query
	var parquetURL string
	if h.service.ParquetSync.IsLocalFileReady() {
		parquetURL = h.service.ParquetSync.GetLocalFilePath()
	} else {
		parquetURL = h.service.ParquetSync.GetRemoteURL()
	}
	options := duckdb.QueryOptions{
		Columns: []string{"authors"}, // Only fetch authors field for efficiency
		Filters: []duckdb.QueryFilter{
			{Column: "date", Operator: ">=", Value: historicalStart.Format("2006-01-02")},
			{Column: "date", Operator: "<=", Value: historicalEnd.Format("2006-01-02")},
		},
		OrderBy: []string{"date DESC"},
		Limit:   500, // Reasonable limit for historical data
		Offset:  0,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	historicalData, err := h.service.DuckDB.QueryParquetWithFilters(ctx, parquetURL, options)
	if err != nil {
		h.logger.Warnf("Failed to query historical parquet data, skipping new author detection: %v", err)
		return []string{}, nil // Return empty list rather than error to not break the main flow
	}
	
	// Extract historical authors from parquet data
	historicalAuthors := make(map[string]bool)
	for _, row := range historicalData {
		if authorsVal, ok := row["authors"]; ok {
			// Handle different possible formats for authors field
			switch authors := authorsVal.(type) {
			case []string:
				for _, author := range authors {
					if cleanAuthor := strings.TrimSpace(author); cleanAuthor != "" {
						historicalAuthors[strings.ToLower(cleanAuthor)] = true
					}
				}
			case string:
				// If authors is stored as a single string, try to parse it
				if cleanAuthor := strings.TrimSpace(authors); cleanAuthor != "" {
					historicalAuthors[strings.ToLower(cleanAuthor)] = true
				}
			case []interface{}:
				// Handle case where authors is stored as []interface{}
				for _, authorInterface := range authors {
					if authorStr, ok := authorInterface.(string); ok {
						if cleanAuthor := strings.TrimSpace(authorStr); cleanAuthor != "" {
							historicalAuthors[strings.ToLower(cleanAuthor)] = true
						}
					}
				}
			}
		}
	}
	
	h.logger.Infof("Found %d historical authors from parquet data", len(historicalAuthors))
	
	// Get current period authors
	currentAuthors := make(map[string]bool)
	for _, memo := range currentMemos {
		for _, author := range memo.AuthorMemoUsernames {
			if cleanAuthor := strings.TrimSpace(author); cleanAuthor != "" {
				currentAuthors[strings.ToLower(cleanAuthor)] = true
			}
		}
	}
	
	// Find new authors (present in current but not in historical)
	var newAuthors []string
	for currentAuthor := range currentAuthors {
		if !historicalAuthors[currentAuthor] {
			newAuthors = append(newAuthors, currentAuthor)
		}
	}
	
	h.logger.Infof("Detected %d new authors: %v", len(newAuthors), newAuthors)
	return newAuthors, nil
}

// resolveAuthorsFromParquetByTitle queries parquet file to find actual authors for a given post title
func (h *handler) resolveAuthorsFromParquetByTitle(title string) ([]string, error) {
	// Check if parquet querying is disabled
	if os.Getenv("DISABLE_PARQUET_QUERY") == "true" {
		h.logger.Debug("Parquet querying disabled via environment variable, skipping author resolution by title")
		return []string{}, nil
	}
	
	// Check if DuckDB service is available
	if h.service.DuckDB == nil {
		h.logger.Debug("DuckDB service not available, skipping author resolution by title")
		return []string{}, nil
	}
	
	// Clean and prepare title for search
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		return []string{}, nil
	}
	
	h.logger.Debugf("Querying parquet for authors of post: %s", cleanTitle)
	
	// Query parquet file for this specific title
	var parquetURL string
	if h.service.ParquetSync.IsLocalFileReady() {
		parquetURL = h.service.ParquetSync.GetLocalFilePath()
	} else {
		parquetURL = h.service.ParquetSync.GetRemoteURL()
	}
	options := duckdb.QueryOptions{
		Columns: []string{"title", "authors"}, // Fetch title and authors
		Filters: []duckdb.QueryFilter{
			{Column: "title", Operator: "LIKE", Value: "%" + cleanTitle + "%"}, // Fuzzy match on title
		},
		OrderBy: []string{"date DESC"},
		Limit:   10, // Small limit since we're looking for exact match
		Offset:  0,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	parquetData, err := h.service.DuckDB.QueryParquetWithFilters(ctx, parquetURL, options)
	if err != nil {
		h.logger.Debugf("Failed to query parquet for title '%s': %v", cleanTitle, err)
		return []string{}, nil // Return empty rather than error to not break flow
	}
	
	// Find exact title match and extract authors
	for _, row := range parquetData {
		if titleVal, ok := row["title"]; ok {
			if titleStr, ok := titleVal.(string); ok {
				// Check for exact match (case-insensitive)
				if strings.EqualFold(strings.TrimSpace(titleStr), cleanTitle) {
					if authorsVal, ok := row["authors"]; ok {
						// Extract authors from different possible formats
						var authors []string
						switch authorsData := authorsVal.(type) {
						case []string:
							authors = authorsData
						case string:
							if strings.TrimSpace(authorsData) != "" {
								authors = []string{authorsData}
							}
						case []interface{}:
							for _, authorInterface := range authorsData {
								if authorStr, ok := authorInterface.(string); ok {
									if cleanAuthor := strings.TrimSpace(authorStr); cleanAuthor != "" {
										authors = append(authors, cleanAuthor)
									}
								}
							}
						}
						
						// Clean and validate authors
						var validAuthors []string
						for _, author := range authors {
							cleanAuthor := strings.TrimSpace(author)
							if cleanAuthor != "" && !strings.Contains(cleanAuthor, ",") {
								validAuthors = append(validAuthors, cleanAuthor)
							}
						}
						
						h.logger.Debugf("Found %d authors for title '%s': %v", len(validAuthors), cleanTitle, validAuthors)
						return validAuthors, nil
					}
				}
			}
		}
	}
	
	h.logger.Debugf("No exact match found in parquet for title: %s", cleanTitle)
	return []string{}, nil
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
	e := &model.Event{
		DiscordEventID:   in.ID,
		DiscordChannelID: in.DiscordChannelID,
		DiscordCreatorID: in.DiscordCreatorID,
		Name:             in.Name,
		Description:      in.Description,
		Date:             in.Date,
		EventType:        evtType,
		Image:            in.Image,
	}
	_, err = h.store.DiscordEvent.Create(h.repo.DB(), e)
	if err != nil {
		h.logger.Error(err, "failed to create event")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// create youtube broadcast
	if evtType == model.DiscordScheduledEventTypeOGIF {
		err = h.service.Youtube.CreateBroadcast(e)
		if err != nil {
			h.logger.Error(err, "failed to create youtube broadcast")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, "bind query failed"))
		return
	}
	query.Standardize()

	limit, offset := query.ToLimitOffset()

	// Default by last 7 days, 0 is get all
	inputDays := c.Query("days")
	if inputDays == "" {
		inputDays = "7"
	}
	days, err := strconv.Atoi(inputDays)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, "invalid days query"))
	}

	topics, total, err := h.controller.Discord.ListDiscordResearchTopics(context.Background(), days, limit, offset)
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

func (h *handler) UserOgifStats(c *gin.Context) {
	discordID := c.Query("discordID")
	if discordID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("discord_id is required"), nil, ""))
		return
	}

	var afterTime time.Time
	after := c.Query("after")
	if after != "" {
		var err error
		afterTime, err = time.Parse(time.RFC3339, after)
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("invalid after time format"), nil, ""))
			return
		}
	}

	stats, err := h.controller.Discord.UserOgifStats(c.Request.Context(), discordID, afterTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("discord_id is required"), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(stats, nil, nil, nil, ""))
}

func (h *handler) OgifLeaderboard(c *gin.Context) {
	var afterTime time.Time
	after := c.Query("after")
	if after != "" {
		var err error
		afterTime, err = time.Parse(time.RFC3339, after)
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("invalid after time format"), nil, ""))
			return
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	leaderboard, err := h.controller.Discord.GetOgifLeaderboard(c.Request.Context(), afterTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(leaderboard, nil, nil, nil, ""))
}

func (h *handler) SweepOgifEvent(c *gin.Context) {
	err := h.controller.Event.SweepOgifEvent(c.Request.Context())
	if err != nil {
		h.logger.Error(err, "failed to sweep events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "events swept successfully"))
}
