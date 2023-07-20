package deliverymetric

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetWeeklyReport(c *gin.Context)
	GetMonthlyReport(c *gin.Context)
	Sync(c *gin.Context)
}
