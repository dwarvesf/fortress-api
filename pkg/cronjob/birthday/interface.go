package birthday

import "github.com/gin-gonic/gin"

type ICronjob interface {
	BirthdayDailyMessage(c *gin.Context)
}
