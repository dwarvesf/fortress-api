package deliverymetrics

import (
	"errors"
	"math"

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
	TotalPoints float32 `json:"total_points"`
	Effort      float32 `json:"effort"`
	AvgPoint    float32 `json:"avg_point"`
	AvgEffort   float32 `json:"avg_effort"`
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
			TotalPoints: decimalToRoundedFloat32(lastWeekReport.SumWeight),
			Effort:      decimalToRoundedFloat32(lastWeekReport.SumEffort),
		},
		CurrentWeek: WeekReport{
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

func decimalToRoundedFloat32(d decimal.Decimal) float32 {
	f, _ := d.Float64()
	rounded := math.Round(f*100) / 100
	return float32(rounded)
}

func roundFloat32To2Decimals(f float32) float32 {
	rounded := math.Round(float64(f)*100) / 100
	return float32(rounded)
}
