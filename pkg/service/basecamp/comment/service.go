package comment

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Service interface {
	Create(projectID int, recordingID int, comment *model.Comment) (err error)
	Gets(projectID int, recordingID int) (res []model.Comment, err error)
}
