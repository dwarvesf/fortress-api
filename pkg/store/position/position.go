package position_store

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

// One get all positions
func (s *store) All() ([]*model.Position, error) {
	var positions []*model.Position
	return positions, s.db.Find(&positions).Error
}
