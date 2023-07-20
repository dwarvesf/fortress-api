package deliverymetricmonthly

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetLast(db *gorm.DB, num int) ([]model.MonthlyDeliveryMetric, error)
	Avg(db *gorm.DB) (model.AvgMonthlyDeliveryMetric, error)
	AvgWithoutLatestMonth(db *gorm.DB) (model.AvgMonthlyDeliveryMetric, error)
}
