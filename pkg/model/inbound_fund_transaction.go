package model

// InboundFundTransaction records amounts moved to the inbound fund
type InboundFundTransaction struct {
	BaseModel

	InvoiceID UUID    `json:"invoice_id"`
	Amount    float64 `json:"amount"`
	Notes     string  `json:"notes"`

	Invoice *Invoice `json:"invoice,omitempty"`
}
