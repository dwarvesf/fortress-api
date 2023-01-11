package dashboard

import "github.com/gin-gonic/gin"

type IHandler interface {
	WorkSurveys(c *gin.Context)
}
