package deliverymetric

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (client *model.DeliveryMetric, err error)
	GetLatest(db *gorm.DB) (*model.DeliveryMetric, error)
	GetLatestWeek(db *gorm.DB) (*time.Time, error)
	GetLatestMonth(db *gorm.DB) (*time.Time, error)
	GetTopWeighMetrics(db *gorm.DB, w *time.Time, limit int) ([]model.DeliveryMetric, error)
	GetTopMonthlyWeighMetrics(db *gorm.DB, m *time.Time, limit int) ([]model.DeliveryMetric, error)

	Create(db *gorm.DB, e []model.DeliveryMetric) (rs []model.DeliveryMetric, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, client model.DeliveryMetric, updatedFields ...string) (a *model.DeliveryMetric, err error)
}
