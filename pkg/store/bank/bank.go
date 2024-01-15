package bank

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get bank account by id
func (s *store) One(db *gorm.DB, id string) (*model.Bank, error) {
	var bankAccount *model.Bank
	return bankAccount, db.Where("id = ?", id).First(&bankAccount).Error
}

type GetBanksInput struct {
	ID        string
	Bin       string
	SwiftCode string
}

// All get all bank account
func (s *store) All(db *gorm.DB, in GetBanksInput) ([]*model.Bank, error) {
	var banks []*model.Bank

	query := db.Where("is_active IS TRUE")
	if in.ID != "" {
		query = query.Where("id = ?", in.ID)
	}
	if in.Bin != "" {
		query = query.Where("bin = ?", in.Bin)
	}
	if in.SwiftCode != "" {
		query = query.Where("swift_code = ?", in.SwiftCode)
	}

	return banks, query.Find(&banks).Error
}

// IsExist check bank account existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM banks WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// OneByBin get bank account by bin
func (s *store) OneByBin(db *gorm.DB, bin string) (bank *model.Bank, err error) {
	var bankAccount *model.Bank
	return bankAccount, db.Where("bin = ?", bin).First(&bankAccount).Error
}

// OneBySwiftCode get bank account by swift code
func (s *store) OneBySwiftCode(db *gorm.DB, code string) (bank *model.Bank, err error) {
	var bankAccount *model.Bank
	return bankAccount, db.Where("swift_code = ?", code).First(&bankAccount).Error
}
