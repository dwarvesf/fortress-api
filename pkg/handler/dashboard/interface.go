package dashboard

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetResourceUtilization(c *gin.Context)
}
