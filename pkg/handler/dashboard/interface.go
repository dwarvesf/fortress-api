package dashboard

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetProjectSizes(c *gin.Context)
	GetWorkSurveys(c *gin.Context)
	GetActionItemReports(c *gin.Context)
	GetEngineeringHealth(c *gin.Context)
	GetAudits(c *gin.Context)
	GetActionItemSquashReports(c *gin.Context)
	GetSummary(c *gin.Context)
	GetResourcesAvailability(c *gin.Context)
	GetEngagementInfo(c *gin.Context)
	GetEngagementInfoDetail(c *gin.Context)
	GetResourceUtilization(c *gin.Context)
}
