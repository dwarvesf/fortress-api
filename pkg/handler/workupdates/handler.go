package workupdates

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	discordsvc "github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	service        *service.Service
	store          *store.Store
	repo           store.DBRepo
	logger         logger.Logger
	config         *config.Config
	holidayService *google.HolidayService
}

// New creates a new work updates handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		service:        service,
		store:          store,
		repo:           repo,
		logger:         logger,
		config:         cfg,
		holidayService: google.NewHolidayService(logger),
	}
}

// ProjectMissingData represents missing timesheet data for a project
type ProjectMissingData struct {
	ProjectID        string                  `json:"projectId"`
	ProjectName      string                  `json:"projectName"`
	NotReviewedCount int                     `json:"notReviewedCount"`
	Contractors      []ContractorMissingData `json:"contractors"`
}

// ContractorMissingData represents missing timesheet data for a contractor
type ContractorMissingData struct {
	ContractorID     string   `json:"contractorId"`
	ContractorName   string   `json:"contractorName"`
	MissingCount     int      `json:"missingCount"`
	MissingDates     []string `json:"missingDates"` // Format: DD/MM
	NotReviewedCount int      `json:"notReviewedCount"`
	TotalHours       float64  `json:"totalHours"`
	AvgHoursPerDay   float64  `json:"avgHoursPerDay"`
}

// GetWorkUpdatesResponse represents the full response
type GetWorkUpdatesResponse struct {
	Month       string                  `json:"month"`
	Projects    []ProjectMissingData    `json:"projects"`
	Contractors []ContractorMissingData `json:"contractors,omitempty"` // used by --extra mode (cross-project aggregation)
}

// deploymentResult holds the result of processing a single deployment
type deploymentResult struct {
	ContractorID     string
	ContractorName   string
	ProjectID        string
	ProjectName      string
	MissingDates     []string
	NotReviewedCount int
	TotalHours       float64
	AvgHoursPerDay   float64
	Error            error
}

// GetWorkUpdates godoc
// @Summary Get work updates completion status
// @Description Check which contractors have missing timesheet entries for a given month
// @Tags Work Updates
// @Accept json
// @Produce json
// @Param month path string true "Month in YYYY-MM format"
// @Param includeReviewStatus query bool false "When true, show only not-reviewed timesheets instead of missing timesheets"
// @Param channelId query string false "Discord channel ID for live progress updates"
// @Success 200 {object} view.Response{data=GetWorkUpdatesResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/work-updates/{month} [get]
func (h *handler) GetWorkUpdates(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "workupdates",
		"method":  "GetWorkUpdates",
	})

	includeReviewStatus := c.Query("includeReviewStatus") == "true"
	channelID := c.Query("channelId")

	// Parse extra hours threshold; 0 means disabled
	var extraHoursThreshold float64
	if v := c.Query("extraHoursThreshold"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed > 0 {
			extraHoursThreshold = parsed
		}
	}

	month := c.Param("month")
	if month == "" {
		l.Debug("month parameter is required")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("month parameter is required (format: YYYY-MM)"), nil, ""))
		return
	}

	// Validate month format
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		l.Debug("invalid month format")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, fmt.Errorf("invalid month format (expected: YYYY-MM)"), nil, ""))
		return
	}

	year := monthTime.Year()
	monthNum := int(monthTime.Month())

	l.Debugf("fetching work updates for month=%s year=%d monthNum=%d", month, year, monthNum)

	// Get working days for the month (weekdays only)
	workingDays := timeutil.GetWorkingDaysInMonth(year, monthNum)

	// Fetch public holidays and makeup working days from Google Calendar
	holidays, makeupDays, err := h.holidayService.GetHolidaysForMonth(year, monthNum)
	if err != nil {
		l.Error(err, "failed to fetch public holidays, proceeding without holiday exclusion")
		holidays = make(map[string]string)
		makeupDays = make(map[string]string)
	}

	l.Debugf("public holidays: %d, makeup working days: %d", len(holidays), len(makeupDays))

	// Remove public holidays from working days
	if len(holidays) > 0 {
		var filtered []time.Time
		for _, d := range workingDays {
			if _, isHoliday := holidays[d.Format("2006-01-02")]; !isHoliday {
				filtered = append(filtered, d)
			}
		}
		workingDays = filtered
	}

	// Add makeup working days (Saturdays designated as working days)
	for dateStr := range makeupDays {
		if t, parseErr := time.Parse("2006-01-02", dateStr); parseErr == nil {
			workingDays = append(workingDays, t)
		}
	}

	// Sort working days
	sort.Slice(workingDays, func(i, j int) bool {
		return workingDays[i].Before(workingDays[j])
	})

	// For current month, only include working days before today (exclude current date)
	now := time.Now()
	if year == now.Year() && monthNum == int(now.Month()) {
		var filtered []time.Time
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		for _, d := range workingDays {
			if d.Before(today) {
				filtered = append(filtered, d)
			}
		}
		workingDays = filtered
	}

	l.Debugf("working days count: %d (after holidays and makeup days)", len(workingDays))

	// Query active deployments
	notionService := h.service.Notion
	if notionService == nil || notionService.TaskOrderLog == nil {
		l.Debug("notion service is not initialized")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, fmt.Errorf("notion service not initialized"), nil, ""))
		return
	}

	ctx := c.Request.Context()
	deployments, err := notionService.TaskOrderLog.QueryActiveDeploymentsByMonth(ctx, month, "")
	if err != nil {
		l.Error(err, "failed to query active deployments")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Debugf("found %d active deployments", len(deployments))

	if len(deployments) == 0 {
		c.JSON(http.StatusOK, view.CreateResponse[any](GetWorkUpdatesResponse{
			Month:    month,
			Projects: []ProjectMissingData{},
		}, nil, nil, nil, ""))
		return
	}

	// Build progress bar for live Discord updates
	var pb *discordsvc.ProgressBar
	if channelID != "" && h.service.Discord != nil {
		// Create initial progress message in the Discord channel
		initEmbed := buildProgressEmbed(0, len(deployments), month, includeReviewStatus)
		msg, sendErr := h.service.Discord.SendChannelMessageComplex(channelID, "", []*discordgo.MessageEmbed{initEmbed}, nil)
		if sendErr != nil {
			l.Error(sendErr, "failed to send initial progress message to Discord")
		} else if msg != nil {
			reporter := discordsvc.NewChannelMessageReporter(h.service.Discord, channelID, msg.ID)
			pb = discordsvc.NewProgressBar(reporter, l)
		}
	}

	// Process deployments concurrently
	extraHoursMode := extraHoursThreshold > 0
	results := h.processDeploymentsConcurrently(ctx, deployments, workingDays, year, monthNum, pb, includeReviewStatus, month, extraHoursMode)

	var response GetWorkUpdatesResponse
	response.Month = month

	if extraHoursThreshold > 0 {
		// Extra hours mode: aggregate hours per contractor across all projects
		contractorMap := make(map[string]*ContractorMissingData)
		for _, r := range results {
			c, ok := contractorMap[r.ContractorID]
			if !ok {
				c = &ContractorMissingData{
					ContractorID:   r.ContractorID,
					ContractorName: r.ContractorName,
				}
				contractorMap[r.ContractorID] = c
			}
			c.TotalHours += r.TotalHours
		}

		for _, c := range contractorMap {
			if c.TotalHours >= extraHoursThreshold {
				response.Contractors = append(response.Contractors, *c)
			}
		}

		sort.Slice(response.Contractors, func(i, j int) bool {
			return response.Contractors[i].TotalHours > response.Contractors[j].TotalHours
		})

		l.Debugf("work updates response: %d contractors with extra hours (>=%.0f)", len(response.Contractors), extraHoursThreshold)
	} else {
		// Group results by project
		projectMap := make(map[string]*ProjectMissingData)
		for _, r := range results {
			if includeReviewStatus {
				if r.NotReviewedCount == 0 {
					continue
				}

				proj, ok := projectMap[r.ProjectID]
				if !ok {
					proj = &ProjectMissingData{
						ProjectID:   r.ProjectID,
						ProjectName: r.ProjectName,
					}
					projectMap[r.ProjectID] = proj
				}

				proj.NotReviewedCount += r.NotReviewedCount
				proj.Contractors = append(proj.Contractors, ContractorMissingData{
					ContractorID:     r.ContractorID,
					ContractorName:   r.ContractorName,
					NotReviewedCount: r.NotReviewedCount,
				})
			} else {
				if len(r.MissingDates) == 0 {
					continue
				}

				proj, ok := projectMap[r.ProjectID]
				if !ok {
					proj = &ProjectMissingData{
						ProjectID:   r.ProjectID,
						ProjectName: r.ProjectName,
					}
					projectMap[r.ProjectID] = proj
				}

				proj.Contractors = append(proj.Contractors, ContractorMissingData{
					ContractorID:   r.ContractorID,
					ContractorName: r.ContractorName,
					MissingCount:   len(r.MissingDates),
					MissingDates:   r.MissingDates,
					TotalHours:     r.TotalHours,
					AvgHoursPerDay: r.AvgHoursPerDay,
				})
			}
		}

		for _, proj := range projectMap {
			sort.Slice(proj.Contractors, func(i, j int) bool {
				return proj.Contractors[i].ContractorName < proj.Contractors[j].ContractorName
			})
			response.Projects = append(response.Projects, *proj)
		}

		sort.Slice(response.Projects, func(i, j int) bool {
			return response.Projects[i].ProjectName < response.Projects[j].ProjectName
		})

		l.Debugf("work updates response: %d projects with missing timesheets", len(response.Projects))
	}

	if pb != nil {
		pb.Delete()
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}

// nameCache provides thread-safe caching for contractor and project name lookups
type nameCache struct {
	contractors sync.Map // contractorPageID -> string
	projects    sync.Map // projectPageID -> string
}

func (h *handler) processDeploymentsConcurrently(
	ctx context.Context,
	deployments []*notion.DeploymentData,
	workingDays []time.Time,
	year, month int,
	pb *discordsvc.ProgressBar,
	includeReviewStatus bool,
	monthStr string,
	extraHoursMode bool,
) []deploymentResult {
	resultsCh := make(chan deploymentResult, len(deployments))
	var wg sync.WaitGroup

	// Shared caches to avoid redundant Notion API calls for the same contractor/project
	cache := &nameCache{}

	// Create leave service once, reuse across all goroutines
	var leaveService *notion.LeaveService
	if !extraHoursMode {
		leaveService = notion.NewLeaveService(h.config, h.store, h.repo, h.logger)
	}

	// Limit concurrency to avoid overwhelming Notion API
	semaphore := make(chan struct{}, 10)

	for _, dep := range deployments {
		wg.Add(1)
		go func(dep *notion.DeploymentData) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := deploymentResult{
				ContractorID: dep.ContractorPageID,
				ProjectID:    dep.ProjectPageID,
			}

			taskOrderLog := h.service.Notion.TaskOrderLog
			timesheet := h.service.Notion.Timesheet

			// Run independent lookups in parallel:
			// - contractor name (cached)
			// - project name (cached)
			// - timesheet query
			// - leave dates (skipped in extra hours mode)
			var innerWg sync.WaitGroup

			// 1. Contractor name (with cache)
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				if cached, ok := cache.contractors.Load(dep.ContractorPageID); ok {
					result.ContractorName = cached.(string)
					return
				}
				name, err := taskOrderLog.GetContractorName(ctx, dep.ContractorPageID)
				if err != nil {
					h.logger.Error(err, fmt.Sprintf("failed to get contractor name: %s", dep.ContractorPageID))
					result.ContractorName = dep.ContractorPageID
				} else {
					result.ContractorName = name
					cache.contractors.Store(dep.ContractorPageID, name)
				}
			}()

			// 2. Project name (with cache)
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				if cached, ok := cache.projects.Load(dep.ProjectPageID); ok {
					result.ProjectName = cached.(string)
					return
				}
				projName, err := taskOrderLog.GetProjectName(ctx, dep.ProjectPageID)
				if err != nil {
					h.logger.Error(err, fmt.Sprintf("failed to get project name: %s", dep.ProjectPageID))
					result.ProjectName = dep.ProjectPageID
				} else {
					result.ProjectName = projName
					cache.projects.Store(dep.ProjectPageID, projName)
				}
			}()

			// 3. Timesheet query
			var timesheetResult *notion.TimesheetQueryResult
			var timesheetErr error
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				timesheetResult, timesheetErr = timesheet.QueryTimesheetsByContractorProjectMonth(
					ctx, dep.ContractorPageID, dep.ProjectPageID, year, month,
				)
			}()

			// 4. Leave dates (skip in extra hours mode — not needed)
			var leaveDates map[string]bool
			if !extraHoursMode && leaveService != nil {
				innerWg.Add(1)
				go func() {
					defer innerWg.Done()
					dates, err := leaveService.QueryAcknowledgedLeaveDatesByContractorMonth(ctx, dep.ContractorPageID, year, month)
					if err != nil {
						h.logger.Error(err, fmt.Sprintf("failed to query leave dates: contractor=%s", dep.ContractorPageID))
						leaveDates = make(map[string]bool)
					} else {
						leaveDates = dates
					}
				}()
			}

			innerWg.Wait()

			// Process timesheet result
			if timesheetErr != nil {
				h.logger.Error(timesheetErr, fmt.Sprintf("failed to query timesheets: contractor=%s project=%s",
					dep.ContractorPageID, dep.ProjectPageID))
				result.Error = timesheetErr
				resultsCh <- result
				return
			}

			result.NotReviewedCount = timesheetResult.NotReviewedCount
			result.TotalHours = timesheetResult.TotalHours
			if daysWorked := len(timesheetResult.DateCounts); daysWorked > 0 {
				result.AvgHoursPerDay = timesheetResult.TotalHours / float64(daysWorked)
			}

			if !extraHoursMode {
				if leaveDates == nil {
					leaveDates = make(map[string]bool)
				}

				// Filter working days to only include days from the deployment start date
				depWorkingDays := workingDays
				if dep.StartDate != nil {
					startDate := time.Date(dep.StartDate.Year(), dep.StartDate.Month(), dep.StartDate.Day(), 0, 0, 0, 0, time.UTC)
					var filtered []time.Time
					for _, d := range workingDays {
						if !d.Before(startDate) {
							filtered = append(filtered, d)
						}
					}
					depWorkingDays = filtered
				}

				result.MissingDates = calculateMissingDates(depWorkingDays, timesheetResult.DateCounts, leaveDates)
			}

			resultsCh <- result
		}(dep)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	total := len(deployments)
	completed := 0
	var results []deploymentResult
	for r := range resultsCh {
		completed++

		if r.Error != nil {
			h.logger.Error(r.Error, fmt.Sprintf("skipping deployment: contractor=%s project=%s", r.ContractorID, r.ProjectID))
		} else {
			results = append(results, r)
		}

		// Send live progress update every 3 completions or at the end
		if pb != nil && (completed%3 == 0 || completed == total) {
			pb.Report(buildProgressEmbed(completed, total, monthStr, includeReviewStatus))
		}
	}

	return results
}

// buildProgressEmbed creates a Discord embed showing processing progress
func buildProgressEmbed(completed, total int, month string, notReviewMode bool) *discordgo.MessageEmbed {
	pct := float64(completed) / float64(total) * 100

	// Build progress bar visual
	filled := int(pct / 5) // 20 chars total
	var sb strings.Builder
	for i := range 20 {
		if i < filled {
			sb.WriteString("█")
		} else {
			sb.WriteString("░")
		}
	}
	bar := sb.String()

	mode := "missing timesheets"
	if notReviewMode {
		mode = "not-reviewed timesheets"
	}

	description := fmt.Sprintf("Checking %s for **%s**...\n\n`%s` %.0f%% (%d/%d deployments)",
		mode, month, bar, pct, completed, total)

	return &discordgo.MessageEmbed{
		Title:       "Processing...",
		Description: description,
		Color:       5793266, // colorBlue
	}
}


// calculateMissingDates returns the list of working days that have no timesheet entry and no leave
func calculateMissingDates(workingDays []time.Time, timesheetDates map[string]int, leaveDates map[string]bool) []string {
	var missing []string
	for _, d := range workingDays {
		dateKey := d.Format("2006-01-02")
		if timesheetDates[dateKey] > 0 {
			continue
		}
		if leaveDates[dateKey] {
			continue
		}
		missing = append(missing, fmt.Sprintf("%d/%d", d.Day(), int(d.Month())))
	}
	return missing
}
