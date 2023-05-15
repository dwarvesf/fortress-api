package icytransaction

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, model *model.IcyTransaction) error {
	return db.Create(model).Error
}
