package projectmemberposition

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create create new project member position
func (s *store) Create(db *gorm.DB, pos *model.ProjectMemberPosition) error {
	return db.Create(&pos).Preload("Position").First(&pos).Error
}

// DeleteByProjectMemberID delete project_member_positions by project_member_id
func (s *store) DeleteByProjectMemberID(db *gorm.DB, memberID string) error {
	return db.Unscoped().Where("project_member_id = ?", memberID).Delete(&model.ProjectMemberPosition{}).Error
}
