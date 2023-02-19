package googlemail

import (
	"errors"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	MissingThreadIDErr = errors.New("missing thread_id")
)

// Service interface contain related google calendar method
type Service interface {
	SendInvoiceMail(invoice *model.Invoice) (msgID string, err error)
	SendInvoiceThankYouMail(invoice *model.Invoice) (err error)
	SendInvoiceOverdueMail(invoice *model.Invoice) (err error)
}
