package role

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// One get all positions
func (s *store) All() ([]*model.Role, error) {
	var roles []*model.Role
	return roles, s.db.Find(&roles).Error
}

// One get 1 one by id
func (s *store) One(id model.UUID) (*model.Role, error) {
	var role *model.Role
	return role, s.db.Where("id = ?", id).First(&role).Error
}
