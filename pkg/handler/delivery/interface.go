package delivery

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetWeeklyReport(c *gin.Context)
}
