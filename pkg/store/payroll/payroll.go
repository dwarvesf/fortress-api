package payroll

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetLatestCommitTime(db *gorm.DB) (*time.Time, error) {
	var dd time.Time
	return &dd, db.Table("payrolls").Where("is_paid IS TRUE").Select("max(due_date) as due_date").Row().Scan(&dd)
}

func (s *store) List(db *gorm.DB, q PayrollInput) ([]model.Payroll, error) {
	pr := []model.Payroll{}
	if !q.ID.IsZero() {
		db = db.Where("id = ?", q.ID)
	}
	if !q.EmployeeID.IsZero() {
		db = db.Where("employee_id = ?", q.EmployeeID)
	}
	if len(q.EmployeeIDs) > 0 {
		db = db.Where("employee_id IN (?)", q.EmployeeIDs)
	}
	if q.Year != 0 {
		db = db.Where("year = ?", q.Year)
	}
	if q.Month != 0 {
		db = db.Where("month = ?", int(q.Month))
	}
	if q.Day != 0 {
		db = db.Where("date_part('day', due_date) = ?", q.Day)
	}
	if q.IsNotCommit {
		db = db.Where("is_paid IS FALSE")
	}
	if q.GetLatest {
		db = db.Order("created_at desc")
	}
	return pr, db.
		Preload("Employee").
		Preload("Employee.EmployeeBaseSalary", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = true")
		}).
		Preload("Employee.EmployeeBaseSalary.Currency").
		Find(&pr).Error
}

func (s *store) Save(db *gorm.DB, pr []model.Payroll) error {
	if len(pr) == 0 {
		return nil
	}
	ids := []string{}
	for _, v := range pr {
		if !v.ID.IsZero() {
			ids = append(ids, v.ID.String())
		}
	}
	tx := db.Begin()
	if len(ids) != 0 {
		dsmt := "DELETE FROM payrolls WHERE id IN ?"
		if err := tx.Exec(dsmt, ids).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, v := range pr {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, v.EmployeeID)
		valueArgs = append(valueArgs, v.Total)
		valueArgs = append(valueArgs, v.AccountedAmount)
		valueArgs = append(valueArgs, v.DueDate)
		valueArgs = append(valueArgs, v.Month)
		valueArgs = append(valueArgs, v.Year)
		valueArgs = append(valueArgs, v.PersonalAmount)
		valueArgs = append(valueArgs, v.ContractAmount)
		valueArgs = append(valueArgs, v.Bonus)
		valueArgs = append(valueArgs, v.BonusExplain)
		valueArgs = append(valueArgs, v.Commission)
		valueArgs = append(valueArgs, v.CommissionExplain)
		valueArgs = append(valueArgs, v.IsPaid)
		valueArgs = append(valueArgs, v.WiseAmount)
		valueArgs = append(valueArgs, v.WiseRate)
		valueArgs = append(valueArgs, v.WiseFee)
		valueArgs = append(valueArgs, v.Notes)
	}

	smt := `INSERT INTO payrolls(employee_id,total,accounted_amount,due_date,month,year,personal_amount,contract_amount,bonus,bonus_explain,commission,commission_explain,is_paid,wise_amount,wise_rate,wise_fee,notes)
		VALUES %s`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	if err := tx.Exec(smt, valueArgs...).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *store) ListDashboardPayrolls(
	db *gorm.DB,
	q PayrollDashboardInput,
) ([]model.Payroll, error) {
	records := []model.Payroll{}
	db = db.Order("created_at DESC").Where("is_paid = TRUE")
	if q.Date != "" {
		db = db.Where("created_at >= ?::date AND created_at < (?::date + '1 month'::interval)", q.Date, q.Date)
	}
	return records, db.Debug().
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			db = db.Preload("EmployeeBaseSalary", func(db *gorm.DB) *gorm.DB {
				if len(q.Paydays) > 0 {
					db = db.Where("payroll_batch IN (?)", q.Paydays)
				}
				return db.Where("is_active = TRUE")
			})
			db = db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
				if len(q.Departments) > 0 {
					db = db.Where("department IN (?)", q.Departments)
				}
				return db.Where("is_active = TRUE")
			})
			return db
		}).
		Preload("Employee.EmployeeBaseSalary", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = true")
		}).
		Preload("Employee.EmployeeBaseSalary.Currency").
		Find(&records).Error
}
