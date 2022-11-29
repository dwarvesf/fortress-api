package stack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all Senitorities
func (s *store) All(db *gorm.DB) ([]*model.Stack, error) {
	var stacks []*model.Stack
	return stacks, db.Find(&stacks).Error
}

// One get 1 stack by id
func (s *store) One(db *gorm.DB, id string) (*model.Stack, error) {
	var stack *model.Stack
	return stack, db.Where("id = ?", id).First(&stack).Error
}

// GetByIDs return list stack by IDs
func (s *store) GetByIDs(db *gorm.DB, ids []string) ([]*model.Stack, error) {
	var stacks []*model.Stack
	return stacks, db.Where("id IN ?", ids).Find(&stacks).Error
}
