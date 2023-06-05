package onleaverequest

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

// Create creates an onleave request record in the database
func (s store) Create(db *gorm.DB, r *model.OnLeaveRequest) (request *model.OnLeaveRequest, err error) {
	return r, db.Create(r).Error
}
