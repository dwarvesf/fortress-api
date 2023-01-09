package dashboard

import (
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	ProjectSizes(c *gin.Context)
	WorkSurveys(c *gin.Context)
}
