package apikey

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByClientID(db *gorm.DB, id string) (*model.Apikey, error) {
	var apikey *model.Apikey
	return apikey, db.Where("client_id = ?", id).
		First(&apikey).Error
}

func (s *store) Create(db *gorm.DB, e *model.Apikey) (apiKey *model.Apikey, err error) {
	return e, db.Create(e).Error
}
