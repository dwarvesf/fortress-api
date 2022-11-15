package project

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, query))
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	projects, total, err := h.store.Project.All(h.repo.DB(), project.GetListProjectInput{
		Status: query.Status,
		Name:   query.Name,
		Type:   query.Type,
	}, query.Pagination)
	if err != nil {
		l.Error(err, "error query project from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToProjectData(projects),
		&view.PaginationResponse{Pagination: query.Pagination, Total: total}, nil, nil))
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
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body))
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, ErrInvalidProjectStatus, body))
		return
	}

	// 2. get update status for project
	rs, err := h.store.Project.UpdateStatus(h.repo.DB(), projectID, body.ProjectStatus)
	if err != nil {
		l.Error(err, "error query update status for project to db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	// 3. return project data
	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProjectStatusResponse(rs), nil, nil, nil))
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
	if err := c.ShouldBind(&body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body))
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	country, err := h.store.Country.One(h.repo.DB(), body.CountryID)
	if err != nil {
		l.Error(err, "failed to get country")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, body))
		return
	}

	p := &model.Project{
		Name:      body.Name,
		Country:   country.Name,
		Type:      model.ProjectType(body.Type),
		Status:    model.ProjectStatus(body.Status),
		StartDate: body.GetStartDate(),
	}

	tx, done := h.repo.NewTransaction()

	if err := h.store.Project.Create(tx.DB(), p); err != nil {
		l.Error(err, "failed to create project")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil))
		return
	}

	// create project account manager
	accountManager := &model.ProjectHead{
		ProjectID:      p.ID,
		EmployeeID:     body.AccountManagerID,
		JoinedDate:     time.Now(),
		CommissionRate: decimal.Zero,
		Position:       model.HeadPositionAccountManager,
	}

	if err := h.store.ProjectHead.Create(tx.DB(), accountManager); err != nil {
		l.Error(err, "failed to create account manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body))
		return
	}

	// create project delivery manager
	deliveryManager := &model.ProjectHead{
		ProjectID:      p.ID,
		EmployeeID:     body.DeliveryManagerID,
		JoinedDate:     time.Now(),
		CommissionRate: decimal.Zero,
		Position:       model.HeadPositionDeliveryManager,
	}

	if err := h.store.ProjectHead.Create(tx.DB(), deliveryManager); err != nil {
		l.Error(err, "failed to create delivery manager")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), body))
		return
	}

	p.Heads = append(p.Heads, *accountManager, *deliveryManager)

	c.JSON(http.StatusOK, view.CreateResponse(view.ToCreateProjectDataResponse(p), nil, done(nil), nil))
}
