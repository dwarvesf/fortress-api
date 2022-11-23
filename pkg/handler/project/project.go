package project

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

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
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
	}
}

// List godoc
// @Summary Get list of project
// @Description Get list of project
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param status query  string false  "Project status"
// @Param name   query  string false  "Project name"
// @Param type   query  string false  "Project type"
// @Param page   query  string false  "Page"
// @Param size   query  string false  "Size"
// @Success 200 {object} view.ProjectListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects [get]
func (h *handler) List(c *gin.Context) {
	query := GetListProjectInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	query.Standardize()

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
		Status: query.Status,
		Name:   query.Name,
		Type:   query.Type,
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
	// 1. parse id from uri, validate id
	projectID := c.Param("id")

	// 1.1 get body request
	var body updateAccountStatusBody
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	// 1.2 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "UpdateProjectStatus",
		"body":    body,
	})

	if !body.ProjectStatus.IsValid() {
		l.Error(ErrInvalidProjectStatus, "invalid value for ProjectStatus")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectStatus, body, ""))
		return
	}

	// 2. get update status for project
	rs, err := h.store.Project.UpdateStatus(h.repo.DB(), projectID, body.ProjectStatus)
	if err != nil {
		l.Error(err, "error query update status for project to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	// 3. return project data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectStatusResponse(rs), nil, nil, nil, ""))
}

// Create godoc
// @Summary	Create new project
// @Description	Create new project
// @Tags Project
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param Body body CreateProjectInput true "body"
// @Success 200 {object} view.CreateProjectData
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects [post]
func (h *handler) Create(c *gin.Context) {
	body := CreateProjectInput{}
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

	p := &model.Project{
		Name:         body.Name,
		CountryID:    body.CountryID,
		Type:         model.ProjectType(body.Type),
		Status:       model.ProjectStatus(body.Status),
		StartDate:    body.GetStartDate(),
		ProjectEmail: body.ProjectEmail,
		ClientEmail:  body.ClientEmail,
		Country:      country,
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

	p.Heads = append(p.Heads, accountManager)

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

		p.Heads = append(p.Heads, deliveryManager)
	}

	// assign members to project
	for _, member := range body.Members {
		slot, code, err := h.assignMemberToProject(tx.DB(), p.ID.String(), member)
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
// @Param page query string false "Page"
// @Param size query string false "Size"
// @Param sort query string false "Sort"
// @Success 200 {object} view.ProjectMemberListResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/:id/members [get]
func (h *handler) GetMembers(c *gin.Context) {
	query := GetListStaffInput{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}
	query.Standardize()

	projectID := c.Param("id")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("invalid project_id"), nil, ""))
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

	exists, err := h.store.Project.Exists(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	if !exists {
		l.Error(ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, nil, ""))
		return
	}

	members, total, err := h.store.ProjectSlot.All(h.repo.DB(), projectslot.GetListProjectSlotInput{
		ProjectID: projectID,
		Status:    query.Status,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "error query project members from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, query, ""))
		return
	}

	heads, err := h.store.ProjectHead.GetByProjectID(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "error query project heads from db")
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
// @Router /project/:id/members/:memberID [delete]
func (h *handler) DeleteMember(c *gin.Context) {
	input := DeleteMemberInput{
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
	projectMember, err := h.store.ProjectMember.One(h.repo.DB(), input.ProjectID, input.MemberID, model.ProjectMemberStatusActive.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "project member not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectMemberNotFound, input.MemberID, ""))
			return
		}
		l.Error(err, "error query member from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input.MemberID, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	slotID := projectMember.ProjectSlotID.String()

	err = h.store.ProjectMemberPosition.HardDeleteByProjectMemberID(tx.DB(), projectMember.ID.String())
	if err != nil {
		l.Error(err, "error delete project member position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectMember.HardDelete(tx.DB(), projectMember.ID.String())
	if err != nil {
		l.Error(err, "error delete project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectSlotPosition.HardDeleteByProjectSlotID(tx.DB(), slotID)
	if err != nil {
		l.Error(err, "error delete project slot position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectSlot.HardDelete(tx.DB(), slotID)
	if err != nil {
		l.Error(err, "error delete project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
		return
	}

	err = h.store.ProjectHead.HardDeleteByPosition(tx.DB(), projectMember.ProjectID.String(), projectMember.EmployeeID.String(), model.HeadPositionTechnicalLead.String())
	if err != nil {
		l.Error(err, "error delete project head")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input.MemberID, ""))
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
// @Param Body body UpdateMemberInput true "Body"
// @Success 200 {object} view.CreateMemberDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [put]
func (h *handler) UpdateMember(c *gin.Context) {
	var body UpdateMemberInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectID, nil, ""))
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

	// check project exists
	exists, err := h.store.Project.Exists(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, body, ""))
		return
	}

	// check seniority exists
	exists, err = h.store.Seniority.Exists(h.repo.DB(), body.SeniorityID.String())
	if err != nil {
		l.Error(err, "failed to check seniority existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(ErrSeniorityNotFound, "cannot find seniority by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrSeniorityNotFound, body, ""))
		return
	}

	// check position existence
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error when finding position")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	positionMap := model.ToPositionMap(positions)
	for _, pID := range body.Positions {
		if _, ok := positionMap[pID]; !ok {
			l.Error(errPositionNotFound(pID.String()), "error position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errPositionNotFound(pID.String()), body, ""))
			return
		}
	}

	// check project slot status
	slot, err := h.store.ProjectSlot.One(h.repo.DB(), body.ProjectSlotID.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(ErrProjectSlotNotFound, "cannot find project slot by id")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectSlotNotFound, body, ""))
			return
		}
		l.Error(err, "failed to get project slot by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if slot.Status == model.ProjectMemberStatusInactive {
		l.Info("slot is inactive")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrSlotIsInactive, body, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if !body.EmployeeID.IsZero() {
		// check project member status
		member, err := h.store.ProjectMember.One(tx.DB(), projectID, body.EmployeeID.String(), model.ProjectMemberStatusActive.String())
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "failed to get project member")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}

		if !member.ID.IsZero() {
			if member.EmployeeID != body.EmployeeID {
				l.Info("employeeID cannot be changed")
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(ErrEmployeeIDCannotBeChanged), body, ""))
				return
			}

			if member.Status == model.ProjectMemberStatusInactive {
				l.Info("member is inactive")
				c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrMemberIsInactive, body, ""))
				return
			}
		}

		// update project member
		member = &model.ProjectMember{
			ProjectID:      model.MustGetUUIDFromString(projectID),
			EmployeeID:     body.EmployeeID,
			SeniorityID:    body.SeniorityID,
			ProjectSlotID:  slot.ID,
			DeploymentType: model.DeploymentType(body.DeploymentType),
			Status:         model.ProjectMemberStatus(body.Status),
			JoinedDate:     body.GetJoinedDate(),
			LeftDate:       body.GetLeftDate(),
			Rate:           body.Rate,
			Discount:       body.Discount,
		}

		if err = h.store.ProjectMember.Upsert(tx.DB(), member); err != nil {
			l.Error(err, "failed to upsert project member")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), member, ""))
			return
		}

		slot.ProjectMember = *member

		// create project member positions
		if err := h.store.ProjectMemberPosition.HardDeleteByProjectMemberID(tx.DB(), member.ID.String()); err != nil {
			l.Error(err, "failed to delete project member positions")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}

		memberPos := make([]model.ProjectMemberPosition, 0, len(body.Positions))

		for _, v := range body.Positions {
			pos := model.ProjectMemberPosition{
				ProjectMemberID: member.ID,
				PositionID:      v,
			}

			if err := h.store.ProjectMemberPosition.Create(tx.DB(), &pos); err != nil {
				l.Error(err, "failed to create project member positions")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
				return
			}

			memberPos = append(memberPos, pos)
		}
		slot.ProjectMember.ProjectMemberPositions = memberPos

		// create project head
		slot.IsLead = body.IsLead
		if body.IsLead {
			if err := h.updateProjectHead(tx, projectID, body.EmployeeID, model.HeadPositionTechnicalLead); err != nil {
				l.Error(err, "failed to update technicalLeads")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
				return
			}
		} else {
			if err := h.store.ProjectHead.DeleteByProjectIDAndPosition(tx.DB(), projectID, model.HeadPositionTechnicalLead.String()); err != nil {
				l.Error(err, "failed to upsert project head")
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
				return
			}
		}
	}

	// update project slot
	slot.ID = body.ProjectSlotID
	slot.SeniorityID = body.SeniorityID
	slot.ProjectID = model.MustGetUUIDFromString(projectID)
	slot.DeploymentType = model.DeploymentType(body.DeploymentType)
	slot.Status = model.ProjectMemberStatus(body.Status)
	slot.Rate = body.Rate
	slot.Discount = body.Discount

	err = h.store.ProjectSlot.Update(tx.DB(), body.ProjectSlotID.String(), slot)
	if err != nil {
		l.Error(err, "failed to create project slot")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// update project slot positions
	if err := h.store.ProjectSlotPosition.HardDeleteByProjectSlotID(tx.DB(), slot.ID.String()); err != nil {
		l.Error(err, "failed to delete project member positions")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	slotPos := make([]model.ProjectSlotPosition, 0, len(body.Positions))

	for _, v := range body.Positions {
		pos := model.ProjectSlotPosition{
			ProjectSlotID: slot.ID,
			PositionID:    v,
		}

		if err := h.store.ProjectSlotPosition.Create(tx.DB(), &pos); err != nil {
			l.Error(err, "failed to create project member positions")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}

		slotPos = append(slotPos, pos)
	}
	slot.ProjectSlotPositions = slotPos

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(slot), nil, done(nil), nil, ""))
}

// AssignMember godoc
// @Summary Assign member into an existing project
// @Description Assign member in an existing project
// @Tags Project
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body AssignMemberInput true "Body"
// @Success 200 {object} view.CreateMemberDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/members [post]
func (h *handler) AssignMember(c *gin.Context) {
	var body AssignMemberInput
	if err := c.ShouldBindJSON(&body); err != nil {
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body, ""))
			return
		}
	}

	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectID, nil, ""))
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

	// get active project member info
	_, err := h.store.ProjectMember.One(h.repo.DB(), projectID, body.EmployeeID.String(), model.ProjectMemberStatusActive.String())
	if err != gorm.ErrRecordNotFound {
		if err == nil {
			l.Error(err, "project member exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrProjectMemberExists, projectID, ""))
			return
		}
		l.Error(err, "failed to query project member")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, projectID, ""))
		return
	}

	// check project exists
	exists, err := h.store.Project.Exists(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "failed to check project existence")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(ErrProjectNotFound, "cannot find project by id")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, body, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	slot, code, err := h.assignMemberToProject(tx.DB(), projectID, body)
	if err != nil {
		l.Error(err, "failed to assign member to project")
		c.JSON(code, view.CreateResponse[any](nil, nil, done(err), body, ""))
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateMemberData(slot), nil, done(nil), nil, ""))
}

func (h *handler) assignMemberToProject(db *gorm.DB, projectID string, req AssignMemberInput) (*model.ProjectSlot, int, error) {
	l := h.logger

	// check seniority existence
	exists, err := h.store.Seniority.Exists(db, req.SeniorityID.String())
	if err != nil {
		l.Error(err, "failed to check seniority existence")
		return nil, http.StatusInternalServerError, err
	}

	if !exists {
		l.Error(ErrSeniorityNotFound, "cannot find seniority by id")
		return nil, http.StatusNotFound, ErrSeniorityNotFound
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
			l.Error(errPositionNotFound(pID.String()), "error position not found")
			return nil, http.StatusNotFound, errPositionNotFound(pID.String())
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

	// create project slot position
	slotPos := make([]model.ProjectSlotPosition, 0, len(req.Positions))

	for _, v := range req.Positions {
		pos := model.ProjectSlotPosition{
			ProjectSlotID: slot.ID,
			PositionID:    v,
		}

		if err := h.store.ProjectSlotPosition.Create(db, &pos); err != nil {
			l.Error(err, "failed to create project member positions")
			return nil, http.StatusInternalServerError, err
		}

		slotPos = append(slotPos, pos)
	}
	slot.ProjectSlotPositions = slotPos

	if !req.EmployeeID.IsZero() {
		// check employee existence
		exists, err = h.store.Employee.Exists(db, req.EmployeeID.String())
		if err != nil {
			l.Error(err, "failed to check employee existence")
			return nil, http.StatusInternalServerError, err
		}

		if !exists {
			l.Error(ErrEmployeeNotFound, "cannot find employee by id")
			return nil, http.StatusNotFound, ErrEmployeeNotFound
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
			if err := h.store.ProjectMemberPosition.Create(db, &model.ProjectMemberPosition{
				ProjectMemberID: member.ID,
				PositionID:      v,
			}); err != nil {
				l.Error(err, "failed to create project member positions")
				return nil, http.StatusInternalServerError, err
			}
		}

		// create project head
		slot.IsLead = req.IsLead
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
// @Success 200 {object} view.ProjectListDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id} [get]
func (h *handler) Details(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectID, nil, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "project",
		"method":  "Details",
		"id":      projectID,
	})

	project, err := h.store.Project.One(h.repo.DB(), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("project not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, nil, ""))
			return
		}
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectData(project), nil, nil, nil, ""))
}

// UpdateGeneralInfo godoc
// @Summary Update general info of the project by id
// @Description Update general info of the project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body UpdateGeneralInfoInput true "Body"
// @Success 200 {object} view.UpdateProjectGeneralInfoResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/general-info [put]
func (h *handler) UpdateGeneralInfo(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectID, nil, ""))
		return
	}

	var body UpdateGeneralInfoInput
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

	// Check project exists
	exists, err := h.store.Project.Exists(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "error check existence of project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(err, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, body, ""))
		return
	}

	// Check country exists
	exists, err = h.store.Country.Exists(h.repo.DB(), body.CountryID.String())
	if err != nil {
		l.Error(err, "error check existence of country")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(err, "country not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrCountryNotFound, body, ""))
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
			l.Error(errStackNotFound(sID.String()), "stack not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errStackNotFound(sID.String()), body, ""))
			return
		}
	}

	_, err = time.Parse("2006-01-02", body.StartDate)
	if body.StartDate != "" && err != nil {
		l.Error(ErrInvalidStartDate, "invalid start date")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidStartDate, body, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Delete all exist employee stack
	if err := h.store.ProjectStack.HardDelete(tx.DB(), projectID); err != nil {
		l.Error(err, "failed to delete project stacks in database")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Create new employee stack
	for _, stackID := range body.Stacks {
		projectStack := &model.ProjectStack{
			ProjectID: model.MustGetUUIDFromString(projectID),
			StackID:   stackID,
		}

		projectStack, err := h.store.ProjectStack.Create(tx.DB(), projectStack)
		if err != nil {
			l.Error(err, "failed to create project stack")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
			return
		}
	}

	rs, err := h.store.Project.UpdateGeneralInfo(tx.DB(), project.UpdateGeneralInfoInput{
		Name:      body.Name,
		StartDate: body.GetStartDate(),
		CountryID: body.CountryID,
	}, projectID)

	if err != nil {
		l.Error(err, "failed to update project information to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectGeneralInfo(rs), nil, done(nil), nil, ""))
}

// UpdateContactInfo godoc
// @Summary Update contact info of the project by id
// @Description Update contact info of the project by id
// @Tags Project
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Project ID"
// @Param Body body UpdateContactInfoInput true "Body"
// @Success 200 {object} view.UpdateProjectContactInfoResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /projects/{id}/contact-info [put]
func (h *handler) UpdateContactInfo(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" || !model.IsUUIDFromString(projectID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectID, nil, ""))
		return
	}

	var body UpdateContactInfoInput
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

	// Check project exists
	exists, err := h.store.Project.Exists(h.repo.DB(), projectID)
	if err != nil {
		l.Error(err, "error when check existence of project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(err, "project not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrProjectNotFound, body, ""))
		return
	}

	// Check account manager exists
	exists, err = h.store.Employee.Exists(h.repo.DB(), body.AccountManagerID.String())

	if err != nil {
		l.Error(err, "error when finding account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(ErrAccountManagerNotFound, "account manager not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrAccountManagerNotFound, body, ""))
		return
	}

	// Check delivery manager exists
	exists, err = h.store.Employee.Exists(h.repo.DB(), body.DeliveryManagerID.String())

	if err != nil {
		l.Error(err, "error when finding delivery manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body, ""))
		return
	}

	if !exists {
		l.Error(ErrDeliveryManagerNotFound, "delivery manager not found")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, ErrDeliveryManagerNotFound, body, ""))
		return
	}

	// Begin transaction
	tx, done := h.repo.NewTransaction()

	// Update Account Manager
	if err := h.updateProjectHead(tx, projectID, body.AccountManagerID, model.HeadPositionAccountManager); err != nil {
		l.Error(err, "failed to update account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Update Delivery Manager
	if err := h.updateProjectHead(tx, projectID, body.DeliveryManagerID, model.HeadPositionDeliveryManager); err != nil {
		l.Error(err, "failed to update delivery manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	// Update email info
	rs, err := h.store.Project.UpdateContactInfo(tx.DB(), project.UpdateContactInfoInput{
		ClientEmail:  body.ClientEmail,
		ProjectEmail: body.ProjectEmail,
	}, projectID)

	if err != nil {
		l.Error(err, "failed to update project information to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectContactInfo(rs), nil, done(nil), nil, ""))
}

func (h *handler) updateProjectHead(tx store.DBRepo, projectID string, memberID model.UUID, position model.HeadPosition) error {
	timeNow := time.Now()
	needCreate := false
	newProjectHead := model.ProjectHead{
		ProjectID:  model.MustGetUUIDFromString(projectID),
		EmployeeID: memberID,
		JoinedDate: timeNow,
		Position:   position,
	}

	oldHead, err := h.store.ProjectHead.One(tx.DB(), projectID, position)

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFailToFindProjectHead
		}

		needCreate = true
	} else if oldHead.EmployeeID != memberID {
		if err := h.store.ProjectHead.UpdateLeftDate(tx.DB(), projectID, oldHead.EmployeeID.String(), position.String(), timeNow); err != nil {
			return ErrFailToUpdateLeftDate
		}

		needCreate = true
	}

	if needCreate {
		err = h.store.ProjectHead.Create(tx.DB(), &newProjectHead)
		if err != nil {
			return ErrFailToCreateProjectHead
		}
	}

	return nil
}
