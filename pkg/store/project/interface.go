package project

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All(input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error)
}
