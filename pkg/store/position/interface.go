package position

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (positions []*model.Position, err error)
	One(db *gorm.DB, id model.UUID) (position *model.Position, err error)
}
