package recruitment

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore --
type IStore interface {
	Save(db *gorm.DB, cv *model.Candidate) error
	Update(db *gorm.DB, cv *model.Candidate) error
	GetByBasecampID(db *gorm.DB, bcID int) (*model.Candidate, error)
	GetApproachCandidateByBasecampID(db *gorm.DB, bcID int) (*model.Candidate, error)
	GetByDuration(db *gorm.DB, from, to time.Time) ([]model.Candidate, error)
	GetAll(db *gorm.DB) ([]model.Candidate, error)
	GetOffered(db *gorm.DB, batchDate, dueDate time.Time) ([]model.Candidate, error)
}
