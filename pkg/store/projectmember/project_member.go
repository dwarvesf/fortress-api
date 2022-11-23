package projectmember

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Delete delete ProjectMember by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Unscoped().Where("id = ?", id).Delete(&model.ProjectMember{}).Error
}

// IsExist check ProjectMember existance
func (s *store) IsExist(db *gorm.DB, id string) (bool, error) {
	var record struct {
		Result bool
	}

	query := db.Raw("SELECT EXISTS (SELECT * FROM project_members WHERE id = ?) as result", id)
	return record.Result, query.Scan(&record).Error
}

// One return a project member by projectID and employeeID
func (s *store) One(db *gorm.DB, projectID string, employeeID string, status string) (*model.ProjectMember, error) {
	var member *model.ProjectMember
	return member, db.Where("project_id = ? AND employee_id = ? AND status = ?", projectID, employeeID, status).Preload("Employee").First(&member).Error
}

// Create using for create new member
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
