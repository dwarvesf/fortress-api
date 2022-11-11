package project

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
}

func New(store *store.Store, service *service.Service, logger logger.Logger) IHandler {
	return &handler{
		store:   store,
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

	projects, total, err := h.store.Project.All(project.GetListProjectInput{
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
