package organization

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}
func (s *store) One(db *gorm.DB, id string) (*model.Organization, error) {
	var organization *model.Organization
	return organization, db.Where("id = ?", id).First(&organization).Error
}

func (s *store) All(db *gorm.DB) ([]*model.Organization, error) {
	var organizaitons []*model.Organization
	return organizaitons, db.Find(&organizaitons).Error
}

func (s *store) OneByCode(db *gorm.DB, code string) (*model.Organization, error) {
	var organization *model.Organization
	return organization, db.Where("code = ?", code).First(&organization).Error
}

// IsExist check organization existence
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM organizations WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
