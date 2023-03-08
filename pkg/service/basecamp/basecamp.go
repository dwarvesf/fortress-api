package basecamp

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/attachment"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/campfire"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/comment"
	mb "github.com/dwarvesf/fortress-api/pkg/service/basecamp/message_board"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/people"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/project"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/recording"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/schedule"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/subscription"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/todo"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/webhook"
)

type Service struct {
	Basecamp *model.Basecamp
	Client   client.Service

	Comment      comment.Service
	Todo         todo.Service
	MsgBoard     mb.Service
	Campfire     campfire.Service
	Recording    recording.Service
	Project      project.Service
	People       people.Service
	Subscription subscription.Service
	Schedule     schedule.Service
	Webhook      webhook.Service
	Attachment   attachment.Service
}

func NewService(bc *model.Basecamp, cfg *config.Config, logger logger.Logger) *Service {
	c, err := client.NewClient(bc, cfg)
	if err != nil {
		logger.Error(err, "init basecamp")
		return nil
	}

	return &Service{
		Basecamp: bc,
		Client:   c,

		Comment:      comment.NewService(c),
		Todo:         todo.NewService(c, cfg),
		MsgBoard:     mb.NewService(c),
		Campfire:     campfire.NewService(c, logger, cfg),
		Recording:    recording.NewService(c),
		Project:      project.NewService(c),
		People:       people.NewService(c),
		Subscription: subscription.NewService(c),
		Schedule:     schedule.NewService(c, logger),
		Webhook:      webhook.NewService(c),
		Attachment:   attachment.NewService(c),
	}
}
