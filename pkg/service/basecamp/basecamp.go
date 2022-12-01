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

func New(bc config.Basecamp) (*BasecampService, error) {
	c, err := client.NewClient(bc)
	if err != nil {
		return nil, err
	}

	return &BasecampService{
		Basecamp: bc,
		Client:   c,

		Comment:      comment.New(c),
		Todo:         todo.New(c),
		Attachment:   attachment.New(c),
		Schedule:     schedule.New(c),
		Subscription: subscription.New(c),
		Account:      account.New(c),
		MsgBoard:     message_board.New(c),
		MsgLine:      message_line.New(c, bc.BotKey),
		Project:      project.New(c),
		Recording:    recording.New(c),
	}, nil
}

func NewTestService() *BasecampService {
	return &BasecampService{}
}
