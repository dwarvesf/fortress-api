package view

import (
	"time"
)

type DeliveryMetricWeeklyReport struct {
	LastWeek    DeliveryMetricWeekReport `json:"last_week"`
	CurrentWeek DeliveryMetricWeekReport `json:"current_week"`

	TotalPointChangePercentage float32 `json:"total_point_change_percentage"`
	EffortChangePercentage     float32 `json:"effort_change_percentage"`
	AvgPointChangePercentage   float32 `json:"avg_point_change_percentage"`
	AvgEffortChangePercentage  float32 `json:"avg_effort_change_percentage"`
}

type DeliveryMetricWeekReport struct {
	Date        *time.Time `json:"date"`
	TotalPoints float32    `json:"total_points"`
	Effort      float32    `json:"effort"`
	AvgPoint    float32    `json:"avg_point"`
	AvgEffort   float32    `json:"avg_effort"`
}
