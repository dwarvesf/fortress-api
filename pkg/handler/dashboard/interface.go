package dashboard

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetResourceUtilization(c *gin.Context)
	GetEngagementInfo(c *gin.Context)
	GetEngagementInfoDetail(c *gin.Context)
}
