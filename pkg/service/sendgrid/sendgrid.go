package sendgrid

import (
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Service interface {
	SendEmail(*model.Email) error
}

type sendgridClient struct {
	client *sendgrid.Client
	cfg    *config.Config
	l      logger.Logger
}

func New(key string, cfg *config.Config, l logger.Logger) Service {
	client := sendgrid.NewSendClient(key)
	return &sendgridClient{
		client: client,
		cfg:    cfg,
		l:      l,
	}
}

func (s *sendgridClient) SendEmail(email *model.Email) error {
	// boundary check to make sure we don't mess up
	if s.cfg.Env != "prod" {
		m := mail.NewEmail("Minh Luu", "leo@dwarvesv.com")
		email.To = []*mail.Email{m}
		email.Bcc = []*mail.Email{}
	}
	m := mail.NewV3Mail()
	m.SetFrom(email.From)
	m.AddContent(mail.NewContent("text/html", email.HTMLContent))
	m.AddCategories(email.Categories...)

	personalization := mail.NewPersonalization()
	personalization.Subject = email.Subject
	personalization.AddTos(email.To...)
	personalization.AddBCCs(email.Bcc...)

	m.AddPersonalizations(personalization)

	s.l.Infof("Sending email", m.Personalizations)
	response, err := s.client.Send(m)
	if err != nil {
		s.l.Error(err, "SendEmail() failed with ")
		return err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted {
		s.l.Error(err, "Email not sent")
		return err
	}

	return nil
}
