package employeeinvitation

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, employeeOrganization *model.EmployeeInvitation) (*model.EmployeeInvitation, error)
	OneByEmployeeID(db *gorm.DB, employeeID string) (*model.EmployeeInvitation, error)
	Save(db *gorm.DB, employeeInvitation *model.EmployeeInvitation) error
}
