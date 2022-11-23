package projectstack

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// TODO: rename filename

type store struct {
}

func New() IStore {
	return &store{}
}

// Create create new one by id
func (s *store) Create(db *gorm.DB, projectStack *model.ProjectStack) (*model.ProjectStack, error) {
	return projectStack, db.Create(&projectStack).Error
}

// HardDelete hard delete all by project id
func (s *store) HardDelete(db *gorm.DB, projectID string) error {
	return db.Table("project_stacks").Unscoped().Where("project_id = ?", projectID).Delete(&model.EmployeeStack{}).Error
}
