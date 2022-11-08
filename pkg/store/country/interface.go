package country

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (countries []*model.Country, err error)
	One(id string) (countries *model.Country, err error)
}
