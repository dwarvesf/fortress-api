package brainerylogs

import "github.com/gin-gonic/gin"

type IHandler interface {
	Create(c *gin.Context)
	GetMetrics(c *gin.Context)
}
