package role

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (roles []*model.Role, err error)
	One(id model.UUID) (role *model.Role, err error)
}
