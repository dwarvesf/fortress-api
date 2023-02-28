package bankaccount

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get bank account by id
func (s *store) One(db *gorm.DB, id string) (*model.BankAccount, error) {
	var bankAccount *model.BankAccount
	return bankAccount, db.Where("id = ?", id).
		Preload("Currency").
		First(&bankAccount).Error
}

// All get all bank account
func (s *store) All(db *gorm.DB) ([]*model.BankAccount, error) {
	var bankAccounts []*model.BankAccount
	return bankAccounts, db.Preload("Currency").Find(&bankAccounts).Error
}

// IsExist check bank account existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM bank_accounts WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
