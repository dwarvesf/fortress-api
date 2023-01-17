package projecthead

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create using for insert new data to project head
func (s *store) Create(db *gorm.DB, projectHead *model.ProjectHead) error {
	return db.Create(projectHead).Preload("Employee").First(projectHead).Error
}

// GetActiveLeadsByProjectID get active project heads by projectID
func (s *store) GetActiveLeadsByProjectID(db *gorm.DB, projectID string) ([]*model.ProjectHead, error) {
	var projectHeads []*model.ProjectHead
	return projectHeads, db.Where("project_id = ? AND (left_date IS NULL OR left_date > now()) AND deleted_at IS NULL", projectID).
		Preload("Employee").
		Find(&projectHeads).Error
}

func (s *store) DeleteByPositionInProject(db *gorm.DB, projectID string, employeeID string, position string) error {
	return db.Unscoped().Where("project_id = ? AND employee_id = ? AND position = ?", projectID, employeeID, position).Delete(&model.ProjectHead{}).Error
}

func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectHead, updatedFields ...string) (*model.ProjectHead, error) {
	head := model.ProjectHead{}
	return &head, db.Model(&head).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}

// One Get one head by project id and position
func (s *store) One(db *gorm.DB, projectID string, position model.HeadPosition) (*model.ProjectHead, error) {
	var projectHead *model.ProjectHead
	return projectHead, db.Where("project_id = ? AND position = ? AND left_date IS NULL", projectID, position).First(&projectHead).Error
}

func (s *store) UpdateLeftDateOfEmployee(db *gorm.DB, employeeID string, projectID string, position string, leftDate time.Time) (*model.ProjectHead, error) {
	head := model.ProjectHead{}
	return &head, db.
		Model(&head).
		Where("employee_id = ? AND project_id = ? AND position = ? AND left_date is NULL", employeeID, projectID, position).
		Update("left_date", leftDate).Error
}
