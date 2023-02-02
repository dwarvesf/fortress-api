package dashboard

import (
	"errors"
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/dashboard/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// GetProjectSizes godoc
// @Summary Get the total number of active member in each project
// @Description Get the total number of active member in each project
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.ProjectSizeResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/sizes [get]
func (h *handler) GetProjectSizes(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetProjectSizes",
	})

	res, err := h.store.Dashboard.GetProjectSizes(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get project sizes")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
}

// GetWorkSurveys godoc
// @Summary Get Work Surveys data for dashboard
// @Description Get Work Surveys data for dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID   query  string false  "Project ID"
// @Success 200 {object} view.WorkSurveyResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/work-surveys [get]
func (h *handler) GetWorkSurveys(c *gin.Context) {
	input := request.WorkSurveysInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID != "" && !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetWorkSurveys",
		"input":   input,
	})

	var project *model.Project
	var workSurveys []*model.WorkSurvey
	var err error

	if input.ProjectID != "" {
		// Check project existence
		project, err = h.store.Project.One(h.repo.DB(), input.ProjectID, false)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "project not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
				return
			}

			l.Error(err, "failed to get project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
			return
		}

		// Get work survey by project ID
		workSurveys, err = h.store.Dashboard.GetWorkSurveysByProjectID(h.repo.DB(), input.ProjectID)
		if err != nil {
			l.Error(err, "failed to get work survey by project ID")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	} else {
		workSurveys, err = h.store.Dashboard.GetAllWorkSurveys(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get work survey")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToWorkSurveyData(project, workSurveys), nil, nil, nil, ""))
}

// GetActionItemReports godoc
// @Summary Get Action items report for dashboard
// @Description Get Action items report for dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID   query  string false  "Project ID"
// @Success 200 {object} view.ActionItemReportResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/action-items [get]
func (h *handler) GetActionItemReports(c *gin.Context) {
	input := request.ActionItemInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID != "" && !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetActionItemReports",
		"input":   input,
	})

	var actionItemReports []*model.ActionItemReport
	var err error

	if input.ProjectID != "" {
		// Check project existence
		isExist, err := h.store.Project.IsExist(h.repo.DB(), input.ProjectID)
		if !isExist {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
			return
		}

		// Get action item report by project ID
		actionItemReports, err = h.store.Dashboard.GetActionItemReportsByProjectID(h.repo.DB(), input.ProjectID)
		if err != nil {
			l.Error(err, "failed to get action item report by project ID")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	} else {
		actionItemReports, err = h.store.Dashboard.GetAllActionItemReports(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get action item report")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToActionItemReportData(actionItemReports), nil, nil, nil, ""))
}

// GetEngineeringHealth godoc
// @Summary Get Enginerring health information for dashboard
// @Description Get Enginerring health information for dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID   query  string false  "Project ID"
// @Success 200 {object} view.EngineeringHealthResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/engineering-healths [get]
func (h *handler) GetEngineeringHealth(c *gin.Context) {
	input := request.WorkSurveysInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID != "" && !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetEngineeringHealth",
		"input":   input,
	})

	var average []*model.AverageEngineeringHealth
	var groups []*model.GroupEngineeringHealth
	var err error

	if input.ProjectID != "" {
		// Check project existence
		project, err := h.store.Project.One(h.repo.DB(), input.ProjectID, false)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "project not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
				return
			}

			l.Error(err, "failed to get project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
			return
		}

		average, err = h.store.Dashboard.AverageEngineeringHealthByProjectNotionID(h.repo.DB(), project.NotionID.String())
		if err != nil {
			l.Error(err, "failed to get average engineering health")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		groups, err = h.store.Dashboard.GroupEngineeringHealthByProjectNotionID(h.repo.DB(), project.NotionID.String())
		if err != nil {
			l.Error(err, "failed to get group engineering health")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	} else {
		average, err = h.store.Dashboard.AverageEngineeringHealth(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get average engineering health")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		groups, err = h.store.Dashboard.GroupEngineeringHealth(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get group engineering health")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEngineeringHealthData(average, groups), nil, nil, nil, ""))
}

// GetAudits godoc
// @Summary Get Audit information for dashboard
// @Description Get Audit information for dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID   query  string false  "Project ID"
// @Success 200 {object} view.AuditResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/audits [get]
func (h *handler) GetAudits(c *gin.Context) {
	input := request.WorkSurveysInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID != "" && !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetAudits",
		"input":   input,
	})

	var average []*model.AverageAudit
	var groups []*model.GroupAudit
	var err error

	if input.ProjectID != "" {
		// Check project existence
		project, err := h.store.Project.One(h.repo.DB(), input.ProjectID, false)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "project not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
				return
			}

			l.Error(err, "failed to get project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
			return
		}

		average, err = h.store.Dashboard.GetAverageAuditByProjectNotionID(h.repo.DB(), project.NotionID.String())
		if err != nil {
			l.Error(err, "failed to get average audits")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		groups, err = h.store.Dashboard.GetGroupAuditByProjectNotionID(h.repo.DB(), project.NotionID.String())
		if err != nil {
			l.Error(err, "failed to get group audits")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	} else {
		average, err = h.store.Dashboard.GetAverageAudit(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get average audits")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		groups, err = h.store.Dashboard.GetGroupAudit(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get group audits")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuditData(average, groups), nil, nil, nil, ""))
}

// GetActionItemSquashReports godoc
// @Summary Get Action items squash report for dashboard
// @Description Get Action items squash report for dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param projectID   query  string false  "Project ID"
// @Success 200 {object} view.ActionItemSquashReportResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/action-item-squash [get]
func (h *handler) GetActionItemSquashReports(c *gin.Context) {
	input := request.ActionItemInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if input.ProjectID != "" && !model.IsUUIDFromString(input.ProjectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetActionItemSquashReports",
		"input":   input,
	})

	var actionItemSquashReports []*model.ActionItemSquashReport
	var err error

	if input.ProjectID != "" {
		// Check project existence
		isExist, err := h.store.Project.IsExist(h.repo.DB(), input.ProjectID)
		if err != nil {
			l.Error(err, "failed to get project by project ID")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		if !isExist {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
			return
		}

		// Get action item report by project ID
		actionItemSquashReports, err = h.store.Dashboard.GetActionItemSquashReportsByProjectID(h.repo.DB(), input.ProjectID)
		if err != nil {
			l.Error(err, "failed to get action item squash report by project ID")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	} else {
		actionItemSquashReports, err = h.store.Dashboard.GetAllActionItemSquashReports(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get action item squash report")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToActionItemSquashReportData(actionItemSquashReports), nil, nil, nil, ""))
}

// GetSummary godoc
// @Summary Get the summary audit info for projects
// @Description Get the summary audit info for projects
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.AuditSummariesResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/summary [get]
func (h *handler) GetSummary(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetSummary",
	})

	summaries, err := h.store.Dashboard.GetAuditSummaries(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get audit summaries")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	summaryMap := make(map[model.UUID][]*model.AuditSummary)

	for _, summary := range summaries {
		summaryMap[summary.ID] = append(summaryMap[summary.ID], summary)
	}

	now := time.Now()
	currentMonth := now.Month()
	currentYear := now.Year()
	firstDayOfLastQuarter := time.Date(currentYear, (currentMonth-1)/3*3+1, 1, 0, 0, 0, 0, time.UTC)
	previousQuarterSizes, err := h.store.Dashboard.GetProjectSizesByStartTime(h.repo.DB(), firstDayOfLastQuarter)
	if err != nil {
		l.Error(err, "failed to get project sizes")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	previousQuarterMap := make(map[model.UUID]int64)
	for _, projectSize := range previousQuarterSizes {
		previousQuarterMap[projectSize.ID] = projectSize.Size
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToAuditSummaries(summaryMap, previousQuarterMap), nil, nil, nil, ""))
}

// GetResourcesAvailability godoc
// @Summary Get resources availability
// @Description Get resources availability
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.ResourceAvailabilityResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/resources/availabilities [get]
func (h *handler) GetResourcesAvailability(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetResourcesAvailability",
	})

	slots, err := h.store.Dashboard.GetPendingSlots(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get pending slots")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	employees, err := h.store.Dashboard.GetAvailableEmployees(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get available employees")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToResourceAvailability(slots, employees), nil, nil, nil, ""))
}

// GetEngagementInfo godoc
// @Summary Get engagement dashboard
// @Description Get engagement dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.GetEngagementDashboardResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/engagement/info [get]
func (h *handler) GetEngagementInfo(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetEngagementInfo",
	})

	events, err := h.store.FeedbackEvent.GetLatestEventByType(h.repo.DB(), model.EventTypeSurvey, model.EventSubtypeEngagement, 4)
	if err != nil {
		l.Error(err, "failed to get engagement events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, nil, ""))
		return
	}

	timeList := make([]time.Time, 0)
	for _, t := range events {
		timeList = append(timeList, *t.StartDate)
	}

	statistic, err := h.store.EmployeeEventQuestion.GetAverageAnswerEngagementByTime(h.repo.DB(), timeList)
	if err != nil {
		l.Error(err, "failed to get engagement statistic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEngagementDashboard(statistic), nil, nil, nil, ""))
}

// GetEngagementInfoDetail godoc
// @Summary Get engagement dashboard
// @Description Get engagement dashboard
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param filter  query  string true  "chapter/seniority/project"
// @Param startDate  query  string true  "startDate"
// @Success 200 {object} view.GetEngagementDashboardDetailResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/engagement/detail [get]
func (h *handler) GetEngagementInfoDetail(c *gin.Context) {
	query := request.GetEngagementDashboardDetailRequest{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	filter := model.EngagementDashboardFilter(query.Filter)
	if !filter.IsValid() {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEngagementDashboardFilter, query, ""))
		return
	}

	startDate, err := time.Parse("2006-01-02", query.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidStartDate, query, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetEngagementInfoDetail",
	})

	events, err := h.store.FeedbackEvent.GetLatestEventByType(h.repo.DB(), model.EventTypeSurvey, model.EventSubtypeEngagement, 4)
	if err != nil {
		l.Error(err, "failed to get engagement events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrEventNotFound, nil, ""))
		return
	}

	timeList := make([]time.Time, 0)
	for _, t := range events {
		timeList = append(timeList, *t.StartDate)
	}

	statistic, err := h.store.EmployeeEventQuestion.GetAverageAnswerEngagementByFilter(h.repo.DB(), filter, &startDate)
	if err != nil {
		l.Error(err, "failed to get engagement statistic")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEngagementDashboardDetails(statistic), nil, nil, nil, ""))
}
