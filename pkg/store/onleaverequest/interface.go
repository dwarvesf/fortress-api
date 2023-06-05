package onleaverequest

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, request *model.OnLeaveRequest) (r *model.OnLeaveRequest, err error)
}
