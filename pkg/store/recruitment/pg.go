package recruitment

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

// New create new pg service
func New() IStore {
	return &store{}
}

func (s *store) Save(db *gorm.DB, cv *model.Candidate) error {
	return db.Save(&cv).Error
}

func (s *store) GetByBasecampID(db *gorm.DB, bcID int) (*model.Candidate, error) {
	var c model.Candidate
	return &c, db.Where("basecamp_todo_id = ?", bcID).First(&c).Error
}

func (s *store) GetApproachCandidateByBasecampID(db *gorm.DB, bcID int) (*model.Candidate, error) {
	var c model.Candidate
	return &c, db.Where("basecamp_todo_id = ? and status = ?", bcID, model.ApproachCandidateStatus).First(&c).Error
}

func (s *store) Update(db *gorm.DB, candidate *model.Candidate) error {
	return db.Save(&candidate).Error
}

func (s *store) GetByDuration(db *gorm.DB, from, to time.Time) ([]model.Candidate, error) {
	var c []model.Candidate
	return c, db.Where("created_at > ? AND created_at < ?", from, to).Find(&c).Error
}

func (s *store) GetAll(db *gorm.DB) ([]model.Candidate, error) {
	var c []model.Candidate
	return c, db.Find(&c).Error
}

func (s *store) GetOffered(db *gorm.DB, batchDate, dueDate time.Time) ([]model.Candidate, error) {
	var c []model.Candidate
	return c, db.Where("offer_start_date > ? AND offer_start_date < ? AND status = ?", batchDate, dueDate, model.HiredCandidateStatus).Find(&c).Error
}
