package project

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/project/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/project/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store      *store.Store
	controller *controller.Controller
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:      store,
		controller: controller,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
	}
}

// List godoc
// @Summary Get list of project
// @Description Get list of project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param status query  []string false  "Project status"
// @Param name   query  string false  "Project name"
// @Param type   query  string false  "Project type"
// @Param page   query  string false  "Page"
// @Param size   query  string false  "Size"
// @Success 200 {object} view.ProjectListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects [get]
func (h *handler) List(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	query := request.GetListProjectInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	query.StandardizeInput()

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "List",
		"query":   query,
	})

	if err := query.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	projects, total, err := h.store.Project.All(h.repo.DB(), project.GetListProjectInput{
		Statuses: query.Status,
		Name:     query.Name,
		Type:     query.Type,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectsData(projects, userInfo),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil, ""))
}

// UpdateProjectStatus godoc
// @Summary Update status for project by id
// @Description Update status for project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param status body model.ProjectStatus true "Project Status"
// @Success 200 {object} view.UpdateProjectStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/status [put]
func (h *handler) UpdateProjectStatus(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	var body request.UpdateAccountStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateProjectStatus",
		"body":    body,
	})

	if !body.ProjectStatus.IsValid() {
		l.Error(errs.ErrInvalidProjectStatus, "invalid value for ProjectStatus")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectStatus, body, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), projectID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, projectID, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	p.Status = body.ProjectStatus
	p.EndDate = nil

	if body.ProjectStatus == model.ProjectStatusClosed {
		p.EndDate = new(time.Time)
		*p.EndDate = time.Now()
	}

	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), projectID, *p, "status", "end_date")
	if err != nil {
		l.Error(err, "failed to update project status")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// TODO: we can open it when needing automation flow to inactivated project member
	//if body.ProjectStatus == model.ProjectStatusClosed {
	//	if err := h.closeProject(tx.DB(), projectID); err != nil {
	//		l.Error(err, "failed to close project")
	//		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
	//		return
	//	}
	//}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectStatusResponse(p), nil, done(nil), nil, ""))
}

// func (h *handler) closeProject(db *gorm.DB, projectID string) error {
// 	err := h.store.ProjectMember.UpdateEndDateByProjectID(db, projectID)
// 	if err != nil {
// 		h.logger.Error(err, "failed to update end_date by project_id")
// 		return err
// 	}

// 	err = h.store.ProjectMember.UpdateSelectedFieldByProjectID(db, projectID,
// 		model.ProjectMember{Status: model.ProjectMemberStatusInactive},
// 		"status")
// 	if err != nil {
// 		h.logger.Error(err, "failed to update status of project_member by project_id")
// 		return err
// 	}

// 	err = h.store.ProjectSlot.UpdateSelectedFieldByProjectID(db, projectID,
// 		model.ProjectSlot{Status: model.ProjectMemberStatusInactive},
// 		"status",
// 	)
// 	if err != nil {
// 		h.logger.Error(err, "failed to update status of project_slot by project_id")
// 		return err
// 	}

// 	return nil
// }

// Create godoc
// @Summary	Create new project
// @Description	Create new project
// @Tags Project
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreateProjectInput true "body"
// @Success 200 {object} view.CreateProjectData
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects [post]
func (h *handler) Create(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	body := request.CreateProjectInput{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "Create",
		"body":    body,
	})

	if err := body.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	country, err := h.store.Country.One(h.repo.DB(), body.CountryID.String())
	if err != nil {
		l.Error(err, "failed to get country")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	var bankAccount *model.BankAccount
	if !body.BankAccountID.IsZero() {
		bankAccount, err = h.store.BankAccount.One(h.repo.DB(), body.BankAccountID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "bank account not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrBankAccountNotFound, body, ""))
				return
			}

			l.Error(err, "failed to get bank account")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	var organization *model.Organization
	if !body.OrganizationID.IsZero() {
		organization, err = h.store.Organization.One(h.repo.DB(), body.OrganizationID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Error(err, "organization not found")
				c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrOrganizationNotFound, body, ""))
				return
			}

			l.Error(err, "failed to get organization")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	if body.Code == "" {
		body.Code = strings.ReplaceAll(strings.ToLower(body.Name), " ", "-")
	}

	exists, err := h.store.Project.IsExistByCode(h.repo.DB(), body.Code)
	if err != nil {
		l.Error(err, "failed to check existence by code")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if exists {
		l.Error(errs.ErrDuplicateProjectCode, "failed to create project")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrDuplicateProjectCode, body, ""))
		return
	}

	var client *model.Client
	if !body.ClientID.IsZero() {
		client, err = h.store.Client.One(h.repo.DB(), body.ClientID.String())
		if err != nil {
			l.Error(err, "client not found")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrClientNotFound, body, ""))
			return
		}
	}

	// Create employee organization
	org, err := h.store.Organization.OneByCode(h.repo.DB(), model.OrganizationCodeDwarves)
	if err != nil {
		l.Error(err, "error invalid organization")
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrOrganizationNotFound, body, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	p := &model.Project{
		Name:         body.Name,
		CountryID:    body.CountryID,
		Type:         model.ProjectType(body.Type),
		Status:       model.ProjectStatus(body.Status),
		StartDate:    body.GetStartDate(),
		ProjectEmail: body.ProjectEmail,
		ClientEmail:  strings.Join(body.ClientEmail, ","),
		Country:      country,
		Code:         body.Code,
		Function:     model.ProjectFunction(body.Function),
		ClientID:     body.ClientID,
	}

	if body.OrganizationID.IsZero() {
		p.OrganizationID = org.ID
	} else {
		p.OrganizationID = body.OrganizationID
	}

	if !body.BankAccountID.IsZero() {
		p.BankAccountID = body.BankAccountID
		p.BankAccount = bankAccount
	}

	if !body.OrganizationID.IsZero() {
		p.OrganizationID = body.OrganizationID
		p.Organization = organization
	}

	tx, done := h.repo.NewTransaction()

	if err := h.store.Project.Create(tx.DB(), p); err != nil {
		l.Error(err, "failed to create project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	p.Client = client

	// Create audit notion id
	if !body.AuditNotionID.IsZero() {
		if _, err := h.store.ProjectNotion.Create(tx.DB(), &model.ProjectNotion{ProjectID: p.ID, AuditNotionID: body.AuditNotionID}); err != nil {
			l.Error(err, "failed to create project notion")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// create project account manager
	ams, err := h.createProjectHeads(tx.DB(), p.ID, model.HeadPositionAccountManager, body.AccountManagers)
	if err != nil {
		l.Error(err, "failed to create account managers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}
	p.Heads = append(p.Heads, ams...)

	// create project delivery manager
	dms, err := h.createProjectHeads(tx.DB(), p.ID, model.HeadPositionDeliveryManager, body.DeliveryManagers)
	if err != nil {
		l.Error(err, "failed to create delivery managers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}
	p.Heads = append(p.Heads, dms...)

	// create project sale persons
	sps, err := h.createProjectHeads(tx.DB(), p.ID, model.HeadPositionSalePerson, body.DeliveryManagers)
	if err != nil {
		l.Error(err, "failed to create sale persons")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}
	p.Heads = append(p.Heads, sps...)

	// assign members to project
	for _, member := range body.Members {
		slot, code, err := h.createSlotsAndAssignMembers(tx.DB(), p, member, userInfo)
		if err != nil {
			l.Error(err, "failed to assign member to project")
			c.JSON(code, view.CreateResponse[any](nil, nil, done(err), member, ""))
			return
		}

		p.Slots = append(p.Slots, *slot)
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateProjectDataResponse(userInfo, p), nil, done(nil), nil, ""))
}

func (h *handler) createProjectHeads(db *gorm.DB, projectID model.UUID, position model.HeadPosition, req []request.ProjectHeadInput) ([]*model.ProjectHead, error) {
	var heads []*model.ProjectHead
	for _, head := range req {
		emp, err := h.store.Employee.One(db, head.EmployeeID.String(), false)
		if err != nil {
			h.logger.Error(err, "failed to get employee by id")
			return nil, err
		}

		head := &model.ProjectHead{
			ProjectID:      projectID,
			EmployeeID:     head.EmployeeID,
			CommissionRate: head.CommissionRate,
			Position:       position,
		}

		if err := h.store.ProjectHead.Create(db, head); err != nil {
			h.logger.Error(err, "failed to create project head")
			return nil, err
		}

		head.Employee = *emp
		heads = append(heads, head)
	}

	return heads, nil
}

// GetMembers godoc
// @Summary Get list members of project
// @Description Get list members of project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string false "Project ID"
// @Param status query string false "Status"
// @Param preload query bool false "Preload data with default value is true"
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Param sort query string false "Sort"
// @Param distinct query bool false "Distinct"
// @Success 200 {object} view.ProjectMemberListResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [get]
func (h *handler) GetMembers(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	query := request.GetListStaffInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}
	query.Standardize()

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler":   "project",
		"method":    "GetMembers",
		"projectID": projectID,
		"query":     query,
	})

	if err := query.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	project, err := h.store.Project.One(h.repo.DB(), projectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, query, ""))
			return
		}
		l.Error(err, "cannot find project by id")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	var pendingSlots []*model.ProjectSlot
	var assignedMembers []*model.ProjectMember

	// Get pending slots
	if query.Status == "" || query.Status == model.ProjectMemberStatusPending.String() {
		pendingSlots, err = h.store.ProjectSlot.GetPendingSlots(h.repo.DB(), projectID, query.Preload)
		if err != nil {
			l.Error(err, "failed to get pending slots")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
			return
		}
	}

	// Get assigned members
	if query.Status != model.ProjectMemberStatusPending.String() {
		assignedMembers, err = h.store.ProjectMember.GetAssignedMembers(h.repo.DB(), projectID, query.Status, query.Preload)
		if err != nil {
			l.Error(err, "failed to get assigned members")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
			return
		}
	}

	// Merge pending slots and assigned members into a slice
	total, members := h.mergeSlotAndMembers(h.repo.DB(), pendingSlots, assignedMembers, query.Pagination)

	heads, err := h.store.ProjectHead.GetActiveLeadsByProjectID(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to get project heads")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectMemberListData(userInfo, members, heads, project, query.Distinct),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil, ""))
}

func (h *handler) mergeSlotAndMembers(db *gorm.DB, slots []*model.ProjectSlot, members []*model.ProjectMember, pagination model.Pagination) (int64, []*model.ProjectMember) {
	results := make([]*model.ProjectMember, 0, len(slots)+len(members))

	for _, slot := range slots {
		member := &model.ProjectMember{
			ProjectID:      slot.ProjectID,
			ProjectSlotID:  slot.ID,
			SeniorityID:    slot.SeniorityID,
			DeploymentType: slot.DeploymentType,
			Status:         slot.Status,
			Rate:           slot.Rate,
			Discount:       slot.Discount,
			Seniority:      &slot.Seniority,
			Note:           slot.Note,
		}

		for _, psPosition := range slot.ProjectSlotPositions {
			member.Positions = append(member.Positions, psPosition.Position)
		}

		results = append(results, member)
	}

	results = append(results, members...)

	total := int64(len(results))

	// Get response by offset and limit
	limit, offset := pagination.ToLimitOffset()
	if offset > len(results) {
		results = []*model.ProjectMember{}
	} else if limit+offset > len(slots) {
		results = results[offset:]
	} else {
		results = results[offset : offset+limit]
	}

	return total, results
}

// DeleteMember godoc
// @Summary Delete member in a project
// @Description Delete member in a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param memberID path string true "Project Member ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /project/{id}/members/{memberID} [delete]
func (h *handler) DeleteMember(c *gin.Context) {
	input := request.DeleteMemberInput{
		ProjectID: c.Param("id"),
		MemberID:  c.Param("memberID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "DeleteMember",
		"body":    input,
	})

	member, err := h.store.ProjectMember.OneByID(h.repo.DB(), input.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project member not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectMemberNotFound, input, ""))
			return
		}
		l.Error(err, "failed to get project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// if projectMember.Status == model.ProjectMemberStatusInactive {
	// 	l.Error(errs.ErrCouldNotDeleteInactiveMember, "can not change information of inactive member")
	// 	c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrCouldNotDeleteInactiveMember, input.MemberID, ""))
	// 	return
	// }

	tx, done := h.repo.NewTransaction()

	err = h.store.ProjectMemberPosition.DeleteByProjectMemberID(tx.DB(), member.ID.String())
	if err != nil {
		l.Error(err, "failed to delete project member position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	err = h.store.ProjectMember.Delete(tx.DB(), member.ID.String())
	if err != nil {
		l.Error(err, "failed to delete project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	err = h.store.ProjectHead.DeleteByPositionInProject(tx.DB(),
		member.ProjectID.String(),
		member.EmployeeID.String(),
		model.HeadPositionTechnicalLead.String())
	if err != nil {
		l.Error(err, "failed to delete project head")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Update project slot status to inactive
	_, err = h.store.ProjectSlot.UpdateSelectedFieldsByID(tx.DB(),
		member.ProjectSlotID.String(),
		model.ProjectSlot{
			Status: model.ProjectMemberStatusInactive,
		},
		"status")
	if err != nil {
		l.Error(err, "failed to update project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// DeleteSlot godoc
// @Summary Delete slot in a project
// @Description Delete slot in a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param slotID path string true "Slot ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /project/{id}/slot/{slotID} [delete]
func (h *handler) DeleteSlot(c *gin.Context) {
	input := request.DeleteSlotInput{
		ProjectID: c.Param("id"),
		SlotID:    c.Param("slotID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "DeleteSlot",
		"body":    input,
	})

	slot, err := h.store.ProjectSlot.One(h.repo.DB(), input.SlotID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project slot not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectSlotNotFound, input, ""))
			return
		}
		l.Error(err, "failed to get project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	slot.Status = model.ProjectMemberStatusInactive

	_, err = h.store.ProjectSlot.UpdateSelectedFieldsByID(h.repo.DB(), input.SlotID, *slot, "status")
	if err != nil {
		l.Error(err, "failed to update project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// UnassignMember godoc
// @Summary Unassign member in a project
// @Description Unassign member in a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param memberID path string true "Employee ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /project/{id}/members/{memberID} [put]
func (h *handler) UnassignMember(c *gin.Context) {
	input := request.UnassignMemberInput{
		ProjectID: c.Param("id"),
		MemberID:  c.Param("memberID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UnassignMember",
		"body":    input,
	})

	// get member info
	projectMember, err := h.store.ProjectMember.GetActiveMemberInProject(h.repo.DB(), input.ProjectID, input.MemberID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project member not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectMemberNotFound, input.MemberID, ""))
			return
		}
		l.Error(err, "failed to get project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input.MemberID, ""))
		return
	}

	// if projectMember.Status == model.ProjectMemberStatusInactive {
	// 	l.Error(errs.ErrCouldNotDeleteInactiveMember, "can not change information of inactive member")
	// 	c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrCouldNotDeleteInactiveMember, input.MemberID, ""))
	// 	return
	// }

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// remove member out of project
	timeNow := time.Now()
	if projectMember.EndDate == nil {
		projectMember.EndDate = &timeNow
	}

	projectMember.Status = model.ProjectMemberStatusInactive

	_, err = h.store.ProjectMember.UpdateSelectedFieldsByID(tx.DB(),
		projectMember.ID.String(),
		*projectMember,
		"end_date",
		"status")
	if err != nil {
		l.Error(err, "failed to update project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// update technical lead if employees is technical lead
	_, err = h.store.ProjectHead.UpdateDateOfEmployee(tx.DB(),
		input.MemberID,
		input.ProjectID,
		model.HeadPositionTechnicalLead.String(),
		projectMember.StartDate,
		projectMember.EndDate)
	if err != nil {
		l.Error(err, "failed to update endDate for technical lead")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// update slot status -> pending
	slot := model.ProjectSlot{
		Status: model.ProjectMemberStatusPending,
	}
	_, err = h.store.ProjectSlot.UpdateSelectedFieldsByID(tx.DB(), projectMember.ProjectSlotID.String(), slot, "status")
	if err != nil {
		l.Error(err, "failed to update project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// UpdateMember godoc
// @Summary Update member in an existing project
// @Description Update member in an existing project
// @Tags Project
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body request.UpdateMemberInput true "Body"
// @Success 200 {object} view.CreateMemberDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [put]
func (h *handler) UpdateMember(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	var body request.UpdateMemberInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateMember",
		"body":    body,
	})

	if err := body.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// check project existence
	p, err := h.store.Project.One(h.repo.DB(), projectID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, body, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// check seniority existence
	exists, err := h.store.Seniority.IsExist(h.repo.DB(), body.SeniorityID.String())
	if err != nil {
		l.Error(err, "failed to check seniority existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrSeniorityNotFound, "cannot find seniority by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrSeniorityNotFound, body, ""))
		return
	}

	// check position existence
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to get all positions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range body.Positions {
		if _, ok := positionMap[pID]; !ok {
			l.Error(errs.ErrPositionNotFoundWithID(pID.String()), "position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrPositionNotFoundWithID(pID.String()), body, ""))
			return
		}
	}

	// check project slot status
	slot, err := h.store.ProjectSlot.One(h.repo.DB(), body.ProjectSlotID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectSlotNotFound, "cannot find project slot by id")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectSlotNotFound, body, ""))
			return
		}
		l.Error(err, "failed to get project slot by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if !body.EmployeeID.IsZero() {
		member, err := h.updateProjectMember(tx.DB(), p, slot.ID.String(), projectID, body, userInfo)
		if err != nil {
			l.Error(err, "failed to update project member")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		slot.ProjectMember = *member
	}

	// update project slot
	slot.SeniorityID = body.SeniorityID
	slot.DeploymentType = model.DeploymentType(body.DeploymentType)
	slot.Status = model.ProjectMemberStatus(body.Status)
	slot.Note = body.Note

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
		slot.Rate = body.Rate
		slot.Discount = body.Discount
	}

	_, err = h.store.ProjectSlot.UpdateSelectedFieldsByID(tx.DB(), body.ProjectSlotID.String(), *slot,
		"seniority_id",
		"deployment_type",
		"status",
		"rate",
		"discount",
		"note",
	)
	if err != nil {
		l.Error(err, "failed to update project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// update project slot positions
	if err := h.store.ProjectSlotPosition.DeleteByProjectSlotID(tx.DB(), slot.ID.String()); err != nil {
		l.Error(err, "failed to delete project member positions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	slotPos := make([]model.ProjectSlotPosition, 0, len(body.Positions))
	for _, v := range body.Positions {
		slotPos = append(slotPos, model.ProjectSlotPosition{
			ProjectSlotID: slot.ID,
			PositionID:    v,
		})
	}

	if err := h.store.ProjectSlotPosition.Create(tx.DB(), slotPos...); err != nil {
		l.Error(err, "failed to create project slot positions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	slot.ProjectSlotPositions = slotPos

	for i, v := range slot.ProjectSlotPositions {
		slot.ProjectSlotPositions[i].Position = positionMap[v.PositionID]
	}

	for i, v := range slot.ProjectMember.ProjectMemberPositions {
		slot.ProjectMember.ProjectMemberPositions[i].Position = positionMap[v.PositionID]
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(userInfo, slot), nil, done(nil), nil, ""))
}

// updateProjectMember flow:
//
// --- start ---
//
//	if input.EmployeeID != nil {
//		if input.ProjectMemberID != nil {
//			update ProjectMember by ProjectMemberID
//		 } else {
//			update ProjectMember by ProjectID and EmployeeID
//		 }
//		 if !input.IsLead || input.EndDate != nil {
//			endDate := input.EndDate
//			if input.EndDate == nil {
//			   endDate = time.Now()
//				   update endDate of projectHead
//				}
//			 } else {
//				create new ProjectHead
//			 }
//		 }
//
// --- end ---
func (h *handler) updateProjectMember(db *gorm.DB, p *model.Project, slotID string, projectID string, input request.UpdateMemberInput, userInfo *model.CurrentLoggedUserInfo) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	var err error

	// check upsell person existence
	var upsellPerson *model.Employee
	if !input.UpsellPersonID.IsZero() {
		upsellPerson, err = h.store.Employee.One(h.repo.DB(), input.UpsellPersonID.String(), false)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				h.logger.Error(errs.ErrEmployeeNotFound, "upsell person not found")
				return nil, err
			}

			h.logger.Error(err, "failed to get upsell person by id")
			return nil, err
		}
	}

	if !input.ProjectMemberID.IsZero() {
		// Update assigned slot
		member, err = h.store.ProjectMember.OneByID(db, input.ProjectMemberID.String())
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				h.logger.Error(errs.ErrProjectMemberNotFound, "project member not found")
				return nil, err
			}
			h.logger.Error(err, "failed to get project member by id")
			return nil, err
		}

		member.SeniorityID = input.SeniorityID
		member.DeploymentType = model.DeploymentType(input.DeploymentType)
		member.StartDate = input.GetStartDate()
		member.Note = input.Note
		member.UpsellPersonID = input.UpsellPersonID

		updateEndDate := false
		inputEndDate := input.GetEndDate()
		if member.EndDate != nil && inputEndDate == nil {
			member.EndDate = nil
			updateEndDate = true
		}

		if member.EndDate == nil && inputEndDate != nil {
			member.EndDate = inputEndDate
			updateEndDate = true
		}

		if member.EndDate != nil && inputEndDate != nil {
			if !member.EndDate.Equal(*inputEndDate) {
				member.EndDate = inputEndDate
				updateEndDate = true
			}
		}

		updateStatus := false
		if member.Status != model.ProjectMemberStatus(input.Status) {
			member.Status = model.ProjectMemberStatus(input.Status)
			updateStatus = true
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
			member.UpsellCommissionRate = input.UpsellCommissionRate
		}

		updateRate := false
		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
			if !member.Rate.Equal(input.Rate) {
				member.Rate = input.Rate
				updateRate = true
			}
			member.Discount = input.Discount
		}

		_, err = h.store.ProjectMember.UpdateSelectedFieldsByID(db, input.ProjectMemberID.String(), *member,
			"start_date",
			"end_date",
			"status",
			"rate",
			"discount",
			"deployment_type",
			"seniority_id",
			"note",
			"upsell_person_id",
			"upsell_commission_rate",
		)
		if err != nil {
			h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to update project member")
			return nil, err
		}

		if updateStatus {
			err = h.controller.Discord.Log(model.LogDiscordInput{
				Type: "project_member_update_status",
				Data: map[string]interface{}{
					"employee_id":         userInfo.UserID,
					"updated_employee_id": member.EmployeeID.String(),
					"project_name":        p.Name,
					"status":              member.Status.String(),
				},
			})
			if err != nil {
				h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to log project member status update")
			}
		}

		if updateRate {
			charRate, _ := member.Rate.Float64()
			rate := utils.FormatMoney(charRate, p.BankAccount.Currency.Name)
			err = h.controller.Discord.Log(model.LogDiscordInput{
				Type: "project_member_update_charge_rate",
				Data: map[string]interface{}{
					"employee_id":         userInfo.UserID,
					"updated_employee_id": member.EmployeeID.String(),
					"project_name":        p.Name,
					"rate":                fmt.Sprintf("%s %s", rate, p.BankAccount.Currency.Name),
				},
			})
			if err != nil {
				h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to log project member charge rate update")
			}
		}

		if updateEndDate {
			endDateLog := "N/A"
			if member.EndDate != nil {
				endDateLog = member.EndDate.Format("2006-01-02")
			}

			err = h.controller.Discord.Log(model.LogDiscordInput{
				Type: "project_member_update_end_date",
				Data: map[string]interface{}{
					"employee_id":         userInfo.UserID,
					"updated_employee_id": member.EmployeeID.String(),
					"project_name":        p.Name,
					"end_date":            endDateLog,
				},
			})
			if err != nil {
				h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to log project member end date update")
			}
		}
	} else {
		// Update pending slot

		// Is slot contains any member?
		member, err = h.store.ProjectMember.OneBySlotID(db, slotID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Fields(logger.Fields{"slotID": slotID}).Error(err, "failed to get project member by slotID")
			return nil, err
		}

		if member != nil && !member.EmployeeID.IsZero() && member.EmployeeID != input.EmployeeID {
			h.logger.
				Fields(logger.Fields{"member": member}).
				Error(errs.ErrSlotAlreadyContainsAnotherMember, "slot already contains another member")
			return nil, errs.ErrSlotAlreadyContainsAnotherMember
		}

		// Is member active in project?
		_, err = h.store.ProjectMember.GetActiveMemberInProject(db, projectID, input.EmployeeID.String())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Fields(logger.Fields{
				"projectID":  projectID,
				"employeeID": input.EmployeeID,
			}).Error(err, "failed to get active member in project")
			return nil, err
		}

		// If member is not active in project, create new project member
		if errors.Is(err, gorm.ErrRecordNotFound) {
			member = &model.ProjectMember{
				ProjectID:      model.MustGetUUIDFromString(projectID),
				EmployeeID:     input.EmployeeID,
				SeniorityID:    input.SeniorityID,
				ProjectSlotID:  model.MustGetUUIDFromString(slotID),
				DeploymentType: model.DeploymentType(input.DeploymentType),
				Status:         model.ProjectMemberStatus(input.Status),
				StartDate:      input.GetStartDate(),
				EndDate:        input.GetEndDate(),
				Note:           input.Note,
				UpsellPersonID: input.UpsellPersonID,
			}

			if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
				member.UpsellCommissionRate = input.UpsellCommissionRate
			}

			if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
				member.Rate = input.Rate
				member.Discount = input.Discount
			}

			if err := h.store.ProjectMember.Create(db, member); err != nil {
				h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to create project member")
				return nil, err
			}

			err = h.controller.Discord.Log(model.LogDiscordInput{
				Type: "project_member_add",
				Data: map[string]interface{}{
					"employee_id":         userInfo.UserID,
					"updated_employee_id": member.EmployeeID.String(),
					"project_name":        p.Name,
					"deployment_type":     member.DeploymentType.String(),
				},
			})
			if err != nil {
				h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to log project member add")
			}
		}
	}

	// update project member positions
	if err := h.store.ProjectMemberPosition.DeleteByProjectMemberID(db, member.ID.String()); err != nil {
		h.logger.Fields(logger.Fields{"memberID": member.ID}).Error(err, "failed to delete project member positions")
		return nil, err
	}

	memberPos := make([]model.ProjectMemberPosition, 0, len(input.Positions))
	for _, v := range input.Positions {
		memberPos = append(memberPos, model.ProjectMemberPosition{
			ProjectMemberID: member.ID,
			PositionID:      v,
		})
	}

	if err := h.store.ProjectMemberPosition.Create(db, memberPos...); err != nil {
		h.logger.Fields(logger.Fields{"positions": memberPos}).Error(err, "failed to create project member positions")
		return nil, err
	}

	member.UpsellPerson = upsellPerson
	member.ProjectMemberPositions = memberPos

	member.IsLead = input.IsLead
	endDate := input.GetEndDate()
	if !input.IsLead {
		// End of lead time
		if endDate == nil {
			endDate = new(time.Time)
			*endDate = time.Now()
		}

		_, err := h.store.ProjectHead.UpdateDateOfEmployee(db,
			input.EmployeeID.String(),
			projectID,
			model.HeadPositionTechnicalLead.String(),
			input.GetStartDate(),
			endDate)
		if err != nil {
			h.logger.Fields(logger.Fields{
				"projectID":  projectID,
				"employeeID": input.EmployeeID,
			}).Error(err, "failed to update end_date of project head")
			return nil, err
		}
	} else {
		// Start of lead time or update lead time
		_, err := h.updateProjectLead(db, projectID, input.EmployeeID, input.GetStartDate(), input.GetEndDate(), input.LeadCommissionRate, userInfo)
		if err != nil {
			h.logger.Fields(logger.Fields{
				"projectID":  projectID,
				"employeeID": input.EmployeeID,
			}).Error(err, "failed to update technicalLeads")
			return nil, err
		}
	}

	return member, nil
}

// AssignMember godoc
// @Summary Assign member into an existing project
// @Description Assign member in an existing project
// @Tags Project
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body request.AssignMemberInput true "Body"
// @Success 200 {object} view.CreateMemberDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [post]
func (h *handler) AssignMember(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	var body request.AssignMemberInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "AssignMember",
		"body":    body,
		"id":      projectID,
	})

	if err := body.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// TODO: uncomment
	// The code has been commented because inactive user can be assigned to project,
	// we do not need to check if member active in project or not

	// // get active project member info
	// if !body.EmployeeID.IsZero() {
	// 	_, err := h.store.ProjectMember.GetActiveMemberInProject(h.repo.DB(), projectID, body.EmployeeID.String())
	// 	if err != gorm.ErrRecordNotFound {
	// 		if err == nil {
	// 			l.Error(err, "project member exists")
	// 			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrProjectMemberExists, projectID, ""))
	// 			return
	// 		}
	// 		l.Error(err, "failed to query project member")
	// 		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
	// 		return
	// 	}
	// }

	// check project existence
	p, err := h.store.Project.One(h.repo.DB(), projectID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// remove commission rate if user does not have permission
	body.RestrictPermission(userInfo)

	tx, done := h.repo.NewTransaction()

	slot, code, err := h.createSlotsAndAssignMembers(tx.DB(), p, body, userInfo)
	if err != nil {
		l.Error(err, "failed to assign member to project")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), body, ""))
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(userInfo, slot), nil, done(nil), nil, ""))
}

func (h *handler) createSlotsAndAssignMembers(db *gorm.DB, p *model.Project, req request.AssignMemberInput, userInfo *model.CurrentLoggedUserInfo) (*model.ProjectSlot, int, error) {
	l := h.logger

	// check seniority existence
	seniority, err := h.store.Seniority.One(db, req.SeniorityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrSeniorityNotFound, "cannot find seniority by id")
			return nil, http.StatusNotFound, errs.ErrSeniorityNotFound
		}
		l.Error(err, "failed to check seniority existence")
		return nil, http.StatusInternalServerError, err
	}

	// check position existence
	positions, err := h.store.Position.All(db)
	if err != nil {
		l.Error(err, "failed to get all position")
		return nil, http.StatusInternalServerError, err
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range req.Positions {
		if _, ok := positionMap[pID]; !ok {
			l.Error(errs.ErrPositionNotFoundWithID(pID.String()), "position not found")
			return nil, http.StatusNotFound, errs.ErrPositionNotFoundWithID(pID.String())
		}
	}

	// create project slot
	slot := &model.ProjectSlot{
		ProjectID:      p.ID,
		DeploymentType: model.DeploymentType(req.DeploymentType),
		Status:         req.GetStatus(),
		SeniorityID:    req.SeniorityID,
		Note:           req.Note,
	}

	if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
		slot.Rate = req.Rate
		slot.Discount = req.Discount
	}

	if err := h.store.ProjectSlot.Create(db, slot); err != nil {
		l.Error(err, "failed to create project slot")
		return nil, http.StatusInternalServerError, err
	}
	slot.Seniority = *seniority

	// create project slot position
	slotPos := make([]model.ProjectSlotPosition, 0, len(req.Positions))

	for _, v := range req.Positions {
		slotPos = append(slotPos, model.ProjectSlotPosition{
			ProjectSlotID: slot.ID,
			PositionID:    v,
		})
	}

	if err := h.store.ProjectSlotPosition.Create(db, slotPos...); err != nil {
		l.Error(err, "failed to create project member positions")
		return nil, http.StatusInternalServerError, err
	}

	for i := range slotPos {
		slotPos[i].Position = positionMap[slotPos[i].PositionID]
	}

	slot.ProjectSlotPositions = slotPos

	// assign member to slot
	if !req.EmployeeID.IsZero() {
		// check employee existence
		exists, err := h.store.Employee.IsExist(db, req.EmployeeID.String())
		if err != nil {
			l.Error(err, "failed to check employee existence")
			return nil, http.StatusInternalServerError, err
		}

		if !exists {
			l.Error(errs.ErrEmployeeNotFound, "cannot find employee by id")
			return nil, http.StatusNotFound, errs.ErrEmployeeNotFound
		}

		// check upsell person existence
		var upsellPerson *model.Employee
		if !req.UpsellPersonID.IsZero() {
			upsellPerson, err = h.store.Employee.One(db, req.UpsellPersonID.String(), false)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					l.Error(err, "failed to get upsell person")
					return nil, http.StatusInternalServerError, err
				}
				l.Error(errs.ErrEmployeeNotFound, "upsell person not found")
				return nil, http.StatusNotFound, errs.ErrEmployeeNotFound
			}
		}

		// create project member
		member := &model.ProjectMember{
			ProjectID:      p.ID,
			EmployeeID:     req.EmployeeID,
			SeniorityID:    req.SeniorityID,
			ProjectSlotID:  slot.ID,
			DeploymentType: model.DeploymentType(req.DeploymentType),
			Status:         req.GetStatus(),
			StartDate:      req.GetStartDate(),
			EndDate:        req.GetEndDate(),
			Note:           req.Note,
			UpsellPersonID: req.UpsellPersonID,
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
			member.UpsellCommissionRate = req.UpsellCommissionRate
		}

		if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectMembersRateEdit) {
			member.Rate = req.Rate
			member.Discount = req.Discount
		}

		if err = h.store.ProjectMember.Create(db, member); err != nil {
			l.Error(err, "failed to create project member")
			return nil, http.StatusInternalServerError, err
		}

		member.UpsellPerson = upsellPerson
		slot.ProjectMember = *member

		// create project member positions
		for _, v := range req.Positions {
			if err := h.store.ProjectMemberPosition.Create(db, model.ProjectMemberPosition{
				ProjectMemberID: member.ID,
				PositionID:      v,
			}); err != nil {
				l.Error(err, "failed to create project member positions")
				return nil, http.StatusInternalServerError, err
			}
		}

		// create project head
		slot.ProjectMember.IsLead = req.IsLead
		if req.IsLead {
			head := &model.ProjectHead{
				ProjectID:  p.ID,
				EmployeeID: req.EmployeeID,
				Position:   model.HeadPositionTechnicalLead,
				StartDate:  *req.GetStartDate(),
			}

			if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
				head.CommissionRate = req.LeadCommissionRate
			}

			if err := h.store.ProjectHead.Create(db, head); err != nil {
				l.Error(err, "failed to create project head")
				return nil, http.StatusInternalServerError, err
			}

			slot.ProjectMember.Head = head
		}

		err = h.controller.Discord.Log(model.LogDiscordInput{
			Type: "project_member_add",
			Data: map[string]interface{}{
				"employee_id":         userInfo.UserID,
				"updated_employee_id": member.EmployeeID.String(),
				"project_name":        p.Name,
				"deployment_type":     member.DeploymentType.String(),
			},
		})
		if err != nil {
			h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to log project member add")
		}
	}

	return slot, http.StatusOK, nil
}

// Details godoc
// @Summary Get details of a project
// @Description Get details of a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Success 200 {object} view.ProjectDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id} [get]
func (h *handler) Details(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "Details",
		"id":      projectID,
	})

	rs, err := h.store.Project.One(h.repo.DB(), projectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) && !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadReadActive) {
		_, ok := userInfo.Projects[rs.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, rs.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}
	}

	if rs.Status == model.ProjectStatusClosed && !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsReadFullAccess) {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectData(rs, userInfo), nil, nil, nil, ""))
}

// UpdateGeneralInfo godoc
// @Summary Update general info of the project by id
// @Description Update general info of the project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body request.UpdateProjectGeneralInfoInput true "Body"
// @Success 200 {object} view.UpdateProjectGeneralInfoResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/general-info [put]
func (h *handler) UpdateGeneralInfo(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	var body request.UpdateProjectGeneralInfoInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if err := body.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateGeneralInfo",
		"id":      projectID,
		"request": body,
	})

	// Check project existence
	p, err := h.store.Project.One(h.repo.DB(), projectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, projectID, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	// Check country existence
	exist, err := h.store.Country.IsExist(h.repo.DB(), body.CountryID.String())
	if err != nil {
		l.Error(err, "error check existence of country")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(err, "country not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrCountryNotFound, body, ""))
		return
	}

	// Check bank account existence
	if !body.BankAccountID.IsZero() {
		exist, err := h.store.BankAccount.IsExist(h.repo.DB(), body.BankAccountID.String())
		if err != nil {
			l.Error(err, "error check existence of bank account")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}

		if !exist {
			l.Error(err, "bank account not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrBankAccountNotFound, body, ""))
			return
		}
	}

	// Check organization existence
	if !body.OrganizationID.IsZero() {
		exist, err := h.store.Organization.IsExist(h.repo.DB(), body.OrganizationID.String())
		if err != nil {
			l.Error(err, "error check existence of organization")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}

		if !exist {
			l.Error(err, "organization not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrOrganizationNotFound, body, ""))
			return
		}
	}

	// Check valid stack id
	_, stacks, err := h.store.Stack.All(h.repo.DB(), "", nil)
	if err != nil {
		l.Error(err, "error when finding stacks")
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

	_, err = time.Parse("2006-01-02", body.StartDate)
	if body.StartDate != "" && err != nil {
		l.Error(errs.ErrInvalidStartDate, "invalid start date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidStartDate, body, ""))
		return
	}

	if !body.ClientID.IsZero() {
		client, err := h.store.Client.One(h.repo.DB(), body.ClientID.String())
		if err != nil {
			l.Error(err, "client not found")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrClientNotFound, body, ""))
			return
		}

		p.Client = client
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Delete all exist employee stack
	if err := h.store.ProjectStack.DeleteByProjectID(tx.DB(), projectID); err != nil {
		l.Error(err, "failed to delete project stacks in database")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Create new employee stack
	for _, stackID := range body.Stacks {
		_, err := h.store.ProjectStack.Create(tx.DB(), &model.ProjectStack{
			ProjectID: model.MustGetUUIDFromString(projectID),
			StackID:   stackID,
		})
		if err != nil {
			l.Error(err, "failed to create project stack")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	p.Name = body.Name
	p.StartDate = body.GetStartDate()
	p.CountryID = body.CountryID
	p.Function = model.ProjectFunction(body.Function)
	p.BankAccountID = body.BankAccountID
	p.ClientID = body.ClientID
	p.OrganizationID = body.OrganizationID

	projectNotion, err := h.store.ProjectNotion.OneByProjectID(tx.DB(), p.ID.String())

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "failed to get project notion")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), projectID, ""))
			return
		} else if !body.AuditNotionID.IsZero() {
			// create new project notion
			_, err := h.store.ProjectNotion.Create(tx.DB(), &model.ProjectNotion{
				ProjectID:     p.ID,
				AuditNotionID: body.AuditNotionID,
			})
			if err != nil {
				l.Error(err, "failed to create project notion")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
				return
			}
		}
	} else {
		projectNotion.AuditNotionID = body.AuditNotionID
		// update audit notion id
		if _, err := h.store.ProjectNotion.UpdateSelectedFieldsByID(tx.DB(), projectNotion.ID.String(), *projectNotion); err != nil {
			l.Error(err, "failed to create project notion")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	// TODO: allow updating client_id
	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), projectID, *p,
		"name",
		"start_date",
		"country_id",
		"function",
		"bank_account_id",
		// "client_id",
		"organization_id",
	)

	if err != nil {
		l.Error(err, "failed to update project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectGeneralInfo(p), nil, done(nil), nil, ""))
}

// UpdateContactInfo godoc
// @Summary Update contact info of the project by id
// @Description Update contact info of the project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body request.UpdateContactInfoInput true "Body"
// @Success 200 {object} view.UpdateProjectContactInfoResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/contact-info [put]
func (h *handler) UpdateContactInfo(c *gin.Context) {
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	var body request.UpdateContactInfoInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateContactInfo",
		"id":      projectID,
		"request": body,
	})

	// Validate client email address
	if err := body.Validate(); err != nil {
		l.Error(err, "invalid input request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	// Check project existence
	p, err := h.store.Project.One(h.repo.DB(), projectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, projectID, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	err = h.updateProjectHeads(tx.DB(), projectID, model.HeadPositionAccountManager, body.AccountManagers, userInfo)
	if err != nil {
		l.Error(err, "failed to update account managers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	err = h.updateProjectHeads(tx.DB(), projectID, model.HeadPositionDeliveryManager, body.DeliveryManagers, userInfo)
	if err != nil {
		l.Error(err, "failed to update delivery managers")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	err = h.updateProjectHeads(tx.DB(), projectID, model.HeadPositionSalePerson, body.SalePersons, userInfo)
	if err != nil {
		l.Error(err, "failed to update sale persons")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Update email info
	p.ClientEmail = strings.Join(body.ClientEmail, ",")
	p.ProjectEmail = body.ProjectEmail

	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), projectID, *p, "client_email", "project_email")
	if err != nil {
		l.Error(err, "failed to update project information to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	heads, err := h.store.ProjectHead.GetActiveLeadsByProjectID(tx.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to get project heads")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	p.Heads = heads

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectContactInfo(p, userInfo), nil, done(nil), nil, ""))
}

func (h *handler) updateProjectHeads(db *gorm.DB, projectID string, position model.HeadPosition, headsInput []request.ProjectHeadInput, userInfo *model.CurrentLoggedUserInfo) error {
	heads, err := h.store.ProjectHead.GetByProjectIDAndPosition(db, projectID, position)
	if err != nil {
		h.logger.Fields(logger.Fields{
			"projectID": projectID,
			"position":  position,
		}).Error(err, "failed to get heads")
		return err
	}

	// create input map
	headInputMap := map[model.UUID]decimal.Decimal{}
	for _, head := range headsInput {
		exists, err := h.store.Employee.IsExist(db, head.EmployeeID.String())
		if err != nil {
			h.logger.Error(err, "failed to check employee existence")
			return err
		}

		if !exists {
			h.logger.Error(errs.ErrEmployeeNotFound, "employee not found")
			return errs.ErrEmployeeNotFound
		}

		headInputMap[head.EmployeeID] = head.CommissionRate
	}

	// update/delete exist heads
	for _, head := range heads {
		if _, ok := headInputMap[head.EmployeeID]; ok {
			if authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
				head.CommissionRate = headInputMap[head.EmployeeID]

				_, err := h.store.ProjectHead.UpdateSelectedFieldsByID(db, head.ID.String(), *head, "commission_rate")
				if err != nil {
					h.logger.Fields(logger.Fields{
						"projectID": projectID,
						"headID":    head.ID.String(),
					}).Error(err, "failed to update head")
					return err
				}
			}

			delete(headInputMap, head.EmployeeID)
		} else {
			if err := h.store.ProjectHead.DeleteByID(db, head.ID.String()); err != nil {
				h.logger.Fields(logger.Fields{
					"projectID": projectID,
					"headID":    head.ID.String(),
				}).Error(err, "failed to delete head")
				return err
			}
		}
	}

	// create new head
	for employeeID, commissionRate := range headInputMap {
		if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
			commissionRate = decimal.Zero
		}

		head := &model.ProjectHead{
			BaseModel: model.BaseModel{
				ID: model.NewUUID(),
			},
			ProjectID:      model.MustGetUUIDFromString(projectID),
			EmployeeID:     employeeID,
			CommissionRate: commissionRate,
			Position:       position,
		}

		if err := h.store.ProjectHead.Create(db, head); err != nil {
			h.logger.AddField("head", head).Error(err, "failed to create head")
			return err
		}
	}

	return nil
}

func (h *handler) updateProjectLead(db *gorm.DB, projectID string, employeeID model.UUID, startDate *time.Time, endDate *time.Time, commissionRate decimal.Decimal, userInfo *model.CurrentLoggedUserInfo) (*model.ProjectHead, error) {
	head, err := h.store.ProjectHead.One(db, projectID, employeeID.String(), model.HeadPositionTechnicalLead)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Fields(logger.Fields{
			"projectID":  projectID,
			"employeeID": employeeID,
		}).Error(err, "failed to get tecihnical lead")
		return nil, err
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectsCommissionRateEdit) {
		commissionRate = decimal.Zero
	} else if head != nil {
		head.CommissionRate = commissionRate
	}

	if err == nil {
		// Update old record
		head.StartDate = *startDate
		head.EndDate = endDate

		_, err := h.store.ProjectHead.UpdateSelectedFieldsByID(db, head.ID.String(), *head,
			"start_date",
			"end_date",
			"commission_rate",
		)
		if err != nil {
			h.logger.Fields(logger.Fields{"head": *head}).Error(err, "failed to update project head")
			return nil, err
		}
	} else {
		// Create new record
		head = &model.ProjectHead{
			ProjectID:      model.MustGetUUIDFromString(projectID),
			EmployeeID:     employeeID,
			CommissionRate: commissionRate,
			StartDate:      *startDate,
			EndDate:        endDate,
			Position:       model.HeadPositionTechnicalLead,
		}

		if err := h.store.ProjectHead.Create(db, head); err != nil {
			h.logger.Fields(logger.Fields{"head": head}).Error(err, "failed to create project head")
			return nil, err
		}
	}

	return head, nil
}

// GetWorkUnits godoc
// @Summary Get list work units of a project
// @Description Get list work units of a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param status query  model.WorkUnitStatus false "status"
// @Success 200 {object} view.ListWorkUnitResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/work-units [get]
func (h *handler) GetWorkUnits(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	input := request.GetListWorkUnitInput{
		ProjectID: c.Param("id"),
	}

	if err := c.ShouldBindQuery(&input.Query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input.Query, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler":   "project",
		"method":    "GetWorkUnits",
		"projectID": input.ProjectID,
		"query":     input.Query,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, input, ""))
			return
		}

		l.Info("failed to check if project exists")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectWorkUnitsReadFullAccess) {
		_, ok := userInfo.Projects[p.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}
	}

	workUnits, err := h.store.WorkUnit.GetByProjectID(h.repo.DB(), input.ProjectID, input.Query.Status)
	if err != nil {
		l.Error(err, "failed to get work units")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input.ProjectID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToWorkUnitList(workUnits, input.ProjectID, p.Code), nil, nil, nil, ""))
}

// CreateWorkUnit godoc
// @Summary Create work unit of a project
// @Description Get work unit of a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param body body request.CreateWorkUnitBody true "Body"
// @Success 200 {object} view.WorkUnitResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/work-units [post]
func (h *handler) CreateWorkUnit(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	input := request.CreateWorkUnitInput{
		ProjectID: c.Param("id"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "CreateWorkUnit",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// Has permission when have work unit create full-access and active in project
	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectWorkUnitsCreateFullAccess) {
		_, ok := userInfo.Projects[p.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		leadMap := map[string]bool{}
		for _, v := range p.Heads {
			if v.IsLead() {
				leadMap[v.EmployeeID.String()] = true
			}
		}

		_, ok = leadMap[userInfo.UserID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMemberIsNotProjectLead, nil, ""))
			return
		}
	}

	tx, done := h.repo.NewTransaction()

	workUnit := &model.WorkUnit{
		Name:      input.Body.Name,
		Type:      model.WorkUnitType(input.Body.Type),
		Status:    model.WorkUnitStatus(input.Body.Status),
		SourceURL: input.Body.URL,
		ProjectID: model.MustGetUUIDFromString(input.ProjectID),
	}

	if err := h.store.WorkUnit.Create(tx.DB(), workUnit); err != nil {
		l.Error(err, "failed to create new work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	stacks, err := h.store.Stack.GetByIDs(tx.DB(), input.Body.Stacks)
	if err != nil {
		l.Error(err, "failed to get stacks")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// create work unit stack
	for _, stack := range stacks {
		wuStack := model.WorkUnitStack{
			StackID:    stack.ID,
			WorkUnitID: workUnit.ID,
		}
		if err := h.store.WorkUnitStack.Create(tx.DB(), &wuStack); err != nil {
			l.Error(err, "failed to create new work unit stack")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
			return
		}

		wuStack.Stack = *stack
		workUnit.WorkUnitStacks = append(workUnit.WorkUnitStacks, &wuStack)
	}

	employees, err := h.store.Employee.GetByIDs(tx.DB(), input.Body.Members)
	if err != nil {
		l.Error(err, "failed to get employees")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// create work unit member
	for _, employee := range employees {
		pMember, err := h.store.ProjectMember.GetActiveMemberInProject(tx.DB(), input.ProjectID, employee.ID.String())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrMemberIsNotActiveInProject, "member is not active in project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrMemberIsNotActiveInProject), input, ""))
			return
		}

		wuMember := model.WorkUnitMember{
			Status:     model.ProjectMemberStatusActive.String(),
			WorkUnitID: workUnit.ID,
			EmployeeID: employee.ID,
			ProjectID:  model.MustGetUUIDFromString(input.ProjectID),
			StartDate:  *pMember.StartDate,
		}
		if err := h.store.WorkUnitMember.Create(tx.DB(), &wuMember); err != nil {
			l.Error(err, "failed to create new work unit member")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
			return
		}

		wuMember.Employee = *employee
		workUnit.WorkUnitMembers = append(workUnit.WorkUnitMembers, &wuMember)
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToWorkUnit(workUnit, p.Code), nil, done(nil), nil, ""))
}

// UpdateWorkUnit godoc
// @Summary Update work unit info
// @Description Update work unit info
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param workUnitID path string true "Work Unit ID"
// @Param Body body request.UpdateWorkUnitInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/work-units/{workUnitID} [put]
func (h *handler) UpdateWorkUnit(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	input := request.UpdateWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input.Body, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateWorkUnit",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectWorkUnitsEditFullAccess) {
		_, ok := userInfo.Projects[p.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		leadMap := map[string]bool{}
		for _, v := range p.Heads {
			if v.IsLead() {
				leadMap[v.EmployeeID.String()] = true
			}
		}

		_, ok = leadMap[userInfo.UserID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMemberIsNotProjectLead, nil, ""))
			return
		}
	}

	// Check Exitsences of elements in input
	status, err := h.checkExitsInUpdateWorkUnitInput(h.repo.DB(), input)

	if err != nil {
		l.Error(err, "err when checking the existence of elements in the input")
		c.JSON(status, view.CreateResponse[any](nil, nil, err, input.Body, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	workUnit := &model.WorkUnit{
		Name:      input.Body.Name,
		Type:      input.Body.Type,
		SourceURL: input.Body.URL,
		ProjectID: model.MustGetUUIDFromString(input.ProjectID),
	}

	_, err = h.store.WorkUnit.UpdateSelectedFieldsByID(tx.DB(), input.WorkUnitID, *workUnit, "name", "type", "source_url", "project_id")
	if err != nil {
		l.Error(err, "failed to update work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Update work unit stack
	if err := h.updateWorkUnitStack(tx.DB(), input.WorkUnitID, input.Body.Stacks); err != nil {
		l.Error(err, "failed to update work unit stack")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Get all active members of work unit
	members, err := h.store.WorkUnitMember.All(tx.DB(), input.WorkUnitID)
	if err != nil {
		l.Error(err, "failed to get all work unit members in database")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	var curMemberIDs []model.UUID

	for _, v := range members {
		curMemberIDs = append(curMemberIDs, v.EmployeeID)
	}

	// Get map for employee in work unit member
	inputMemberIDs := map[model.UUID]string{}
	for _, member := range input.Body.Members {
		inputMemberIDs[member] = member.String()
	}

	// Get delete member id list
	var deleteMemberIDs []string

	for _, v := range curMemberIDs {
		_, ok := inputMemberIDs[v]
		if !ok {
			deleteMemberIDs = append(deleteMemberIDs, v.String())
		} else {
			delete(inputMemberIDs, v)
		}
	}

	// Get create member id list
	var createMemberIDs []model.UUID

	for id := range inputMemberIDs {
		createMemberIDs = append(createMemberIDs, id)
	}

	// Delete work unit members
	if status, err = h.deleteWorkUnit(tx.DB(), input.WorkUnitID, deleteMemberIDs); err != nil {
		l.Error(err, "failed to remove work unit member in database")
		c.JSON(status, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Create new work unit member
	if status, err := h.createWorkUnit(tx.DB(), input.ProjectID, input.WorkUnitID, createMemberIDs); err != nil {
		l.Error(err, "failed to create new work unit member")
		c.JSON(status, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

func (h *handler) checkExitsInUpdateWorkUnitInput(db *gorm.DB, input request.UpdateWorkUnitInput) (int, error) {
	// Check project existence
	exists, err := h.store.Project.IsExist(db, input.ProjectID)
	if err != nil {
		return http.StatusInternalServerError, errs.ErrFailToCheckInputExistence
	}

	if !exists {
		return http.StatusNotFound, errs.ErrProjectNotFound
	}

	// Check work unit existence
	exists, err = h.store.WorkUnit.IsExists(h.repo.DB(), input.WorkUnitID)
	if err != nil {
		return http.StatusInternalServerError, errs.ErrFailToCheckInputExistence
	}

	if !exists {
		return http.StatusNotFound, errs.ErrProjectNotFound
	}

	// Check stack existence
	_, stacks, err := h.store.Stack.All(h.repo.DB(), "", nil)
	if err != nil {
		return http.StatusInternalServerError, errs.ErrFailToCheckInputExistence
	}

	stackMap := model.ToStackMap(stacks)
	for _, sID := range input.Body.Stacks {
		_, ok := stackMap[sID]
		if !ok {
			return http.StatusNotFound, errs.ErrStackNotFoundWithID(sID.String())
		}
	}

	return 0, nil
}

func (h *handler) updateWorkUnitStack(db *gorm.DB, workUnitID string, stackIDs []model.UUID) error {
	// Delete all exist work unit stack
	if err := h.store.WorkUnitStack.DeleteByWorkUnitID(db, workUnitID); err != nil {
		return errs.ErrFailToDeleteWorkUnitStack
	}

	// Create new work unit stack
	for _, stackID := range stackIDs {
		err := h.store.WorkUnitStack.Create(db, &model.WorkUnitStack{
			WorkUnitID: model.MustGetUUIDFromString(workUnitID),
			StackID:    stackID,
		})
		if err != nil {
			return errs.ErrFailedToCreateWorkUnitStack
		}
	}

	return nil
}

func (h *handler) deleteWorkUnit(db *gorm.DB, workUnitID string, deleteMemberIDList []string) (int, error) {
	now := time.Now()
	for _, deleteMemberID := range deleteMemberIDList {
		workUnitMember, err := h.store.WorkUnitMember.One(db,
			workUnitID,
			deleteMemberID,
			model.WorkUnitMemberStatusActive.String())

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, errs.ErrInvalidInActiveMember
		}
		if err != nil {
			return http.StatusInternalServerError, errs.ErrFailedToGetWorkUnitMember
		}

		deleteMember := &model.WorkUnitMember{
			Status:  model.WorkUnitMemberStatusInactive.String(),
			EndDate: &now,
		}

		if _, err = h.store.WorkUnitMember.UpdateSelectedFieldsByID(db, workUnitMember.ID.String(), *deleteMember, "status", "end_date"); err != nil {
			return http.StatusInternalServerError, errs.ErrFailedToUpdateWorkUnitMember
		}

		if err = h.store.WorkUnitMember.SoftDeleteByWorkUnitID(db, workUnitMember.ID.String(), deleteMemberID); err != nil {
			return http.StatusInternalServerError, errs.ErrFailedToSoftDeleteWorkUnitMember
		}
	}

	return 0, nil
}

func (h *handler) createWorkUnit(db *gorm.DB, projectID string, workUnitID string, createMemberIDList []model.UUID) (int, error) {
	for _, createMemberID := range createMemberIDList {
		_, err := h.store.ProjectMember.GetActiveMemberInProject(db, projectID, createMemberID.String())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusBadRequest, errs.ErrMemberIsInactive
		}

		if err != nil {
			return http.StatusInternalServerError, errs.ErrFailedToGetProjectMember
		}

		now := time.Now()
		wuMember := model.WorkUnitMember{
			Status:     model.ProjectMemberStatusActive.String(),
			WorkUnitID: model.MustGetUUIDFromString(workUnitID),
			EmployeeID: createMemberID,
			ProjectID:  model.MustGetUUIDFromString(projectID),
			StartDate:  now,
		}

		if err := h.store.WorkUnitMember.Create(db, &wuMember); err != nil {
			return http.StatusInternalServerError, errs.ErrFailedToCreateWorkUnitMember
		}
	}

	return 0, nil
}

// ArchiveWorkUnit godoc
// @Summary Archive an active work unit of a project
// @Description Archive an active work unit of a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param workUnitID path string true  "Work Unit ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/work-units/{workUnitID}/archive [put]
func (h *handler) ArchiveWorkUnit(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	input := request.ArchiveWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "ArchiveWorkUnit",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectWorkUnitsEditFullAccess) {
		_, ok := userInfo.Projects[p.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		leadMap := map[string]bool{}
		for _, v := range p.Heads {
			if v.IsLead() {
				leadMap[v.EmployeeID.String()] = true
			}
		}

		_, ok = leadMap[userInfo.UserID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMemberIsNotProjectLead, nil, ""))
			return
		}
	}

	workUnit, err := h.store.WorkUnit.One(h.repo.DB(), input.WorkUnitID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "work unit not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrWorkUnitNotFound, nil, ""))
		return
	}
	if err != nil {
		l.Error(err, "failed to get one work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	workUnit.Status = model.WorkUnitStatusArchived

	// update work unit status -> 'archived'
	_, err = h.store.WorkUnit.UpdateSelectedFieldsByID(tx.DB(), input.WorkUnitID, *workUnit, "status")
	if err != nil {
		l.Error(err, "failed to update work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	wuMembers, err := h.store.WorkUnitMember.GetByWorkUnitID(tx.DB(), input.WorkUnitID)
	if err != nil {
		l.Error(err, "failed to get work unit members")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// update work unit member: end_date = now() and status = 'inactive'
	timeNow := time.Now()
	for _, member := range wuMembers {
		member.EndDate = &timeNow
		member.Status = model.ProjectMemberStatusInactive.String()

		_, err := h.store.WorkUnitMember.UpdateSelectedFieldsByID(tx.DB(), member.ID.String(), *member, "end_date", "status")
		if err != nil {
			l.Error(err, "failed to get work unit members")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// UnarchiveWorkUnit godoc
// @Summary Unarchive an archive work unit of a project
// @Description Unarchive an archive work unit of a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param workUnitID path string true "Work Unit ID"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/work-units/{workUnitID}/unarchive [put]
func (h *handler) UnarchiveWorkUnit(c *gin.Context) {
	// 0. Get current logged in user data
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, userInfo.UserID, ""))
		return
	}

	input := request.ArchiveWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UnarchiveWorkUnit",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		l.Error(err, "validate failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	p, err := h.store.Project.One(h.repo.DB(), input.ProjectID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(errs.ErrProjectNotFound, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !authutils.HasPermission(userInfo.Permissions, model.PermissionProjectWorkUnitsEditFullAccess) {
		_, ok := userInfo.Projects[p.ID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
			return
		}

		leadMap := map[string]bool{}
		for _, v := range p.Heads {
			if v.IsLead() {
				leadMap[v.EmployeeID.String()] = true
			}
		}

		_, ok = leadMap[userInfo.UserID]
		if !ok || !model.IsUserActiveInProject(userInfo.UserID, p.ProjectMembers) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrMemberIsNotProjectLead, nil, ""))
			return
		}
	}

	workUnit, err := h.store.WorkUnit.One(h.repo.DB(), input.WorkUnitID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Error(err, "work unit not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrWorkUnitNotFound, nil, ""))
		return
	}
	if err != nil {
		l.Error(err, "failed to get one work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	workUnit.Status = model.WorkUnitStatusActive

	// update work unit status -> 'active'
	_, err = h.store.WorkUnit.UpdateSelectedFieldsByID(tx.DB(), input.WorkUnitID, *workUnit, "status")
	if err != nil {
		l.Error(err, "failed to update work unit")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	wuMembers, err := h.store.WorkUnitMember.GetByWorkUnitID(tx.DB(), input.WorkUnitID)
	if err != nil {
		l.Error(err, "failed to get work unit members")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// check member status in project and update work unit member
	for _, member := range wuMembers {
		// _, err := h.store.ProjectMember.GetActiveMemberInProject(tx.DB(), input.ProjectID, member.EmployeeID.String(), false)

		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// 	l.Error(err, "member is not active in project")
		// 	c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrMemberIsInactive), nil, ""))
		// 	return
		// }
		// if err != nil {
		// 	l.Error(err, "failed to get one project member")
		// 	c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		// 	return
		// }

		member.EndDate = nil
		member.Status = model.ProjectMemberStatusActive.String()

		_, err = h.store.WorkUnitMember.UpdateSelectedFieldsByID(tx.DB(), member.ID.String(), *member, "end_date", "status")
		if err != nil {
			l.Error(err, "failed to get work unit members")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
}

// UpdateSendingSurveyState godoc
// @Summary Update allows sending survey for project by id
// @Description Update allows sending survey for project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param allowsSendingSurvey query bool false "Allows sending survey"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/sending-survey-state [put]
func (h *handler) UpdateSendingSurveyState(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	query := request.UpdateSendingSurveyInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateSendingSurveyState",
		"query":   query,
	})

	p, err := h.store.Project.One(h.repo.DB(), projectID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, projectID, ""))
			return
		}

		l.Error(err, "failed to get project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	p.AllowsSendingSurvey = query.AllowsSendingSurvey
	_, err = h.store.Project.UpdateSelectedFieldsByID(h.repo.DB(), projectID, *p, "allows_sending_survey")
	if err != nil {
		l.Error(err, "failed to update project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// UploadAvatar godoc
// @Summary Upload avatar of project by id
// @Description Upload avatar of project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param id path string true "Project ID"
// @Param Authorization header string true "jwt token"
// @Param file formData file true "avatar upload"
// @Success 200 {object} view.ProjectContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/upload-avatar [post]
func (h *handler) UploadAvatar(c *gin.Context) {
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
		"handler": "project",
		"method":  "UploadAvatar",
		"params":  params,
		// "body":    body,
	})

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/projects/%s/images/%s", h.config.Google.GCSBucketName, params.ID, fileName)
	gcsPath := fmt.Sprintf("projects/%s/images/%s", params.ID, fileName)
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

	// 2.2 check project existed
	existed, err := h.store.Project.IsExist(tx.DB(), params.ID)
	if err != nil {
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}
	if !existed {
		l.Info("project not existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrProjectNotExisted), nil, ""))
		return
	}

	// 2.3 upload to GCS
	multipart, err := file.Open()
	if err != nil {
		l.Error(err, "error in open file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	err = h.service.Google.UploadContentGCS(multipart, gcsPath)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// 3. update avatar field
	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), params.ID, model.Project{
		Avatar: filePath,
	}, "avatar")
	if err != nil {
		l.Error(err, "error in update avatar")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToProjectContentData(filePath), nil, done(nil), nil, ""))
}

// SyncProjectMemberStatus godoc
// @Summary Sync project member status
// @Description Sync project member status
// @Tags Project
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /cron-jobs/sync-project-member-status [put]
func (h *handler) SyncProjectMemberStatus(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "SyncProjectMemberStatus",
	})

	err := h.store.ProjectMember.UpdateEndDateOverdueMemberToInActive(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to update end date overdue member status to inactive")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.store.ProjectMember.UpdateMemberInClosedProjectToInActive(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to update member in closed/paused project status to inactive")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.store.ProjectMember.UpdateLeftMemberToInActive(h.repo.DB())
	if err != nil {
		l.Error(err, "failed to update left member project status to inactive")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
