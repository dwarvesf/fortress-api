package audit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (audit []*model.Audit, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.Audit) (audit *model.Audit, err error)
	Update(db *gorm.DB, audit *model.Audit) (a *model.Audit, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, audit model.Audit, updatedFields ...string) (a *model.Audit, err error)
}
