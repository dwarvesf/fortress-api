package messageboard

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

// Service -- message board service
type Service interface {
	Create(message *model.Message, projectID, messageBoardID int) (err error)
	GetList(projectID int, messageBoardID int) (messages []model.Message, err error)
	Get(projectID int, messageID int) (message model.Message, err error)
}
