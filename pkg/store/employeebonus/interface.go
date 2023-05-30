package employeebonus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore is an interface that abstract database method for bonus
type IStore interface {
	GetByUserID(db *gorm.DB, id model.UUID) ([]model.EmployeeBonus, error)
}
