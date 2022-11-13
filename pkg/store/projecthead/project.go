package projecthead

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) IStore {
	return &store{
		db: db,
	}
}

// Create create new project
func (s *store) Create(projectHead *model.ProjectHead) error {
	return s.db.Create(projectHead).Preload("Employee").First(projectHead).Error
}
