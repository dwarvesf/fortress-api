package chapter

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() (chapters []*model.Chapter, err error)
}
