package workunitmember

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new WorkUnitMember
func (s *store) Create(db *gorm.DB, wum *model.WorkUnitMember) error {
	return db.Create(&wum).Error
}
