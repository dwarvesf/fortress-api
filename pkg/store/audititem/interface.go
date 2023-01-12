package audititem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (auditItems []*model.AuditItem, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.AuditItem) (auditItem *model.AuditItem, err error)
	Update(db *gorm.DB, auditItem *model.AuditItem) (ac *model.AuditItem, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, auditItem model.AuditItem, updatedFields ...string) (ac *model.AuditItem, err error)
}
