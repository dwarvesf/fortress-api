package apikeyrole

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, e *model.ApikeyRole) (*model.ApikeyRole, error)
}
