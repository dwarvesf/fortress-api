package webhook

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Service interface {
	FindWebHook(projectID int, hookID int) (result *model.Hook, err error)
	UpdateWebHook(projectID int, hookID int, hookBody model.Hook) (err error)
}
