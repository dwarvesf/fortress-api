package position

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (positions []*model.Position, err error)
	One(db *gorm.DB, id model.UUID) (position *model.Position, err error)
	Update(db *gorm.DB, position *model.Position) (p *model.Position, err error)
	Create(db *gorm.DB, position *model.Position) (p *model.Position, err error)
	Delete(db *gorm.DB, id string) (err error)
}
