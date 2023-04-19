package model

import (
	"time"

	"gorm.io/datatypes"
)

// Expense for slack command
type Expense struct {
	BaseModel

	EmployeeID              UUID           `json:"employee_id"`
	CurrencyID              UUID           `json:"currency_id"`
	Amount                  int            `json:"amount"`
	IssuedDate              time.Time      `sql:"default: now()" json:"issued_date"`
	Reason                  string         `json:"reason"`
	InvoiceImageURL         string         `json:"invoice_image_url"`
	Metadata                datatypes.JSON `json:"metadata"`
	BasecampID              int            `json:"basecamp_id"`
	AccountingTransactionID *UUID          `json:"accounting_transaction_id"`
}
