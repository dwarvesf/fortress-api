package client

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) OneByID(db *gorm.DB, id string) (*model.Client, error) {
	var client *model.Client
	return client, db.Where("id = ?", id).Preload("Contacts").First(&client).Error
}
