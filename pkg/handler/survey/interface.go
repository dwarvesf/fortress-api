package survey

import "github.com/gin-gonic/gin"

type IHandler interface {
	ListSurvey(c *gin.Context)
	GetSurveyDetail(c *gin.Context)
	GetSurveyReviewDetail(c *gin.Context)
	SendSurvey(c *gin.Context)
	CreateSurvey(c *gin.Context)
	DeleteSurvey(c *gin.Context)
	DeleteSurveyTopic(c *gin.Context)
	GetSurveyTopicDetail(c *gin.Context)
	UpdateTopicReviewers(c *gin.Context)
	MarkDone(c *gin.Context)
	DeleteTopicReviewers(c *gin.Context)
}
