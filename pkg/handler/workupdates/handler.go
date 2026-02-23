package workupdates

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
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
}

// GetWorkUpdatesResponse represents the full response
type GetWorkUpdatesResponse struct {
	Month    string               `json:"month"`
	Projects []ProjectMissingData `json:"projects"`
}

// deploymentResult holds the result of processing a single deployment
type deploymentResult struct {
	ContractorID     string
	ContractorName   string
	ProjectID        string
	ProjectName      string
	MissingDates     []string
	NotReviewedCount int
	Error            error
}

// GetWorkUpdates godoc
// @Summary Get work updates completion status
// @Description Check which contractors have missing timesheet entries for a given month
// @Tags Work Updates
// @Accept json
// @Produce json
// @Param month path string true "Month in YYYY-MM format"
// @Param includeReviewStatus query bool false "Include not-reviewed timesheet counts"
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

	// For current month, only include working days up to today
	now := time.Now()
	if year == now.Year() && monthNum == int(now.Month()) {
		var filtered []time.Time
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		for _, d := range workingDays {
			if !d.After(today) {
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

	// Process deployments concurrently
	results := h.processDeploymentsConcurrently(ctx, deployments, workingDays, year, monthNum)

	// Group results by project
	projectMap := make(map[string]*ProjectMissingData)
	for _, r := range results {
		if includeReviewStatus {
			// Not-review mode: only show contractors with not-reviewed entries
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
			// Default mode: show contractors with missing timesheets
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
			})
		}
	}

	// Convert to slice and sort
	var projects []ProjectMissingData
	for _, proj := range projectMap {
		// Sort contractors alphabetically
		sort.Slice(proj.Contractors, func(i, j int) bool {
			return proj.Contractors[i].ContractorName < proj.Contractors[j].ContractorName
		})
		projects = append(projects, *proj)
	}

	// Sort projects alphabetically
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ProjectName < projects[j].ProjectName
	})

	response := GetWorkUpdatesResponse{
		Month:    month,
		Projects: projects,
	}

	l.Debugf("work updates response: %d projects with missing timesheets", len(projects))

	c.JSON(http.StatusOK, view.CreateResponse[any](response, nil, nil, nil, ""))
}

func (h *handler) processDeploymentsConcurrently(
	ctx context.Context,
	deployments []*notion.DeploymentData,
	workingDays []time.Time,
	year, month int,
) []deploymentResult {
	resultsCh := make(chan deploymentResult, len(deployments))
	var wg sync.WaitGroup

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

			// Fetch contractor name
			name, err := taskOrderLog.GetContractorName(ctx, dep.ContractorPageID)
			if err != nil {
				h.logger.Error(err, fmt.Sprintf("failed to get contractor name: %s", dep.ContractorPageID))
				result.ContractorName = dep.ContractorPageID
			} else {
				result.ContractorName = name
			}

			// Fetch project name
			projName, err := taskOrderLog.GetProjectName(ctx, dep.ProjectPageID)
			if err != nil {
				h.logger.Error(err, fmt.Sprintf("failed to get project name: %s", dep.ProjectPageID))
				result.ProjectName = dep.ProjectPageID
			} else {
				result.ProjectName = projName
			}

			// Query timesheet entries
			timesheetResult, err := timesheet.QueryTimesheetsByContractorProjectMonth(
				ctx, dep.ContractorPageID, dep.ProjectPageID, year, month,
			)
			if err != nil {
				h.logger.Error(err, fmt.Sprintf("failed to query timesheets: contractor=%s project=%s",
					dep.ContractorPageID, dep.ProjectPageID))
				result.Error = err
				resultsCh <- result
				return
			}

			result.NotReviewedCount = timesheetResult.NotReviewedCount

			// Query leave dates from Notion
			leaveDates := h.getApprovedLeaveDates(ctx, dep.ContractorPageID, year, month)

			// Calculate missing dates
			result.MissingDates = calculateMissingDates(workingDays, timesheetResult.DateCounts, leaveDates)

			resultsCh <- result
		}(dep)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []deploymentResult
	for r := range resultsCh {
		if r.Error != nil {
			h.logger.Error(r.Error, fmt.Sprintf("skipping deployment: contractor=%s project=%s", r.ContractorID, r.ProjectID))
			continue
		}
		results = append(results, r)
	}

	return results
}

// getApprovedLeaveDates returns a set of date strings (YYYY-MM-DD) for approved leave days
func (h *handler) getApprovedLeaveDates(ctx context.Context, contractorPageID string, year, month int) map[string]bool {
	leaveService := notion.NewLeaveService(h.config, h.store, h.repo, h.logger)
	if leaveService == nil {
		h.logger.Debug("leave service not available, skipping leave date query")
		return make(map[string]bool)
	}

	leaveDates, err := leaveService.QueryAcknowledgedLeaveDatesByContractorMonth(ctx, contractorPageID, year, month)
	if err != nil {
		h.logger.Error(err, fmt.Sprintf("failed to query leave dates: contractor=%s", contractorPageID))
		return make(map[string]bool)
	}

	return leaveDates
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
