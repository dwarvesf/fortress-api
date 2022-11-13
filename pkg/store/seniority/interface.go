package seniority

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() ([]*model.Seniority, error)
	One(id model.UUID) (seniorities *model.Seniority, err error)
}
