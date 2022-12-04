package employeechapter

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeeChapter *model.EmployeeChapter) (*model.EmployeeChapter, error)
	DeleteByEmployeeID(db *gorm.DB, employeeID string) error
}
