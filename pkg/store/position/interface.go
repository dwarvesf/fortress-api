package position

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (positions []*model.Position, err error)
	One(id model.UUID) (position *model.Position, err error)
}
