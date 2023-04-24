package invoice

import "errors"

var (
	ErrBankAccountNotFound             = errors.New("bank account not found")
	ErrCouldNotGetTheLatestInvoice     = errors.New("could not get the latest invoice")
	ErrCouldNotGetTheNextInvoiceNumber = errors.New("could not get the next invoice number")
	ErrInvoiceNotFound                 = errors.New("invoice not found")
	ErrInvoiceStatusAlready            = errors.New("invoice status already")
	ErrProjectNotFound                 = errors.New("project not found")
	ErrSenderNotFound                  = errors.New("sender not found")
)
