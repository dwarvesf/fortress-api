package dashboard

import (
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	ProjectSizes(c *gin.Context)
	WorkSurveys(c *gin.Context)
	GetActionItemReports(c *gin.Context)
	EngineeringHealth(c *gin.Context)
	Audits(c *gin.Context)
}
