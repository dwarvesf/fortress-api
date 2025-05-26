package model

import "time"

// InboundFundTransaction records amounts moved to the inbound fund
type InboundFundTransaction struct {
	BaseModel

	InvoiceID      UUID        `json:"invoice_id"`
	Amount         VietnamDong `json:"amount"`
	Notes          string      `json:"notes"`
	ConversionRate float64
	PaidAt         *time.Time

	Invoice *Invoice `json:"invoice,omitempty"`
}
