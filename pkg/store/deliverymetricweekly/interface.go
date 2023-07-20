package deliverymetricweekly

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetLast(db *gorm.DB, num int) ([]model.WeeklyDeliveryMetric, error)
	Avg(db *gorm.DB) (model.AvgWeeklyDeliveryMetric, error)
	AvgWithoutLatestWeek(db *gorm.DB) (model.AvgWeeklyDeliveryMetric, error)
	AvgByMonth(db *gorm.DB, month time.Time) ([]model.AvgMonthWeeklyDeliveryMetric, error)
}
