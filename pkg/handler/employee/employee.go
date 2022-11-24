package employee

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/view"
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

// List godoc
// @Summary Get the list of employees
// @Description Get the list of employees with pagination and workingStatus
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param workingStatus query  string false  "Working Status"
// @Param preload query bool false "Preload"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Success 200 {object} view.EmployeeListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees [get]
func (h *handler) List(c *gin.Context) {
	query := GetListEmployeeQuery{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	if err := query.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	query.Standardize()

	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "List",
		"params":  query,
	})

	employees, total, err := h.store.Employee.All(h.repo.DB(), employee.GetAllInput{
		WorkingStatus: query.WorkingStatus,
		Preload:       query.Preload,
		Keyword:       query.Keyword,
		PositionID:    query.PositionID,
		StackID:       query.StackID,
		ProjectID:     query.ProjectID,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEmployeeListData(employees),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil, ""))
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params, ""))
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
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, params, ""))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, params, ""))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(rs), nil, nil, nil, ""))
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body UpdateWorkingStatusInput
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	employee.WorkingStatus = body.EmployeeStatus
	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employee, "working_status")
	if err != nil {
		l.Error(err, "failed to update employee status")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(employee), nil, nil, nil, ""))
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
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body UpdateGeneralInfoInput
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

	// check line manager existence
	if !body.LineManagerID.IsZero() {
		exist, err := h.store.Employee.IsExist(h.repo.DB(), body.LineManagerID.String())
		if err != nil {
			l.Error(err, "error when finding line manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}

		if !exist {
			l.Error(ErrLineManagerNotFound, "error line manager not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrLineManagerNotFound, body, ""))
			return
		}
	}

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3. update information and return
	employee.FullName = body.FullName
	employee.PersonalEmail = body.Email
	employee.PhoneNumber = body.Phone
	employee.LineManagerID = body.LineManagerID
	employee.DiscordID = body.DiscordID
	employee.GithubID = body.GithubID
	employee.NotionID = body.NotionID

	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employee,
		"full_name",
		"personal_email",
		"phone_number",
		"line_manager_id",
		"discord_id",
		"github_id",
		"notion_id")
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateGeneralInfoEmployeeData(employee), nil, nil, nil, ""))
}

// Create godoc
// @Summary Create new employee
// @Description Create new employee
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Body body CreateEmployeeInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees [post]
func (h *handler) Create(c *gin.Context) {
	// 1. parse eml data from body
	var req CreateEmployeeInput

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	if !model.WorkingStatus(req.Status).IsValid() {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeStatus, req, ""))
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

	// Check position existence
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error when finding position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	positionsReq := make([]model.Position, 0)
	positionMap := model.ToPositionMap(positions)
	for _, pID := range req.Positions {
		_, ok := positionMap[pID]
		if !ok {
			l.Error(errPositionNotFound(pID.String()), "error position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errPositionNotFound(pID.String()), req, ""))
			return
		}

		positionsReq = append(positionsReq, positionMap[pID])
	}

	sen, err := h.store.Seniority.One(h.repo.DB(), req.SeniorityID)
	if err != nil {
		l.Error(err, "error invalid seniority")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrSeniorityNotfound, req, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	role, err := h.store.Role.One(h.repo.DB(), req.RoleID)
	if err != nil {
		l.Error(err, "error invalid role")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrRoleNotfound, req, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
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
		WorkingStatus: model.WorkingStatus(req.Status),
		JoinedDate:    &now,
		SeniorityID:   sen.ID,
	}

	// 2.1 check employee exists -> raise error
	_, err = h.store.Employee.OneByTeamEmail(h.repo.DB(), eml.TeamEmail)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "error eml exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrEmployeeExisted, req, ""))
			return
		}
		l.Error(err, "error store new eml")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}
	tx, done := h.repo.NewTransaction()
	// 2.2 store employee
	eml, err = h.store.Employee.Create(tx.DB(), eml)
	if err != nil {
		l.Error(err, "failed to create new employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), req, ""))
		return
	}

	// 2.3 create employee position
	for _, p := range positionsReq {
		_, err = h.store.EmployeePosition.Create(tx.DB(), &model.EmployeePosition{
			EmployeeID: eml.ID,
			PositionID: p.ID,
		})
		if err != nil {
			l.Error(err, "failed to create new employee position")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), req, ""))
			return
		}
	}

	// 2.4 create employee role
	_, err = h.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
		EmployeeID: eml.ID,
		RoleID:     role.ID,
	})
	if err != nil {
		l.Error(err, "failed to create new employee position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), req, ""))
		return
	}

	// 3. return employee
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToEmployeeData(eml), nil, done(nil), nil, ""))
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
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body UpdateSkillsInput
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Check chapter existence
	exist, err := h.store.Chapter.IsExist(h.repo.DB(), body.Chapter.String())
	if err != nil {
		l.Error(err, "failed to check chapter existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(ErrChapterNotFound, "chapter not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrChapterNotFound, body, ""))
		return
	}

	// Check seniority existence
	exist, err = h.store.Seniority.IsExist(h.repo.DB(), body.Seniority.String())
	if err != nil {
		l.Error(err, "failed to check seniority existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(ErrSeniorityNotFound, "seniority not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrSeniorityNotFound, body, ""))
		return
	}

	// Check stack existence
	stacks, err := h.store.Stack.All(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get all stacks")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	stackMap := model.ToStackMap(stacks)
	for _, sID := range body.Stacks {
		_, ok := stackMap[sID]
		if !ok {
			l.Error(errPositionNotFound(sID.String()), "stack not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errPositionNotFound(sID.String()), body, ""))
			return
		}
	}

	// Check position existence
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get all positions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range body.Positions {
		_, ok := positionMap[pID]

		if !ok {
			l.Error(errPositionNotFound(pID.String()), "position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errPositionNotFound(pID.String()), body, ""))
			return
		}
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Delete all exist employee positions
	if err := h.store.EmployeePosition.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		l.Error(err, "failed to delete employee position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Create new employee position
	for _, positionID := range body.Positions {
		_, err := h.store.EmployeePosition.Create(tx.DB(), &model.EmployeePosition{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			PositionID: positionID,
		})
		if err != nil {
			l.Error(err, "failed to create employee position")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// Delete all exist employee stack
	if err := h.store.EmployeeStack.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		l.Error(err, "failed to delete employee stack in database")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Create new employee stack
	for _, stackID := range body.Stacks {
		_, err := h.store.EmployeeStack.Create(tx.DB(), &model.EmployeeStack{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			StackID:    stackID,
		})
		if err != nil {
			l.Error(err, "failed to create employee stack")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// Update employee information
	employee.ChapterID = body.Chapter
	employee.SeniorityID = body.Seniority

	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *employee, "chapter_id", "seniority_id")
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateSkillEmployeeData(employee), nil, done(nil), nil, ""))
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
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidEmployeeID, nil, ""))
		return
	}

	var body UpdatePersonalInfoInput
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	employee.DateOfBirth = body.DoB
	employee.Gender = body.Gender
	employee.Address = body.Address
	employee.PersonalEmail = body.PersonalEmail

	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employee,
		"date_of_birth",
		"gender",
		"address",
		"personal_email")
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdatePersonalEmployeeData(employee), nil, nil, nil, ""))
}

// UploadContent godoc
// @Summary Upload content of employee by id
// @Description Upload content of employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/upload-content [post]
func (h *handler) UploadContent(c *gin.Context) {
	// 1.1 parse id from uri, validate id
	var params struct {
		ID string `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params, ""))
		return
	}

	// 1.2 get upload file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, file, ""))
		return
	}

	// 1.3 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UploadContent",
		"params":  params,
		// "body":    body,
	})

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePrePath := "employees/" + params.ID
	fileType := ""

	// 2.1 validate
	if !fileExtension.Valid() {
		l.Info("invalid file extension")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrInvalidFileExtension, nil, ""))
		return
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrInvalidFileSize, nil, ""))
			return
		}
		filePrePath = filePrePath + "/images"
		fileType = "image"
	}
	if fileExtension == model.ContentExtensionPdf {
		if fileSize > model.MaxFileSizePdf {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrInvalidFileSize, nil, ""))
			return
		}
		filePrePath = filePrePath + "/docs"
		fileType = "document"
	}

	tx, done := h.repo.NewTransaction()

	// 2.2 check file name exist
	_, err = h.store.Content.GetByPath(tx.DB(), filePrePath+"/"+fileName)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Error(err, "error query content from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	if err == nil {
		l.Info("file already existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, ErrFileAlreadyExisted, nil, ""))
		done(ErrFileAlreadyExisted)
		return
	}

	// 2.3 check employee existed
	employee, err := h.store.Employee.One(tx.DB(), params.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			done(err)
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	content, err := h.store.Content.Create(tx.DB(), model.Content{
		Type:       fileType,
		Extension:  fileExtension.String(),
		Path:       filePrePath + "/" + fileName,
		EmployeeID: employee.ID,
		UploadBy:   employee.ID,
	})
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	multipart, err := file.Open()
	if err != nil {
		l.Error(err, "error in open file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	// 3. Upload to GCS
	err = h.service.Google.UploadContentGCS(multipart, content.Path)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	done(nil)

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.config.Google.GCSBucketName, content.Path)), nil, nil, nil, ""))
}
