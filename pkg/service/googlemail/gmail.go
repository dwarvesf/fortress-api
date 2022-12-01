package googlemail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"golang.org/x/oauth2"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

const (
	gmailURL = `https://www.googleapis.com/gmail/v1/users/`
)

type googleMailThread struct {
	ID       string              `json:"id"`
	Messages []googleMailMessage `json:"messages"`
}

// GoogleMailMessage --
type googleMailMessage struct {
	ID       string   `json:"id"`
	ThreadID string   `json:"threadId"`
	Payload  *payload `json:"payload"`
}

type payload struct {
	Headers []header `json:"headers"`
}

type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type googleService struct {
	apiKey               string
	config               *oauth2.Config
	token                *oauth2.Token
	StartHistoryId       uint64
	templatePath         string
	Env                  string
	teamEmailToken       string
	teamEmailID          string
	accountingEmailToken string
	accountingEmailID    string
}

// NewGoogleService function return Google service
func New(env, apiKey string, config *oauth2.Config, templatePath,
	teamEmailToken, teamEmailID,
	accountingEmailToken, accountingEmailID string) GoogleMailService {
	if env != "prod" {
		templatePath = filepath.Join(os.Getenv("GOPATH"), templatePath)
	}
	return &googleService{
		apiKey:               apiKey,
		config:               config,
		Env:                  env,
		templatePath:         templatePath,
		teamEmailToken:       teamEmailToken,
		teamEmailID:          teamEmailID,
		accountingEmailToken: accountingEmailToken,
		accountingEmailID:    accountingEmailID,
	}
}

func (g *googleService) SendPayrollPaidMail(p *model.Payroll, tax float64) (err error) {
	g.token = &oauth2.Token{
		RefreshToken: g.teamEmailToken,
	}
	if err := g.ensureToken(); err != nil {
		return err
	}

	if !email(p.Employee.TeamEmail) {
		return errors.New("email invalid")
	}

	funcMap := p.GetPaidSuccessfulEmailFuncMap(tax, g.Env)

	templateMail := "paidPayroll.tpl"

	mail, err := composeMailContent(
		&mailParseInfo{
			teamEmail,
			templateMail,
			p,
			funcMap,
		}, g.templatePath)
	if err != nil {
		return err
	}

	id := g.teamEmailID

	_, err = g.sendEmail(mail, id)
	return err
}

func (g *googleService) ensureToken() error {
	if !g.token.Valid() {
		tks := g.config.TokenSource(context.Background(), g.token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}
	return nil
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
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return msgID, errors.New(string(resBody))
	}

	payload := &struct {
		ThreadID string `json:"threadId"`
	}{}

	if err := json.NewDecoder(res.Body).Decode(payload); err != nil {
		return msgID, err
	}

	return payload.ThreadID, nil
}

// SendInvoiceMail
func (g *googleService) SendInvoiceMail(invoice *model.Invoice) (msgID string, err error) {
	g.token = &oauth2.Token{
		RefreshToken: g.accountingEmailToken,
	}
	if err = g.ensureToken(); err != nil {
		return
	}

	if !email(invoice.Email) {
		return "", errors.New("email invalid")
	}

	lastDayOfMonth := utils.LastDayOfMonth(invoice.Year, time.Month(invoice.Month))
	addresses, err := gatherAddresses(invoice.CCs)
	if err != nil {
		return "", err
	}
	funcMap := template.FuncMap{
		"formatDate": func(t *time.Time) string {
			return utils.FormatDatetime(*t)
		},
		"lastDayOfMonth": func() string {
			return utils.FormatDatetime(lastDayOfMonth)
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
		"toString": func(month time.Month) string {
			return month.String()
		},
	}

	encodedEmail, err := composeMailContent(
		&mailParseInfo{
			accountingUser,
			"invoice.tpl",
			&invoice,
			funcMap,
		}, g.templatePath)
	if err != nil {
		return
	}
	id := g.accountingEmailID

	threadID, err := g.sendEmail(encodedEmail, id)
	if err != nil {
		return
	}

	return threadID, nil
}

func (g *googleService) SendInvoiceThankyouMail(invoice *model.Invoice) (err error) {

	g.token = &oauth2.Token{
		RefreshToken: g.accountingEmailToken,
	}
	if err := g.ensureToken(); err != nil {
		return err
	}
	if invoice.ThreadID == "" {
		return errors.New("missing thread_id")
	}

	id := g.accountingEmailID
	thread, err := g.getEmailThread(invoice.ThreadID, id)
	if err != nil {
		return err
	}

	invoice.MessageID, invoice.References, err = getMessageIDFromThread(thread)
	if err != nil {
		return err
	}

	if !email(invoice.Email) {
		return errors.New("email invalid")
	}

	addresses, err := gatherAddresses(invoice.CCs)
	if err != nil {
		return err
	}
	funcMap := template.FuncMap{
		"gatherAddress": func() string {
			return addresses
		},
		"toString": func(month time.Month) string {
			return month.String()
		},
	}

	encodedEmail, err := composeMailContent(
		&mailParseInfo{
			accountingUser,
			"invoiceThankyou.tpl",
			&invoice,
			funcMap,
		}, g.templatePath)
	if err != nil {
		return err
	}

	_, err = g.sendEmail(encodedEmail, id)
	return err
}
