package dashboard

import (
	"errors"
	"net/http"

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

// ProjectSizes godoc
// @Summary Get the total number of active member in each project
// @Description Get the total number of active member in each project
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.ProjectSizeResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/projects/sizes [get]
func (h *handler) ProjectSizes(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "ProjectSizes",
	})

	res, err := h.store.Dashboard.GetProjectSizes(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get project sizes")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
}

// WorkSurveys godoc
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
// @Router /dashboards/work-surveys [get]
func (h *handler) WorkSurveys(c *gin.Context) {
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
		"method":  "WorkSurveys",
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
// @Success 200 {object} view.ActionItemReportData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/action-items [get]
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
