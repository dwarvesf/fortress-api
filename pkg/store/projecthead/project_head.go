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

// BatchCreate create multiple project heads in one transaction
func (s *store) BatchCreate(db *gorm.DB, heads []*model.ProjectHead) ([]*model.ProjectHead, error) {
	return heads, db.Create(&heads).Error
}

// GetActiveLeadsByProjectID get active project heads by projectID
func (s *store) GetActiveLeadsByProjectID(db *gorm.DB, projectID string) ([]*model.ProjectHead, error) {
	var projectHeads []*model.ProjectHead

	now := time.Now()
	return projectHeads, db.Where("project_id = ? AND (end_date IS NULL OR end_date > ?)", projectID, now).
		Order("position").
		Preload("Employee").
		Find(&projectHeads).Error
}

func (s *store) DeleteByPositionInProject(db *gorm.DB, projectID string, employeeID string, position string) error {
	return db.Unscoped().Where("project_id = ? AND employee_id = ? AND position = ?", projectID, employeeID, position).Delete(&model.ProjectHead{}).Error
}

func (s *store) DeleteByID(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ?", id).Delete(&model.ProjectHead{}).Error
}

func (s *store) UpdateSelectedFieldsByID(db *gorm.DB, id string, updateModel model.ProjectHead, updatedFields ...string) (*model.ProjectHead, error) {
	head := model.ProjectHead{}
	return &head, db.Model(&head).Where("id = ?", id).Select(updatedFields).Updates(&updateModel).Error
}

// One Get one head by project id and position
func (s *store) One(db *gorm.DB, projectID string, employeeID string, position model.HeadPosition) (*model.ProjectHead, error) {
	var projectHead *model.ProjectHead
	return projectHead, db.
		Where("project_id = ?", projectID).
		Where("employee_id = ?", employeeID).
		Where("position = ?", position).
		First(&projectHead).Error
}

func (s *store) UpdateDateOfEmployee(db *gorm.DB, employeeID string, projectID string, position string, startDate *time.Time, endDate *time.Time) (*model.ProjectHead, error) {
	head := model.ProjectHead{}
	return &head, db.
		Model(&head).
		Where("employee_id = ? AND project_id = ? AND position = ?", employeeID, projectID, position).
		Updates(map[string]interface{}{
			"start_date": startDate,
			"end_date":   endDate,
		}).Error
}

func (s *store) GetByProjectIDAndPosition(db *gorm.DB, projectID string, position model.HeadPosition) ([]*model.ProjectHead, error) {
	var heads []*model.ProjectHead
	return heads, db.Where("project_id = ? AND position = ?", projectID, position).Find(&heads).Error
}
