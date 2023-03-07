package expense

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type pgService struct {
}

func New() IStore {
	return &pgService{}
}

func (s *pgService) Create(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Create(&e).Error
}

func (s *pgService) Delete(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Delete(&e).Error
}

func (s *pgService) Update(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Model(&model.Expense{}).Updates(&e).Error
}

func (s *pgService) GetValuation(db *gorm.DB, y int) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw("select * from vw_expense where year = ?", y).Find(&res).Error
}

func (s *pgService) GetByQuery(db *gorm.DB, q *ExpenseQuery) (*model.Expense, error) {
	e := &model.Expense{}
	if q.BasecampID != 0 {
		db = db.Where("basecamp_id = ?", q.BasecampID)
	}

	return e, db.First(e).Error
}
