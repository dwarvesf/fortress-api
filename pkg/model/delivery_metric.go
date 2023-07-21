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

type LeaderBoard struct {
	Date  *time.Time
	Items []LeaderBoardItem
}

type MonthlyLeaderBoard struct {
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

type WeeklyReport struct {
	LastWeek    WeekReport `json:"last_week"`
	CurrentWeek WeekReport `json:"current_week"`

	TotalPointChangePercentage float32 `json:"total_point_change_percentage"`
	EffortChangePercentage     float32 `json:"effort_change_percentage"`
	AvgPointChangePercentage   float32 `json:"avg_point_change_percentage"`
	AvgEffortChangePercentage  float32 `json:"avg_effort_change_percentage"`
}

type WeekReport struct {
	Date        *time.Time `json:"date"`
	TotalPoints float32    `json:"total_points"`
	Effort      float32    `json:"effort"`
	AvgPoint    float32    `json:"avg_point"`
	AvgEffort   float32    `json:"avg_effort"`
}

type MonthlyReport struct {
	Reports []MonthReport `json:"reports"`
}

type MonthReport struct {
	Month       *time.Time `json:"date"`
	TotalWeight float32    `json:"total_weight"`
	Effort      float32    `json:"effort"`

	AvgWeight       float32 `json:"avg_weight"`
	AvgEffort       float32 `json:"avg_effort"`
	AvgWeeklyWeight float32 `json:"avg_weekly_weight"`
	AvgWeeklyEffort float32 `json:"avg_weekly_effort"`

	TotalPointChangePercentage      float32 `json:"total_point_change_percentage"`
	EffortChangePercentage          float32 `json:"effort_change_percentage"`
	AvgWeeklyPointChangePercentage  float32 `json:"avg_weekly_point_change_percentage"`
	AvgWeeklyEffortChangePercentage float32 `json:"avg_weekly_effort_change_percentage"`
}
