package expense

import (
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct {
}

func New() IStore {
	return &store{}
}

func (s *store) Create(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Create(&e).Error
}

func (s *store) Delete(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Delete(&e).Error
}

func (s *store) Update(db *gorm.DB, e *model.Expense) (*model.Expense, error) {
	return e, db.Model(&model.Expense{}).Where("id = ?", e.ID).Updates(&e).First(&e).Error
}

func (s *store) GetValuation(db *gorm.DB, y int) (*model.CurrencyView, error) {
	res := &model.CurrencyView{}
	return res, db.Raw("select * from vw_expense where year = ?", y).Find(&res).Error
}

func (s *store) GetByQuery(db *gorm.DB, q *ExpenseQuery) (*model.Expense, error) {
	e := &model.Expense{}
	if q.BasecampID != 0 {
		db = db.Where("basecamp_id = ?", q.BasecampID)
	}
	if q.TaskProvider != "" {
		db = db.Where("task_provider = ?", q.TaskProvider)
	}
	if q.TaskRef != "" {
		db = db.Where("task_ref = ?", q.TaskRef)
	}

	return e, db.First(e).Error
}
