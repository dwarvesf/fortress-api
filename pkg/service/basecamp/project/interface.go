package project

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type ProjectService interface {
	GetAll() ([]model.Project, error)
	Get(id int) (model.Project, error)
}
