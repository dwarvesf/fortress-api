package model

import "github.com/sendgrid/sendgrid-go/helpers/mail"

type Email struct {
	HTMLContent string
	Subject     string
	From        *mail.Email
	To          []*mail.Email
	Bcc         []*mail.Email
	Categories  []string
}

// TaskOrderConfirmationEmail represents email data for task order confirmation
type TaskOrderConfirmationEmail struct {
	ContractorName string
	TeamEmail      string
	Month          string   // YYYY-MM format
	Clients        []TaskOrderClient
	InvoiceDueDay  string   // Invoice due date (e.g., "10th", "25th")
	Milestones     []string // Client milestones for the month
}

// TaskOrderClient represents a client in the task order
type TaskOrderClient struct {
	Name    string
	Country string
}

// TaskOrderRawEmail represents email data for task order confirmation with raw content
type TaskOrderRawEmail struct {
	ContractorName string
	TeamEmail      string
	Month          string // YYYY-MM format
	RawContent     string // Plain text content from Order page body
}
