package view

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type DeliveryMetricWeeklyReport struct {
	LastWeek    DeliveryMetricWeekReport `json:"lastWeek"`
	CurrentWeek DeliveryMetricWeekReport `json:"currentWeek"`

	TotalPointChangePercentage float32 `json:"totalPointChangePercentage"`
	EffortChangePercentage     float32 `json:"effortChangePercentage"`
	AvgPointChangePercentage   float32 `json:"avgPointChangePercentage"`
	AvgEffortChangePercentage  float32 `json:"avgEffortChangePercentage"`
}

type DeliveryMetricWeekReport struct {
	Date        *time.Time `json:"date"`
	TotalPoints float32    `json:"totalPoints"`
	Effort      float32    `json:"effort"`
	AvgPoint    float32    `json:"avgPoint"`
	AvgEffort   float32    `json:"avgEffort"`
}

type DeliveryLeaderBoardResponse struct {
	Data WeeklyLeaderBoard `json:"data"`
}

type WeeklyLeaderBoard struct {
	Date  *time.Time        `json:"date"`
	Items []LeaderBoardItem `json:"items"`
}

type LeaderBoardItem struct {
	EmployeeID      string          `json:"employeeID"`
	EmployeeName    string          `json:"employeeName"`
	Points          decimal.Decimal `json:"points"`
	Effectiveness   decimal.Decimal `json:"effectiveness"`
	DiscordID       string          `json:"discordID"`
	DiscordUsername string          `json:"discordUsername"`
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

func ToDeliveryMetricMonthlyReport(current model.MonthReport, prev model.MonthReport) *DeliveryMetricMonthlyReport {
	return &DeliveryMetricMonthlyReport{
		CurrentMonth: DeliveryMetricMonthlyReportItem{
			Month:           current.Month,
			TotalWeight:     current.TotalWeight,
			Effort:          current.Effort,
			AvgWeight:       current.AvgWeight,
			AvgEffort:       current.AvgEffort,
			AvgWeeklyWeight: current.AvgWeeklyWeight,
			AvgWeeklyEffort: current.AvgWeeklyEffort,
		},
		LastMonth: DeliveryMetricMonthlyReportItem{
			Month:           prev.Month,
			TotalWeight:     prev.TotalWeight,
			Effort:          prev.Effort,
			AvgWeight:       prev.AvgWeight,
			AvgEffort:       prev.AvgEffort,
			AvgWeeklyWeight: prev.AvgWeeklyWeight,
			AvgWeeklyEffort: prev.AvgWeeklyEffort,
		},

		TotalPointChangePercentage:      current.TotalPointChangePercentage,
		EffortChangePercentage:          current.EffortChangePercentage,
		AvgWeightChangePercentage:       current.AvgWeightChangePercentage,
		AvgEffortChangePercentage:       current.AvgEffortChangePercentage,
		AvgWeeklyPointChangePercentage:  current.AvgWeeklyPointChangePercentage,
		AvgWeeklyEffortChangePercentage: current.AvgWeeklyEffortChangePercentage,
	}
}

type DeliveryMetricMonthlyReport struct {
	CurrentMonth DeliveryMetricMonthlyReportItem `json:"currentMonth"`
	LastMonth    DeliveryMetricMonthlyReportItem `json:"lastMonth"`

	TotalPointChangePercentage      float32 `json:"totalPointChangePercentage"`
	EffortChangePercentage          float32 `json:"effortChangePercentage"`
	AvgWeightChangePercentage       float32 `json:"avgWeightChangePercentage"`
	AvgEffortChangePercentage       float32 `json:"avgEffortChangePercentage"`
	AvgWeeklyPointChangePercentage  float32 `json:"avgWeeklyPointChangePercentage"`
	AvgWeeklyEffortChangePercentage float32 `json:"avgWeeklyEffortChangePercentage"`
}

type DeliveryMetricMonthlyReportItem struct {
	Month       *time.Time `json:"date"`
	TotalWeight float32    `json:"totalWeight"`
	Effort      float32    `json:"effort"`

	AvgWeight       float32 `json:"avgWeight"`
	AvgEffort       float32 `json:"avgEffort"`
	AvgWeeklyWeight float32 `json:"avgWeeklyWeight"`
	AvgWeeklyEffort float32 `json:"avgWeeklyEffort"`
}

type MonthlyLeaderBoard struct {
	Date  *time.Time        `json:"date"`
	Items []LeaderBoardItem `json:"items"`
}
