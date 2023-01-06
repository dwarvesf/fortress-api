package dashboard

import (
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	ProjectSizes(c *gin.Context)
}
