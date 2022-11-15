package stack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// All get all Senitorities
func (s *store) All(db *gorm.DB) ([]*model.Stack, error) {
	var stacks []*model.Stack
	return stacks, db.Find(&stacks).Error
}
