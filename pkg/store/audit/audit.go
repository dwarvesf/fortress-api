package audit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// One get audit by id
func (s *store) One(db *gorm.DB, id string) (*model.Audit, error) {
	var audit *model.Audit
	return audit, db.Where("id = ?", id).First(&audit).Error
}

// All get all audit
func (s *store) All(db *gorm.DB) ([]*model.Audit, error) {
	var audit []*model.Audit
	return audit, db.Find(&audit).Error
}

// Delete delete 1 audit by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Audit{}).Error
}

// Create creates a new audit
func (s *store) Create(db *gorm.DB, e *model.Audit) (audit *model.Audit, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, audit *model.Audit) (*model.Audit, error) {
	return audit, db.Model(&audit).Where("id = ?", audit.ID).Updates(&audit).First(&audit).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.Audit, updatedFields ...string) (*model.Audit, error) {
	audit := model.Audit{}
	return &audit, db.Model(&audit).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}

// ResetActionItem reset action item in audit table
func (s *store) ResetActionItem(db *gorm.DB) error {
	return db.Model(&model.Audit{}).Where("deleted_at IS NULL").Updates(map[string]interface{}{"action_item": 0}).Error
}
