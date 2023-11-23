package role

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all positions
func (s *store) All(db *gorm.DB) ([]*model.Role, error) {
	var roles []*model.Role
	return roles, db.Where("is_show IS TRUE").Order("level").Find(&roles).Error
}

// GetByLevel get by input level
func (s *store) GetByLevel(db *gorm.DB, level int64) ([]*model.Role, error) {
	var roles []*model.Role
	return roles, db.Where("level >= ? AND is_show IS TRUE", level).Find(&roles).Error
}

func (s *store) GetByCode(db *gorm.DB, code string) (*model.Role, error) {
	var role *model.Role
	return role, db.Where("code = ?", code).First(&role).Error
}

func (s *store) GetByIDs(db *gorm.DB, ids []model.UUID) ([]*model.Role, error) {
	var roles []*model.Role
	return roles, db.Where("id IN ?", ids).Find(&roles).Error
}

// One get 1 one by id
func (s *store) One(db *gorm.DB, id model.UUID) (*model.Role, error) {
	var role *model.Role
	return role, db.Where("id = ?", id).First(&role).Error
}

// IsExist check the existence of employee
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM roles WHERE id = ?) as result", id)

	return result.Result, query.Scan(&result).Error
}
