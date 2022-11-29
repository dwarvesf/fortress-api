package workunitstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new WorkUnitStack
func (s *store) Create(db *gorm.DB, wus *model.WorkUnitStack) error {
	return db.Create(&wus).Error
}
