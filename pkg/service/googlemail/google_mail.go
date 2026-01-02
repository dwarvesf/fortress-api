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
	"github.com/dwarvesf/fortress-api/pkg/utils"
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

	// Verify accounting@d.foundation alias before sending
	id := g.appConfig.Google.AccountingEmailID
	verified, err := g.IsAliasVerified(id, "accounting@d.foundation")
	if err != nil || !verified {
		return "", ErrAliasNotVerified
	}

	// Support comma-separated emails (e.g., "a@b.com, c@d.com")
	// First email goes to "To", rest go to "CC"
	var validEmails []string
	for _, email := range strings.Split(invoice.Email, ",") {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		if !mailutils.Email(email) {
			return "", fmt.Errorf("email invalid: %s", email)
		}
		validEmails = append(validEmails, email)
	}
	if len(validEmails) == 0 {
		return "", errors.New("no valid email provided")
	}
	// First email as primary recipient (To)
	invoice.Email = validEmails[0]
	// Additional emails from comma-separated list
	additionalEmails := validEmails[1:]

	lastDayOfMonth := timeutil.LastDayOfMonth(invoice.Month, invoice.Year)

	// Build CC list: additional emails + existing CC + accounting@d.foundation
	var ccList []string
	ccList = append(ccList, additionalEmails...)
	if len(invoice.CC) > 0 && string(invoice.CC) != "\u0000" && !strings.EqualFold(string(invoice.CC), "null") {
		var existingCC []string
		if err := json.Unmarshal(invoice.CC, &existingCC); err != nil {
			return "", err
		}
		ccList = append(ccList, existingCC...)
	}
	// Add accounting@d.foundation if not already present
	hasAccounting := false
	for _, cc := range ccList {
		if cc == "accounting@d.foundation" {
			hasAccounting = true
			break
		}
	}
	if !hasAccounting {
		ccList = append(ccList, "accounting@d.foundation")
	}

	addresses := strings.Join(ccList, ", ")

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
			return "Description: " + invoice.Description
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

	// Verify accounting@d.foundation alias before sending
	id := g.appConfig.Google.AccountingEmailID
	verified, err := g.IsAliasVerified(id, "accounting@d.foundation")
	if err != nil || !verified {
		return ErrAliasNotVerified
	}

	// Try to get thread if ThreadID exists, but don't fail if it doesn't exist
	// This allows sending thank you emails even when switching Gmail accounts
	if invoice.ThreadID != "" {
		thread, err := g.service.Users.Threads.Get(id, invoice.ThreadID).Do()
		if err == nil {
			// Successfully got thread, extract message ID for reply threading
			invoice.MessageID, invoice.References, err = getMessageIDFromThread(thread)
			if err != nil {
				// Failed to extract message ID, will send as new email instead
				invoice.MessageID = ""
				invoice.References = ""
			}
		}
		// If thread fetch fails (404), continue to send as new email
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

	// Verify accounting@d.foundation alias before sending
	id := g.appConfig.Google.AccountingEmailID
	verified, err := g.IsAliasVerified(id, "accounting@d.foundation")
	if err != nil || !verified {
		return ErrAliasNotVerified
	}
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
	if len(i.CC) == 0 || string(i.CC) == "\u0000" || strings.EqualFold(string(i.CC), "null") {
		ccList = []string{}
	} else {
		if err := json.Unmarshal(i.CC, &ccList); err != nil {
			return err
		}
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

// SendPayrollPaidMail after paid a payroll for a user to notify that
// we have been paid their payroll
func (g *googleService) SendPayrollPaidMail(p *model.Payroll) (err error) {
	if g.appConfig.Env == "local" {
		p.Employee.TeamEmail = "quang@d.foundation"
	}

	if err := g.ensureToken(g.appConfig.Google.TeamGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	if !mailutils.Email(p.Employee.TeamEmail) {
		return errors.New("email invalid")
	}

	funcMap := g.getPaidSuccessfulEmailFuncMap(p)
	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			teamEmail,
			"paidPayroll.tpl",
			p,
			funcMap,
		})
	if err != nil {
		return err
	}

	id := g.appConfig.Google.TeamEmailID
	_, err = g.sendEmail(encodedEmail, id)
	return err
}

// SendTaskOrderConfirmationMail sends a monthly task order confirmation email
func (g *googleService) SendTaskOrderConfirmationMail(data *model.TaskOrderConfirmationEmail) error {
	// Use accounting refresh token
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}
	if err := g.prepareService(); err != nil {
		return err
	}

	// Verify accounting alias
	id := g.appConfig.Google.AccountingEmailID
	verified, err := g.IsAliasVerified(id, "accounting@d.foundation")
	if err != nil || !verified {
		return fmt.Errorf("accounting@d.foundation alias not verified for user %s", id)
	}

	// Parse template
	content, err := composeTaskOrderConfirmationContent(g.appConfig, data)
	if err != nil {
		return fmt.Errorf("failed to compose email content: %w", err)
	}

	// Send email
	_, err = g.service.Users.Messages.Send(id, &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(content)),
	}).Do()
	if err != nil {
		return fmt.Errorf("failed to send task order confirmation email: %w", err)
	}

	return nil
}

// ToPaidSuccessfulEmailContent to parse the payroll object
// into template when sending email after payroll is paid
func (g *googleService) getPaidSuccessfulEmailFuncMap(p *model.Payroll) map[string]interface{} {
	// the salary will be the contract(companyAccountAmount in DB)
	// plus the base salary(personalAccountAmount in DB)

	var addresses string = "quang@d.foundation"
	if g.appConfig.Env == "prod" {
		addresses = "quang@d.foundation, hr@d.foundation"
	}

	return template.FuncMap{
		"ccList": func() string {
			return addresses
		},
		"userFirstName": func() string {
			return p.Employee.GetFirstNameFromFullName()
		},
		"currency": func() string {
			return p.Employee.BaseSalary.Currency.Symbol
		},
		"currencyName": func() string {
			return p.Employee.BaseSalary.Currency.Name
		},
		"formattedCurrentMonth": func() string {
			fm := time.Month(int(p.Month))
			return fm.String()
		},
		"formattedBaseSalaryAmount": func() string {
			return utils.FormatNumber(p.BaseSalaryAmount)
		},
		"formattedTotalAllowance": func() string {
			return utils.FormatNumber(int64(p.TotalAllowance))
		},
		"formattedSalaryAdvance": func() string {
			return utils.FormatNumber(int64(p.SalaryAdvanceAmount))
		},
		"haveBonusOrCommission": func() bool {
			return len(p.CommissionExplains) > 0 || len(p.ProjectBonusExplains) > 0
		},
		"haveCommission": func() bool {
			return len(p.CommissionExplains) > 0
		},
		"haveBonus": func() bool {
			return len(p.ProjectBonusExplains) > 0
		},
		"commissionExplain": func() []model.CommissionExplain {
			return p.CommissionExplains
		},
		"projectBonusExplains": func() []model.ProjectBonusExplain {
			return p.ProjectBonusExplains
		},
	}
}

// SendInvitationMail ...
func (g *googleService) SendInvitationMail(invitation *model.InvitationEmail) (err error) {
	if err := g.ensureToken(g.appConfig.Google.TeamGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	// Verify team@d.foundation alias before sending
	id := g.appConfig.Google.TeamEmailID
	verified, err := g.IsAliasVerified(id, "hr@d.foundation")
	if err != nil || !verified {
		return ErrAliasNotVerified
	}

	invitation.Link = strings.Replace(invitation.Link, "=", "=3D", -1)

	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			teamEmail,
			"invitation.tpl",
			&invitation,
			map[string]interface{}{},
		})
	if err != nil {
		return err
	}

	_, err = g.sendEmail(encodedEmail, id)
	return err
}

// SendOffboardingMail ...
func (g *googleService) SendOffboardingMail(offboarding *model.OffboardingEmail) (err error) {
	if err := g.ensureToken(g.appConfig.Google.TeamGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	// Verify team@d.foundation alias before sending
	id := g.appConfig.Google.TeamEmailID
	verified, err := g.IsAliasVerified(id, "hr@d.foundation")
	if err != nil || !verified {
		return ErrAliasNotVerified
	}

	encodedEmail, err := composeMailContent(g.appConfig,
		&MailParseInfo{
			teamEmail,
			"offboarding_keep_fwd_email.tpl",
			&offboarding,
			map[string]interface{}{},
		})
	if err != nil {
		return err
	}

	_, err = g.sendEmail(encodedEmail, id)
	return err
}

// ListSendAsAliases lists all SendAs aliases for the given user
func (g *googleService) ListSendAsAliases(userId string) ([]*gmail.SendAs, error) {
	if g.service == nil {
		return nil, errors.New("gmail service not initialized")
	}

	result, err := g.service.Users.Settings.SendAs.List(userId).Do()
	if err != nil {
		return nil, err
	}

	return result.SendAs, nil
}

// GetSendAsAlias gets a specific SendAs alias for the given user
func (g *googleService) GetSendAsAlias(userId, email string) (*gmail.SendAs, error) {
	if g.service == nil {
		return nil, errors.New("gmail service not initialized")
	}

	sendAs, err := g.service.Users.Settings.SendAs.Get(userId, email).Do()
	if err != nil {
		return nil, err
	}

	return sendAs, nil
}

// CreateSendAsAlias creates a new SendAs alias for the given user
func (g *googleService) CreateSendAsAlias(userId, email, displayName string) (*gmail.SendAs, error) {
	if g.service == nil {
		return nil, errors.New("gmail service not initialized")
	}

	sendAs := &gmail.SendAs{
		SendAsEmail:    email,
		DisplayName:    displayName,
		ReplyToAddress: email,
		TreatAsAlias:   true,
	}

	created, err := g.service.Users.Settings.SendAs.Create(userId, sendAs).Do()
	if err != nil {
		return nil, err
	}

	return created, nil
}

// VerifySendAsAlias resends the verification email for a SendAs alias
func (g *googleService) VerifySendAsAlias(userId, email string) error {
	if g.service == nil {
		return errors.New("gmail service not initialized")
	}

	err := g.service.Users.Settings.SendAs.Verify(userId, email).Do()
	return err
}

// IsAliasVerified checks if a SendAs alias is verified
func (g *googleService) IsAliasVerified(userId, email string) (bool, error) {
	if g.service == nil {
		return false, errors.New("gmail service not initialized")
	}

	sendAs, err := g.service.Users.Settings.SendAs.Get(userId, email).Do()
	if err != nil {
		return false, err
	}

	return sendAs.VerificationStatus == "accepted", nil
}
