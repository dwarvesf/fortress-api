package service

import (
	"gorm.io/gorm"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	FindOperationByMonth(db *gorm.DB, month time.Month) ([]*model.OperationalService, error)
}
