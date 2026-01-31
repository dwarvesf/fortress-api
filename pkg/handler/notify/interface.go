package notify

import "github.com/gin-gonic/gin"

// IHandler defines the interface for notification handlers
type IHandler interface {
	// PreviewExtraPaymentNotification previews contractors to be notified for extra payments
	PreviewExtraPaymentNotification(c *gin.Context)
	// SendExtraPaymentNotification sends extra payment notification emails
	SendExtraPaymentNotification(c *gin.Context)
	// SendOneExtraPaymentNotification sends extra payment notification to a single contractor by page ID
	SendOneExtraPaymentNotification(c *gin.Context)
}
