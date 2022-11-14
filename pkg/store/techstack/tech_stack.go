package techstack

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

// One get all Senitorities
func (s *store) All() ([]*model.TechStack, error) {
	var techStacks []*model.TechStack
	return techStacks, s.db.Find(&techStacks).Error
}
