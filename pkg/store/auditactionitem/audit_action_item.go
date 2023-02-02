package auditactionitem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all audit action item
func (s *store) All(db *gorm.DB) ([]*model.AuditActionItem, error) {
	var auditActionItems []*model.AuditActionItem
	return auditActionItems, db.Find(&auditActionItems).Error
}

// All get all audit action item
func (s *store) AllByAuditID(db *gorm.DB, auditID string) ([]*model.AuditActionItem, error) {
	var auditActionItems []*model.AuditActionItem
	return auditActionItems, db.Where("audit_id = ?", auditID).Find(&auditActionItems).Error
}

// Delete delete 1 audit action item by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.AuditActionItem{}).Error
}

// DeleteByAuditID delete 1 audit action item by id
func (s *store) DeleteByAuditID(db *gorm.DB, auditID string) error {
	return db.Where("id = ?", auditID).Delete(&model.AuditActionItem{}).Error
}

// Create creates a new audit action item
func (s *store) Create(db *gorm.DB, e *model.AuditActionItem) (auditActionItem *model.AuditActionItem, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, auditActionItem *model.AuditActionItem) (*model.AuditActionItem, error) {
	return auditActionItem, db.Model(&auditActionItem).Where("id = ?", auditActionItem.ID).Updates(&auditActionItem).First(&auditActionItem).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.AuditActionItem, updatedFields ...string) (*model.AuditActionItem, error) {
	auditActionItem := model.AuditActionItem{}
	return &auditActionItem, db.Model(&auditActionItem).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
