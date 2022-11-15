package position

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// All get all positions
func (s *store) All(db *gorm.DB) ([]*model.Position, error) {
	var positions []*model.Position
	return positions, db.Find(&positions).Error
}

// One get 1 one by id
func (s *store) One(db *gorm.DB, id model.UUID) (*model.Position, error) {
	var pos *model.Position
	return pos, db.Where("id = ?", id).First(&pos).Error
}
