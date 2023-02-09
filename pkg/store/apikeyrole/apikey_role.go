package apikeyrole

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, e *model.ApikeyRole) (*model.ApikeyRole, error) {
	return e, db.Create(e).Error
}
