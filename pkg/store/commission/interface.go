package commission

import (
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IStore is an interface that abstract database method for commission
type IStore interface {
	// Create new row for table user_commissions, save the commission
	// for an user for specific invoice
	Create(db *gorm.DB, us []model.EmployeeCommission) error

	Get(db *gorm.DB, q Query) ([]model.EmployeeCommission, error)
	MarkPaid(db *gorm.DB, ids model.UUID) error
}

type Query struct {
	EmployeeID string
	FromDate   *time.Time
	ToDate     *time.Time
	IsPaid     bool
}
