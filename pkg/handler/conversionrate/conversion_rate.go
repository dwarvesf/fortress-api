package conversionrate

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

func (h *handler) List(c *gin.Context) {
	//l := h.logger.Fields(logger.Fields{
	//	"handler": "conversionRate",
	//	"method":  "List",
	//})

	//clients, err := h.controller.Client.List(c)
	//if err != nil {
	//	l.Error(err, "failed to get client list")
	//	c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
	//	return
	//}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "conversionRate",
		"method":  "Sync",
	})

	err := h.controller.ConversionRate.Sync(c)
	if err != nil {
		l.Error(err, "failed to sync")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
