package employee

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
}

// New returns a handler
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

// UpdateEmployeeStatus godoc
// @Summary Update account status by employee id
// @Description Update account status by employee id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param employeeStatus body model.AccountStatus true "Employee Status"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.UpdataEmployeeStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employee/{id}/employee-status [post]
func (h *handler) UpdateEmployeeStatus(c *gin.Context) {
	// 1. parse id from uri, validate id
	var params struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	type updateAccountStatusBody struct {
		EmployeeStatus model.AccountStatus `json:"employeeStatus"`
	}

	// 1.1 get body request
	var body updateAccountStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body))
			return
		}
	}

	// 1.2 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateEmployeeStatus",
		"params":  params,
	})

	if !body.EmployeeStatus.Valid() {
		l.Error(ErrInvalidEmployeeStatus, "invalid value for EmployeeStatus")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeStatus, body))
		return
	}

	// 2. get update account status for employee
	employee, err := h.store.Employee.UpdateEmployeeStatus(params.ID, body.EmployeeStatus)
	if err != nil {
		l.Error(err, "error query update account status employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	// 3. return status reonse
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeListData(employee), nil, nil, nil))
}

// GetProfile godoc
// @Summary Get profile information of employee
// @Description Get profile information of employee
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.ProfileDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employee/profile [get]
func (h *handler) GetProfile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "GetProfile",
	})

	employee, err := h.store.Employee.One(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToProfileData(employee), nil, nil, nil))
}
