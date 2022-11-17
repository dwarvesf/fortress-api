package projectslotposition

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Create(db *gorm.DB, pos *model.ProjectSlotPosition) error
	HardDeleteByProjectSlotID(db *gorm.DB, id string) error
}
