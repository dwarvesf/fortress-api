package organization

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (organization *model.Organization, err error)
	OneByCode(db *gorm.DB, code string) (organization *model.Organization, err error)
	All(db *gorm.DB) ([]*model.Organization, error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
}
