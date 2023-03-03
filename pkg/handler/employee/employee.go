package employee

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/controller/employee"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

// List godoc
// @Summary Get the list of employees
// @Description Get the list of employees with pagination and workingStatus
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.GetListEmployeeInput true "Body"
// @Success 200 {object} view.EmployeeListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/search [post]
func (h *handler) List(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	var body request.GetListEmployeeInput

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	body.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "List",
		"params":  body,
	})

	workingStatuses, err := h.getWorkingStatusInput(body.WorkingStatuses, userInfo)
	if err != nil {
		l.Error(err, "failed to get working status")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	}

	requestBody := employee.GetListEmployeeInput{
		Pagination: body.Pagination,

		WorkingStatuses: body.WorkingStatuses,
		Preload:         body.Preload,
		Positions:       body.Positions,
		Stacks:          body.Stacks,
		Projects:        body.Projects,
		Chapters:        body.Chapters,
		Seniorities:     body.Seniorities,
		Organizations:   body.Organizations,
		LineManagers:    body.LineManagers,
		Keyword:         body.Keyword,
	}

	employees, total, err := h.controller.Employee.List(workingStatuses, requestBody, userInfo)
	if err != nil {
		l.Error(err, "failed to get list employees")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEmployeeListData(employees, userInfo),
		&view.PaginationResponse{Pagination: body.Pagination, Total: total}, nil, nil, ""))
}

func (h *handler) getWorkingStatusInput(input []string, userInfo *model.CurrentLoggedUserInfo) ([]string, error) {
	// user who do not have permission
	if !authutils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadFilterByAllStatuses) {
		if len(input) == 0 {
			return []string{
				model.WorkingStatusOnBoarding.String(),
				model.WorkingStatusProbation.String(),
				model.WorkingStatusFullTime.String(),
				model.WorkingStatusContractor.String(),
			}, nil
		}

		var result []string

		for _, v := range input {
			if v != model.WorkingStatusLeft.String() {
				result = append(result, v)
			}
		}

		return result, nil
	}

	// user who have permission
	if len(input) == 0 {
		return []string{
			model.WorkingStatusOnBoarding.String(),
			model.WorkingStatusProbation.String(),
			model.WorkingStatusFullTime.String(),
			model.WorkingStatusContractor.String(),
			model.WorkingStatusLeft.String(),
		}, nil
	}

	return input, nil
}

// Details godoc
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
func (h *handler) Details(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	// 1. parse id from uri, validate id
	id := c.Param("id")

	// 1.1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "Details",
		"id":      id,
	})

	rs, err := h.controller.Employee.Details(id, userInfo)
	if err != nil {
		l.Error(err, "failed to get detail employees")
		errs.ConvertControllerErr(c, err)
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToOneEmployeeData(rs, userInfo), nil, nil, nil, ""))
}

// UpdateEmployeeStatus godoc
// @Summary Update account status by employee id
// @Description Update account status by employee id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param employeeStatus body model.WorkingStatus true "Employee Status"
// @Success 200 {object} view.UpdateEmployeeStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/employee-status [put]
func (h *handler) UpdateEmployeeStatus(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body request.UpdateWorkingStatusInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateEmployeeStatus",
		"id":      employeeID,
	})

	if err := body.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	emp, err := h.controller.Employee.UpdateEmployeeStatus(employeeID, employee.UpdateWorkingStatusInput{
		EmployeeStatus: body.EmployeeStatus,
	})
	if err != nil {
		l.Error(err, "failed to get detail employees")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(emp), nil, nil, nil, ""))
}

// UpdateGeneralInfo godoc
// @Summary Update general info of the employee by id
// @Description Update general info of the employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body request.UpdateEmployeeGeneralInfoInput true "Body"
// @Success 200 {object} view.UpdateGeneralEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/general-info [put]
func (h *handler) UpdateGeneralInfo(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body request.UpdateEmployeeGeneralInfoInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateGeneralInfo",
		"request": body,
	})

	requestBody := employee.UpdateEmployeeGeneralInfoInput{
		FullName:        body.FullName,
		Email:           body.Email,
		Phone:           body.Phone,
		LineManagerID:   body.LineManagerID,
		DisplayName:     body.DisplayName,
		GithubID:        body.GithubID,
		NotionID:        body.NotionID,
		NotionName:      body.NotionName,
		NotionEmail:     body.NotionEmail,
		DiscordID:       body.DiscordID,
		DiscordName:     body.DiscordName,
		LinkedInName:    body.LinkedInName,
		LeftDate:        body.LeftDate,
		JoinedDate:      body.JoinedDate,
		OrganizationIDs: body.OrganizationIDs,
		ReferredBy:      body.ReferredBy,
	}

	emp, err := h.controller.Employee.UpdateGeneralInfo(h.logger, employeeID, requestBody)
	if err != nil {
		l.Error(err, "failed to update general info for employee")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateGeneralInfoEmployeeData(emp), nil, nil, nil, ""))
}

// Create godoc
// @Summary Create new employee
// @Description Create new employee
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Body body request.CreateEmployeeInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees [post]
func (h *handler) Create(c *gin.Context) {
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 1. parse eml data from body
	var input request.CreateEmployeeInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// 1.1 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "Create",
		"input":   input,
	})

	requestBody := employee.CreateEmployeeInput{
		FullName:      input.FullName,
		DisplayName:   input.DisplayName,
		TeamEmail:     input.TeamEmail,
		PersonalEmail: input.PersonalEmail,
		Positions:     input.Positions,
		Salary:        input.Salary,
		SeniorityID:   input.SeniorityID,
		Roles:         input.Roles,
		Status:        input.Status,
		ReferredBy:    input.ReferredBy,
	}

	eml, err := h.controller.Employee.Create(userID, requestBody)
	if err != nil {
		l.Error(err, "failed to create employee")
		errs.ConvertControllerErr(c, err)
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(eml), nil, nil, nil, ""))
}

// UpdateSkills godoc
// @Summary Update Skill for employee by id
// @Description Update Skill for employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param Body body request.UpdateSkillsInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.UpdateSkillsEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/skills [put]
func (h *handler) UpdateSkills(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body request.UpdateSkillsInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateSkills",
		"request": body,
	})

	requestBody := employee.UpdateSkillsInput{
		Positions:       body.Positions,
		LeadingChapters: body.LeadingChapters,
		Chapters:        body.Chapters,
		Seniority:       body.Seniority,
		Stacks:          body.Stacks,
	}
	emp, err := h.controller.Employee.UpdateSkills(h.logger, employeeID, requestBody)
	if err != nil {
		l.Error(err, "failed to update skills")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateSkillEmployeeData(emp), nil, nil, nil, ""))
}

// UpdatePersonalInfo godoc
// @Summary Update personal info of the employee by id
// @Description Update personal info of the employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body request.UpdatePersonalInfoInput true "Body"
// @Success 200 {object} view.UpdatePersonalEmployeeResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/personal-info [put]
func (h *handler) UpdatePersonalInfo(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body request.UpdatePersonalInfoInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdatePersonalInfo",
		"request": body,
	})

	requestBody := employee.UpdatePersonalInfoInput{
		DoB:              body.DoB,
		Gender:           body.Gender,
		PlaceOfResidence: body.PlaceOfResidence,
		Address:          body.Address,
		PersonalEmail:    body.PersonalEmail,
		Country:          body.Country,
		City:             body.City,
	}

	if isValid := h.validateCountryAndCity(h.repo.DB(), body.Country, body.City); !isValid {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, nil, ""))
		return
	}

	emp, err := h.controller.Employee.UpdatePersonalInfo(employeeID, requestBody)
	if err != nil {
		l.Error(err, "failed to update personal info")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdatePersonalEmployeeData(emp), nil, nil, nil, ""))
}

func (h *handler) validateCountryAndCity(db *gorm.DB, countryName string, city string) bool {
	if countryName == "" && city == "" {
		return true
	}

	if countryName == "" && city != "" {
		return false
	}

	country, err := h.store.Country.OneByName(db, countryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return false
	}

	if city != "" && !slices.Contains([]string(country.Cities), city) {
		return false
	}

	return true
}

// UploadAvatar godoc
// @Summary Upload avatar of employee by id
// @Description Upload avatar of employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param Authorization header string true "jwt token"
// @Param file formData file true "avatar upload"
// @Success 200 {object} view.EmployeeContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/upload-avatar [post]
func (h *handler) UploadAvatar(c *gin.Context) {
	// 1.1 get userID
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	uuidUserID, err := model.UUIDFromString(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 1.2 parse id from uri, validate id
	var params struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params, ""))
		return
	}

	// 1.3 get upload file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, file, ""))
		return
	}

	// 1.4 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UploadAvatar",
		"params":  params,
		// "body":    body,
	})

	filePath, err := h.controller.Employee.UploadAvatar(uuidUserID, file, employee.UploadAvatarInput{
		ID: params.ID,
	})
	if err != nil {
		l.Error(err, "failed to update avatar")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(filePath), nil, nil, nil, ""))
}

// UpdateRole godoc
// @Summary Update role by employee id
// @Description Update role by employee id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param roleID body model.UUID true "Account role ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/roles [put]
func (h *handler) UpdateRole(c *gin.Context) {
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var input request.UpdateRoleInput

	input.EmployeeID = c.Param("id")
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UpdateRole",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	inputRequest := employee.UpdateRoleInput{
		EmployeeID: input.EmployeeID,
		Body: employee.UpdateRoleBody{
			Roles: input.Body.Roles,
		},
	}

	err = h.controller.Employee.UpdateRole(userID, inputRequest)
	if err != nil {
		l.Error(err, "failed to update role")
		errs.ConvertControllerErr(c, err)
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// GetLineManagers godoc
// @Summary Get the list of line managers
// @Description Get the list of line managers
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.LineManagersResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /line-managers [get]
func (h *handler) GetLineManagers(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler":  "employee",
		"method":   "GetLineManagers",
		"userInfo": userInfo.UserID,
	})

	managers, err := h.controller.Employee.GetLineManagers(userInfo)
	if err != nil {
		l.Error(err, "failed to get line managers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToBasicEmployees(managers), nil, nil, nil, ""))
}
