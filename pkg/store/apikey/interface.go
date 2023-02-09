package apikey

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByClientID(db *gorm.DB, id string) (*model.Apikey, error)
	Create(db *gorm.DB, e *model.Apikey) (apiKey *model.Apikey, err error)
}
