package payroll

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

// New create new pg service
func New() IStore {
	return &store{}
}

func (s *store) GetSalary(db *gorm.DB, year int, month int) ([]model.Employee, error) {
	var employees []model.Employee

	return employees, db.Table("employee").
		Where("status = <> ?", model.WorkingStatusLeft).
		Preload("Expense", func(db *gorm.DB) *gorm.DB {
			return db.Where("date_part('year', issued_date) = ? AND date_part('month', issued_date) = ?", year, month)
		}).
		Group("employee.id").
		Find(&employees).
		Error
}

func (s *store) Create(db *gorm.DB, p *model.Payroll) error {
	return db.Create(p).Error
}

func (s *store) Get(db *gorm.DB, employeeId string, month, year int) (*model.Payroll, error) {
	res := &model.Payroll{}
	return res, db.Where("employee_id = ? AND year = ? AND month = ?", employeeId, year, month).First(&res).Error
}

func (s *store) GetList(db *gorm.DB, q GetListPayrollInput) ([]model.Payroll, error) {
	var res []model.Payroll
	payrollQuery := db.
		Preload("Employee", func(db *gorm.DB) *gorm.DB {
			return db.Order("display_name asc")
		}).
		Preload("Employee.BaseSalary.Currency")

	if q.ID != "" {
		payrollQuery = payrollQuery.Where("payrolls.id = ?", q.ID)
	}
	if q.UserID != "" {
		payrollQuery = payrollQuery.Where("payrolls.employee_id = ?", q.UserID)
	}
	if q.Month != 0 {
		payrollQuery = payrollQuery.Where("payrolls.month = ?", q.Month)
	}
	if q.Year != 0 {
		payrollQuery = payrollQuery.Where("payrolls.year = ?", q.Year)
	}
	if q.Day != 0 {
		payrollQuery = payrollQuery.Where("date_part('day', due_date) = ?", q.Day)
	}

	err := payrollQuery.
		Order(`
		year desc,
		month desc`).
		Find(&res).
		Error
	if err != nil {
		return nil, err
	}

	for i, v := range res {
		var baseSalary model.BaseSalary
		if err := db.
			Preload("Currency").
			Where("employee_id = ?", v.EmployeeID).
			Order("effective_date desc").
			First(&baseSalary).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		res[i].Employee.BaseSalary = baseSalary
	}

	return res, nil
}

func (s *store) UpdateSpecificFields(db *gorm.DB, id string, fields map[string]interface{}) error {
	return db.Model(&model.Payroll{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (s *store) GetLatestCommitTime(db *gorm.DB) (time.Time, error) {
	var DueDate time.Time
	return DueDate, db.Table("payrolls").Select("max(due_date) as due_date").Row().Scan(&DueDate)
}

func (s *store) InsertList(db *gorm.DB, payrolls []model.Payroll) error {
	if len(payrolls) <= 0 {
		return fmt.Errorf("payrolls cannot be empty")
	}

	var valueStrings []string
	var valueArgs []interface{}

	for _, payroll := range payrolls {
		valueStrings = append(valueStrings, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")

		valueArgs = append(valueArgs, model.NewUUID())
		valueArgs = append(valueArgs, payroll.EmployeeID)
		valueArgs = append(valueArgs, payroll.Total)
		valueArgs = append(valueArgs, payroll.Month)
		valueArgs = append(valueArgs, payroll.Year)
		valueArgs = append(valueArgs, payroll.CommissionAmount)
		valueArgs = append(valueArgs, payroll.CommissionExplain)
		valueArgs = append(valueArgs, payroll.UserRankSnapshot)
		valueArgs = append(valueArgs, payroll.TotalExplain)
		valueArgs = append(valueArgs, payroll.ProjectBonusAmount)
		valueArgs = append(valueArgs, payroll.DueDate)
		valueArgs = append(valueArgs, payroll.ProjectBonusExplain)
		valueArgs = append(valueArgs, payroll.IsPaid)
		valueArgs = append(valueArgs, payroll.ConversionAmount)
		valueArgs = append(valueArgs, payroll.BaseSalaryAmount)
		valueArgs = append(valueArgs, payroll.ContractAmount)
		valueArgs = append(valueArgs, payroll.SalaryAdvanceAmount)
	}

	smt := `INSERT INTO payrolls(id, employee_id, total, month, year, commission_amount, commission_explain, employee_rank_snapshot, total_explain, project_bonus_amount, due_date, project_bonus_explain, is_paid, conversion_amount, base_salary_amount, contract_amount, salary_advance_amount) VALUES %s`
	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	tx := db.Begin()
	err := tx.Exec(smt, valueArgs...).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
