package auditactionitem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (auditActionItems []*model.AuditActionItem, err error)
	AllByAuditID(db *gorm.DB, auditID string) (auditActionItems []*model.AuditActionItem, err error)
	Delete(db *gorm.DB, id string) (err error)
	DeleteByAuditID(db *gorm.DB, auditID string) (err error)
	Create(db *gorm.DB, e *model.AuditActionItem) (auditActionItem *model.AuditActionItem, err error)
	Update(db *gorm.DB, auditActionItem *model.AuditActionItem) (ac *model.AuditActionItem, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, auditActionItem model.AuditActionItem, updatedFields ...string) (ac *model.AuditActionItem, err error)
}
