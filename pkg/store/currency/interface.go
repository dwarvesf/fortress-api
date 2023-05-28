package currency

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByName(db *gorm.DB, name string) (*model.Currency, error)
	GetList(db *gorm.DB) ([]model.Currency, error)
}
