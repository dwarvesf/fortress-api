package position

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

// All get all positions
func (s *store) All() ([]*model.Position, error) {
	var positions []*model.Position
	return positions, s.db.Find(&positions).Error
}

// One get 1 one by id
func (s *store) One(id model.UUID) (*model.Position, error) {
	var pos *model.Position
	return pos, s.db.Where("id = ?", id).First(&pos).Error
}
