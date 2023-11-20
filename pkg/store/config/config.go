package config

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) One(db *gorm.DB, id string) (*model.Config, error) {
	var config *model.Config
	return config, db.Where("id = ?", id).
		First(&config).Error
}

func (s *store) OneByKey(db *gorm.DB, key string) (*model.Config, error) {
	var config *model.Config
	return config, db.Where("key = ?", key).
		First(&config).Error
}

func (s *store) Save(db *gorm.DB, config *model.Config) (err error) {
	return db.Save(&config).Error
}
