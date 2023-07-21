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

func ToDeliveryMetricLeaderBoard(board *model.WeeklyLeaderBoard) *WeeklyLeaderBoard {
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
