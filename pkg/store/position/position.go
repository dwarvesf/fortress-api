package position

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

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

// Update update the position
func (s *store) Update(db *gorm.DB, position *model.Position) (*model.Position, error) {
	return position, db.Model(&model.Position{}).Where("id = ?", position.ID).Updates(&position).First(&position).Error
}

// Delete delete ProjectMember by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Position{}).Error
}

// Create create new position
func (s *store) Create(db *gorm.DB, position *model.Position) (*model.Position, error) {
	return position, db.Create(position).Error
}
