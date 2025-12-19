package timesheet

import "github.com/gin-gonic/gin"

type IHandler interface {
	// LogHours creates a new timesheet entry
	LogHours(c *gin.Context)

	// GetEntries retrieves timesheet entries for a user
	GetEntries(c *gin.Context)

	// GetWeeklySummary retrieves weekly summary for a user
	GetWeeklySummary(c *gin.Context)
}
