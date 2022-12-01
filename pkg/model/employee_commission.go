package model

import "time"

type EmployeeCommission struct {
	BaseModel

	ProjectID                   UUID
	ProjectName                 string
	InvoiceID                   UUID
	InvoiceItemID               UUID
	EmployeeID                  UUID
	ProjectCommissionObjectID   UUID
	ProjectCommissionReceiverID UUID
	Percentage                  int
	Amount                      int64
	ConversionRate              float64
	IsPaid                      bool
	Formula                     string
	Note                        string
	PaidAt                      *time.Time
}
