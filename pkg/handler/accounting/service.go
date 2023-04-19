package accounting

import "github.com/gin-gonic/gin"

type IHandler interface {
	CreateAccountingTodo(c *gin.Context)
}
