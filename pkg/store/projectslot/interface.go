package projectslot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetPendingSlots(db *gorm.DB, projectID string, preload bool) ([]*model.ProjectSlot, error)
	One(db *gorm.DB, id string) (*model.ProjectSlot, error)
	Create(db *gorm.DB, slot *model.ProjectSlot) error
	Delete(db *gorm.DB, id string) (err error)

	UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectSlot, updatedFields ...string) (*model.ProjectSlot, error)
	UpdateSelectedFieldByProjectID(db *gorm.DB, projectID string, updateModel model.ProjectSlot, updatedField string) error
}
