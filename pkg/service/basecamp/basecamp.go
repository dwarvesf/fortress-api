package basecamp

import (
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/attachment"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/campfire"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/comment"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/messageboard"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/people"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/project"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/recording"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/schedule"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/subscription"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/todo"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/webhook"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	store  *store.Store
	repo   store.DBRepo
	config *config.Config
	logger logger.Logger

	Basecamp *model.Basecamp
	Client   client.Service

	Attachment   attachment.Service
	Campfire     campfire.Service
	Comment      comment.Service
	MsgBoard     messageboard.Service
	People       people.Service
	Project      project.Service
	Recording    recording.Service
	Schedule     schedule.Service
	Subscription subscription.Service
	Todo         todo.Service
	Webhook      webhook.Service

	Wise wise.IService
}

func New(store *store.Store, repo store.DBRepo, cfg *config.Config, bc *model.Basecamp, logger logger.Logger) *Service {
	c, err := client.NewClient(bc, cfg)
	if err != nil {
		logger.Error(err, "init basecamp service")
		return nil
	}

	return &Service{
		store:        store,
		repo:         repo,
		config:       cfg,
		logger:       logger,
		Basecamp:     bc,
		Client:       c,
		Attachment:   attachment.NewService(c),
		Campfire:     campfire.NewService(c, logger, cfg),
		Comment:      comment.NewService(c),
		MsgBoard:     messageboard.NewService(c),
		People:       people.NewService(c),
		Project:      project.NewService(c),
		Recording:    recording.NewService(c),
		Schedule:     schedule.NewService(c, logger),
		Subscription: subscription.NewService(c),
		Todo:         todo.NewService(c, cfg),
		Webhook:      webhook.NewService(c),
		Wise:         wise.New(cfg, logger),
	}
}

func (s *Service) BuildCommentMessage(bucketID, recordID int, content string, msgType string) model.BasecampCommentMessage {
	var cmtPayload *model.Comment
	switch msgType {
	case model.CommentMsgTypeFailed:
		cmtPayload = s.buildFailedComment(content)
	case model.CommentMsgTypeCompleted:
		cmtPayload = s.buildCompletedComment(content)
	default:
		cmtPayload = &model.Comment{Content: content}
	}

	return model.BasecampCommentMessage{
		RecordingID: recordID,
		ProjectID:   bucketID,
		Payload:     cmtPayload,
	}
}

func (s *Service) BasecampMention(basecampID int) (res string, err error) {
	if basecampID == consts.AutoBotID {
		return fmt.Sprintf(`<bc-attachment sgid="%s" content-type="application/vnd.basecamp.mention"></bc-attachment>`, consts.AutoBotSgID), nil
	}

	employee, err := s.store.Employee.OneByBasecampID(s.repo.DB(), basecampID)
	if err != nil {
		return
	}

	if employee.BasecampAttachableSGID == "" {
		u, err := s.People.GetByID(basecampID)
		if err != nil {
			return res, err
		}
		employee.BasecampAttachableSGID = u.AttachableSgID

		if _, err = s.store.Employee.UpdateSelectedFieldsByID(s.repo.DB(), employee.ID.String(), *employee, "basecamp_attachable_sgid"); err != nil {
			return res, err
		}
	}

	return fmt.Sprintf(`<bc-attachment sgid="%s" content-type="application/vnd.basecamp.mention"></bc-attachment>`, employee.BasecampAttachableSGID), nil
}

func (s *Service) buildFailedComment(content string) *model.Comment {
	if s.config.Env == "prod" {
		m, _ := s.BasecampMention(consts.HuyNguyenBasecampID)
		return &model.Comment{Content: fmt.Sprintf(`<img width="17" class="thread-entry__icon" src="https://3.basecamp-static.com/assets/icons/thread_events/uncompleted-6066b80e80b6463243d7773fa67373b62e2a7d159ba12a17c94b1e18b30a5770.svg"><div><em>%s</em> %s</div>`, content, m)}
	}
	return &model.Comment{Content: fmt.Sprintf(`<img width="17" class="thread-entry__icon" src="https://3.basecamp-static.com/assets/icons/thread_events/uncompleted-6066b80e80b6463243d7773fa67373b62e2a7d159ba12a17c94b1e18b30a5770.svg"><div><em>%s</em></div>`, content)}
}

func (s *Service) buildCompletedComment(content string) *model.Comment {
	return &model.Comment{Content: fmt.Sprintf(`<img width="17" class="thread-entry__icon" src="https://3.basecamp-static.com/assets/icons/thread_events/completed-12705cf5fc372d800bba74c8133d705dc43a12c939a8477099749e2ef056e739.svg"><div><em>%s</em></div>`, content)}
}
