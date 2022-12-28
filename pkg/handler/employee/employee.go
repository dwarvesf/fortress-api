package employee

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/request"
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
// @Param Body body request.GetListEmployeeInput true "Body"
// @Success 200 {object} view.EmployeeListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/search [post]
func (h *handler) List(c *gin.Context) {
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

	employees, total, err := h.store.Employee.All(h.repo.DB(), employee.EmployeeFilter{
		WorkingStatuses: body.WorkingStatuses,
		Preload:         body.Preload,
		Keyword:         body.Keyword,
		Positions:       body.Positions,
		Stacks:          body.Stacks,
		Projects:        body.Projects,
		Chapters:        body.Chapters,
		Seniorities:     body.Seniorities,
		JoinedDateSort:  model.SortOrderDESC,
	}, body.Pagination)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEmployeeListData(employees),
		&view.PaginationResponse{Pagination: body.Pagination, Total: total}, nil, nil, ""))
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
	id := c.Param("id")

	// 1.1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "One",
		"id":      id,
	})

	// 2. get employee from store
	var rs *model.Employee
	var err error

	if model.IsUUIDFromString(id) {
		rs, err = h.store.Employee.One(h.repo.DB(), id, true)
	} else {
		rs, err = h.store.Employee.OneByUsername(h.repo.DB(), id, true)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, id, ""))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, id, ""))
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
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

	// check line manager existence
	if !body.LineManagerID.IsZero() {
		exist, err := h.store.Employee.IsExist(h.repo.DB(), body.LineManagerID.String())
		if err != nil {
			l.Error(err, "error when finding line manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}

		if !exist {
			l.Error(errs.ErrLineManagerNotFound, "error line manager not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrLineManagerNotFound, body, ""))
			return
		}
	}

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3. update information and return

	if strings.TrimSpace(body.FullName) != "" {
		employee.FullName = body.FullName
	}

	if strings.TrimSpace(body.Email) != "" {
		employee.TeamEmail = body.Email
	}

	if strings.TrimSpace(body.Phone) != "" {
		employee.PhoneNumber = body.Phone
	}

	if strings.TrimSpace(body.GithubID) != "" {
		employee.GithubID = body.GithubID
	}

	if strings.TrimSpace(body.NotionID) != "" {
		employee.NotionID = body.NotionID
	}

	if strings.TrimSpace(body.NotionName) != "" {
		employee.NotionName = body.NotionName
	}

	if strings.TrimSpace(body.NotionEmail) != "" {
		employee.NotionEmail = body.NotionEmail
	}

	if strings.TrimSpace(body.DiscordID) != "" {
		employee.DiscordID = body.DiscordID
	}

	if strings.TrimSpace(body.DiscordName) != "" {
		employee.DiscordName = body.DiscordName
	}

	if strings.TrimSpace(body.LinkedInName) != "" {
		employee.LinkedInName = body.LinkedInName
	}

	employee.LineManagerID = body.LineManagerID

	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employee,
		"full_name",
		"team_email",
		"phone_number",
		"line_manager_id",
		"discord_id",
		"discord_name",
		"github_id",
		"notion_id",
		"notion_name",
		"notion_email",
		"linkedin_name",
	)
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
// @Param Body body request.CreateEmployeeInput true "Body"
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeData
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees [post]
func (h *handler) Create(c *gin.Context) {
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

	// 1.2 prepare employee data
	now := time.Now()

	// Check position existence
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error when finding position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	positionsReq := make([]model.Position, 0)
	positionMap := model.ToPositionMap(positions)
	for _, pID := range input.Positions {
		_, ok := positionMap[pID]
		if !ok {
			l.Error(errs.ErrPositionNotFoundWithID(pID.String()), "error position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrPositionNotFoundWithID(pID.String()), input, ""))
			return
		}

		positionsReq = append(positionsReq, positionMap[pID])
	}

	sen, err := h.store.Seniority.One(h.repo.DB(), input.SeniorityID)
	if err != nil {
		l.Error(err, "error invalid seniority")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrSeniorityNotfound, input, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	role, err := h.store.Role.One(h.repo.DB(), input.RoleID)
	if err != nil {
		l.Error(err, "error invalid role")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrRoleNotfound, input, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// get the username
	eml := &model.Employee{
		BaseModel: model.BaseModel{
			ID: model.NewUUID(),
		},
		FullName:      input.FullName,
		DisplayName:   input.DisplayName,
		TeamEmail:     input.TeamEmail,
		PersonalEmail: input.PersonalEmail,
		WorkingStatus: model.WorkingStatus(input.Status),
		JoinedDate:    &now,
		SeniorityID:   sen.ID,
		Username:      strings.Split(input.TeamEmail, "@")[0],
	}

	// 2.1 check employee exists -> raise error
	_, err = h.store.Employee.OneByTeamEmail(h.repo.DB(), eml.TeamEmail)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "error eml exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmployeeExisted, input, ""))
			return
		}
		l.Error(err, "error store new eml")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	_, err = h.store.Employee.OneByUsername(h.repo.DB(), eml.Username, false)
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "username exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmployeeExisted, input, ""))
			return
		}
		l.Error(err, "failed to check username existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()
	// 2.2 store employee
	eml, err = h.store.Employee.Create(tx.DB(), eml)
	if err != nil {
		l.Error(err, "failed to create new employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
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
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
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
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Check chapter existence
	chapters, err := h.store.Chapter.All(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get all chapters")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	chapterMap := model.ToChapterMap(chapters)
	for _, sID := range body.Chapters {
		_, ok := chapterMap[sID]
		if !ok {
			l.Error(errs.ErrChapterNotFoundWithID(sID.String()), "chapter not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrChapterNotFoundWithID(sID.String()), body, ""))
			return
		}
	}

	// Check seniority existence
	exist, err := h.store.Seniority.IsExist(h.repo.DB(), body.Seniority.String())
	if err != nil {
		l.Error(err, "failed to check seniority existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(errs.ErrSeniorityNotFound, "seniority not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrSeniorityNotFound, body, ""))
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
			l.Error(errs.ErrStackNotFoundWithID(sID.String()), "stack not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrStackNotFoundWithID(sID.String()), body, ""))
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
			l.Error(errs.ErrPositionNotFoundWithID(pID.String()), "position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrPositionNotFoundWithID(pID.String()), body, ""))
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

	// Delete all exist employee stack
	if err := h.store.EmployeeChapter.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
		l.Error(err, "failed to delete employee chapter in database")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Create new employee stack
	for _, chapterID := range body.Chapters {
		_, err := h.store.EmployeeChapter.Create(tx.DB(), &model.EmployeeChapter{
			EmployeeID: model.MustGetUUIDFromString(employeeID),
			ChapterID:  chapterID,
		})
		if err != nil {
			l.Error(err, "failed to create employee chapter")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// Remove all chapter lead by employee
	leadingChapters, err := h.store.Chapter.GetAllByLeadID(tx.DB(), employeeID)
	if err != nil {
		l.Error(err, "failed to get list chapter lead by the employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	for _, lChapter := range leadingChapters {
		if err := h.store.Chapter.UpdateChapterLead(tx.DB(), lChapter.ID.String(), nil); err != nil {
			l.Error(err, "failed to remove chapter lead")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// Create new chapter
	leader := model.MustGetUUIDFromString(employeeID)
	for _, lChapter := range body.LeadingChapters {
		if err := h.store.Chapter.UpdateChapterLead(tx.DB(), lChapter.String(), &leader); err != nil {
			l.Error(err, "failed to remove chapter lead")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// Update employee information
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

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	employee.DateOfBirth = body.DoB
	employee.Gender = body.Gender
	employee.Address = body.Address
	employee.PlaceOfResidence = body.PlaceOfResidence
	employee.PersonalEmail = body.PersonalEmail
	employee.Country = body.Country
	employee.City = body.City

	employee, err = h.store.Employee.Update(h.repo.DB(), employee)
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
// @Param file formData file true "content upload"
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
	filePath := "employees/" + params.ID
	fileType := ""

	// 2.1 validate
	if !fileExtension.Valid() {
		l.Info("invalid file extension")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
			return
		}
		filePath = filePath + "/images"
		fileType = "image"
	}
	if fileExtension == model.ContentExtensionPdf {
		if fileSize > model.MaxFileSizePdf {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
			return
		}
		filePath = filePath + "/docs"
		fileType = "document"
	}
	filePath = filePath + "/" + fileName

	tx, done := h.repo.NewTransaction()

	// 2.2 check file name exist
	_, err = h.store.Content.OneByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Error(err, "error query content from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	if err == nil {
		l.Info("file already existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrFileAlreadyExisted, nil, ""))
		done(errs.ErrFileAlreadyExisted)
		return
	}

	// 2.3 check employee existed
	employee, err := h.store.Employee.One(tx.DB(), params.ID, false)
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
		Path:       fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.config.Google.GCSBucketName, filePath),
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
	err = h.service.Google.UploadContentGCS(multipart, filePath)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	done(nil)

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(content.Path), nil, nil, nil, ""))
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
	userID, err := utils.GetUserIDFromContext(c)
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

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	fileType := "image"
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/employees/%s/images/%s", h.config.Google.GCSBucketName, params.ID, fileName)
	gcsPath := fmt.Sprintf("employees/%s/images/%s", params.ID, fileName)

	// 2.1 validate
	if !fileExtension.ImageValid() {
		l.Info("invalid file extension")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
			return
		}
	}

	tx, done := h.repo.NewTransaction()

	// 2.2 check employee existed
	employee, err := h.store.Employee.One(tx.DB(), params.ID, false)
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

	// 2.3 check file name exist
	_, err = h.store.Content.OneByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Error(err, "error query content from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		// not found => create and upload content to GCS
		_, err = h.store.Content.Create(tx.DB(), model.Content{
			Type:       fileType,
			Extension:  fileExtension.String(),
			Path:       filePath,
			EmployeeID: employee.ID,
			UploadBy:   uuidUserID,
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

		err = h.service.Google.UploadContentGCS(multipart, gcsPath)
		if err != nil {
			l.Error(err, "error in upload file")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			done(err)
			return
		}
	}

	// 3. update avatar field
	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employee.ID.String(), model.Employee{
		Avatar: filePath,
	}, "avatar")
	if err != nil {
		l.Error(err, "error in update avatar")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	done(nil)

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(filePath), nil, nil, nil, ""))
}

// AddMentee godoc
// @Summary Add mentee for a mentor
// @Description Add mentee for a mentor
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body request.AddMenteeInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/mentees [post]
func (h *handler) AddMentee(c *gin.Context) {
	employeeID := c.Param("id")
	if employeeID == "" || !model.IsUUIDFromString(employeeID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidEmployeeID, nil, ""))
		return
	}

	var input request.AddMenteeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "AddMentee",
		"input":   input,
	})

	// Check employee existence
	mentor, err := h.store.Employee.One(h.repo.DB(), employeeID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrEmployeeNotFound, "employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, input, ""))
			return
		}

		l.Error(err, "error when finding employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if mentor.WorkingStatus == model.WorkingStatusLeft {
		l.Error(errs.ErrEmployeeLeft, "employee is left")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmployeeLeft, input, ""))
		return
	}

	// Check mentee existence
	mentee, err := h.store.Employee.One(h.repo.DB(), input.MenteeID.String(), false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrMenteeNotFound, "mentor not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMenteeNotFound, input, ""))
			return
		}

		l.Error(err, "error when finding mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if mentee.WorkingStatus == model.WorkingStatusLeft {
		l.Error(errs.ErrMenteeLeft, "mentee is left")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrMenteeLeft, input, ""))
		return
	}

	if employeeID == input.MenteeID.String() {
		l.Error(errs.ErrCouldNotMentorThemselves, "employee could not be their own mentor")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrCouldNotMentorThemselves, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	// Validate Mentee
	employeeMentee, err := h.store.EmployeeMentee.OneByMenteeID(tx.DB(), employeeID, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "failed to get employee mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) && employeeMentee.MentorID == input.MenteeID {
		l.Error(errs.ErrCouldNotMentorTheirMentor, "employee could not be mentor of their mentor")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(errs.ErrCouldNotMentorTheirMentor), input, ""))
		return
	}

	// Remove old mentor
	if err := h.store.EmployeeMentee.Delete(tx.DB(), input.MenteeID.String()); err != nil {
		l.Error(err, "failed to delete employee mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "failed to get employee mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	_, err = h.store.EmployeeMentee.Create(tx.DB(), &model.EmployeeMentee{
		MenteeID: input.MenteeID,
		MentorID: model.MustGetUUIDFromString(employeeID),
	})
	if err != nil {
		l.Error(err, "failed to create employee mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// DeleteMentee godoc
// @Summary Delete mentee of a mentor
// @Description Delete mentee of a mentor
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param menteeID path string true "Mentee ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /employees/{id}/mentees/{menteeID} [delete]
func (h *handler) DeleteMentee(c *gin.Context) {
	input := request.DeleteMenteeInput{
		MentorID: c.Param("id"),
		MenteeID: c.Param("menteeID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "DeleteMentee",
		"input":   input,
	})

	// Check employee existence
	exist, err := h.store.Employee.IsExist(h.repo.DB(), input.MentorID)
	if err != nil {
		l.Error(err, "error when finding employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exist {
		l.Error(errs.ErrEmployeeNotFound, "employee not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, input, ""))
		return
	}

	// Check mentee existence
	exist, err = h.store.Employee.IsExist(h.repo.DB(), input.MenteeID)
	if err != nil {
		l.Error(err, "error when finding mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exist {
		l.Error(errs.ErrMenteeNotFound, "mentee not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMenteeNotFound, input, ""))
		return
	}

	// Delete employee mentee
	if err := h.store.EmployeeMentee.DeleteByMentorIDAndMenteeID(h.repo.DB(), input.MentorID, input.MenteeID); err != nil {
		l.Error(err, "failed to delete employee mentee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
