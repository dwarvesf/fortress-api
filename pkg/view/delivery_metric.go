package view

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
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

type DeliveryLeaderBoardResponse struct {
	Data WeeklyLeaderBoard `json:"data"`
}

type WeeklyLeaderBoard struct {
	Date  *time.Time        `json:"date"`
	Items []LeaderBoardItem `json:"items"`
}

type LeaderBoardItem struct {
	EmployeeID      string          `json:"employee_id"`
	EmployeeName    string          `json:"employee_name"`
	Points          decimal.Decimal `json:"points"`
	Effectiveness   decimal.Decimal `json:"effectiveness"`
	DiscordID       string          `json:"discord_id"`
	DiscordUsername string          `json:"discord_username"`
	Rank            int             `json:"rank"`
}

func ToDeliveryMetricLeaderBoard(board *model.LeaderBoard) *WeeklyLeaderBoard {
	items := make([]LeaderBoardItem, 0, len(board.Items))
	// Get user info
	for _, m := range board.Items {
		items = append(items, LeaderBoardItem{
			EmployeeID:      m.EmployeeID,
			EmployeeName:    m.EmployeeName,
			Points:          m.Points,
			Effectiveness:   m.Effectiveness,
			DiscordID:       m.DiscordID,
			DiscordUsername: m.DiscordUsername,
			Rank:            m.Rank,
		})
	}

	return &WeeklyLeaderBoard{
		Date:  board.Date,
		Items: items,
	}
}

func ToDeliveryMetricWeeklyReport(in *model.WeeklyReport) *DeliveryMetricWeeklyReport {
	return &DeliveryMetricWeeklyReport{
		LastWeek: DeliveryMetricWeekReport{
			Date:        in.LastWeek.Date,
			TotalPoints: in.LastWeek.TotalPoints,
			Effort:      in.LastWeek.Effort,
			AvgPoint:    in.LastWeek.AvgPoint,
			AvgEffort:   in.LastWeek.AvgEffort,
		},
		CurrentWeek: DeliveryMetricWeekReport{
			Date:        in.CurrentWeek.Date,
			TotalPoints: in.CurrentWeek.TotalPoints,
			Effort:      in.CurrentWeek.Effort,
			AvgPoint:    in.CurrentWeek.AvgPoint,
			AvgEffort:   in.CurrentWeek.AvgEffort,
		},
		TotalPointChangePercentage: in.TotalPointChangePercentage,
		EffortChangePercentage:     in.EffortChangePercentage,
		AvgPointChangePercentage:   in.AvgPointChangePercentage,
		AvgEffortChangePercentage:  in.AvgEffortChangePercentage,
	}
}

func ToDeliveryMetricMonthlyReport(in *model.MonthlyReport) *DeliveryMetricMonthlyReport {
	return &DeliveryMetricMonthlyReport{}
}

type DeliveryMetricMonthlyReport struct {
	CurrentMonth DeliveryMetricMonthlyReportItem `json:"current_month"`
	LastMonth    DeliveryMetricMonthlyReportItem `json:"last_month"`

	TotalPointChangePercentage      float32 `json:"total_point_change_percentage"`
	EffortChangePercentage          float32 `json:"effort_change_percentage"`
	AvgWeeklyPointChangePercentage  float32 `json:"avg_weekly_point_change_percentage"`
	AvgWeeklyEffortChangePercentage float32 `json:"avg_weekly_effort_change_percentage"`
}

type DeliveryMetricMonthlyReportItem struct {
	Month       *time.Time `json:"date"`
	TotalWeight float32    `json:"total_weight"`
	Effort      float32    `json:"effort"`

	AvgWeight       float32 `json:"avg_weight"`
	AvgEffort       float32 `json:"avg_effort"`
	AvgWeeklyWeight float32 `json:"avg_weekly_weight"`
	AvgWeeklyEffort float32 `json:"avg_weekly_effort"`
}

type MonthlyLeaderBoard struct {
	Date  *time.Time        `json:"date"`
	Items []LeaderBoardItem `json:"items"`
}
