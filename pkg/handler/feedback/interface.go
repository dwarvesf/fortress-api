package feedback

import "github.com/gin-gonic/gin"

type IHandler interface {
	List(c *gin.Context)
	Detail(c *gin.Context)
	Submit(c *gin.Context)
	CountUnreadFeedback(c *gin.Context)
}
