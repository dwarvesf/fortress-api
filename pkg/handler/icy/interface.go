package icy

import "github.com/gin-gonic/gin"

type IHandler interface {
	Accounting(c *gin.Context)
}
