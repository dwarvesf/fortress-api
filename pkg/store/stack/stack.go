package stack

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
func (s *store) All() ([]*model.Stack, error) {
	var stacks []*model.Stack
	return stacks, s.db.Find(&stacks).Error
}
