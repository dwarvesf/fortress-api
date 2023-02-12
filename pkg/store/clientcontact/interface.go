package clientcontact

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (clientContact *model.ClientContact, err error)
	IsExist(db *gorm.DB, id string) (exists bool, err error)
	All(db *gorm.DB) (clientContacts []*model.ClientContact, err error)
	Delete(db *gorm.DB, id string) (err error)
	DeleteByClientID(db *gorm.DB, clientID string) (err error)
	Create(db *gorm.DB, e *model.ClientContact) (clientContact *model.ClientContact, err error)
	Update(db *gorm.DB, clientContact *model.ClientContact) (a *model.ClientContact, err error)
	UpdateSelectedFieldsByID(db *gorm.DB, id string, clientContact model.ClientContact, updatedFields ...string) (a *model.ClientContact, err error)
}
