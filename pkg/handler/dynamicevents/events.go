package dynamicevents

import (
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/dynamicevents/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
)

func (h *handler) Events(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "dynamic-events",
			"method":  "Events",
		},
	)

	var request request.DynamicEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		l.Error(err, "failed to bind input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, "invalid input"))
		return
	}

	data := model.DynamicEvent{
		Data:      request.Data,
		EventType: request.Type,
		Timestamp: time.Now(),
	}

	if err := h.controller.DynamicEvents.CreateEvents(c.Request.Context(), data); err != nil {
		l.Error(err, "failed to create events")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "successfully subscribed"))
}
