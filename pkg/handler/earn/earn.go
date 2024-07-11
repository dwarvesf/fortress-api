package earn

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
	controller *controller.Controller
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
		controller: controller,
	}
}

// ListEarn godoc
// @Summary List of earns from memo
// @Description List of earns from memo
// @Tags Earn
// @Accept  json
// @Produce  json
// @Success 200 {object} ListEarnResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /earn [get]
func (h *handler) ListEarn(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "vault",
		"method":  "ListEarn",
	})

	earns, err := h.controller.Earn.ListEarn(c.Request.Context())
	if err != nil {
		l.Error(err, "get list earn failed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	}

	c.JSON(http.StatusOK, view.CreateResponse(view.ToEarns(earns), nil, nil, nil, ""))
}
