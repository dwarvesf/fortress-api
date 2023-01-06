package dashboard

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkUnitDistribution(c *gin.Context)
}
