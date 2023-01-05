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
	"github.com/dwarvesf/fortress-api/pkg/handler/project/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/project/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/store/projectslot"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

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
	query := request.GetListProjectInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	query.StandardizeInput()

	// TODO: can we move this to middleware ?
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

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectsData(projects),
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

	// TODO: can we move this to middleware ?
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

	project, err := h.store.Project.One(h.repo.DB(), projectID, false)
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

	project.Status = body.ProjectStatus
	_, err = h.store.Project.UpdateSelectedFieldsByID(h.repo.DB(), projectID, *project, "status")
	if err != nil {
		l.Error(err, "error query update status for project to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectStatusResponse(project), nil, nil, nil, ""))
}

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
	body := request.CreateProjectInput{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// TODO: can we move this to middleware ?
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

	p := &model.Project{
		Name:         body.Name,
		CountryID:    body.CountryID,
		Type:         model.ProjectType(body.Type),
		Status:       model.ProjectStatus(body.Status),
		StartDate:    body.GetStartDate(),
		ProjectEmail: body.ProjectEmail,
		ClientEmail:  body.ClientEmail,
		Country:      country,
		Code:         body.Code,
	}

	tx, done := h.repo.NewTransaction()

	if err := h.store.Project.Create(tx.DB(), p); err != nil {
		l.Error(err, "failed to create project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// create project account manager
	accountManager := model.ProjectHead{
		ProjectID:      p.ID,
		EmployeeID:     body.AccountManagerID,
		JoinedDate:     time.Now(),
		CommissionRate: decimal.Zero,
		Position:       model.HeadPositionAccountManager,
	}

	if err := h.store.ProjectHead.Create(tx.DB(), &accountManager); err != nil {
		l.Error(err, "failed to create account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	p.Heads = append(p.Heads, &accountManager)

	// create project delivery manager
	if !body.DeliveryManagerID.IsZero() {
		deliveryManager := model.ProjectHead{
			ProjectID:      p.ID,
			EmployeeID:     body.DeliveryManagerID,
			JoinedDate:     time.Now(),
			CommissionRate: decimal.Zero,
			Position:       model.HeadPositionDeliveryManager,
		}

		if err := h.store.ProjectHead.Create(tx.DB(), &deliveryManager); err != nil {
			l.Error(err, "failed to create delivery manager")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		p.Heads = append(p.Heads, &deliveryManager)
	}

	// assign members to project
	for _, member := range body.Members {
		slot, code, err := h.createSlotInProject(tx.DB(), p.ID.String(), member)
		if err != nil {
			l.Error(err, "failed to assign member to project")
			c.JSON(code, view.CreateResponse[any](nil, nil, done(err), member, ""))
			return
		}

		p.Slots = append(p.Slots, *slot)
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateProjectDataResponse(p), nil, done(nil), nil, ""))
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
// @Success 200 {object} view.ProjectMemberListResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [get]
func (h *handler) GetMembers(c *gin.Context) {
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

	// TODO: can we move this to middleware ?
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

	exists, err := h.store.Project.IsExist(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
		return
	}

	members, total, err := h.store.ProjectSlot.All(h.repo.DB(), projectslot.GetListProjectSlotInput{
		ProjectID: projectID,
		Status:    query.Status,
		Preload:   query.Preload,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "failed to get project members")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	heads, err := h.store.ProjectHead.GetActiveLeadsByProjectID(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to get project heads")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectMemberListData(members, heads),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil, ""))
}

// DeleteMember godoc
// @Summary Delete member in a project
// @Description Delete member in a project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "project ID"
// @Param memberID path string true "employee ID"
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

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "DeleteMember",
		"body":    input.MemberID,
	})

	// get member info
	projectMember, err := h.store.ProjectMember.One(h.repo.DB(), input.ProjectID, input.MemberID, false)
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

	slotID := projectMember.ProjectSlotID.String()

	err = h.store.ProjectMemberPosition.DeleteByProjectMemberID(tx.DB(), projectMember.ID.String())
	if err != nil {
		l.Error(err, "failed to delete project member position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectMember.Delete(tx.DB(), projectMember.ID.String())
	if err != nil {
		l.Error(err, "error delete project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectSlotPosition.DeleteByProjectSlotID(tx.DB(), slotID)
	if err != nil {
		l.Error(err, "error delete project slot position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectSlot.Delete(tx.DB(), slotID)
	if err != nil {
		l.Error(err, "error delete project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectHead.DeleteByPositionInProject(tx.DB(),
		projectMember.ProjectID.String(),
		projectMember.EmployeeID.String(),
		model.HeadPositionTechnicalLead.String())
	if err != nil {
		l.Error(err, "error delete project head")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, done(nil), nil, "ok"))
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
	// TODO: add test
	input := request.UnassignMemberInput{
		ProjectID: c.Param("id"),
		MemberID:  c.Param("memberID"),
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UnassignMember",
		"body":    input,
	})

	// get member info
	projectMember, err := h.store.ProjectMember.One(h.repo.DB(), input.ProjectID, input.MemberID, false)
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
	projectMember.LeftDate = &timeNow
	projectMember.Status = model.ProjectMemberStatusInactive

	_, err = h.store.ProjectMember.UpdateSelectedFieldsByID(tx.DB(),
		projectMember.ID.String(),
		*projectMember,
		"left_date",
		"status")
	if err != nil {
		l.Error(err, "failed to update project member")
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

	// TODO: can we move this to middleware ?
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
	exists, err := h.store.Project.IsExist(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, body, ""))
		return
	}

	// check seniority existence
	exists, err = h.store.Seniority.IsExist(h.repo.DB(), body.SeniorityID.String())
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

	// if slot.Status == model.ProjectMemberStatusInactive {
	// 	l.Info("slot is inactive")
	// 	c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrSlotIsInactive, body, ""))
	// 	return
	// }

	tx, done := h.repo.NewTransaction()

	if !body.EmployeeID.IsZero() {
		member, err := h.assignMemberToProject(tx.DB(), slot.ID.String(), projectID, body)
		if err != nil {
			l.Error(err, "failed to assign member to project")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		slot.ProjectMember = *member
	}

	// update project slot
	slot.SeniorityID = body.SeniorityID
	slot.DeploymentType = model.DeploymentType(body.DeploymentType)
	slot.Status = model.ProjectMemberStatus(body.Status)
	slot.Rate = body.Rate
	slot.Discount = body.Discount

	_, err = h.store.ProjectSlot.UpdateSelectedFieldsByID(tx.DB(), body.ProjectSlotID.String(), *slot,
		"seniority_id",
		"deployment_type",
		"status",
		"rate",
		"discount")
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
		l.Error(err, "failed to create project member positions")
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

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(slot), nil, done(nil), nil, ""))
}

func (h *handler) assignMemberToProject(db *gorm.DB, slotID string, projectID string, input request.UpdateMemberInput) (*model.ProjectMember, error) {
	// check is slot contains any member?
	member, err := h.store.ProjectMember.OneBySlotID(db, slotID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Fields(logger.Fields{"slotID": slotID}).Error(err, "failed to get project member by slotID")
		return nil, err
	}

	// if member.Status == model.ProjectMemberStatusInactive {
	// 	h.logger.Fields(logger.Fields{"member": member}).Error(errs.ErrSlotIsInactive, "slot is inactive")
	// 	return nil, errs.ErrSlotIsInactive
	// }

	if !member.EmployeeID.IsZero() && member.EmployeeID != input.EmployeeID {
		h.logger.
			Fields(logger.Fields{"member": member}).
			Error(errs.ErrSlotAlreadyContainsAnotherMember, "slot already contains another member")
		return nil, errs.ErrSlotAlreadyContainsAnotherMember
	}

	// check is member active in project?
	member, err = h.store.ProjectMember.One(db, projectID, input.EmployeeID.String(), true)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Fields(logger.Fields{
			"projectID":  projectID,
			"employeeID": input.EmployeeID,
		}).Error(err, "failed to get project member")
		return nil, err
	}

	// if member is not active in project, create new member;
	// else check condition and update member
	if errors.Is(err, gorm.ErrRecordNotFound) {
		member = &model.ProjectMember{
			ProjectID:      model.MustGetUUIDFromString(projectID),
			EmployeeID:     input.EmployeeID,
			SeniorityID:    input.SeniorityID,
			ProjectSlotID:  model.MustGetUUIDFromString(slotID),
			DeploymentType: model.DeploymentType(input.DeploymentType),
			Status:         model.ProjectMemberStatus(input.Status),
			JoinedDate:     input.GetJoinedDate(),
			LeftDate:       input.GetLeftDate(),
			Rate:           input.Rate,
			Discount:       input.Discount,
		}

		if err := h.store.ProjectMember.Create(db, member); err != nil {
			h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to create project member")
			return nil, err
		}
	} else {
		if member.ProjectSlotID.String() != slotID {
			h.logger.Fields(logger.Fields{
				"member.ProjectSlotID": member.ProjectSlotID,
				"slotID":               slotID,
			}).Info("slotID cannot be changed")
			return nil, errs.ErrSlotIDCannotBeChanged
		}

		// if member.Status == model.ProjectMemberStatusInactive {
		// 	h.logger.Fields(logger.Fields{"status": member.Status}).Info("member is inactive")
		// 	return nil, errs.ErrMemberIsInactive
		// }

		member.SeniorityID = input.SeniorityID
		member.DeploymentType = model.DeploymentType(input.DeploymentType)
		member.Status = model.ProjectMemberStatus(input.Status)
		member.LeftDate = input.GetLeftDate()
		member.Rate = input.Rate
		member.Discount = input.Discount

		_, err := h.store.ProjectMember.UpdateSelectedFieldsByID(db, member.ID.String(), *member,
			"left_date",
			"status",
			"rate",
			"discount",
			"deployment_type",
			"seniority_id")
		if err != nil {
			h.logger.Fields(logger.Fields{"member": member}).Error(err, "failed to update project member")
			return nil, err
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

	member.ProjectMemberPositions = memberPos

	// update project head
	member.IsLead = input.IsLead
	if input.IsLead {
		if _, err := h.updateProjectHead(db, projectID, input.EmployeeID, model.HeadPositionTechnicalLead); err != nil {
			h.logger.Fields(logger.Fields{
				"projectID":  projectID,
				"employeeID": input.EmployeeID,
			}).Error(err, "failed to update technicalLeads")
			return nil, err
		}
	} else {
		_, err := h.store.ProjectHead.UpdateLeftDateOfEmployee(db,
			input.EmployeeID.String(),
			projectID,
			model.HeadPositionTechnicalLead.String())
		if err != nil {
			h.logger.Fields(logger.Fields{
				"projectID":  projectID,
				"employeeID": input.EmployeeID,
			}).Error(err, "failed to update left_date of project head")
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

	// TODO: can we move this to middleware ?
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
	// 	_, err := h.store.ProjectMember.One(h.repo.DB(), projectID, body.EmployeeID.String(), false)
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
	exists, err := h.store.Project.IsExist(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, body, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	slot, code, err := h.createSlotInProject(tx.DB(), projectID, body)
	if err != nil {
		l.Error(err, "failed to assign member to project")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), body, ""))
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(slot), nil, done(nil), nil, ""))
}

func (h *handler) createSlotInProject(db *gorm.DB, projectID string, req request.AssignMemberInput) (*model.ProjectSlot, int, error) {
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
		l.Error(err, "error when finding position")
		return nil, http.StatusInternalServerError, err
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range req.Positions {
		if _, ok := positionMap[pID]; !ok {
			l.Error(errs.ErrPositionNotFoundWithID(pID.String()), "error position not found")
			return nil, http.StatusNotFound, errs.ErrPositionNotFoundWithID(pID.String())
		}
	}

	// create project slot
	slot := &model.ProjectSlot{
		ProjectID:      model.MustGetUUIDFromString(projectID),
		DeploymentType: model.DeploymentType(req.DeploymentType),
		Status:         model.ProjectMemberStatus(req.Status),
		Rate:           req.Rate,
		Discount:       req.Discount,
		SeniorityID:    req.SeniorityID,
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

		// create project member
		member := &model.ProjectMember{
			ProjectID:      model.MustGetUUIDFromString(projectID),
			EmployeeID:     req.EmployeeID,
			SeniorityID:    req.SeniorityID,
			ProjectSlotID:  slot.ID,
			DeploymentType: model.DeploymentType(req.DeploymentType),
			Status:         model.ProjectMemberStatus(req.Status),
			JoinedDate:     req.GetJoinedDate(),
			LeftDate:       req.GetLeftDate(),
			Rate:           req.Rate,
			Discount:       req.Discount,
		}

		if err = h.store.ProjectMember.Create(db, member); err != nil {
			l.Error(err, "failed to create project member")
			return nil, http.StatusInternalServerError, err
		}

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
			if err := h.store.ProjectHead.Create(db, &model.ProjectHead{
				ProjectID:      model.MustGetUUIDFromString(projectID),
				EmployeeID:     req.EmployeeID,
				CommissionRate: decimal.Zero,
				Position:       model.HeadPositionTechnicalLead,
				JoinedDate:     time.Now(),
			}); err != nil {
				l.Error(err, "failed to create project head")
				return nil, http.StatusInternalServerError, err
			}
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
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidProjectID, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
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

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectData(rs), nil, nil, nil, ""))
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

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateGeneralInfo",
		"id":      projectID,
		"request": body,
	})

	// Check project existence
	project, err := h.store.Project.One(h.repo.DB(), projectID, true)
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

	// Check valid stack id
	stacks, err := h.store.Stack.All(h.repo.DB())
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

	project.Name = body.Name
	project.StartDate = body.GetStartDate()
	project.CountryID = body.CountryID

	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), projectID, *project,
		"name",
		"start_date",
		"country_id")

	if err != nil {
		l.Error(err, "failed to update project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectGeneralInfo(project), nil, done(nil), nil, ""))
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

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateContactInfo",
		"id":      projectID,
		"request": body,
	})

	// Check project existence
	project, err := h.store.Project.One(h.repo.DB(), projectID, true)
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

	// Check account manager exists
	exist, err := h.store.Employee.IsExist(h.repo.DB(), body.AccountManagerID.String())
	if err != nil {
		l.Error(err, "failed to find account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(errs.ErrAccountManagerNotFound, "account manager not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrAccountManagerNotFound, body, ""))
		return
	}

	// Check delivery manager exists
	exist, err = h.store.Employee.IsExist(h.repo.DB(), body.DeliveryManagerID.String())
	if err != nil {
		l.Error(err, "error when finding delivery manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exist {
		l.Error(errs.ErrDeliveryManagerNotFound, "delivery manager not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrDeliveryManagerNotFound, body, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Update Account Manager
	_, err = h.updateProjectHead(tx.DB(), projectID, body.AccountManagerID, model.HeadPositionAccountManager)
	if err != nil {
		l.Error(err, "failed to update account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Update Delivery Manager
	_, err = h.updateProjectHead(tx.DB(), projectID, body.DeliveryManagerID, model.HeadPositionDeliveryManager)
	if err != nil {
		l.Error(err, "failed to update delivery manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Update email info
	project.ClientEmail = body.ClientEmail
	project.ProjectEmail = body.ProjectEmail

	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), projectID, *project, "client_email", "project_email")
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

	project.Heads = heads

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectContactInfo(project), nil, done(nil), nil, ""))
}

func (h *handler) updateProjectHead(db *gorm.DB, projectID string, employeeID model.UUID, position model.HeadPosition) (*model.ProjectHead, error) {
	timeNow := time.Now()

	head, err := h.store.ProjectHead.One(db, projectID, position)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Fields(logger.Fields{
			"projectID": projectID,
			"position":  position,
		}).Error(err, "failed to get one project head")
		return nil, err
	}

	// - For delivery-manager & account-manager:
	// 		if employee is the head, do nothing;
	// 		else update left_date of old head and create new head
	// - For technical-lead:
	//		if employee is the head, do nothing;
	//		else create new head
	if err == nil {
		if head.EmployeeID == employeeID {
			return head, nil
		}

		if position == model.HeadPositionAccountManager || position == model.HeadPositionDeliveryManager {
			head.LeftDate = &timeNow
			_, err := h.store.ProjectHead.UpdateSelectedFieldsByID(db, head.ID.String(), *head, "left_date")
			if err != nil {
				h.logger.Fields(logger.Fields{"head": *head}).Error(err, "failed to update project head")
				return nil, err
			}
		}
	}

	head = &model.ProjectHead{
		ProjectID:  model.MustGetUUIDFromString(projectID),
		EmployeeID: employeeID,
		JoinedDate: timeNow,
		Position:   position,
	}
	if err := h.store.ProjectHead.Create(db, head); err != nil {
		h.logger.Fields(logger.Fields{"head": head}).Error(err, "failed to create project head")
		return nil, err
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
	input := request.GetListWorkUnitInput{
		ProjectID: c.Param("id"),
	}

	if err := c.ShouldBindQuery(&input.Query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input.Query, ""))
		return
	}

	// TODO: can we move this to middleware ?
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

	project, err := h.store.Project.One(h.repo.DB(), input.ProjectID, false)
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

	workUnits, err := h.store.WorkUnit.GetByProjectID(h.repo.DB(), input.ProjectID, input.Query.Status)
	if err != nil {
		l.Error(err, "failed to get work units")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input.ProjectID, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToWorkUnitList(workUnits, input.ProjectID, project.Code), nil, nil, nil, ""))
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
	input := request.CreateWorkUnitInput{
		ProjectID: c.Param("id"),
	}
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
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

	project, err := h.store.Project.One(h.repo.DB(), input.ProjectID, false)
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
		_, err = h.store.ProjectMember.One(tx.DB(), input.ProjectID, employee.ID.String(), false)
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
			JoinedDate: time.Now(),
		}
		if err := h.store.WorkUnitMember.Create(tx.DB(), &wuMember); err != nil {
			l.Error(err, "failed to create new work unit member")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
			return
		}

		wuMember.Employee = *employee
		workUnit.WorkUnitMembers = append(workUnit.WorkUnitMembers, &wuMember)
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToWorkUnit(workUnit, project.Code), nil, done(nil), nil, ""))
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
	input := request.UpdateWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input.Body, ""))
		return
	}

	// TODO: can we move this to middleware ?
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
	stacks, err := h.store.Stack.All(h.repo.DB())
	if err != nil {
		return http.StatusInternalServerError, errs.ErrFailToCheckInputExistence
	}

	stackMap := model.ToStackMap(stacks)
	for _, sID := range input.Body.Stacks {
		_, ok := stackMap[sID]
		if !ok {
			return http.StatusNotFound, errs.ErrPositionNotFoundWithID(sID.String())
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
			Status:   model.WorkUnitMemberStatusInactive.String(),
			LeftDate: &now,
		}

		if _, err = h.store.WorkUnitMember.UpdateSelectedFieldsByID(db, workUnitMember.ID.String(), *deleteMember, "status", "left_date"); err != nil {
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
		_, err := h.store.ProjectMember.One(db, projectID, createMemberID.String(), false)
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
			JoinedDate: now,
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
	input := request.ArchiveWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	// TODO: can we move this to middleware ?
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

	exists, err := h.store.Project.IsExist(h.repo.DB(), input.ProjectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
		return
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

	// update work unit member: left_date = now() and status = 'inactive'
	timeNow := time.Now()
	for _, member := range wuMembers {
		member.LeftDate = &timeNow
		member.Status = model.ProjectMemberStatusInactive.String()

		_, err := h.store.WorkUnitMember.UpdateSelectedFieldsByID(tx.DB(), member.ID.String(), *member, "left_date", "status")
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
	input := request.ArchiveWorkUnitInput{
		ProjectID:  c.Param("id"),
		WorkUnitID: c.Param("workUnitID"),
	}

	// TODO: can we move this to middleware ?
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

	exists, err := h.store.Project.IsExist(h.repo.DB(), input.ProjectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if !exists {
		l.Error(errs.ErrProjectNotFound, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrProjectNotFound, nil, ""))
		return
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
		// _, err := h.store.ProjectMember.One(tx.DB(), input.ProjectID, member.EmployeeID.String(), false)

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

		member.LeftDate = nil
		member.Status = model.ProjectMemberStatusActive.String()

		_, err = h.store.WorkUnitMember.UpdateSelectedFieldsByID(tx.DB(), member.ID.String(), *member, "left_date", "status")
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

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateSendingSurveyState",
		"query":   query,
	})

	project, err := h.store.Project.One(h.repo.DB(), projectID, false)
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

	project.AllowsSendingSurvey = query.AllowsSendingSurvey
	_, err = h.store.Project.UpdateSelectedFieldsByID(h.repo.DB(), projectID, *project, "allows_sending_survey")
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
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	if !existed {
		l.Info("project not existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrProjectNotExisted, nil, ""))
		done(err)
		return
	}

	// 2.3 upload to GCS
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

	// 3. update avatar field
	_, err = h.store.Project.UpdateSelectedFieldsByID(tx.DB(), params.ID, model.Project{
		Avatar: filePath,
	}, "avatar")
	if err != nil {
		l.Error(err, "error in update avatar")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	done(nil)

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToProjectContentData(filePath), nil, nil, nil, ""))
}
