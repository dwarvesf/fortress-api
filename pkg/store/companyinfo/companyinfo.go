package companyinfo

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get companyInfo by id
func (s *store) One(db *gorm.DB, id string) (*model.CompanyInfo, error) {
	var companyInfo *model.CompanyInfo
	return companyInfo, db.Where("id = ?", id).First(&companyInfo).Error
}

// IsExist check client contact existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM company_infos WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}

// All get all companyInfo
func (s *store) All(db *gorm.DB) ([]*model.CompanyInfo, error) {
	var companyInfo []*model.CompanyInfo
	return companyInfo, db.Find(&companyInfo).Error
}
