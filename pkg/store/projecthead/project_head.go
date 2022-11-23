package projecthead

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// GetByProjectID get all project heads by projectID
func (s *store) GetByProjectID(db *gorm.DB, projectID string) ([]*model.ProjectHead, error) {
	var projectHeads []*model.ProjectHead
	return projectHeads, db.Where("project_id = ? AND deleted_at IS NULL", projectID).Find(&projectHeads).Error
}

// Upsert create new member or update existing head
func (s *store) Upsert(db *gorm.DB, head *model.ProjectHead) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "project_id",
			},
			{
				Name: "employee_id",
			},
			{
				Name: "position",
			},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"commission_rate",
			"position",
			"joined_date",
		}),
	}).
		Create(&head).
		Preload("Employee").
		First(&head).Error
}

func (s *store) DeleteByProjectIDAndPosition(db *gorm.DB, projectID string, pos string) error {
	return db.Where("project_id = ? AND position = ?", projectID, pos).Delete(&model.ProjectHead{}).Error
}

func (s *store) DeleteByPositionInProject(db *gorm.DB, projectID string, employeeID string, position string) error {
	return db.Unscoped().Where("project_id = ? AND employee_id = ? AND position = ?", projectID, employeeID, position).Delete(&model.ProjectHead{}).Error
}

func (s *store) UpdateLeftDate(db *gorm.DB, projectID string, employeeID string, position string, timeNow time.Time) error {
	return db.Model(&model.ProjectHead{}).
		Where("project_id = ? AND employee_id = ? AND position = ?", projectID, employeeID, position).Update("left_date", timeNow).Error
}

// One Get one head by project id and position
func (s *store) One(db *gorm.DB, projectID string, position model.HeadPosition) (*model.ProjectHead, error) {
	var projectHead *model.ProjectHead
	return projectHead, db.Where("project_id = ? AND position = ? AND left_date IS NULL", projectID, position).First(&projectHead).Error
}
