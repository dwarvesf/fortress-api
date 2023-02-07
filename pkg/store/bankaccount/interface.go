package bankaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (bankAccount *model.BankAccount, err error)
	All(db *gorm.DB) (bankAccounts []*model.BankAccount, err error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
}
