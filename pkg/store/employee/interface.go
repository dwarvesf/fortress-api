package employee_store

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (employees []*model.Employee, err error)
	One(id string) (employee *model.Employee, err error)
	OneByTeamEmail(teamEmail string) (employee *model.Employee, err error)
	UpdateEmployeeStatus(employeeID string, accountStatusID model.AccountStatus) (employee *model.Employee, err error)
}
