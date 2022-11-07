package position_store

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (positions []*model.Position, err error)
}
