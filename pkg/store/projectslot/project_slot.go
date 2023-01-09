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

// All get all projects by query and pagination
func (s *store) All(db *gorm.DB, input GetListProjectSlotInput, pagination model.Pagination) ([]*model.ProjectSlot, int64, error) {
	query := db.Table("project_slots").Where("project_slots.deleted_at IS NULL")
	var total int64

	query = query.Where("project_slots.project_id = ?", input.ProjectID).
		Joins("LEFT JOIN project_members pm ON pm.project_slot_id = project_slots.id AND pm.project_id = ?", input.ProjectID)

	if input.Status == model.ProjectMemberStatusPending.String() {
		query = query.Where("project_slots.status = ?", input.Status)
	}

	if input.Status == model.ProjectMemberStatusActive.String() || input.Status == model.ProjectMemberStatusOnBoarding.String() {
		query = query.Where("project_slots.status = ? AND pm.status = ? ", input.Status, input.Status)
	}

	if input.Status == model.ProjectMemberStatusInactive.String() {
		query = query.Where("pm.status = ? ", input.Status)
	}

	query = query.Count(&total)

	if pagination.Sort != "" {
		query = query.Order(pagination.Sort)
	} else {
		query = query.Order("updated_at DESC")
	}

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	if input.Status == "active" {
		input.Preload = true
	}

	query = query.Offset(offset).
		Preload("ProjectMember.Employee", "deleted_at IS NULL")

	if input.Status != "" {
		query.Preload("ProjectMember", "deleted_at IS NULL AND status = ?", input.Status)
	} else {
		query.Preload("ProjectMember", "deleted_at IS NULL")
	}

	if input.Preload {
		query = query.Preload("Seniority", "deleted_at IS NULL").
			Preload("ProjectMember.Seniority", "deleted_at IS NULL").
			Preload("ProjectMember.ProjectMemberPositions", "deleted_at IS NULL").
			Preload("ProjectMember.ProjectMemberPositions.Position", "deleted_at IS NULL")
	}

	var slots []*model.ProjectSlot
	return slots, total, query.Find(&slots).Error
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
