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
