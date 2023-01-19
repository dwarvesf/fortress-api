package projectslot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Delete delete ProjectSlot by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ?", id).Delete(&model.ProjectSlot{}).Error
}

// One get 1 one by id
func (s *store) One(db *gorm.DB, id string) (*model.ProjectSlot, error) {
	var slot *model.ProjectSlot
	return slot, db.Where("id = ?", id).Preload("Seniority", "deleted_at IS NULL").First(&slot).Error
}

// Create create new project slot
func (s *store) Create(db *gorm.DB, slot *model.ProjectSlot) error {
	return db.Create(&slot).Error
}

// UpdateSelectedFieldsByID just update selected fields by id
func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectSlot, updatedFields ...string) (*model.ProjectSlot, error) {
	slot := model.ProjectSlot{}
	return &slot, db.Model(&slot).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}

// UpdateSelectedFieldByProjectID just update selected field by projectID
func (s *store) UpdateSelectedFieldByProjectID(db *gorm.DB, projectID string, updateModel model.ProjectSlot, updatedField string) error {
	return db.Model(&model.ProjectSlot{}).
		Where("project_id = ?", projectID).
		Select(updatedField).
		Updates(updateModel).Error
}

func (s *store) GetPendingSlots(db *gorm.DB, projectID string, preload bool) ([]*model.ProjectSlot, error) {
	query := db.Where("project_id = ? AND status = ?", projectID, model.ProjectMemberStatusPending).Order("created_at DESC")

	if preload {
		query = query.Preload("Seniority", "deleted_at IS NULL").
			Preload("ProjectSlotPositions", "deleted_at IS NULL").
			Preload("ProjectSlotPositions.Position", "deleted_at IS NULL")
	}

	var slots []*model.ProjectSlot
	return slots, query.Find(&slots).Error
}
