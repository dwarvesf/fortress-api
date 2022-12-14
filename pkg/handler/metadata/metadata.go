package metadata

import (
	"errors"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/handler/metadata/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
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

// WorkingStatuses godoc
// @Summary Get list values for working status
// @Description Get list values for working status
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} []view.MetaData
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/working-status [get]
func (h *handler) WorkingStatuses(c *gin.Context) {
	// return list values for working status
	// hardcode for now since we dont need db storage for this
	res := []view.MetaData{
		{
			Code: model.WorkingStatusLeft.String(),
			Name: "Left",
		},
		{
			Code: model.WorkingStatusOnBoarding.String(),
			Name: "On Boarding",
		},
		{
			Code: model.WorkingStatusProbation.String(),
			Name: "Probation",
		},
		{
			Code: model.WorkingStatusFullTime.String(),
			Name: "Full-time",
		},
		{
			Code: model.WorkingStatusContractor.String(),
			Name: "Contractor",
		},
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
}

// Seniorities godoc
// @Summary Get list values for sentitorities
// @Description Get list values for sentitorities
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.SeniorityResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/seniorities [get]
func (h *handler) Seniorities(c *gin.Context) {
	// 1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "Seniorities",
	})

	// 2 query seniorities from db
	seniorities, err := h.store.Seniority.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error query seniorities from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3 return array of seniorities
	c.JSON(http.StatusOK, view.CreateResponse[any](seniorities, nil, nil, nil, ""))
}

// Chapters godoc
// @Summary Get list values for chapters
// @Description Get list values for chapters
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.ChapterResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/chapters [get]
func (h *handler) Chapters(c *gin.Context) {
	// 1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "Chapters",
	})

	// 2 query chapters from db
	chapters, err := h.store.Chapter.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error query chapters from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3 return array of chapters
	c.JSON(http.StatusOK, view.CreateResponse[any](chapters, nil, nil, nil, ""))
}

// AccountRoles godoc
// @Summary Get list values for account roles
// @Description Get list values for account roles
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.AccountRoleResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/account-roles [get]
func (h *handler) AccountRoles(c *gin.Context) {
	// 1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "AccountRoles",
	})

	// 2 query roles from db
	roles, err := h.store.Role.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error query roles from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3 return array of roles
	c.JSON(http.StatusOK, view.CreateResponse[any](roles, nil, nil, nil, ""))
}

// ProjectStatuses godoc
// @Summary Get list values for project statuses
// @Description Get list values for project statuses
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} []view.MetaData
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/project-statuses [get]
func (h *handler) ProjectStatuses(c *gin.Context) {
	// return list values for project statuses
	// hardcode for now since we don't need db storage for this
	res := []view.MetaData{
		{
			Code: "on-boarding",
			Name: "On Boarding",
		},
		{
			Code: "paused",
			Name: "Paused",
		},
		{
			Code: "active",
			Name: "Active",
		},
		{
			Code: "closed",
			Name: "Closed",
		},
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil, ""))
}

// Positions godoc
// @Summary Get list values for positions
// @Description Get list values for positions
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.PositionResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/positions [get]
func (h *handler) Positions(c *gin.Context) {
	// 1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "Positions",
	})

	// 2 query positions from db
	positions, err := h.store.Position.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error query positions from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 3 return array of positions
	c.JSON(http.StatusOK, view.CreateResponse[any](positions, nil, nil, nil, ""))
}

// GetCountries godoc
// @Summary Get all countries
// @Description Get all countries
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.CountriesResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/countries [get]
func (h *handler) GetCountries(c *gin.Context) {
	countries, err := h.store.Country.All(h.repo.DB())
	if err != nil {
		h.logger.Fields(logger.Fields{
			"handler": "metadata",
			"method":  "GetCountries",
		}).Error(err, "failed to get all countries")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(countries, nil, nil, nil, ""))
}

// GetCities godoc
// @Summary Get list cities by country
// @Description Get list cities by country
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.CitiesResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/countries/{country_id}/cities [get]
func (h *handler) GetCities(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "GetCities",
	})

	countryID := c.Param("country_id")
	if countryID == "" {
		l.Info("country_id is empty")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("country_id is empty"), nil, ""))
		return
	}

	country, err := h.store.Country.One(h.repo.DB(), countryID)
	if err != nil {
		l.AddField("countryID", countryID).Error(err, "failed to get cities")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(country.Cities, nil, nil, nil, ""))
}

// Stacks godoc
// @Summary Get list values for stacks
// @Description Get list values for stacks
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.StackResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/stacks [get]
func (h *handler) Stacks(c *gin.Context) {
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "Stacks",
	})

	// 1 query stacks from db
	stacks, err := h.store.Stack.All(h.repo.DB())
	if err != nil {
		l.Error(err, "error query Stacks from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 2 return array of account statuses
	c.JSON(http.StatusOK, view.CreateResponse[any](stacks, nil, nil, nil, ""))
}

// UpdateStack godoc
// @Summary Update stack information by ID
// @Description Update stack information by ID
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Stack ID"
// @Param Body body request.UpdateStackBody true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/stacks/{id} [put]
func (h *handler) UpdateStack(c *gin.Context) {
	var input request.UpdateStackInput

	input.ID = c.Param("id")
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "UpdateStack",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	stack, err := h.store.Stack.One(h.repo.DB(), input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "Stack not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrStackNotFound, input, ""))
			return
		}

		l.Error(err, "failed to get stacks")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	stack.Name = input.Body.Name
	stack.Avatar = input.Body.Avatar
	stack.Code = input.Body.Code

	stack, err = h.store.Stack.Update(h.repo.DB(), stack)
	if err != nil {
		l.Error(err, "error query Stacks from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 2 return array of account statuses
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// CreateStack godoc
// @Summary Create new stack
// @Description Create new stack
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreateStackInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/stacks [post]
func (h *handler) CreateStack(c *gin.Context) {
	var input request.CreateStackInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "CreateStack",
		"input":   input,
	})

	stack := &model.Stack{
		Name:   input.Name,
		Code:   input.Code,
		Avatar: input.Avatar,
	}

	stack, err := h.store.Stack.Create(h.repo.DB(), stack)
	if err != nil {
		l.Error(err, "error query Stacks from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 2 return array of account statuses
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// DeleteStack godoc
// @Summary Delete stack by ID
// @Description Delete stack by ID
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Stack ID"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/stacks/{id} [delete]
func (h *handler) DeleteStack(c *gin.Context) {
	stackID := c.Param("id")

	if stackID == "" || !model.IsUUIDFromString(stackID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidStackID, stackID, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "DeleteStack",
		"stackID": stackID,
	})

	if err := h.store.Stack.Delete(h.repo.DB(), stackID); err != nil {
		l.Error(err, "error query Stacks from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 2 return array of account statuses
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// UpdatePosition godoc
// @Summary Update position information by ID
// @Description Update position information by ID
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Position ID"
// @Param Body body request.UpdatePositionBody true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/positions/{id} [put]
func (h *handler) UpdatePosition(c *gin.Context) {
	var input request.UpdatePositionInput

	input.ID = c.Param("id")
	if err := c.ShouldBindJSON(&input.Body); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "UpdatePosition",
		"input":   input,
	})

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	position, err := h.store.Position.One(h.repo.DB(), model.MustGetUUIDFromString(input.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Error(err, "Position not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrPositionNotFound, input, ""))
			return
		}

		l.Error(err, "failed to get position")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	position.Name = input.Body.Name
	position.Code = input.Body.Code

	position, err = h.store.Position.Update(h.repo.DB(), position)
	if err != nil {
		l.Error(err, "error query Positions from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// CreatePosition godoc
// @Summary Create new position
// @Description Create new position
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreatePositionInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/positions [post]
func (h *handler) CreatePosition(c *gin.Context) {
	var input request.CreatePositionInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "CreatePosition",
		"input":   input,
	})

	position := &model.Position{
		Name: input.Name,
		Code: input.Code,
	}

	position, err := h.store.Position.Create(h.repo.DB(), position)
	if err != nil {
		l.Error(err, "error query Positions from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// DeletePosition godoc
// @Summary Delete position by ID
// @Description Delete position by ID
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Position ID"
// @Success 200 {object} view.MessageResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/positions/{id} [delete]
func (h *handler) DeletePosition(c *gin.Context) {
	positionID := c.Param("id")

	if positionID == "" || !model.IsUUIDFromString(positionID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidPositionID, positionID, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler":    "metadata",
		"method":     "DeletePosition",
		"positionID": positionID,
	})

	if err := h.store.Position.Delete(h.repo.DB(), positionID); err != nil {
		l.Error(err, "error query Position from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// GetQuestions godoc
// @Summary Get list question by category and subcategory
// @Description Get list question by category and subcategory
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Param category query model.EventType true "Category"
// @Param subcategory query model.EventSubtype true "Subcategory"
// @Success 200 {object} view.GetQuestionResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/questions [get]
func (h *handler) GetQuestions(c *gin.Context) {
	var input GetQuestionsInput

	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "metadata",
		"method":  "GetQuestions",
		"input":   input,
	})

	rs, err := h.store.Question.AllByCategory(h.repo.DB(), input.Category, input.Subcategory)
	if err != nil {
		l.Error(err, "failed to get question from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToListQuestion(rs), nil, nil, nil, ""))
}
