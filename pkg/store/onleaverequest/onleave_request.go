package onleaverequest

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

type GetOnLeaveInput struct {
	Date string
}

// Create creates an on-leave request record in the database
func (s *store) Create(db *gorm.DB, r *model.OnLeaveRequest) (request *model.OnLeaveRequest, err error) {
	return r, db.Create(r).Error
}

func (s *store) All(db *gorm.DB, input GetOnLeaveInput) ([]*model.OnLeaveRequest, error) {
	var chapters []*model.OnLeaveRequest
	query := db.
		Preload("Creator").
		Preload("Creator.SocialAccounts").
		Preload("Approver").
		Preload("Approver.SocialAccounts")

	if input.Date != "" {
		query = query.Where("start_date <= ? AND ? <= end_date", input.Date, input.Date)
	}

	return chapters, query.Find(&chapters).Error
}
