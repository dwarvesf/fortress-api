package dashboard

import (
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

// WorkUnitDistribution godoc
// @Summary Get work unit distribution for dashboard
// @Description Get work unit distribution for dashboard
// @Tags Dashboard
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param sortRequired query model.SortOrder true "Sort type"
// @Success 200 {object} view.WorkUnitDistributionResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /dashboards//work-unit-distribution [get]
func (h *handler) WorkUnitDistribution(c *gin.Context) {
	input := request.WorkUnitDistributionInput{}
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "dashboard",
		"method":  "WorkUnitDistribution",
		"input":   input,
	})

	if input.SortRequired != "" && !input.SortRequired.IsValid() {
		l.Error(errs.ErrInvalidSortRequiredValue, "invalid value for sortRequired")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidSortRequiredValue, input, ""))
		return
	}

	// get all employee fulltime
	employees, err := h.store.Employee.GetByWorkingStatus(h.repo.DB(), model.WorkingStatusFullTime, true)
	if err != nil {
		l.Error(err, "failed to get employees by working status")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](http.StatusInternalServerError, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToWorkUnitDistributionDataList(employees, input.SortRequired), nil, nil, nil, ""))
}
