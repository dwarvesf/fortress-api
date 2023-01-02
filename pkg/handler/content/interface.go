package content

import "github.com/gin-gonic/gin"

type IHandler interface {
	UploadContent(c *gin.Context)
}
