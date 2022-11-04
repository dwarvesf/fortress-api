package employee_store

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (employees []*model.Employee, err error)
	One(id string) (employee *model.Employee, err error)
}
