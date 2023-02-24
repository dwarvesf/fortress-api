package employee

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdatePersonalInfoInput struct {
	DoB              *time.Time
	Gender           string
	PlaceOfResidence string
	Address          string
	PersonalEmail    string
	Country          string
	City             string
}

func (r *controller) UpdatePersonalInfo(employeeID string, body UpdatePersonalInfoInput) (*model.Employee, error) {
	emp, err := r.store.Employee.One(r.repo.DB(), employeeID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEmployeeNotFound
		}
		return nil, err
	}

	// validate personal email
	_, err = r.store.Employee.OneByEmail(r.repo.DB(), body.PersonalEmail)
	if emp.PersonalEmail != body.PersonalEmail && body.PersonalEmail != "" && !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			return nil, err
		}
		return nil, err
	}

	emp.DateOfBirth = body.DoB
	emp.Gender = body.Gender
	emp.Address = body.Address
	emp.PlaceOfResidence = body.PlaceOfResidence
	emp.PersonalEmail = body.PersonalEmail
	emp.Country = body.Country
	emp.City = body.City

	emp, err = r.store.Employee.Update(r.repo.DB(), emp)
	if err != nil {
		return nil, err
	}

	return emp, nil
}
