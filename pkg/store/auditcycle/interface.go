package auditcycle

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (auditCycle *model.AuditCycle, err error)
	All(db *gorm.DB) (auditCycles []*model.AuditCycle, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.AuditCycle) (auditCycle *model.AuditCycle, err error)
	Update(db *gorm.DB, auditCycle *model.AuditCycle) (ac *model.AuditCycle, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, auditCycle model.AuditCycle, updatedFields ...string) (ac *model.AuditCycle, err error)
	ResetActionItem(db *gorm.DB) (err error)
}
