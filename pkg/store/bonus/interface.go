package bonus

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// Store is an interface that abstract database method for bonus
type IStore interface {
	GetByUserID(db *gorm.DB, id model.UUID) ([]model.EmployeeBonus, error)
}
