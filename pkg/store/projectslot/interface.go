package projectslot

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	All(db *gorm.DB, input GetListProjectSlotInput, pagination model.Pagination) ([]*model.ProjectSlot, int64, error)
}
