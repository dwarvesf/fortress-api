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
	TaskProvider            string         `json:"task_provider"`
	TaskRef                 string         `json:"task_ref"`
	TaskBoard               string         `json:"task_board"`
	TaskAttachmentURL       string         `json:"task_attachment_url"`
	TaskAttachments         datatypes.JSON `json:"task_attachments"`
	AccountingTransactionID *UUID          `json:"accounting_transaction_id"`
}
