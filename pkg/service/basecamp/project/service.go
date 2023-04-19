package project

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Service interface {
	GetAll() (result []model.Project, err error)
	Get(id int) (result model.Project, err error)
}
