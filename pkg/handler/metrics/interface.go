package metrics

import "github.com/gin-gonic/gin"

// IMetricsHandler defines the interface for metrics endpoints
type IMetricsHandler interface {
	Metrics(c *gin.Context)
}