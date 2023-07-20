package deliverymetrics

import (
	"errors"
	"math"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

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

func (c controller) GetWeeklyReport() (*WeeklyReport, error) {
	return GetWeeklyReport(c.store, c.repo.DB())
}

func GetWeeklyReport(s *store.Store, db *gorm.DB) (*WeeklyReport, error) {
	// Get data of latest week
	metrics, err := s.WeeklyDeliveryMetric.GetLast(db, 2)
	if err != nil {
		return nil, err
	}
	if len(metrics) < 2 {
		return nil, errors.New("not enough data")
	}

	currentReport := metrics[0]
	lastWeekReport := metrics[1]

	report := &WeeklyReport{
		LastWeek: WeekReport{
			Date:        lastWeekReport.Date,
			TotalPoints: decimalToRoundedFloat32(lastWeekReport.SumWeight),
			Effort:      decimalToRoundedFloat32(lastWeekReport.SumEffort),
		},
		CurrentWeek: WeekReport{
			Date:        currentReport.Date,
			TotalPoints: decimalToRoundedFloat32(currentReport.SumWeight),
			Effort:      decimalToRoundedFloat32(currentReport.SumEffort),
		},
	}

	// Avg
	avgMetric, err := s.WeeklyDeliveryMetric.Avg(db)
	if err != nil {
		return nil, err
	}
	report.CurrentWeek.AvgPoint = decimalToRoundedFloat32(avgMetric.Weight)
	report.CurrentWeek.AvgEffort = decimalToRoundedFloat32(avgMetric.Effort)

	avgWithoutLatestWeek, err := s.WeeklyDeliveryMetric.AvgWithoutLatestWeek(db)
	if err != nil {
		return nil, err
	}
	report.LastWeek.AvgPoint = decimalToRoundedFloat32(avgWithoutLatestWeek.Weight)
	report.LastWeek.AvgEffort = decimalToRoundedFloat32(avgWithoutLatestWeek.Effort)

	// Compare data of current week and last week
	report.TotalPointChangePercentage = roundFloat32To2Decimals(
		(report.CurrentWeek.TotalPoints - report.LastWeek.TotalPoints) / report.LastWeek.TotalPoints * 100)
	report.EffortChangePercentage = roundFloat32To2Decimals(
		(report.CurrentWeek.Effort - report.LastWeek.Effort) / report.LastWeek.Effort * 100)
	report.AvgPointChangePercentage = roundFloat32To2Decimals(
		(report.CurrentWeek.AvgPoint - report.LastWeek.AvgPoint) / report.LastWeek.AvgPoint * 100)
	report.AvgEffortChangePercentage = roundFloat32To2Decimals(
		(report.CurrentWeek.AvgEffort - report.LastWeek.AvgEffort) / report.LastWeek.AvgEffort * 100)

	return report, nil
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

func (c controller) GetMonthlyReport() (*MonthlyReport, error) {
	return GetMonthlyReport(c.store, c.repo.DB(), 2) // 2 latest months
}

func GetMonthlyReport(s *store.Store, db *gorm.DB, monthNumToTake int) (*MonthlyReport, error) {
	// Get data of latest month
	metrics, err := s.MonthlyDeliveryMetric.GetLast(db, monthNumToTake)
	if err != nil {
		return nil, err
	}
	if len(metrics) < 2 {
		return nil, errors.New("not enough data")
	}

	reports := make([]MonthReport, 0, monthNumToTake)
	for _, m := range metrics {
		r := MonthReport{
			Month:       m.Month,
			TotalWeight: decimalToRoundedFloat32(m.SumWeight),
			Effort:      decimalToRoundedFloat32(m.SumEffort),
		}

		// Avg monthly
		avgMetric, err := s.MonthlyDeliveryMetric.Avg(db)
		if err != nil {
			return nil, err
		}
		r.AvgWeight = decimalToRoundedFloat32(avgMetric.Weight)
		r.AvgEffort = decimalToRoundedFloat32(avgMetric.Effort)

		// Avg month weekly
		if m.Month != nil {
			avgMonthWeekly, err := s.WeeklyDeliveryMetric.AvgByMonth(db, *m.Month)
			if err != nil {
				return nil, err
			}
			if len(avgMonthWeekly) == 0 {
				return nil, errors.New("missing month weekly data")
			}
			r.AvgWeeklyWeight = decimalToRoundedFloat32(avgMonthWeekly[0].Weight)
			r.AvgWeeklyEffort = decimalToRoundedFloat32(avgMonthWeekly[0].Effort)
		}

		reports = append(reports, r)
	}

	// Calculate change with previous month
	for i := 0; i < len(reports)-1; i++ {
		currentReport := reports[i]
		prevMonthReport := reports[i+1]

		currentReport.TotalPointChangePercentage = roundFloat32To2Decimals(
			(currentReport.TotalWeight - prevMonthReport.TotalWeight) / prevMonthReport.TotalWeight * 100)
		currentReport.EffortChangePercentage = roundFloat32To2Decimals(
			(currentReport.Effort - prevMonthReport.Effort) / prevMonthReport.Effort * 100)
		currentReport.AvgWeeklyPointChangePercentage = roundFloat32To2Decimals(
			(currentReport.AvgWeeklyWeight - prevMonthReport.AvgWeeklyWeight) / prevMonthReport.AvgWeeklyWeight * 100)
		currentReport.AvgWeeklyEffortChangePercentage = roundFloat32To2Decimals(
			(currentReport.AvgWeeklyEffort - prevMonthReport.AvgWeeklyEffort) / prevMonthReport.AvgWeeklyEffort * 100)

		// Update new value to slice
		reports[i] = currentReport
	}

	return &MonthlyReport{
		Reports: reports,
	}, nil
}

func decimalToRoundedFloat32(d decimal.Decimal) float32 {
	f, _ := d.Float64()
	rounded := math.Round(f*100) / 100
	return float32(rounded)
}

func roundFloat32To2Decimals(f float32) float32 {
	rounded := math.Round(float64(f)*100) / 100
	return float32(rounded)
}
