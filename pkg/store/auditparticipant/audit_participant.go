package auditparticipant

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all audit participant
func (s *store) All(db *gorm.DB) ([]*model.AuditParticipant, error) {
	var auditParticipants []*model.AuditParticipant
	return auditParticipants, db.Find(&auditParticipants).Error
}

// Delete delete 1 audit participant by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.AuditParticipant{}).Error
}

// Create creates a new audit participant
func (s *store) Create(db *gorm.DB, e *model.AuditParticipant) (auditParticipant *model.AuditParticipant, err error) {
	return e, db.Create(e).Error
}

// Update update all value (including nested model)
func (s *store) Update(db *gorm.DB, auditParticipant *model.AuditParticipant) (*model.AuditParticipant, error) {
	return auditParticipant, db.Model(&auditParticipant).Where("id = ?", auditParticipant.ID).Updates(&auditParticipant).First(&auditParticipant).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.AuditParticipant, updatedFields ...string) (*model.AuditParticipant, error) {
	auditParticipant := model.AuditParticipant{}
	return &auditParticipant, db.Model(&auditParticipant).Where("id = ?", id).Select(updatedFields).Updates(updateModel).Error
}
