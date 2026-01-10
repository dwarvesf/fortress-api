package invoiceemail

import "github.com/gin-gonic/gin"

// IHandler defines the interface for invoice email handler
type IHandler interface {
	// ProcessInvoiceEmails processes incoming invoice emails (cron endpoint)
	ProcessInvoiceEmails(c *gin.Context)
}
