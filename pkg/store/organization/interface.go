package organization

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (organizaiton *model.Organization, err error)
	All(db *gorm.DB) ([]*model.Organization, error)
}
