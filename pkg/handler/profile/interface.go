package profile

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetProfile(c *gin.Context)
	UpdateInfo(c *gin.Context)
}
