package client

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/client/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/client/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
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

// Create godoc
// @Summary Create new client
// @Description Create new client
// @Tags Client
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreateClientInput true "Body"
// @Success 200 {object} view.CreateClientResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /clients [post]
func (h *handler) Create(c *gin.Context) {
	input := request.CreateClientInput{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "client",
		"method":  "Create",
		"request": input,
	})

	client, err := h.controller.Client.Create(c, input)
	if err != nil {
		l.Error(err, "failed to create client")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](client, nil, nil, nil, ""))
}

// List godoc
// @Summary Get all clients
// @Description Get all clients
// @Tags Client
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.CreateClientInput true "Body"
// @Success 200 {object} view.GetListClientResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /clients [get]
func (h *handler) List(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "client",
		"method":  "List",
	})

	clients, err := h.controller.Client.List(c)
	if err != nil {
		l.Error(err, "failed to get client list")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](clients, nil, nil, nil, ""))
}

// Detail godoc
// @Summary Get client detail by id
// @Description Get client detail by id
// @Tags Client
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.GetDetailClientResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /clients/{id} [get]
func (h *handler) Detail(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" || !model.IsUUIDFromString(clientID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidClientID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "client",
		"method":  "Detail",
	})

	client, err := h.controller.Client.Detail(c, clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrClientNotFound, nil, ""))
			return
		}

		l.Error(err, "failed to get client detail")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](client, nil, nil, nil, ""))
}

// Update godoc
// @Summary Update client by id
// @Description Update client by id
// @Tags Client
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.UpdateClientInput true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /clients/{id} [put]
func (h *handler) Update(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" || !model.IsUUIDFromString(clientID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidClientID, nil, ""))
		return
	}

	input := request.UpdateClientInput{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "client",
		"method":  "Update",
		"request": input,
	})

	errCode, err := h.controller.Client.Update(c, clientID, input)
	if err != nil {
		l.Error(err, "failed to update client")
		c.JSON(errCode, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

// Delete godoc
// @Summary Delete client by id
// @Description Delete client by id
// @Tags Client
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /clients/{id} [delete]
func (h *handler) Delete(c *gin.Context) {
	clientID := c.Param("id")
	if clientID == "" || !model.IsUUIDFromString(clientID) {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidClientID, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler":  "client",
		"method":   "UpdateInfo",
		"clientID": clientID,
	})

	errCode, err := h.controller.Client.Delete(c, clientID)
	if err != nil {
		l.Error(err, "failed to delete client")
		c.JSON(errCode, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
