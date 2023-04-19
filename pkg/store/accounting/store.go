package accounting

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore implement calendar method
type IStore interface {
	CreateTransaction(db *gorm.DB, transaction *model.AccountingTransaction) error
	GetAccountingTransactions(db *gorm.DB) ([]model.AccountingTransaction, error)
	GetAccountingCategories(db *gorm.DB) ([]model.AccountingCategory, error)
	DeleteTransaction(db *gorm.DB, t *model.AccountingTransaction) error
	CreateMultipleTransaction(db *gorm.DB, transactions []*model.AccountingTransaction) error
}
