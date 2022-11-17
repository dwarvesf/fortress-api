package projectslot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB, input GetListProjectSlotInput, pagination model.Pagination) ([]*model.ProjectSlot, int64, error)
	One(db *gorm.DB, id string) (*model.ProjectSlot, error)
	Update(db *gorm.DB, id string, slot *model.ProjectSlot) (*model.ProjectSlot, error)
	Create(db *gorm.DB, slot *model.ProjectSlot) error
	HardDelete(db *gorm.DB, id string) (err error)
}
