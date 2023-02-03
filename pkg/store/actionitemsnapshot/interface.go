package actionitemsnapshot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (actionItemSnapshot *model.ActionItemSnapshot, err error)
	All(db *gorm.DB) (actionItemSnapshots []*model.ActionItemSnapshot, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.ActionItemSnapshot) (actionItemSnapshot *model.ActionItemSnapshot, err error)
	Update(db *gorm.DB, actionItemSnapshot *model.ActionItemSnapshot) (ac *model.ActionItemSnapshot, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, actionItemSnapshot model.ActionItemSnapshot, updatedFields ...string) (ac *model.ActionItemSnapshot, err error)
	OneByAuditCycleIDAndTime(db *gorm.DB, auditCycleID string, today string) (actionItemSnapshot *model.ActionItemSnapshot, err error)
}
