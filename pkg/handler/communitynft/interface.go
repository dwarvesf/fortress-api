package communitynft

import "github.com/gin-gonic/gin"

type IHandler interface {
	GetNftMetadata(c *gin.Context)
}
