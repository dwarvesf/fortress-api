package auditparticipant

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (auditParticipants []*model.AuditParticipant, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.AuditParticipant) (auditParticipant *model.AuditParticipant, err error)
	Update(db *gorm.DB, auditParticipant *model.AuditParticipant) (ac *model.AuditParticipant, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, auditParticipant model.AuditParticipant, updatedFields ...string) (ac *model.AuditParticipant, err error)
}
