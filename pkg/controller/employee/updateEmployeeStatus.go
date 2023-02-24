package employee

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateWorkingStatusInput struct {
	EmployeeStatus model.WorkingStatus
}

func (r *controller) UpdateEmployeeStatus(employeeID string, body UpdateWorkingStatusInput) (*model.Employee, error) {
	emp, err := r.store.Employee.One(r.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	emp.WorkingStatus = body.EmployeeStatus
	emp.LeftDate = nil

	now := time.Now()
	if body.EmployeeStatus == model.WorkingStatusLeft {
		emp.LeftDate = &now
	}

	_, err = r.store.Employee.UpdateSelectedFieldsByID(r.repo.DB(), employeeID, *emp, "working_status", "left_date")
	if err != nil {
		return nil, err
	}

	return emp, err
}
