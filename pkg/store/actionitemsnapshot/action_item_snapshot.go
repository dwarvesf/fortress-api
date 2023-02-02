package actionitemsnapshot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get action item snapshot by id
func (s *store) One(db *gorm.DB, id string) (*model.ActionItemSnapshot, error) {
	var actionItemSnapshot *model.ActionItemSnapshot
	return actionItemSnapshot, db.Where("id = ?", id).First(&actionItemSnapshot).Error
}

// All get all action item snapshot
func (s *store) All(db *gorm.DB) ([]*model.ActionItemSnapshot, error) {
	var actionItemSnapshots []*model.ActionItemSnapshot
	return actionItemSnapshots, db.Find(&actionItemSnapshots).Error
}

// Delete delete 1 action item snapshot by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.ActionItemSnapshot{}).Error
}

// Create creates a new action item snapshot
func (s *store) Create(db *gorm.DB, e *model.ActionItemSnapshot) (actionItemSnapshot *model.ActionItemSnapshot, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, actionItemSnapshot *model.ActionItemSnapshot) (*model.ActionItemSnapshot, error) {
	return actionItemSnapshot, db.Model(&actionItemSnapshot).Where("id = ?", actionItemSnapshot.ID).Updates(&actionItemSnapshot).First(&actionItemSnapshot).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ActionItemSnapshot, updatedFields ...string) (*model.ActionItemSnapshot, error) {
	actionItemSnapshot := model.ActionItemSnapshot{}
	return &actionItemSnapshot, db.Model(&actionItemSnapshot).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
