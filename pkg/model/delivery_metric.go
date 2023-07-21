package model

import (
	"time"

	"github.com/shopspring/decimal"
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

type DeliveryMetrics []DeliveryMetric

type WeeklyLeaderBoard struct {
	Date  *time.Time
	Items []LeaderBoardItem
}

type LeaderBoardItem struct {
	EmployeeID      string
	EmployeeName    string
	Points          decimal.Decimal
	Effectiveness   decimal.Decimal
	DiscordID       string
	DiscordUsername string
	Rank            int
}
