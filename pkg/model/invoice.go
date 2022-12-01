package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvoiceAlreadyMarkedAsPaid  = errors.New("invoice has already been marked as paid")
	ErrInvoiceAlreadyMarkedAsError = errors.New("invoice has already been marked as error")
)

type InvoiceStatus string

const (
	InvoiceDraft     InvoiceStatus = "draft"
	InvoiceSent      InvoiceStatus = "sent"
	InvoiceOverdue   InvoiceStatus = "overdue"
	InvoicePaid      InvoiceStatus = "paid"
	InvoiceError     InvoiceStatus = "error"
	InvoiceScheduled InvoiceStatus = "scheduled"
)

func (i InvoiceStatus) String() string {
	return string(i)
}

//Invoice contain company information
type Invoice struct {
	BaseModel

	Number           string
	InvoicedAt       *time.Time
	DueAt            *time.Time
	PaidAt           *time.Time
	FailedAt         *time.Time
	UpdatedAt        *time.Time
	StatusID         InvoiceStatus
	Email            string
	CC               JSON
	Description      string
	Note             string
	SubTotal         int64
	Tax              int64
	Discount         int64
	Total            int64
	ConversionAmount VietnamDong
	InvoiceFileURL   string
	ErrorInvoiceID   *UUID
	LineItems        JSON
	Month            int
	Year             int
	SentByID         *UUID
	SentBy           *Employee
	ThreadID         string
	ScheduledDate    *time.Time
	ConversionRate   float64

	BankID UUID
	Bank   BankAccount

	ProjectID UUID
	Project   *Project

	InvoiceFileContent []byte   `gorm:"-"` // we not store this in db
	MessageID          string   `gorm:"-"`
	References         string   `gorm:"-"`
	TodoAttachment     string   `gorm:"-"`
	CCs                []string `gorm:"-"`
}

func (i *Invoice) Validate() error {
	if i == nil {
		return errors.New("nil structure")
	}
	if i.Project == nil {
		return errors.New("missing project")
	}
	if i.Bank.Currency.Name == "" {
		return errors.New("missing bank info")
	}
	return nil
}

func GatherAddresses(CCs JSON) (string, error) {
	ccList := []string{}
	if err := json.Unmarshal(CCs, &ccList); err != nil {
		return "", err
	}
	for _, v := range ccList {
		if v == "" {
			continue
		}
	}
	return strings.Join(ccList, ", "), nil
}
