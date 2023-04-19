package invoice

import "errors"

var (
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceStatusAlready = errors.New("invoice status already")
)
