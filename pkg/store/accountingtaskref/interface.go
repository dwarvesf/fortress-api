package accountingtaskref

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"gorm.io/gorm"
)

type IStore interface {
	Create(db *gorm.DB, ref *model.AccountingTaskRef) error
	FindByProjectMonthYear(db *gorm.DB, projectID string, month, year int, group string) ([]*model.AccountingTaskRef, error)
}
