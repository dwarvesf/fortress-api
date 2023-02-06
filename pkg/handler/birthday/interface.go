package birthday

import "github.com/gin-gonic/gin"

type IBirthday interface {
	BirthdayDailyMessage(c *gin.Context)
}
