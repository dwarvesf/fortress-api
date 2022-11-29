package workunit

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new WorkUnit
func (s *store) Create(db *gorm.DB, workUnit *model.WorkUnit) error {
	return db.Create(&workUnit).Error
}
