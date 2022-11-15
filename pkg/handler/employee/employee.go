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
	repo    store.DBRepo
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
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
// @Param       workingStatus   query  string false  "Working Status"
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

	employees, total, err := h.store.Employee.Search(h.repo.DB(), employee.SearchFilter{
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
// @Param Authorization header string true "jwt token"
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
	rs, err := h.store.Employee.One(h.repo.DB(), params.ID)
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
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param employeeStatus body model.AccountStatus true "Employee Status"
// @Success 200 {object} view.UpdataEmployeeStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/employee-status [put]
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
	rs, err := h.store.Employee.UpdateEmployeeStatus(h.repo.DB(), params.ID, body.EmployeeStatus)
	if err != nil {
		l.Error(err, "error query update account status employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, params))
		return
	}

	// 3. return status reonse
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(rs), nil, nil, nil))
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

	rs, err := h.store.Employee.One(h.repo.DB(), userID)
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

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToProfileData(rs), nil, nil, nil))
}

// UpdateGeneralInfo godoc
// @Summary Update general info of the employee by id
// @Description Update general info of the employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body UpdateGeneralInfoInput true "Body"
// @Success 200 {object} view.UpdateGeneralEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/general-info [put]
func (h *handler) UpdateGeneralInfo(c *gin.Context) {
	employeeID := c.Param("id")

	var body UpdateGeneralInfoInput
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
		_, err := h.store.Employee.One(h.repo.DB(), body.LineManagerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(ErrLineManagerNotFound, "error line manager not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrLineManagerNotFound, body))
				return
			}
			l.Error(err, "error when finding line manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
			return
		}
	}

	// 3. update information and return
	rs, err := h.store.Employee.UpdateGeneralInfo(h.repo.DB(), employee.UpdateGeneralInfoInput{
		FullName:      body.Fullname,
		Email:         body.Email,
		Phone:         body.Phone,
		LineManagerID: body.LineManagerID,
		DiscordID:     body.DiscordID,
		GithubID:      body.GithubID,
		NotionID:      body.NotionID,
	}, employeeID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrEmployeeNotFound, "error employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, body))
			return
		}

		l.Error(err, "error update employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateGeneralInfoEmployeeData(rs), nil, nil, nil))
}

// Create godoc
// @Summary Create new employee
// @Description Create new employee
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Body body CreateEmployee true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employee [post]
func (h *handler) Create(c *gin.Context) {
	// 1. parse eml data from body
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

	// 1.2 prepare eml data
	now := time.Now()

	pos, err := h.store.Position.One(h.repo.DB(), req.PositionID)
	if err != nil {
		l.Error(err, "error invalid position")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrPositionNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	sen, err := h.store.Seniority.One(h.repo.DB(), req.SeniorityID)
	if err != nil {
		l.Error(err, "error invalid seniority")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrSeniorityNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	role, err := h.store.Role.One(h.repo.DB(), req.RoleID)
	if err != nil {
		l.Error(err, "error invalid role")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrRoleNotfound, req))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	eml := &model.Employee{
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
	_, err = h.store.Employee.OneByTeamEmail(h.repo.DB(), eml.TeamEmail)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "error eml exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrEmployeeExisted, req))
			return
		}
		l.Error(err, "error store new eml")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 2.2 store employee
	eml, err = h.store.Employee.Create(h.repo.DB(), eml)
	if err != nil {
		l.Error(err, "error store new eml")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(eml), nil, nil, nil))
}

// UpdateSkills godoc
// @Summary Update Skill for employee by id
// @Description Update Skill for employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param Body body UpdateSkillsInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.UpdateSkillsEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/skills [put]
func (h *handler) UpdateSkills(c *gin.Context) {
	// 1. parse id from uri, validate id
	employeeID := c.Param("id")

	// 2. parse json body from request
	var body UpdateSkillsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, employeeID))
			return
		}
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateSkills",
		"request": body,
	})

	// 3. update info and return
	rs, err := h.store.Employee.UpdateSkills(h.repo.DB(), employee.UpdateSkillsInput{
		Positions: body.Positions,
		Chapter:   body.Chapter,
		Seniority: body.Seniority,
		Stacks:    body.Stacks,
	}, employeeID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrEmployeeNotFound, "error employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, body))
			return
		}

		l.Error(err, "error update employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateSkillEmployeeData(rs), nil, nil, nil))
}

// UpdatePersonalInfo godoc
// @Summary Update personal info of the employee by id
// @Description Update personal info of the employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body UpdatePersonalInfoInput true "Body"
// @Success 200 {object} view.UpdatePersonalEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/personal-info [put]
func (h *handler) UpdatePersonalInfo(c *gin.Context) {
	employeeID := c.Param("id")

	var body UpdatePersonalInfoInput
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

	// 3. update informations and rerurn
	rs, err := h.store.Employee.UpdatePersonalInfo(h.repo.DB(), employee.UpdatePersonalInfoInput{
		DoB:           body.DoB,
		Gender:        body.Gender,
		Address:       body.Address,
		PersonalEmail: body.PersonalEmail,
	}, employeeID)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrEmployeeNotFound, "error employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, body))
			return
		}

		l.Error(err, "error update employee to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdatePersonalEmployeeData(rs), nil, nil, nil))
}
