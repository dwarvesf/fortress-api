package employeeeventtopic

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetByEmployeeIDWithPagination(db *gorm.DB, employeeID string, input GetByEmployeeIDInput, pagination model.Pagination) (eTopics []*model.EmployeeEventTopic, total int64, err error)
}
