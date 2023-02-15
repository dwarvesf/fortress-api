package apikey

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByClientID(db *gorm.DB, id string) (*model.APIKey, error) {
	var apikey *model.APIKey
	return apikey, db.Where("client_id = ?", id).
		First(&apikey).Error
}

func (s *store) Create(db *gorm.DB, e *model.APIKey) (apiKey *model.APIKey, err error) {
	return e, db.Create(e).Error
}
