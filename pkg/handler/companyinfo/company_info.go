package companyinfo

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
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

// List godoc
// @Summary Get all company info
// @Description Get all company info
// @id get list of company info
// @Tags Client
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} view.GetListCompanyInfoResponse
// @Failure 500 {object} ErrorResponse
// @Router /company-infos [get]
func (h *handler) List(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "companyInfo",
		"method":  "List",
	})

	companyInfos, err := h.controller.CompanyInfo.List(c)
	if err != nil {
		l.Error(err, "failed to get company info list")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToCompanyInfos(companyInfos), nil, nil, nil, ""))
}
