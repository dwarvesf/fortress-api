package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type WeeklyDeliveryMetric struct {
	Date      *time.Time
	SumWeight decimal.Decimal
	SumEffort decimal.Decimal
}

type AvgWeeklyDeliveryMetric struct {
	Weight decimal.Decimal
	Effort decimal.Decimal
}
