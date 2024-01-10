package withdrawal

import "github.com/gin-gonic/gin"

type IHandler interface {
	CheckWithdrawCondition(c *gin.Context)
}
