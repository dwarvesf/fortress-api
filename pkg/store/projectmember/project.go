package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type store struct {
}

func New() IStore {
	return &store{}
}

// GetByProjectIDAndEmployeeID return a project member by projectID and employeeID
func (s *store) GetByProjectIDAndEmployeeID(db *gorm.DB, projectID string, employeeID string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("project_id = ? AND employee_id = ?", projectID, employeeID).Preload("Employee").First(&member).Error
}

// Create create new member
func (s *store) Create(db *gorm.DB, member *model.ProjectMember) error {
	return db.Create(&member).Preload("Employee").First(&member).Error
}

// Upsert create new member or update existing member
func (s *store) Upsert(db *gorm.DB, member *model.ProjectMember) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "project_slot_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"employee_id",
			"seniority_id",
			"deployment_type",
			"status",
			"joined_date",
			"left_date",
			"rate",
			"discount",
		}),
	}).
		Create(&member).
		Preload("Employee").
		First(&member).Error
}
