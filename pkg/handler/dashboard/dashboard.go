package dashboard

import (
	"fmt"
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

// GetResourceUtilization godoc
// @Summary Get dashboard resource utilization
// @Description Get dashboard resource utilization
// @Tags Dashboard
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.GetDashboardResourceUtilizationResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards/resources/utilization [get]
func (h *handler) GetResourceUtilization(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "GetResourceUtilization",
	})

	res, err := h.store.Dashboard.GetResourceUtilizationByYear(h.repo.DB(), time.Now().Year())
	if err != nil {
		l.Error(err, "failed to get resource utilization by year")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
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

	fmt.Println(startDate)

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
