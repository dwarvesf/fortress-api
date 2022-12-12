package feedback

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	Detail(c *gin.Context)
	ListSurvey(c *gin.Context)
	GetSurveyDetail(c *gin.Context)
	GetSurveyReviewDetail(c *gin.Context)
	SendPerformanceReview(c *gin.Context)
	Submit(c *gin.Context)
	CreateSurvey(c *gin.Context)
	DeleteSurvey(c *gin.Context)
	DeleteSurveyTopic(c *gin.Context)
}
