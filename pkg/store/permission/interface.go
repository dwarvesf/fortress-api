package permission

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	GetByEmployeeID(employeeID string) (permissions []*model.Permission, err error)
}
