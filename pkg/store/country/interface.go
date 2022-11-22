package country

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	All(db *gorm.DB) (countries []*model.Country, err error)
	One(db *gorm.DB, id string) (countries *model.Country, err error)
	Exists(db *gorm.DB, id string) (bool, error)
}
