package employee

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
}

func New(store *store.Store, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
		service: service,
		logger:  logger,
	}
}

func (h *handler) List(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "employee list"})
}

// One godoc
// @Summary Get employee by id
// @Description Get employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Success 200 {object} view.EmployeeListData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employee/{id} [get]
func (h *handler) One(c *gin.Context) {
	// 1. parse id from uri, validate id
	var params struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	// 1.1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "One",
		"params":  params,
	})

	// 2. get employee from store
	employee, err := h.store.Employee.One(params.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, params))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeListData(employee), nil, nil, nil))
}
