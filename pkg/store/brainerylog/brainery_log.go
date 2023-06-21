package brainerylog

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates a brainery log record in the database
func (s *store) Create(db *gorm.DB, b *model.BraineryLog) (braineryLog *model.BraineryLog, err error) {
	return b, db.Create(b).Error
}
