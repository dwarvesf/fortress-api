package conversionrate

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByCurrencyID(db *gorm.DB, id string) (*model.ConversionRate, error)
	GetList(db *gorm.DB) ([]model.ConversionRate, error)
	Update(db *gorm.DB, cr *model.ConversionRate) error
}
