package currency

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type currencyService struct{}

// New create new pg service
func New() IStore {
	return &currencyService{}
}

func (c currencyService) GetByName(db *gorm.DB, name string) (*model.Currency, error) {
	currency := &model.Currency{}
	return currency, db.Where("name = ?", name).First(currency).Error
}

func (c currencyService) GetList(db *gorm.DB) ([]model.Currency, error) {
	currencies := []model.Currency{}
	return currencies, db.Find(&currencies).Error
}
