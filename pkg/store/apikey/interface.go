package apikey

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByClientID(db *gorm.DB, id string) (*model.APIKey, error)
	Create(db *gorm.DB, e *model.APIKey) (apiKey *model.APIKey, err error)
}
