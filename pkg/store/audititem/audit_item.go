package audititem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all audit item
func (s *store) All(db *gorm.DB) ([]*model.AuditItem, error) {
	var auditItems []*model.AuditItem
	return auditItems, db.Find(&auditItems).Error
}

// Delete delete 1 audit item by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.AuditItem{}).Error
}

// Create creates a new audit item
func (s *store) Create(db *gorm.DB, e *model.AuditItem) (auditItem *model.AuditItem, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, auditItem *model.AuditItem) (*model.AuditItem, error) {
	return auditItem, db.Model(&auditItem).Where("id = ?", auditItem.ID).Updates(&auditItem).First(&auditItem).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.AuditItem, updatedFields ...string) (*model.AuditItem, error) {
	auditItem := model.AuditItem{}
	return &auditItem, db.Model(&auditItem).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
