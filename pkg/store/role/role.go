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
	return roles, db.Find(&roles).Error
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
