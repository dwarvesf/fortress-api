package seniority

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All() ([]*model.Seniority, error)
}
