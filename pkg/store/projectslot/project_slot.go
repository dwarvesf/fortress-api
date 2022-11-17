package projectslot

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// All get all projects with filter and pagination
func (s *store) All(db *gorm.DB, input GetListProjectSlotInput, pagination model.Pagination) ([]*model.ProjectSlot, int64, error) {
	query := db.Debug().Table("project_slots").Where("project_slots.deleted_at IS NULL")
	var total int64

	query = query.Where("project_slots.project_id = ?", input.ProjectID).
		Joins("LEFT JOIN project_members pm ON pm.project_slot_id = project_slots.id AND pm.project_id = ?", input.ProjectID)

	switch input.Status {
	case model.ProjectMemberStatusPending.String():
		query = query.Where("project_slots.status = ? AND pm.id IS NULL", input.Status)

	case model.ProjectMemberStatusActive.String(),
		model.ProjectMemberStatusOnBoarding.String(),
		model.ProjectMemberStatusInactive.String():
		query = query.Where("project_slots.status = ? AND pm.id IS NOT NULL", input.Status)
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

	query = query.Preload("ProjectMember", "deleted_at IS NULL").
		Preload("ProjectMember.Seniority", "deleted_at IS NULL").
		Preload("ProjectMember.Employee", "deleted_at IS NULL").
		Preload("ProjectMember.Employee.EmployeePositions", "deleted_at IS NULL").
		Preload("ProjectMember.Employee.EmployeePositions.Position", "deleted_at IS NULL").
		Offset(offset)

	var slots []*model.ProjectSlot
	return slots, total, query.Find(&slots).Error
}

// One get 1 one by id
func (s *store) One(db *gorm.DB, id string) (*model.ProjectSlot, error) {
	var slot *model.ProjectSlot
	return slot, db.Where("id = ?", id).First(&slot).Error
}

// Update update existing slot
func (s *store) Update(db *gorm.DB, id string, slot *model.ProjectSlot) (*model.ProjectSlot, error) {
	return slot, db.Table("project_slots").Where("id = ?", id).Updates(slot).Error
}

// Create create new project slot
func (s *store) Create(db *gorm.DB, slot *model.ProjectSlot) error {
	return db.Create(&slot).Error
}
