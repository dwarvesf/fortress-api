package googlemail

import (
	"errors"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	ErrMissingThreadID = errors.New("missing thread_id")
	ErrInvalidEmail    = errors.New("invalid email")
)

// IService interface contain related google calendar method
type IService interface {
	SendInvoiceMail(invoice *model.Invoice) (msgID string, err error)
	SendInvoiceThankYouMail(invoice *model.Invoice) (err error)
	SendInvoiceOverdueMail(invoice *model.Invoice) (err error)
}
