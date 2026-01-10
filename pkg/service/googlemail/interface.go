package googlemail

import (
	"context"
	"errors"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"google.golang.org/api/gmail/v1"
)

var (
	ErrMissingThreadID     = errors.New("missing thread_id")
	ErrInvalidEmail        = errors.New("invalid email")
	ErrEmptyMessageThread  = errors.New("empty message thread")
	ErrCannotFindMessageID = errors.New("cannot find message id")
	ErrAliasNotVerified    = errors.New("sendas alias not verified")
	ErrMessageNotFound     = errors.New("message not found")
	ErrAttachmentNotFound  = errors.New("attachment not found")
	ErrLabelNotFound       = errors.New("label not found")
)

// InboxMessage represents a simplified email message from the inbox
type InboxMessage struct {
	ID        string
	ThreadID  string
	Subject   string
	From      string
	To        string
	Date      time.Time
	Snippet   string
	LabelIDs  []string
	HasPDF    bool
	PDFPartID string // Part ID of the first PDF attachment
}

// MessageAttachment represents an email attachment
type MessageAttachment struct {
	PartID   string
	Filename string
	MimeType string
	Size     int64
	Data     []byte // Populated only when fetched
}

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

	// Inbox reading methods for invoice email listener
	ListInboxMessages(ctx context.Context, query string, maxResults int64) ([]InboxMessage, error)
	GetMessage(ctx context.Context, messageID string) (*InboxMessage, error)
	GetAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error)
	AddLabel(ctx context.Context, messageID, labelID string) error
	GetOrCreateLabel(ctx context.Context, labelName string) (string, error)
}
