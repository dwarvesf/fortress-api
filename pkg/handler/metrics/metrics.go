package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler provides metrics endpoints
type Handler struct{}

// New creates a new metrics handler
func New() *Handler {
	return &Handler{}
}

// Metrics godoc
// @Summary Get Prometheus metrics
// @Description Endpoint for Prometheus to scrape metrics from the application
// @Tags metrics
// @Accept json
// @Produce text/plain
// @Success 200 {string} string "Metrics in Prometheus format"
// @Router /metrics [get]
func (h *Handler) Metrics(c *gin.Context) {
	handler := promhttp.Handler()
	handler.ServeHTTP(c.Writer, c.Request)
}