package employeeorganization

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeeOrganization *model.EmployeeOrganization) (*model.EmployeeOrganization, error)
	DeleteByEmployeeID(db *gorm.DB, employeeID string) (err error)
}
