package deliverymetricmonthly

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetLast(db *gorm.DB, num int) ([]model.MonthlyDeliveryMetric, error)
	AvgTo(db *gorm.DB, maxMonth *time.Time) (model.AvgMonthlyDeliveryMetric, error)
}
