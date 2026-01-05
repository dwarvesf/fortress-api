package googlemail

import (
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"google.golang.org/api/gmail/v1"
)

var (
	ErrMissingThreadID     = errors.New("missing thread_id")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrEmptyMessageThread  = errors.New("empty message thread")
	ErrCannotFindMessageID = errors.New("cannot find message id")
	ErrAliasNotVerified    = errors.New("sendas alias not verified")
)

// IService interface contain related google calendar method
type IService interface {
	SendInvitationMail(invitation *model.InvitationEmail) (err error)
	SendInvoiceMail(invoice *model.Invoice) (msgID string, err error)
	SendInvoiceOverdueMail(invoice *model.Invoice) (err error)
	SendInvoiceThankYouMail(invoice *model.Invoice) (err error)
	SendPayrollPaidMail(p *model.Payroll) (err error)
	SendOffboardingMail(offboarding *model.OffboardingEmail) (err error)
	SendTaskOrderConfirmationMail(data *model.TaskOrderConfirmationEmail) error
	SendTaskOrderRawContentMail(data *model.TaskOrderRawEmail) error

	// SendAs management methods
	ListSendAsAliases(userId string) ([]*gmail.SendAs, error)
	GetSendAsAlias(userId, email string) (*gmail.SendAs, error)
	CreateSendAsAlias(userId, email, displayName string) (*gmail.SendAs, error)
	VerifySendAsAlias(userId, email string) error
	IsAliasVerified(userId, email string) (bool, error)
}
