package googlemail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/mailutils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

type googleService struct {
	config         *oauth2.Config
	token          *oauth2.Token
	StartHistoryId uint64
	service        *gmail.Service
	appConfig      *config.Config
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config) IService {
	return &googleService{
		config:    config,
		appConfig: appConfig,
	}
}

func (g *googleService) prepareService() error {
	client := g.config.Client(context.Background(), g.token)
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("Get Gmail Service Failed " + err.Error())
	}

	g.service = service

	return nil
}

// SendInvoiceMail function to send invoice email
func (g *googleService) SendInvoiceMail(invoice *model.Invoice) (msgID string, err error) {
	err = g.filterReceiver(invoice)
	if err != nil {
		return "", err
	}

	if err = g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return "", err
	}

	if err := g.prepareService(); err != nil {
		return "", err
	}

	if !mailutils.Email(invoice.Email) {
		return "", errors.New("email invalid")
	}

	lastDayOfMonth := timeutil.LastDayOfMonth(invoice.Month, invoice.Year)
	addresses, err := model.GatherAddresses(invoice.CC)
	if err != nil {
		return "", err
	}
	funcMap := template.FuncMap{
		"formatDate": func(t *time.Time) string {
			return timeutil.FormatDatetime(*t)
		},
		"lastDayOfMonth": func() string {
			return timeutil.FormatDatetime(lastDayOfMonth)
		},
		"encodedPDF": func() string {
			return base64.StdEncoding.EncodeToString(invoice.InvoiceFileContent)
		},
		"description": func() string {
			if invoice.Description == "" {
				return ""
			}
			return "Description: " + invoice.Description + "."
		},
		"gatherAddress": func() string {
			return addresses
		},
		"toString": func(month int) string {
			return time.Month(month).String()
		},
	}

	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			accountingUser,
			"invoice.tpl",
			&invoice,
			funcMap,
		})
	if err != nil {
		return
	}
	id := g.appConfig.Google.AccountingEmailID

	threadID, err := g.sendEmail(encodedEmail, id)
	if err != nil {
		return
	}

	return threadID, nil
}

// SendInvoiceThankYouMail function send thank you email
func (g *googleService) SendInvoiceThankYouMail(invoice *model.Invoice) (err error) {
	err = g.filterReceiver(invoice)
	if err != nil {
		return err
	}

	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	if invoice.ThreadID == "" {
		return ErrMissingThreadID
	}

	id := g.appConfig.Google.AccountingEmailID
	thread, err := g.service.Users.Threads.Get(id, invoice.ThreadID).Do()
	if err != nil {
		return err
	}

	invoice.MessageID, invoice.References, err = getMessageIDFromThread(thread)
	if err != nil {
		return err
	}

	if !mailutils.Email(invoice.Email) {
		return ErrInvalidEmail
	}

	addresses, err := model.GatherAddresses(invoice.CC)
	if err != nil {
		return err
	}
	funcMap := template.FuncMap{
		"gatherAddress": func() string {
			return addresses
		},
		"toString": func(month int) string {
			return time.Month(month).String()
		},
	}

	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			accountingUser,
			"invoiceThankyou.tpl",
			&invoice,
			funcMap,
		})
	if err != nil {
		return err
	}

	_, err = g.sendEmail(encodedEmail, id)
	return err
}

// SendInvoiceOverdueMail function send overdue email
func (g *googleService) SendInvoiceOverdueMail(invoice *model.Invoice) error {
	if err := g.filterReceiver(invoice); err != nil {
		return err
	}

	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if invoice.ThreadID == "" {
		return ErrMissingThreadID
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	id := g.appConfig.Google.AccountingEmailID
	thread, err := g.getEmailThread(invoice.ThreadID, id)
	if err != nil {
		return err
	}

	invoice.MessageID, invoice.References, err = getMessageIDFromThread(thread)
	if err != nil {
		return err
	}

	if !mailutils.Email(invoice.Email) {
		return ErrInvalidEmail
	}

	addresses, err := model.GatherAddresses(invoice.CC)
	if err != nil {
		return err
	}
	funcMap := template.FuncMap{
		"gatherAddress": func() string {
			return addresses
		},
		"toString": func(month int) string {
			return time.Month(month).String()
		},
	}

	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			accountingUser,
			"invoiceOverdue.tpl",
			&invoice,
			funcMap,
		})
	if err != nil {
		return err
	}

	_, err = g.sendEmail(encodedEmail, id)
	return err
}

func (g *googleService) ensureToken(refreshToken string) error {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	if !g.token.Valid() {
		tks := g.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}

		g.token = tok
	}

	return nil
}

func (g *googleService) sendEmail(encodedEmail, id string) (msgID string, err error) {
	rs, err := g.service.Users.Messages.Send(id, &gmail.Message{
		Raw: encodedEmail,
	}).Do()
	if err != nil {
		return
	}

	return rs.ThreadId, nil
}

func (g *googleService) getEmailThread(threadID, id string) (*gmail.Thread, error) {
	thread, err := g.service.Users.Threads.Get(id, threadID).Do()
	if err != nil {
		return nil, err
	}

	return thread, nil
}

func getMessageIDFromThread(thread *gmail.Thread) (msgID, references string, err error) {
	if len(thread.Messages) == 0 {
		return "", "", ErrEmptyMessageThread
	}

	for _, v := range thread.Messages[len(thread.Messages)-1].Payload.Headers {
		if strings.ToLower(v.Name) == "message-id" {
			msgID = v.Value
		}
		if strings.ToLower(v.Name) == "references" {
			references = v.Value
		}
	}

	if msgID == "" {
		return "", "", ErrCannotFindMessageID
	}

	return msgID, fmt.Sprintf(`%s %s`, references, msgID), nil
}

func (g *googleService) filterReceiver(i *model.Invoice) error {
	if g.appConfig.Env == "prod" {
		return nil
	}

	if !mailutils.IsDwarvesMail(i.Email) {
		i.Email = g.appConfig.Invoice.TestEmail
	}

	var ccList []string
	if err := json.Unmarshal(i.CC, &ccList); err != nil {
		return err
	}

	for idx := range ccList {
		if ccList[idx] == "" {
			continue
		}
		if !mailutils.IsDwarvesMail(ccList[idx]) {
			ccList[idx] = g.appConfig.Invoice.TestEmail
		}
	}

	b, err := json.Marshal(ccList)
	if err != nil {
		return err
	}

	var js model.JSON
	if err := js.Scan(b); err != nil {
		return err
	}
	i.CC = js

	return nil
}
