package message_board

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type MsgBoardService interface {
	Create(projectID int64, msgBoardID int64, msg *model.Message) error
}
