package feedback

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	Detail(c *gin.Context)
	ListSurvey(c *gin.Context)
	GetSurveyDetail(c *gin.Context)
}
