package googlemail

import (
	"errors"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	MissingThreadIDErr = errors.New("missing thread_id")
)

// IService interface contain related google calendar method
type IService interface {
	SendInvoiceMail(invoice *model.Invoice) (msgID string, err error)
	SendInvoiceThankYouMail(invoice *model.Invoice) (err error)
	SendInvoiceOverdueMail(invoice *model.Invoice) (err error)
}
