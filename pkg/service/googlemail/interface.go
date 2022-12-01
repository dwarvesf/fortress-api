package googlemail

import "github.com/dwarvesf/fortress-api/pkg/model"

type GoogleMailService interface {
	SendPayrollPaidMail(p *model.Payroll, tax float64) (err error)
	SendInvoiceMail(invoice *model.Invoice) (msgID string, err error)
	SendInvoiceThankyouMail(invoice *model.Invoice) (err error)
}
