package metadata

import (
	"errors"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/model"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
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
// @Router /metadata/countries/:country_id/cities [get]
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
