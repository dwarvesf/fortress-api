package role

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (roles []*model.Role, err error)
	One(db *gorm.DB, id model.UUID) (role *model.Role, err error)
}
