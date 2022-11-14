package employee

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IStore interface {
	Search(query SearchFilter, pagination model.Pagination) (employees []*model.Employee, total int64, err error)
	One(id string) (employee *model.Employee, err error)
	OneByTeamEmail(teamEmail string) (employee *model.Employee, err error)
	UpdateEmployeeStatus(employeeID string, accountStatusID model.AccountStatus) (employee *model.Employee, err error)
	UpdateGeneralInfo(body EditGeneralInfo, id string) (*model.Employee, error)
	UpdatePersonalInfo(body EditPersonalInfo, id string) (employee *model.Employee, err error)
}
