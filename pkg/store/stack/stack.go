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
func (s *store) GetByIDs(db *gorm.DB, ids []model.UUID) ([]*model.Stack, error) {
	var stacks []*model.Stack
	return stacks, db.Where("id IN ?", ids).Find(&stacks).Error
}

// Update update the stack
func (s *store) Update(db *gorm.DB, stack *model.Stack) (*model.Stack, error) {
	return stack, db.Model(&model.Stack{}).Where("id = ?", stack.ID).Updates(&stack).First(&stack).Error
}

// Delete delete ProjectMember by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Stack{}).Error
}

// Create create new stack
func (s *store) Create(db *gorm.DB, stack *model.Stack) (*model.Stack, error) {
	return stack, db.Create(stack).Error
}
