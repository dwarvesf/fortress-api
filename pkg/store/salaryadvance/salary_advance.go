package salaryadvance

import (
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) One(db *gorm.DB, id string) (*model.SalaryAdvance, error) {
	var salaryAdvance *model.SalaryAdvance
	return salaryAdvance, db.Where("id = ?", id).
		First(&salaryAdvance).Error
}

func (s *store) ListNotPayBackByEmployeeID(db *gorm.DB, employeeID string) ([]model.SalaryAdvance, error) {
	var advanceSalaries []model.SalaryAdvance
	return advanceSalaries, db.Where("employee_id = ?", employeeID).Where("is_paid_back = ?", false).Find(&advanceSalaries).Error
}

func (s *store) Save(db *gorm.DB, salaryAdvance *model.SalaryAdvance) (err error) {
	return db.Save(&salaryAdvance).Error
}

func (s *store) ListAggregatedSalaryAdvance(db *gorm.DB, idPaid *bool, paging model.Pagination, order model.SortOrder) (report []model.AggregatedSalaryAdvance, err error) {
	query := db.
		Table("salary_advance_histories").
		Select("employee_id, sum(amount_icy) as amount_icy, sum(amount_usd) as amount_usd")

	if idPaid != nil {
		query = query.Where("is_paid_back = ?", idPaid)
	}

	if paging.Sort != "" {
		query = query.Order(fmt.Sprintf("%s %s", paging.Sort, order))
	}

	limit, offset := paging.ToLimitOffset()
	if limit != 0 {
		query = query.Limit(limit).Offset(offset)
	}

	return report, query.Group("employee_id").Find(&report).Error
}

func (s *store) TotalAggregatedSalaryAdvance(db *gorm.DB, idPaid *bool) (int64, int64, float64, error) {
	queryCurrency := db.
		Table("salary_advance_histories").
		Select("sum(amount_icy) as total_icy, sum(amount_usd) as total_usd")

	queryCount := db.
		Table("salary_advance_histories")

	if idPaid != nil {
		queryCurrency = queryCurrency.Where("is_paid_back = ?", idPaid)
		queryCount = queryCount.Where("is_paid_back = ?", idPaid)
	}

	var (
		count   int64
		nullIcy sql.NullInt64
		nullUsd sql.NullFloat64
	)

	if err := queryCurrency.Row().Scan(&nullIcy, &nullUsd); err != nil {
		return 0, 0, 0, err
	}

	if err := queryCount.Group("employee_id").Count(&count).Error; err != nil {
		return 0, 0, 0, err
	}

	return count, nullIcy.Int64, nullUsd.Float64, nil
}
