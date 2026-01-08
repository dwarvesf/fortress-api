package contractorpayables

import "github.com/gin-gonic/gin"

// IHandler defines the interface for contractor payables handler
type IHandler interface {
	PreviewCommit(c *gin.Context)
	Commit(c *gin.Context)
}
