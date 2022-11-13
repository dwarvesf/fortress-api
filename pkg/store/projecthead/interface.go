package projecthead

import "github.com/dwarvesf/fortress-api/pkg/model"

type IStore interface {
	Create(projectHead *model.ProjectHead) error
}
