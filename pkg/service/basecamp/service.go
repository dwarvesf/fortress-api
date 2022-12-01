package basecamp

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/account"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/attachment"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/comment"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/message_board"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/message_line"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/project"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/recording"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/schedule"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/subscription"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/todo"
)

type BasecampService struct {
	Basecamp config.Basecamp
	Client   client.ClientService

	Comment      comment.CommentService
	Todo         todo.TodoService
	Attachment   attachment.AttachmentService
	Schedule     schedule.ScheduleService
	Subscription subscription.SubscriptionService
	Account      account.AccountService
	MsgBoard     message_board.MsgBoardService
	MsgLine      message_line.MsgLineService
	Project      project.ProjectService
	Recording    recording.RecordingService
}
