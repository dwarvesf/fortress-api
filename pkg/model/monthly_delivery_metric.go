package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type MonthlyDeliveryMetric struct {
	Month     *time.Time
	SumWeight decimal.Decimal
	SumEffort decimal.Decimal
}

type AvgMonthlyDeliveryMetric struct {
	Weight decimal.Decimal
	Effort decimal.Decimal
}
