package commission

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	GetProjectCommissionObjectByMemberID(db *gorm.DB, mid model.UUID) (*model.ProjectCommissionObject, error)
	CreateEmployeeCommissions(db *gorm.DB, comms []model.EmployeeCommission) error
	ListEmployeeCommissions(db *gorm.DB, employeeID model.UUID, isPaid bool) ([]model.EmployeeCommission, error)
	CloseEmployeeCommission(db *gorm.DB, id model.UUID) error
}
