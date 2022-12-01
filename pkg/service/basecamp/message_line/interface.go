package message_line

type MsgLineService interface {
	CreateMsgLine(projectID int64, campfireID int64, line string) error
}
