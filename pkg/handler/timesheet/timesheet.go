package timesheet

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/timesheet/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/timesheet/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
	controller *controller.Controller
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
		controller: controller,
	}
}

// LogHours godoc
// @Summary Log timesheet hours
// @Description Create a new timesheet entry in Notion
// @Tags Timesheet
// @Accept json
// @Produce json
// @Param body body request.LogHoursRequest true "Timesheet entry details"
// @Success 200 {object} view.TimesheetEntryResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /timesheets [post]
func (h *handler) LogHours(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "timesheet",
		"method":  "LogHours",
	})

	var req request.LogHoursRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "failed to bind request")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, "invalid request"))
		return
	}

	if err := req.Validate(); err != nil {
		l.Errorf(err, "validation failed", "body", req)
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// Create timesheet service
	timesheetService := notionSvc.NewTimesheetService(h.config, h.store, h.repo, h.logger)
	if timesheetService == nil {
		l.Error(errors.New("failed to create timesheet service"), "notion secret may not be configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrTimesheetDBNotFound, nil, ""))
		return
	}

	// Find contractor by Discord ID
	contractorID, err := timesheetService.GetContractorPageIDByDiscordID(c.Request.Context(), req.DiscordID)
	if err != nil {
		l.Error(err, "failed to find contractor")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrContractorNotFound, nil,
			"contractor not found for Discord ID: "+req.DiscordID))
		return
	}

	// Create timesheet entry
	entry := notionSvc.TimesheetEntry{
		ProjectID:    req.ProjectID,
		ContractorID: contractorID,
		Date:         &req.Date,
		TaskType:     req.TaskType,
		Hours:        req.Hours,
		ProofOfWorks: req.ProofOfWorks,
		TaskOrderID:  req.TaskOrderID,
	}

	pageID, err := timesheetService.CreateTimesheetEntry(c.Request.Context(), entry)
	if err != nil {
		l.Error(err, "failed to create timesheet entry")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to log hours"))
		return
	}

	entry.PageID = pageID
	c.JSON(http.StatusOK, view.CreateResponse(
		view.ToTimesheetEntry(entry),
		nil, nil, nil,
		"timesheet entry created successfully",
	))
}

// GetEntries godoc
// @Summary Get timesheet entries
// @Description Retrieve timesheet entries for a user by Discord ID
// @Tags Timesheet
// @Accept json
// @Produce json
// @Param discord_id query string true "Discord ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} view.TimesheetEntriesResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /timesheets [get]
func (h *handler) GetEntries(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "timesheet",
		"method":  "GetEntries",
	})

	discordID := c.Query("discord_id")
	if discordID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			errs.ErrEmptyDiscordID, nil, ""))
		return
	}

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "invalid start_date format"))
			return
		}
		startDate = &t
	}

	if ed := c.Query("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "invalid end_date format"))
			return
		}
		endDate = &t
	}

	// Create timesheet service
	timesheetService := notionSvc.NewTimesheetService(h.config, h.store, h.repo, h.logger)
	if timesheetService == nil {
		l.Error(errors.New("failed to create timesheet service"), "notion secret may not be configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrTimesheetDBNotFound, nil, ""))
		return
	}

	// Get Discord username from Discord ID for querying Notion
	discordUsername, err := h.getDiscordUsername(c, discordID)
	if err != nil {
		l.Error(err, "failed to get discord username")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "discord user not found"))
		return
	}

	entries, err := timesheetService.QueryTimesheetByDiscord(
		c.Request.Context(),
		discordUsername,
		startDate,
		endDate,
	)
	if err != nil {
		l.Error(err, "failed to query timesheet entries")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(
		view.ToTimesheetEntries(entries),
		nil, nil, nil,
		"entries retrieved successfully",
	))
}

// GetWeeklySummary godoc
// @Summary Get weekly timesheet summary
// @Description Retrieve weekly hours summary for a user
// @Tags Timesheet
// @Accept json
// @Produce json
// @Param discord_id query string true "Discord ID"
// @Param week_offset query int false "Week offset from current week (0=current, -1=last week)"
// @Success 200 {object} view.TimesheetWeeklySummaryResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /timesheets/weekly-summary [get]
func (h *handler) GetWeeklySummary(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "timesheet",
		"method":  "GetWeeklySummary",
	})

	discordID := c.Query("discord_id")
	if discordID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
			errs.ErrEmptyDiscordID, nil, ""))
		return
	}

	// Calculate week boundaries
	weekOffset := 0
	if wo := c.Query("week_offset"); wo != "" {
		if parsed, err := strconv.Atoi(wo); err == nil {
			weekOffset = parsed
		}
	}

	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}

	// Start of week (Monday)
	startOfWeek := now.AddDate(0, 0, -(weekday-1)+(weekOffset*7))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	endOfWeek = time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, endOfWeek.Location())

	// Create timesheet service
	timesheetService := notionSvc.NewTimesheetService(h.config, h.store, h.repo, h.logger)
	if timesheetService == nil {
		l.Error(errors.New("failed to create timesheet service"), "notion secret may not be configured")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrTimesheetDBNotFound, nil, ""))
		return
	}

	// Get Discord username from Discord ID for querying Notion
	discordUsername, err := h.getDiscordUsername(c, discordID)
	if err != nil {
		l.Error(err, "failed to get discord username")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "discord user not found"))
		return
	}

	entries, err := timesheetService.QueryTimesheetByDiscord(
		c.Request.Context(),
		discordUsername,
		&startOfWeek,
		&endOfWeek,
	)
	if err != nil {
		l.Error(err, "failed to query timesheet entries")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Calculate summary
	summary := calculateWeeklySummary(entries, startOfWeek, endOfWeek)

	c.JSON(http.StatusOK, view.CreateResponse(summary, nil, nil, nil, "weekly summary retrieved"))
}

// getDiscordUsername retrieves Discord username from Discord ID
func (h *handler) getDiscordUsername(c *gin.Context, discordID string) (string, error) {
	var discordAccount struct {
		DiscordUsername string
	}
	err := h.repo.DB().WithContext(c.Request.Context()).
		Table("discord_accounts").
		Select("discord_username").
		Where("discord_id = ? AND deleted_at IS NULL", discordID).
		First(&discordAccount).Error
	if err != nil {
		return "", err
	}
	return discordAccount.DiscordUsername, nil
}

// calculateWeeklySummary calculates aggregated weekly summary from entries
func calculateWeeklySummary(entries []notionSvc.TimesheetEntry, startOfWeek, endOfWeek time.Time) view.TimesheetWeeklySummary {
	summary := view.TimesheetWeeklySummary{
		StartDate:  startOfWeek,
		EndDate:    endOfWeek,
		TotalHours: 0,
		ByTaskType: make(map[string]float64),
		ByProject:  make(map[string]float64),
		EntryCount: len(entries),
	}

	for _, entry := range entries {
		summary.TotalHours += entry.Hours
		summary.ByTaskType[entry.TaskType] += entry.Hours
		if entry.ProjectID != "" {
			summary.ByProject[entry.ProjectID] += entry.Hours
		}
	}

	return summary
}
