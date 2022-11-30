package workunitstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, wus *model.WorkUnitStack) error
	DeleteByWorkUnitID(db *gorm.DB, workUnitID string) error
}
