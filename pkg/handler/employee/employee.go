package employee

import (
	"net/http"

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

// Update godoc
// @Summary Edit employee by id
// @Description Edit employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param fullName body string true "fullName" maxlength(99)
// @Param email body string true "email"
// @Param phone body string true "phone" minlength(10)  maxlength(12)
// @Param lineManager body model.UUID true "lineManager"
// @Param discordID body string true "discordID"
// @Param githubID body string true "githubID"
// @Param role body model.UUID true "role"
// @Param chapter body model.UUID true "chapter"
// @Param seniority body model.UUID  true "seniority"
// @Param stack body []model.UUID true "stack"
// @Param DoB body time.Time true "DoB"
// @Param gender body string true "gender"
// @Param address body string true "address" maxlength(200)
// @Param personalEmail body string true "personalEmail"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EditEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/edit [post]
func (h *handler) Update(c *gin.Context) {
	// 1. parse id from uri, validate id
	employeeID := c.Param("id")
	requestType := c.Param("type")

	// 2. parse json body from request
	switch requestType {
	case "edit-general-info":
		var body EditGeneralInfo
		if err := c.ShouldBindJSON(&body); err != nil {
			if err != nil {
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, requestType))
				return
			}
		}

		// TODO: can we move this to middleware ?
		l := h.logger.Fields(logger.Fields{
			"handler": "employee",
			"method":  "Update",
			"request": body,
		})

		// 3. update informations and rerurn
		employee, err := h.store.Employee.UpdateGeneralInfo(employee.EditGeneralInfo{
			Fullname:    body.Fullname,
			Email:       body.Email,
			Phone:       body.Phone,
			LineManager: body.LineManager,
			DiscordID:   body.DiscordID,
			GithubID:    body.GithubID,
		}, employeeID)

		if err != nil {
			l.Error(err, "error update employee to db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
			return
		}
		c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))

	case "edit-skills":
		var body EditSkills
		if err := c.ShouldBindJSON(&body); err != nil {
			if err != nil {
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, requestType))
				return
			}
		}

		// TODO: can we move this to middleware ?
		l := h.logger.Fields(logger.Fields{
			"handler": "employee",
			"method":  "Update",
			"request": body,
		})

		// 3. update informations and rerurn
		employee, err := h.store.Employee.UpdateSkills(employee.EditSkills{
			Role:      body.Role,
			Chapter:   body.Chapter,
			Seniority: body.Seniority,
			Stack:     body.Stack,
		}, employeeID)

		if err != nil {
			l.Error(err, "error update employee to db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
			return
		}

		c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))
	case "edit-personal-info":
		var body EditPersonalInfo
		if err := c.ShouldBindJSON(&body); err != nil {
			if err != nil {
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, requestType))
				return
			}
		}

		// TODO: can we move this to middleware ?
		l := h.logger.Fields(logger.Fields{
			"handler": "employee",
			"method":  "Update",
			"request": body,
		})

		// 3. update informations and rerurn
		employee, err := h.store.Employee.UpdatePersonalInfo(employee.EditPersonalInfo{
			DoB:           body.DoB,
			Gender:        body.Gender,
			Address:       body.Address,
			PersonalEmail: body.PersonalEmail,
		}, employeeID)

		if err != nil {
			l.Error(err, "error update employee to db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
			return
		}

		c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil))
	default:
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEditType, requestType))
		return
	}
}
