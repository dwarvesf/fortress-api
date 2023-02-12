package client

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (client *model.Client, err error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
	All(db *gorm.DB) (clients []*model.Client, err error)
	Delete(db *gorm.DB, id string) (err error)
	Create(db *gorm.DB, e *model.Client) (client *model.Client, err error)
	Update(db *gorm.DB, client *model.Client) (a *model.Client, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, client model.Client, updatedFields ...string) (a *model.Client, err error)
}
