package project

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	All(input GetListProjectInput, pagination model.Pagination) ([]*model.Project, int64, error)
	UpdateStatus(projectID string, projectStatus model.ProjectStatus) (*model.Project, error)
	Create(project *model.Project) error
}
