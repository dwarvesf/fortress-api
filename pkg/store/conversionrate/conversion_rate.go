package conversionrate

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New create new pg service
func New() IStore {
	return &store{}
}

func (c *store) GetByCurrencyID(db *gorm.DB, id string) (*model.ConversionRate, error) {
	rs := &model.ConversionRate{}
	return rs, db.Where("id = ?", id).Preload("Currency").First(rs).Error
}

func (c *store) GetList(db *gorm.DB) ([]model.ConversionRate, error) {
	var rs []model.ConversionRate
	return rs, db.Preload("Currency").Find(&rs).Error
}

func (s *store) Update(db *gorm.DB, cr *model.ConversionRate) error {
	return db.Model(&cr).Where("currency_id = ?", cr.CurrencyID).Updates(&cr).Error
}
