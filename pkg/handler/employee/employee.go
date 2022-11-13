package employee

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
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

// List godoc
// @Summary Get the list of employees
// @Description Get the list of employees with pagination and workingStatus
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param       workingStatus   query  string true  "Working Status"
// @Param       page   query  string true  "Page"
// @Param       size   query  string true  "Size"
// @Success 200 {object} view.EmployeeListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees [get]
func (h *handler) List(c *gin.Context) {
	query := GetListEmployeeQuery{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query))
		return
	}
	query.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "List",
		"params":  query,
	})

	employees, total, err := h.store.Employee.Search(employee.SearchFilter{
		WorkingStatus: query.WorkingStatus,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEmployeeListData(employees),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil))
}

// One godoc
// @Summary Get employee by id
// @Description Get employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id} [get]
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
	rs, err := h.store.Employee.One(params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, params))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(rs), nil, nil, nil))
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
// @Success 200 {object} view.UpdateEmployeeStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/employee-status [post]
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
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))
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
// @Router /profile [get]
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// UpdateGeneralInfo godoc
// @Summary Edit general info of the employee by id
// @Description Edit general info of the employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param fullName body string true "fullName" maxlength(99)
// @Param email body string true "email"
// @Param phone body string true "phone" minlength(10)  maxlength(12)
// @Param lineManagerID body string true "lineManager"
// @Param discordID body string true "discordID"
// @Param githubID body string true "githubID"
// @Param notionID body string true "notionID"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EditEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/general-info [put]
func (h *handler) UpdateGeneralInfo(c *gin.Context) {
	employeeID := c.Param("id")

	var body EditGeneralInfo
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, employeeID))
			return
		}
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateGeneralInfo",
		"request": body,
	})

	// check line manager existence
	if body.LineManagerID != "" {
		_, err := h.store.Employee.One(body.LineManagerID)
		if err != nil {
			l.Error(ErrCantFindLineManager, "error when finding line manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, ErrCantFindLineManager, body))
			return
		}
	}

	// 3. update informations and rerurn
	employee, err := h.store.Employee.UpdateGeneralInfo(employee.EditGeneralInfo{
		Fullname:      body.Fullname,
		Email:         body.Email,
		Phone:         body.Phone,
		LineManagerID: body.LineManagerID,
		DiscordID:     body.DiscordID,
		GithubID:      body.GithubID,
		NotionID:      body.NotionID,
	}, employeeID)

	if err != nil {
		l.Error(err, "error update employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))
}

// Create godoc
// @Summary Create new employee
// @Description Create new employee
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param fullName body string true "fullName" maxlength(99)
// @Param displayName body string true "displayName"
// @Param teamEmail body string true "teamEmail"
// @Param personalEmail body string true "personalEmail"
// @Param positionID body string true "positionID"
// @Param roleID body string true "roleID"
// @Param salary body int true "salary"
// @Param seniorityID body string true "seniorityID"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employee [post]
func (h *handler) Create(c *gin.Context) {
	// 1. parse employee data from body
	var req CreateEmployee

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 1.1 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "One",
		"params":  req,
	})

	// 1.2 prepare employee data
	now := time.Now()
	pos, err := h.store.Position.One(req.PositionID)
	if err != nil {
		l.Error(err, "error invalid position")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrPositionNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}
	sen, err := h.store.Seniority.One(req.SeniorityID)
	if err != nil {
		l.Error(err, "error invalid seniority")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrSeniorityNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}
	role, err := h.store.Role.One(req.RoleID)
	if err != nil {
		l.Error(err, "error invalid role")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrRoleNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}
	employee := &model.Employee{
		BaseModel: model.BaseModel{
			ID: model.NewUUID(),
		},
		FullName:      req.FullName,
		DisplayName:   req.DisplayName,
		TeamEmail:     req.TeamEmail,
		PersonalEmail: req.PersonalEmail,
		WorkingStatus: model.WorkingStatusProbation,
		JoinedDate:    &now,
		AccountStatus: model.AccountStatusOnBoarding,
		SeniorityID:   sen.ID,
		Positions:     []model.Position{*pos},
		Roles:         []model.Role{*role},
	}

	// 2.1 check employee exists -> raise error
	_, err = h.store.Employee.OneByTeamEmail(employee.TeamEmail)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "error employee exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrEmployeeExisted, req))
			return
		}
		l.Error(err, "error store new employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}
	// 2.2 store employee
	employee, err = h.store.Employee.Create(employee)
	if err != nil {
		l.Error(err, "error store new employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))
}
