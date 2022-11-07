package metadata

import (
	"errors"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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

// WorkingStatus godoc
// @Summary Get list values for working status
// @Description Get list values for working status
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} []view.WorkingStatusData
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/working-status [get]
func (h *handler) WorkingStatus(c *gin.Context) {
	// return list values for working status
	// hardcode for now since we dont need db storage for this
	res := []view.WorkingStatusData{
		{
			ID:   "left",
			Name: "Left",
		},
		{
			ID:   "probation",
			Name: "Probation",
		},
		{
			ID:   "full-time",
			Name: "Full-time",
		},
		{
			ID:   "contractor",
			Name: "Contractor",
		},
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](res, nil, nil, nil))
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
		"handler": "Seniorities",
		"method":  "All",
	})

	// 2 query seniorities from db
	seniorities, err := h.store.Seniority.All()
	if err != nil {
		l.Error(err, "error query seniorities from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// 3 return array of seniorities
	c.JSON(http.StatusOK, view.CreateResponse[any](seniorities, nil, nil, nil))
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
		"handler": "Chapters",
		"method":  "All",
	})

	// 2 query chapters from db
	chapters, err := h.store.Chapter.All()
	if err != nil {
		l.Error(err, "error query chapters from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// 3 return array of chapters
	c.JSON(http.StatusOK, view.CreateResponse[any](chapters, nil, nil, nil))
}

// Roles godoc
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
		"handler": "AccountRoles",
		"method":  "All",
	})

	// 2 query roles from db
	roles, err := h.store.Role.All()
	if err != nil {
		l.Error(err, "error query roles from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// 3 return array of roles
	c.JSON(http.StatusOK, view.CreateResponse[any](roles, nil, nil, nil))
}

// AccountStatuses godoc
// @Summary Get list values for account statuses
// @Description Get list values for account statuses
// @Tags Metadata
// @Accept  json
// @Produce  json
// @Success 200 {object} view.AccountStatusResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /metadata/account-statuses [get]
func (h *handler) AccountStatuses(c *gin.Context) {
	// 1 prepare the logger
	// TODO: can we move this to middleware ?
	l := h.logger.Fields(logger.Fields{
		"handler": "AccountStatuses",
		"method":  "All",
	})

	// 2 query accountStatuses from db
	accountStatuses, err := h.store.AccountStatus.All()
	if err != nil {
		l.Error(err, "error query accountStatuses from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// 3 return array of account statuses
	c.JSON(http.StatusOK, view.CreateResponse[any](accountStatuses, nil, nil, nil))
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
		"handler": "Positions",
		"method":  "All",
	})

	// 2 query positions from db
	positions, err := h.store.Position.All()
	if err != nil {
		l.Error(err, "error query positions from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	// 3 return array of positions
	c.JSON(http.StatusOK, view.CreateResponse[any](positions, nil, nil, nil))
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
	countries, err := h.store.Country.All()
	if err != nil {
		h.logger.Fields(logger.Fields{
			"handler": "metadata",
			"method":  "GetCountries",
		}).Error(err, "failed to get all countries")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(countries, nil, nil, nil))
}

// GetCountries godoc
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
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errors.New("country_id is empty"), nil))
		return
	}

	country, err := h.store.Country.One(countryID)
	if err != nil {
		l.AddField("countryID", countryID).Error(err, "failed to get cities")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse(country.Cities, nil, nil, nil))
}
