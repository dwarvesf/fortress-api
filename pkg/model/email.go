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
	Month          string // YYYY-MM format
	Clients        []TaskOrderClient
}

// TaskOrderClient represents a client in the task order
type TaskOrderClient struct {
	Name    string
	Country string
}
