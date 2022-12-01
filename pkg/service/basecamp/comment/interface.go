package comment

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type CommentService interface {
	Create(recordingID int64, projectID int64, comment *model.Comment) error
	Gets(recordingID int64, projectID int64) ([]model.Comment, error)
}
