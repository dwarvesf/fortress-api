package valuation

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	GetAccountReceivable(db *gorm.DB, year string) (total *model.CurrencyView, err error)
	GetLiabilities(db *gorm.DB, year string) (res []model.Liability, total *model.CurrencyView, err error)
	GetRevenue(db *gorm.DB, year string) (total *model.CurrencyView, err error)
	GetInvestment(db *gorm.DB, year string) (total *model.CurrencyView, err error)

	// GetAssetAmount return total amount of current holding assets
	GetAssetAmount(db *gorm.DB, year string) (total float64, err error)
	GetExpense(db *gorm.DB, year string) (total *model.CurrencyView, err error)
	GetPayroll(db *gorm.DB, year string) (total float64, err error)
}
