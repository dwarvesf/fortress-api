package actionitem

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (actionItem *model.ActionItem, err error)
	All(db *gorm.DB) (actionItems []*model.ActionItem, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.ActionItem) (actionItem *model.ActionItem, err error)
	Update(db *gorm.DB, actionItem *model.ActionItem) (ac *model.ActionItem, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, actionItem model.ActionItem, updatedFields ...string) (ac *model.ActionItem, err error)
}
