package onleaverequest

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, r *model.OnLeaveRequest) (request *model.OnLeaveRequest, err error)
	All(db *gorm.DB, input GetOnLeaveInput) ([]*model.OnLeaveRequest, error)
	GetByNocodbID(db *gorm.DB, nocodbID int) (*model.OnLeaveRequest, error)
	GetByNotionPageID(db *gorm.DB, notionPageID string) (*model.OnLeaveRequest, error)
	Delete(db *gorm.DB, id string) error
}
