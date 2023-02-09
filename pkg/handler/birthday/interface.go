package birthday

import "github.com/gin-gonic/gin"

type IHandler interface {
	BirthdayDailyMessage(c *gin.Context)
}
