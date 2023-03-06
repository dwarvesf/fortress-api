package auth

import "github.com/gin-gonic/gin"

type IHandler interface {
	Auth(c *gin.Context)
	GetLoginURL(c *gin.Context)
	Me(c *gin.Context)
	CreateAPIKey(c *gin.Context)
}
