package googlemail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"golang.org/x/oauth2"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/mailutils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

const (
	gmailURL = `https://www.googleapis.com/gmail/v1/users/`
)

type googleService struct {
	apiKey         string
	config         *oauth2.Config
	token          *oauth2.Token
	StartHistoryId uint64
	appConfig      *config.Config
}

// New function return Google service
func New(apiKey string, config *oauth2.Config, appConfig *config.Config) Service {
	return &googleService{
		apiKey:    apiKey,
		config:    config,
		appConfig: appConfig,
	}
}

// SendInvoiceMail function to send invoice email
func (g *googleService) SendInvoiceMail(invoice *model.Invoice) (msgID string, err error) {
	err = g.filterReceiver(invoice)
	if err != nil {
		return "", err
	}

	g.token = &oauth2.Token{
		RefreshToken: g.appConfig.Google.AccountingGoogleRefreshToken,
	}
	if err = g.ensureToken(); err != nil {
		return
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

	g.token = &oauth2.Token{
		RefreshToken: g.appConfig.Google.AccountingGoogleRefreshToken,
	}
	if err := g.ensureToken(); err != nil {
		return err
	}
	if invoice.ThreadID == "" {
		return MissingThreadIDErr
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
		return errors.New("email invalid")
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

	g.token = &oauth2.Token{
		RefreshToken: g.appConfig.Google.AccountingGoogleRefreshToken,
	}
	if err := g.ensureToken(); err != nil {
		return err
	}
	if invoice.ThreadID == "" {
		return MissingThreadIDErr
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
		return errors.New("email invalid")
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

func (g *googleService) ensureToken() error {
	if !g.token.Valid() {
		tks := g.config.TokenSource(oauth2.NoContext, g.token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}
	return nil
}

func buildEmailRequest(encodedEmail, id string, g *googleService) (*http.Request, error) {
	url := getGMailURL(g.apiKey, id, "messages/send")

	jsonContent, err := json.Marshal(map[string]interface{}{
		"raw": encodedEmail,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonContent))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.token.AccessToken))

	return req, nil
}

func buildGetEmailRequest(mailID, id string, g *googleService) (*http.Request, error) {
	url := getGMailURL(g.apiKey, id, "messages/"+mailID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.token.AccessToken))

	return req, nil
}

func buildGetEmailThreadRequest(threadID, id string, g *googleService) (*http.Request, error) {
	url := getGMailURL(g.apiKey, id, "threads/"+threadID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.token.AccessToken))

	return req, nil
}

func getGMailURL(apiKey, id, action string) string {
	return fmt.Sprintf(
		"%v%v/%v?key=%v",
		gmailURL,
		id,
		action,
		apiKey,
	)
}

func (g *googleService) sendEmail(encodedEmail, id string) (msgID string, err error) {
	req, err := buildEmailRequest(encodedEmail, id, g)
	if err != nil {
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return msgID, errors.New(res.Status)
	}

	payload := &struct {
		ThreadID string `json:"threadId"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		return msgID, err
	}

	return payload.ThreadID, nil
}

func (g *googleService) getEmail(mailID, id string) (*model.GoogleMailMessage, error) {
	msg := &model.GoogleMailMessage{}
	req, err := buildGetEmailRequest(mailID, id, g)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(msg); err != nil {
		return nil, err
	}
	return msg, err
}

func (g *googleService) getEmailThread(threadID, id string) (*model.GoogleMailThread, error) {
	thread := &model.GoogleMailThread{}
	req, err := buildGetEmailThreadRequest(threadID, id, g)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(thread); err != nil {
		return nil, err
	}
	return thread, err
}

func getMessageIDFromThread(thread *model.GoogleMailThread) (msgID, references string, err error) {
	if len(thread.Messages) == 0 {
		return "", "", errors.New("empty message thread")
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
		return "", "", errors.New("can't find message_id")
	}

	return msgID, fmt.Sprintf(`%s %s`, references, msgID), nil
}

func (g *googleService) filterReceiver(i *model.Invoice) error {
	if g.appConfig.Env == "prod" {
		return nil
	}
	if !mailutils.IsDwarvesMail(i.Email) {
		i.Email = "huynh@dwarvesv.com"
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
			ccList[idx] = "huynh@dwarvesv.com"
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
