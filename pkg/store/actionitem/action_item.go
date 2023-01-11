package actionitem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all action item
func (s *store) All(db *gorm.DB) ([]*model.ActionItem, error) {
	var actionItems []*model.ActionItem
	return actionItems, db.Find(&actionItems).Error
}

// Delete delete 1 action item by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.ActionItem{}).Error
}

// Create creates a new action item
func (s *store) Create(db *gorm.DB, e *model.ActionItem) (actionItem *model.ActionItem, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, actionItem *model.ActionItem) (*model.ActionItem, error) {
	return actionItem, db.Model(&actionItem).Where("id = ?", actionItem.ID).Updates(&actionItem).First(&actionItem).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ActionItem, updatedFields ...string) (*model.ActionItem, error) {
	actionItem := model.ActionItem{}
	return &actionItem, db.Model(&actionItem).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
