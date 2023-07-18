package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type DeliveryMetric struct {
	BaseModel

	Weight        decimal.Decimal
	Effort        decimal.Decimal
	Effectiveness decimal.Decimal
	EmployeeID    UUID
	ProjectID     UUID
	Date          *time.Time
	Ref           int
}
