package deliverymetric

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (client *model.DeliveryMetric, err error)
	Create(db *gorm.DB, e *model.DeliveryMetric) (client *model.DeliveryMetric, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, client model.DeliveryMetric, updatedFields ...string) (a *model.DeliveryMetric, err error)
}
