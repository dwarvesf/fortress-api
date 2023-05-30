package expense

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Create(db *gorm.DB, e *model.Expense) (*model.Expense, error)
	Delete(db *gorm.DB, e *model.Expense) (*model.Expense, error)
	Update(db *gorm.DB, e *model.Expense) (*model.Expense, error)
	GetValuation(db *gorm.DB, y int) (*model.CurrencyView, error)
	GetByQuery(db *gorm.DB, q *ExpenseQuery) (*model.Expense, error)
}

type ExpenseQuery struct {
	BasecampID int
}
