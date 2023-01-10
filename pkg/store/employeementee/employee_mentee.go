package employeementee

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// OneByMenteeID get 1 employee mentee by mentee
func (s *store) OneByMenteeID(db *gorm.DB, menteeID string, preload bool) (*model.EmployeeMentee, error) {
	query := db.Where("mentee_id = ?", menteeID)

	if preload {
		query = query.Preload("Mentee", "deleted_at IS NULL and left_date IS NULL")
	}

	var employeeMentee *model.EmployeeMentee
	return employeeMentee, query.First(&employeeMentee).Error
}

// Delete delete 1 employee mentee by id
func (s *store) Delete(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.EmployeeMentee{}).Error
}

// DeleteByMentorIDAndMenteeID delete 1 employee mentee by menber id and mentee id
func (s *store) DeleteByMentorIDAndMenteeID(db *gorm.DB, mentorID string, menteeID string) error {
	return db.Where("mentor_id = ? AND mentee_id = ?", mentorID, menteeID).Delete(&model.EmployeeMentee{}).Error
}

// Create creates a new employee mentee
func (s *store) Create(db *gorm.DB, e *model.EmployeeMentee) (employee *model.EmployeeMentee, err error) {
	return e, db.Create(e).Error
}

// IsExiest check the existence of a employee mentee by mentorID and menteeID
func (s *store) IsExist(db *gorm.DB, mentorID string, menteeID string) (bool, error) {
	type res struct {
		Result bool
	}

	result := res{}
	query := db.Raw("SELECT EXISTS (SELECT * FROM employee_mentees WHERE mentor_ID = ? AND mentee_ID = ? AND deleted_at IS NULL) as result", mentorID, menteeID)

	return result.Result, query.Scan(&result).Error
}
