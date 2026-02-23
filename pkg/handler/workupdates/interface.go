package workupdates

import "github.com/gin-gonic/gin"

// IHandler defines the interface for work updates handlers
type IHandler interface {
	// GetWorkUpdates returns timesheet completion status for a given month
	GetWorkUpdates(c *gin.Context)
}
