package basecamp

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	appmodel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

// Provider implements taskprovider.InvoiceProvider backed by Basecamp.
type Provider struct {
	svc *basecamp.Service
	cfg *config.Config
}

// New creates a Basecamp-backed invoice provider.
func New(svc *basecamp.Service, cfg *config.Config) *Provider {
	if svc == nil {
		return nil
	}
	return &Provider{svc: svc, cfg: cfg}
}

func (p *Provider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderBasecamp
}

func (p *Provider) EnsureTask(ctx context.Context, input taskprovider.CreateInvoiceTaskInput) (*taskprovider.InvoiceTaskRef, error) {
	if input.Invoice == nil {
		return nil, errors.New("missing invoice data")
	}

	bucketID, todoID, err := p.ensureInvoiceTodo(input.Invoice)
	if err != nil {
		return nil, err
	}

	return &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderBasecamp,
		ExternalID: strconv.Itoa(todoID),
		BucketID:   bucketID,
		TodoID:     todoID,
	}, nil
}

func (p *Provider) UploadAttachment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceAttachmentInput) (*taskprovider.InvoiceAttachmentRef, error) {
	if input.FileName == "" || len(input.Content) == 0 {
		return nil, errors.New("invalid attachment payload")
	}

	sgid, err := p.svc.Attachment.Create(input.ContentType, input.FileName, input.Content)
	if err != nil {
		return nil, err
	}

	markup := fmt.Sprintf(`<bc-attachment sgid="%s" caption="Invoice attachment"></bc-attachment>`, sgid)
	return &taskprovider.InvoiceAttachmentRef{ExternalID: sgid, Markup: markup}, nil
}

func (p *Provider) PostComment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceCommentInput) error {
	if ref == nil {
		return errors.New("missing invoice task reference")
	}

	msg := p.svc.BuildCommentMessage(ref.BucketID, ref.TodoID, input.Message, input.Type)
	return p.svc.Comment.Create(msg.ProjectID, msg.RecordingID, msg.Payload)
}

func (p *Provider) CompleteTask(ctx context.Context, ref *taskprovider.InvoiceTaskRef) error {
	if ref == nil {
		return errors.New("missing invoice task reference")
	}

	return p.svc.Todo.Complete(ref.BucketID, ref.TodoID)
}

func (p *Provider) ensureInvoiceTodo(iv *appmodel.Invoice) (bucketID int, todoID int, err error) {
	if iv.Project == nil {
		return 0, 0, errors.New("missing project info")
	}

	accountingID := defaultAccountingProjectID(p.cfg)
	accountingTodoID := defaultAccountingTodoSetID(p.cfg)
	playgroundProjectID := defaultPlaygroundProjectID(p.cfg)
	playgroundTodoID := defaultPlaygroundTodoSetID(p.cfg)

	if p.cfg != nil && p.cfg.Env != "prod" {
		accountingID = playgroundProjectID
		accountingTodoID = playgroundTodoID
	}

	re := regexp.MustCompile(`Accounting \| ([A-Za-z]+) ([0-9]{4})`)

	todoLists, err := p.svc.Todo.GetLists(accountingID, accountingTodoID)
	if err != nil {
		return 0, 0, err
	}

	var todoList *bcModel.TodoList
	var latestListDate time.Time

	for i := range todoLists {
		info := re.FindStringSubmatch(todoLists[i].Title)
		if len(info) == 3 {
			month, err := timeutil.GetMonthFromString(info[1])
			if err != nil {
				continue
			}
			year, _ := strconv.Atoi(info[2])
			listDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			if listDate.After(latestListDate) {
				todoList = &todoLists[i]
				latestListDate = listDate
			}
		}
	}

	if todoList == nil {
		month := iv.Month + 1
		if month > 12 {
			month = 1
		}

		todoList, err = p.svc.Todo.CreateList(
			accountingID,
			accountingTodoID,
			bcModel.TodoList{Name: fmt.Sprintf(`Accounting | %v %v`, time.Month(month).String(), iv.Year)},
		)
		if err != nil {
			return 0, 0, err
		}
	}

	todoGroup, err := p.svc.Todo.FirstOrCreateGroup(accountingID, todoList.ID, "In")
	if err != nil {
		return 0, 0, err
	}

	todoItem, err := p.svc.Todo.FirstOrCreateInvoiceTodo(accountingID, todoGroup.ID, iv)
	if err != nil {
		return 0, 0, err
	}

	return accountingID, todoItem.ID, nil
}

func defaultAccountingProjectID(cfg *config.Config) int {
	if cfg != nil && cfg.Basecamp.AccountingProjectID != 0 {
		return cfg.Basecamp.AccountingProjectID
	}
	return consts.AccountingID
}

func defaultAccountingTodoSetID(cfg *config.Config) int {
	if cfg != nil && cfg.Basecamp.AccountingTodoSetID != 0 {
		return cfg.Basecamp.AccountingTodoSetID
	}
	return consts.AccountingTodoID
}

func defaultPlaygroundProjectID(cfg *config.Config) int {
	if cfg != nil && cfg.Basecamp.PlaygroundProjectID != 0 {
		return cfg.Basecamp.PlaygroundProjectID
	}
	return consts.PlaygroundID
}

func defaultPlaygroundTodoSetID(cfg *config.Config) int {
	if cfg != nil && cfg.Basecamp.PlaygroundTodoSetID != 0 {
		return cfg.Basecamp.PlaygroundTodoSetID
	}
	return consts.PlaygroundTodoID
}
