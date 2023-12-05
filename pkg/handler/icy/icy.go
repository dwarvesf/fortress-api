package icy

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	controller *controller.Controller
	logger     logger.Logger
}

func New(
	controller *controller.Controller,
	logger logger.Logger,
) IHandler {
	return &handler{
		controller: controller,
		logger:     logger,
	}
}

func (h *handler) Accounting(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "icy",
		"method":  "Accounting",
	})

	accounting, err := h.controller.Icy.Accounting()
	if err != nil {
		l.Error(err, "failed to get icy accounting")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, "failed to get icy accounting"))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](accounting, nil, nil, nil, ""))
}
