package auth

import "github.com/gin-gonic/gin"

type IHandler interface {
	Auth(c *gin.Context)
	Me(c *gin.Context)
	CreateAPIkey(c *gin.Context)
}
