package commission

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type store struct{}

func New() IStore {
	return &store{}
}

func (s *store) GetProjectCommissionObjectByMemberID(db *gorm.DB, mid model.UUID) (*model.ProjectCommissionObject, error) {
	obj := model.ProjectCommissionObject{}
	return &obj, db.Where("project_member_id = ?", mid).
		Preload("ProjectCommission").
		Preload("ProjectCommissionReceiver").First(&obj).Error
}

func (s *store) CreateEmployeeCommissions(db *gorm.DB, comms []model.EmployeeCommission) error {
	if len(comms) == 0 {
		return nil
	}

	valueStrings := []string{}
	valueArgs := []interface{}{}

	for _, v := range comms {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, v.ProjectID)
		valueArgs = append(valueArgs, v.ProjectName)
		valueArgs = append(valueArgs, v.InvoiceID)
		valueArgs = append(valueArgs, v.InvoiceItemID)
		valueArgs = append(valueArgs, v.EmployeeID)
		valueArgs = append(valueArgs, v.ProjectCommissionObjectID)
		valueArgs = append(valueArgs, v.ProjectCommissionReceiverID)
		valueArgs = append(valueArgs, v.Percentage)
		valueArgs = append(valueArgs, v.Amount)
		valueArgs = append(valueArgs, v.ConversionRate)
		valueArgs = append(valueArgs, v.IsPaid)
		valueArgs = append(valueArgs, v.Formula)
		valueArgs = append(valueArgs, v.Note)
		valueArgs = append(valueArgs, v.PaidAt)
	}

	smt := `INSERT INTO employee_commissions(project_id,project_name,invoice_id,invoice_item_id,employee_id,project_commission_object_id,project_commission_receiver_id,percentage,amount,conversion_rate,is_paid,formula,note,paid_at)
		VALUES %s`

	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))

	tx := db.Begin()
	if err := tx.Exec(smt, valueArgs...).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *store) ListEmployeeCommissions(db *gorm.DB, employeeID model.UUID, isPaid bool) ([]model.EmployeeCommission, error) {
	comms := []model.EmployeeCommission{}
	return comms, db.Where("employee_id = ? AND is_paid = ?", employeeID, isPaid).Find(&comms).Error
}

func (s *store) CloseEmployeeCommission(db *gorm.DB, id model.UUID) error {
	return db.Table("employee_commissions").Where("id = ?", id).Update("is_paid", true).Error
}
