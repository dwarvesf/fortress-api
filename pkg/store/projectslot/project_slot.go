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
	var total int64

	query := db.Table("project_slots").
		Where("project_slots.deleted_at IS NULL").
		Where("project_slots.project_id = ?", input.ProjectID)

	switch input.Status {
	case model.ProjectMemberStatusPending.String():
		query = query.Where("project_slots.status = ?", input.Status)

	case model.ProjectMemberStatusActive.String(),
		model.ProjectMemberStatusOnBoarding.String(),
		model.ProjectMemberStatusInactive.String():
		query = query.Joins("LEFT JOIN project_members pm ON pm.project_slot_id = project_slots.id").
			Where("pm.deleted_at IS NULL").
			Where("pm.status = ? ", input.Status)
	}

	query = query.Count(&total)

	if pagination.Sort != "" {
		query = query.Order(pagination.Sort)
	} else {
		query = query.Order("created_at DESC")
	}

	limit, offset := pagination.ToLimitOffset()
	if pagination.Page > 0 {
		query = query.Limit(limit)
	}

	query = query.Offset(offset)

	if input.Preload {
		query = query.Preload("Seniority", "deleted_at IS NULL").
			Preload("ProjectMember", "deleted_at IS NULL AND status = ?", input.Status).
			Preload("ProjectMember.Employee", "deleted_at IS NULL").
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

func (s *store) GetAssignedSlots(db *gorm.DB, projectID string, preload bool) ([]*model.ProjectSlot, error) {
	query := db.Joins("JOIN project_members pm ON pm.project_slot_id = project_slots.id").
		Joins("LEFT JOIN seniorities s ON pm.seniority_id = s.id").
		Joins(`LEFT JOIN project_heads ph ON pm.status = ?
			AND pm.project_id = ph.project_id 
			AND pm.employee_id = ph.employee_id 
			AND ph.deleted_at IS NULL
			AND (ph.left_date IS NULL OR ph.left_date > now())
			AND ph.position = ?
		`, model.ProjectMemberStatusActive, model.HeadPositionTechnicalLead).
		Where("pm.deleted_at IS NULL AND project_slots.deleted_at IS NULL AND project_slots.project_id = ?", projectID).
		Order("pm.left_date DESC, ph.created_at, s.level DESC").
		Preload("ProjectMember", "deleted_at IS NULL").
		Preload("ProjectMember.Employee", "deleted_at IS NULL")

	if preload {
		query = query.Preload("ProjectMember.Seniority", "deleted_at IS NULL").
			Preload("ProjectMember.ProjectMemberPositions", "deleted_at IS NULL").
			Preload("ProjectMember.ProjectMemberPositions.Position", "deleted_at IS NULL")

	}

	var slots []*model.ProjectSlot
	return slots, query.Find(&slots).Error
}
