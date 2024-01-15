package bank

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (bank *model.Bank, err error)
	OneByBin(db *gorm.DB, bin string) (bank *model.Bank, err error)
	OneBySwiftCode(db *gorm.DB, code string) (bank *model.Bank, err error)
	All(db *gorm.DB, in GetBanksInput) ([]*model.Bank, error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
}
