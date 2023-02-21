package errs

import "errors"

var (
	ErrInvalidDueAt         = errors.New("invalid due at")
	ErrInvalidPaidAt        = errors.New("invalid paid at")
	ErrInvalidScheduledDate = errors.New("invalid scheduled date")
	ErrInvalidInvoiceStatus = errors.New("invalid invoice status")
	ErrSenderNotFound       = errors.New("sender not found")
	ErrBankAccountNotFound  = errors.New("bank account not found")
	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidInvoiceID     = errors.New("invalid invoice id")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvalidEmailDomain   = errors.New("invalid email domain")
	ErrInvalidProjectID     = errors.New("invalid project id")
	ErrInvoiceStatusAlready = errors.New("invoice status already")
)
