package projectslotposition

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetByProjectSlotID(db *gorm.DB, memberID string) ([]*model.ProjectSlotPosition, error) {
	var pos []*model.ProjectSlotPosition
	return pos, db.Where("project_slot_id = ?", memberID).Preload("Position").Find(&pos).Error
}

func (s *store) Create(db *gorm.DB, pos *model.ProjectSlotPosition) error {
	return db.Create(&pos).Preload("Position").First(&pos).Error
}

func (s *store) DeleteByProjectSlotID(db *gorm.DB, slotID string) error {
	return db.Unscoped().Where("project_slot_id = ?", slotID).Delete(&model.ProjectSlotPosition{}).Error
}
