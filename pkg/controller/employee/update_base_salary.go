package employee

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UpdateBaseSalaryInput struct {
	ContractAmount        int64
	CompanyAccountAmount  int64
	PersonalAccountAmount int64
	InsuranceAmount       int64
	CurrencyCode          string
	Batch                 int
	EffectiveDate         *time.Time
}

func (r *controller) UpdateBaseSalary(l logger.Logger, employeeID string, body UpdateBaseSalaryInput) (*model.BaseSalary, error) {
	currency, err := r.store.Currency.GetByName(r.repo.DB(), body.CurrencyCode)
	if err != nil {
		return nil, err
	}

	exists, err := r.store.Employee.IsExist(r.repo.DB(), employeeID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrEmployeeNotFound
	}

	bs, err := r.store.BaseSalary.OneByEmployeeID(r.repo.DB(), employeeID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	euuid, err := model.UUIDFromString(employeeID)
	if err != nil {
		return nil, err
	}

	newBS := &model.BaseSalary{
		BaseModel: model.BaseModel{
			ID: bs.ID,
		},
		EmployeeID:            euuid,
		ContractAmount:        body.ContractAmount,
		CompanyAccountAmount:  body.CompanyAccountAmount,
		PersonalAccountAmount: body.PersonalAccountAmount,
		InsuranceAmount:       model.NewVietnamDong(body.InsuranceAmount),
		CurrencyID:            currency.ID,
		Batch:                 body.Batch,
		EffectiveDate:         body.EffectiveDate,
	}

	err = r.store.BaseSalary.Save(r.repo.DB(), newBS)
	if err != nil {
		return nil, err
	}

	return newBS, nil
}
