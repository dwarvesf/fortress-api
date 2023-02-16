package employee

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
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
	// 0. Get current logged in user data
	userInfo, err := utils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
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

	filter := employee.EmployeeFilter{
		Preload:        body.Preload,
		Keyword:        body.Keyword,
		Positions:      body.Positions,
		Stacks:         body.Stacks,
		Chapters:       body.Chapters,
		Seniorities:    body.Seniorities,
		Organizations:  body.Organizations,
		LineManagers:   body.LineManagers,
		JoinedDateSort: model.SortOrderDESC,
		Projects:       body.Projects,
	}

	// If user don't have this permission, they can only see employees in the project that they are in
	if !utils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadReadActive) {
		projectIDs := make([]string, 0)
		for _, p := range userInfo.Projects {
			projectIDs = append(projectIDs, p.Code)
		}

		filter.Projects = []string{""}
		if len(projectIDs) > 0 {
			filter.Projects = projectIDs
		}
	}

	workingStatuses, err := h.getWorkingStatusInput(c, body.WorkingStatuses)
	if err != nil {
		l.Error(err, "failed to get working status input")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	filter.WorkingStatuses = workingStatuses

	employees, total, err := h.store.Employee.All(h.repo.DB(), filter, body.Pagination)
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEmployeeListData(employees, userInfo),
		&view.PaginationResponse{Pagination: body.Pagination, Total: total}, nil, nil, ""))
}

func (h *handler) getWorkingStatusInput(c *gin.Context, input []string) ([]string, error) {
	userInfo, err := utils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		h.logger.Error(err, "failed to get userID from context")
		return nil, err
	}

	// user who do not have permission
	if !utils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadFilterByAllStatuses) {
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
	userInfo, err := utils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
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

	// 2. get employee from store
	rs, err := h.store.Employee.One(h.repo.DB(), id, true)
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

	if rs.WorkingStatus == model.WorkingStatusLeft && !utils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadFullAccess) {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
		return
	}

	mentees, err := h.store.Employee.GetMenteesByID(h.repo.DB(), rs.ID.String())
	if err != nil {
		l.Error(err, "error query mentees from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, id, ""))
		return
	}

	if len(mentees) > 0 {
		rs.Mentees = mentees
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

	emp, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
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

	emp.WorkingStatus = body.EmployeeStatus
	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *emp, "working_status")
	if err != nil {
		l.Error(err, "failed to update employee status")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
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

	tx, done := h.repo.NewTransaction()

	// check line manager existence
	if !body.LineManagerID.IsZero() {
		exist, err := h.store.Employee.IsExist(tx.DB(), body.LineManagerID.String())
		if err != nil {
			l.Error(err, "error when finding line manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done((err)), body, ""))
			return
		}

		if !exist {
			l.Error(errs.ErrLineManagerNotFound, "error line manager not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrLineManagerNotFound), body, ""))
			return
		}
	}

	// check referrer existence
	if !body.ReferredBy.IsZero() {
		exist, err := h.store.Employee.IsExist(tx.DB(), body.ReferredBy.String())
		if err != nil {
			l.Error(err, "error when finding referrer")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done((err)), body, ""))
			return
		}

		if !exist {
			l.Error(errs.ErrReferrerNotFound, "error referrer not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrReferrerNotFound), body, ""))
			return
		}
	}

	emp, err := h.store.Employee.One(tx.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrEmployeeNotFound), nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// 3. update information and return

	if strings.TrimSpace(body.FullName) != "" {
		emp.FullName = body.FullName
	}

	if strings.TrimSpace(body.Email) != "" {
		emp.TeamEmail = body.Email
	}

	if strings.TrimSpace(body.Phone) != "" {
		emp.PhoneNumber = body.Phone
	}

	if strings.TrimSpace(body.GithubID) != "" {
		emp.GithubID = body.GithubID
	}

	if strings.TrimSpace(body.NotionID) != "" {
		emp.NotionID = body.NotionID
	}

	if strings.TrimSpace(body.NotionName) != "" {
		emp.NotionName = body.NotionName
	}

	if strings.TrimSpace(body.NotionEmail) != "" {
		emp.NotionEmail = body.NotionEmail
	}

	if strings.TrimSpace(body.DiscordID) != "" {
		emp.DiscordID = body.DiscordID
	}

	if strings.TrimSpace(body.DiscordName) != "" {
		emp.DiscordName = body.DiscordName
	}

	if strings.TrimSpace(body.LinkedInName) != "" {
		emp.LinkedInName = body.LinkedInName
	}

	if strings.TrimSpace(body.DisplayName) != "" {
		emp.DisplayName = body.DisplayName
	}

	if strings.TrimSpace(body.JoinedDate) != "" {
		joinedDate, err := time.Parse("2006-01-02", body.JoinedDate)
		if err != nil {
			l.Error(errs.ErrInvalidJoinedDate, "invalid join date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidJoinedDate, body, ""))
			return
		}
		emp.JoinedDate = &joinedDate
	}

	if strings.TrimSpace(body.LeftDate) != "" {
		leftDate, err := time.Parse("2006-01-02", body.LeftDate)
		if err != nil {
			l.Error(errs.ErrInvalidLeftDate, "invalid left date")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidLeftDate, body, ""))
			return
		}
		emp.LeftDate = &leftDate
	}

	if emp.JoinedDate != nil && emp.LeftDate != nil {
		if emp.LeftDate.Before(*emp.JoinedDate) {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrLeftDateBeforeJoinedDate, body, ""))
			return
		}
	}

	emp.LineManagerID = body.LineManagerID
	emp.ReferredBy = body.ReferredBy

	if err := h.updateSocialAccounts(tx.DB(), body, emp.ID); err != nil {
		l.Error(err, "failed to update social accounts")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *emp,
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
		"display_name",
		"joined_date",
		"left_date",
		"referred_by",
	)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	if len(body.OrganizationIDs) > 0 {
		// Check organizations existence
		organizations, err := h.store.Organization.All(tx.DB())
		if err != nil {
			l.Error(err, "failed to get all organizations")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		orgMaps := model.ToOrganizationMap(organizations)
		for _, sID := range body.OrganizationIDs {
			_, ok := orgMaps[sID]
			if !ok {
				l.Error(errs.ErrOrganizationNotFoundWithID(sID.String()), "organization not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrOrganizationNotFoundWithID(sID.String())), body, ""))
				return
			}
		}

		// Delete all exist employee organizations
		if err := h.store.EmployeeOrganization.DeleteByEmployeeID(tx.DB(), employeeID); err != nil {
			l.Error(err, "failed to delete employee organization")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		// Create new employee position
		for _, orgID := range body.OrganizationIDs {
			_, err := h.store.EmployeeOrganization.Create(tx.DB(), &model.EmployeeOrganization{
				EmployeeID:     model.MustGetUUIDFromString(employeeID),
				OrganizationID: orgID,
			})
			if err != nil {
				l.Error(err, "failed to create employee organization")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
				return
			}
		}
	}

	emp, err = h.store.Employee.One(tx.DB(), employeeID, true)
	if err != nil {
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateGeneralInfoEmployeeData(emp), nil, done(nil), nil, ""))
}

func (h *handler) updateSocialAccounts(db *gorm.DB, input request.UpdateEmployeeGeneralInfoInput, employeeID model.UUID) error {
	l := h.logger.Fields(logger.Fields{
		"handler":    "employee",
		"method":     "updateSocialAccounts",
		"input":      input,
		"employeeID": employeeID,
	})

	accounts, err := h.store.SocialAccount.GetByEmployeeID(db, employeeID.String())
	if err != nil {
		l.Error(err, "failed to get social accounts by employeeID")
		return err
	}

	accountsInput := map[model.SocialAccountType]model.SocialAccount{
		model.SocialAccountTypeGitHub: {
			Type:       model.SocialAccountTypeGitHub,
			EmployeeID: employeeID,
			AccountID:  input.GithubID,
			Name:       input.GithubID,
		},
		model.SocialAccountTypeNotion: {
			Type:       model.SocialAccountTypeNotion,
			EmployeeID: employeeID,
			AccountID:  input.NotionID,
			Name:       input.NotionName,
			Email:      input.NotionEmail,
		},
		model.SocialAccountTypeDiscord: {
			Type:       model.SocialAccountTypeDiscord,
			EmployeeID: employeeID,
			AccountID:  input.DiscordID,
			Name:       input.DiscordName,
		},
		model.SocialAccountTypeLinkedIn: {
			Type:       model.SocialAccountTypeLinkedIn,
			EmployeeID: employeeID,
			AccountID:  input.LinkedInName,
			Name:       input.LinkedInName,
		},
	}

	for _, account := range accounts {
		delete(accountsInput, account.Type)

		switch account.Type {
		case model.SocialAccountTypeGitHub:
			account.AccountID = input.GithubID
			account.Name = input.GithubID
		case model.SocialAccountTypeNotion:
			account.Name = input.NotionName
			account.Email = input.NotionEmail
		case model.SocialAccountTypeDiscord:
			account.Name = input.DiscordName
		case model.SocialAccountTypeLinkedIn:
			account.AccountID = input.LinkedInName
			account.Name = input.LinkedInName
		default:
			continue
		}

		if _, err := h.store.SocialAccount.UpdateSelectedFieldsByID(db, account.ID.String(), *account, "account_id", "name", "email"); err != nil {
			l.Errorf(err, "failed to update social account %s", account.ID)
			return err
		}
	}

	for _, account := range accountsInput {
		if _, err := h.store.SocialAccount.Create(db, &account); err != nil {
			l.AddField("account", account).Error(err, "failed to create social account")
			return err
		}
	}

	return nil
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
	userID, err := utils.GetUserIDFromContext(c, h.config)
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

	loggedInUser, err := h.store.Employee.One(h.repo.DB(), userID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

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

	if role.Level <= loggedInUser.EmployeeRoles[0].Role.Level {
		l.Error(errs.ErrInvalidAccountRole, "failed to update role, invalid role")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidAccountRole, input, ""))
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

	if !input.ReferredBy.IsZero() {
		exists, err := h.store.Employee.IsExist(h.repo.DB(), input.ReferredBy.String())
		if err != nil {
			l.Error(err, "failed to getting referrer")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
			return
		}

		if !exists {
			l.Error(errs.ErrReferrerNotFound, "referrer not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrReferrerNotFound, input, ""))
			return
		}

		eml.ReferredBy = input.ReferredBy
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

	_, err = h.store.Employee.One(h.repo.DB(), eml.Username, false)
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

	// Create employee organization
	org, err := h.store.Organization.OneByCode(tx.DB(), model.OrganizationCodeDwarves)
	if err != nil {
		l.Error(err, "error invalid organization")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrOrganizationNotFound), input, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if _, err := h.store.EmployeeOrganization.Create(tx.DB(), &model.EmployeeOrganization{EmployeeID: eml.ID, OrganizationID: org.ID}); err != nil {
		l.Error(err, "error invalid organization")
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

	emp, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
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
	_, stacks, err := h.store.Stack.All(h.repo.DB(), "", nil)
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
	emp.SeniorityID = body.Seniority

	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *emp, "chapter_id", "seniority_id")
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateSkillEmployeeData(emp), nil, done(nil), nil, ""))
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

	emp, err := h.store.Employee.One(h.repo.DB(), employeeID, true)
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

	if isValid := h.validateCountryAndCity(h.repo.DB(), body.Country, body.City); !isValid {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, body, ""))
		return
	}

	emp.DateOfBirth = body.DoB
	emp.Gender = body.Gender
	emp.Address = body.Address
	emp.PlaceOfResidence = body.PlaceOfResidence
	emp.PersonalEmail = body.PersonalEmail
	emp.Country = body.Country
	emp.City = body.City

	emp, err = h.store.Employee.Update(h.repo.DB(), emp)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
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

	l := h.logger.Fields(logger.Fields{
		"handler":     "profile",
		"method":      "validateCountryAndCity",
		"countryName": countryName,
		"city":        city,
	})

	country, err := h.store.Country.OneByName(db, countryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("country not found")
			return false
		}
		l.Error(err, "failed to get country by code")
		return false
	}

	if city != "" && !slices.Contains([]string(country.Cities), city) {
		l.Info("city does not belong to country")
		return false
	}

	return true
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
	emp, err := h.store.Employee.One(tx.DB(), params.ID, false)
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
		EmployeeID: emp.ID,
		UploadBy:   emp.ID,
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
	userID, err := utils.GetUserIDFromContext(c, h.config)
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
	emp, err := h.store.Employee.One(tx.DB(), params.ID, false)
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
			EmployeeID: emp.ID,
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
	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), emp.ID.String(), model.Employee{
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
	userID, err := utils.GetUserIDFromContext(c, h.config)
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

	loggedInUser, err := h.store.Employee.One(h.repo.DB(), userID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	empl, err := h.store.Employee.One(h.repo.DB(), input.EmployeeID, false)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(errs.ErrEmployeeNotFound, "reviewer not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrEmployeeNotFound, nil, ""))
		return
	}

	if err != nil {
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Check role exists
	newRole, err := h.store.Role.One(h.repo.DB(), input.Body.RoleID)
	if err != nil {
		l.Error(err, "error when finding role")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if empl.EmployeeRoles[0].Role.Level == loggedInUser.EmployeeRoles[0].Role.Level {
		l.Error(errs.ErrInvalidAccountRole, "failed to update role, invalid role")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrCouldNotAssignRoleForSameLevelEmployee, input, ""))
		return
	}

	if newRole.Level <= loggedInUser.EmployeeRoles[0].Role.Level {
		l.Error(errs.ErrInvalidAccountRole, "failed to update role, invalid role")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidAccountRole, input, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	if err := h.store.EmployeeRole.HardDeleteByEmployeeID(tx.DB(), input.EmployeeID); err != nil {
		l.Error(err, "failed to delete employee role")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	_, err = h.store.EmployeeRole.Create(tx.DB(), &model.EmployeeRole{
		EmployeeID: model.MustGetUUIDFromString(input.EmployeeID),
		RoleID:     input.Body.RoleID,
	})
	if err != nil {
		l.Error(err, "failed to create employee role")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
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
	userInfo, err := utils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler":  "employee",
		"method":   "GetLineManagers",
		"userInfo": userInfo.UserID,
	})

	var managers []*model.Employee

	if utils.HasPermission(userInfo.Permissions, model.PermissionEmployeesReadLineManagerFullAccess) {
		managers, err = h.store.Employee.GetLineManagers(h.repo.DB())
		if err != nil {
			l.Error(err, "failed to get line managers")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
			return
		}
	} else {
		managers, err = h.store.Employee.GetLineManagersOfPeers(h.repo.DB(), userInfo.UserID)
		if err != nil {
			l.Error(err, "failed to get line managers of peers")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToBasicEmployees(managers), nil, nil, nil, ""))
}
