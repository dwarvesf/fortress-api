package auditcycle

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get audit cycle by id
func (s *store) One(db *gorm.DB, id string) (*model.AuditCycle, error) {
	var auditCycle *model.AuditCycle
	return auditCycle, db.Where("id = ?", id).First(&auditCycle).Error
}

// All get all audit cycle
func (s *store) All(db *gorm.DB) ([]*model.AuditCycle, error) {
	var auditCycles []*model.AuditCycle
	return auditCycles, db.Preload("Project", "deleted_at IS NULL").Find(&auditCycles).Error
}

// Delete delete 1 audit cycle by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.AuditCycle{}).Error
}

// Create creates a new audit cycle
func (s *store) Create(db *gorm.DB, e *model.AuditCycle) (auditCycle *model.AuditCycle, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, auditCycle *model.AuditCycle) (*model.AuditCycle, error) {
	return auditCycle, db.Model(&auditCycle).Where("id = ?", auditCycle.ID).Updates(&auditCycle).First(&auditCycle).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.AuditCycle, updatedFields ...string) (*model.AuditCycle, error) {
	auditCycle := model.AuditCycle{}
	return &auditCycle, db.Model(&auditCycle).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// ResetActionItem reset action_item high medium low to 0 in audit_cycle table
func (s *store) ResetActionItem(db *gorm.DB) error {
	return db.Model(&model.AuditCycle{}).Where("deleted_at IS NULL").Updates(map[string]interface{}{"action_item_high": 0, "action_item_medium": 0, "action_item_low": 0}).Error
}
