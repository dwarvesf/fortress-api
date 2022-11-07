package metadata_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
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
