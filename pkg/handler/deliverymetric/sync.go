package deliverymetric

import (
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *handler) Sync(c *gin.Context) {
	l := h.logger.Fields(
		logger.Fields{
			"handler": "deliverymetric",
			"method":  "Sync",
		},
	)

	err := h.controller.DeliveryMetric.Sync()
	if err != nil {
		l.Errorf(err, "failed to create delivery metric")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}
