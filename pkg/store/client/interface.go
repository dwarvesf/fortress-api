package client

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	OneByID(db *gorm.DB, id string) (*model.Client, error)
}
