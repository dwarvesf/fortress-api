package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusError     InvoiceStatus = "error"
	InvoiceStatusScheduled InvoiceStatus = "scheduled"
)

func (i InvoiceStatus) IsValid() bool {
	switch i {
	case InvoiceStatusDraft,
		InvoiceStatusSent,
		InvoiceStatusOverdue,
		InvoiceStatusPaid,
		InvoiceStatusError,
		InvoiceStatusScheduled:
		return true
	}
	return false
}

func (i InvoiceStatus) String() string {
	return string(i)
}

// Invoice contain company information
type Invoice struct {
	BaseModel

	Number           string
	InvoicedAt       *time.Time
	DueAt            *time.Time
	PaidAt           *time.Time
	FailedAt         *time.Time
	Status           InvoiceStatus
	Email            string
	CC               JSON
	Description      string
	Note             string
	SubTotal         int64
	Tax              int64
	Discount         int64
	Total            int64
	ConversionAmount int64
	InvoiceFileURL   string
	ErrorInvoiceID   *UUID
	LineItems        JSON
	Month            int
	Year             int
	SentBy           *UUID
	Sender           *Employee `gorm:"foreignKey:sent_by;"`
	ThreadID         string
	ScheduledDate    *time.Time
	ConversionRate   float64

	BankID UUID
	Bank   *BankAccount

	ProjectID UUID
	Project   *Project

	InvoiceFileContent []byte `gorm:"-"` // we not store this in db
	MessageID          string `gorm:"-"`
	References         string `gorm:"-"`
	TodoAttachment     string `gorm:"-"`
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
	var ccList []string
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

type InvoiceItem struct {
	Quantity    float64 `json:"quantity"`
	UnitCost    int64   `json:"unitCost"`
	Discount    int64   `json:"discount"`
	Cost        int64   `json:"cost"`
	Description string  `json:"description"`
	IsExternal  bool    `json:"isExternal"`
}

func GetInfoItems(lineItems JSON) ([]InvoiceItem, error) {
	var items []InvoiceItem

	if len(lineItems) == 0 || string(lineItems) == "null" {
		return items, nil
	}

	if err := json.Unmarshal(lineItems, &items); err != nil {
		return nil, err
	}
	return items, nil
}
