package config

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	One(db *gorm.DB, id string) (config *model.Config, err error)
	OneByKey(db *gorm.DB, key string) (config *model.Config, err error)
	Save(db *gorm.DB, salaryAdvance *model.Config) (err error)
}
