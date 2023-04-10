package service

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	FindOperationByMonth(db *gorm.DB, month time.Month) ([]*model.OperationalService, error)
}
