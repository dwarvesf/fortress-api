package basecamp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	appmodel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

var invoiceRegex = regexp.MustCompile(`^\s*(.+?)\s*\|\s*([0-9\.\,]+)\s*\|\s*([a-zA-Z]{3})`)

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
	return &taskprovider.InvoiceAttachmentRef{
		ExternalID: sgid,
		Markup:     markup,
		Meta: map[string]any{
			"provider": "basecamp",
			"sgid":     sgid,
		},
	}, nil
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

func (p *Provider) CreateMonthlyPlan(ctx context.Context, input taskprovider.CreateAccountingPlanInput) (*taskprovider.AccountingPlanRef, error) {
	if p == nil || p.svc == nil || p.cfg == nil {
		return nil, errors.New("basecamp provider not configured")
	}

	projectID, todoSetID := p.accountingProjectAndSet()
	list, err := p.svc.Todo.CreateList(projectID, todoSetID, bcModel.TodoList{
		Name: input.Label,
	})
	if err != nil {
		return nil, err
	}

	groupInName := p.cfg.AccountingIntegration.Basecamp.GroupIn
	if groupInName == "" {
		groupInName = "In"
	}
	groupOutName := p.cfg.AccountingIntegration.Basecamp.GroupOut
	if groupOutName == "" {
		groupOutName = "Out"
	}

	groupIn, err := p.svc.Todo.CreateGroup(projectID, list.ID, bcModel.TodoGroup{Name: groupInName})
	if err != nil {
		return nil, err
	}
	groupOut, err := p.svc.Todo.CreateGroup(projectID, list.ID, bcModel.TodoGroup{Name: groupOutName})
	if err != nil {
		return nil, err
	}

	return &taskprovider.AccountingPlanRef{
		Provider: taskprovider.ProviderBasecamp,
		BoardID:  strconv.Itoa(projectID),
		ListID:   strconv.Itoa(list.ID),
		Month:    input.Month,
		Year:     input.Year,
		GroupLookup: map[taskprovider.AccountingGroup]string{
			taskprovider.AccountingGroupIn:  strconv.Itoa(groupIn.ID),
			taskprovider.AccountingGroupOut: strconv.Itoa(groupOut.ID),
		},
	}, nil
}

func (p *Provider) CreateAccountingTodo(ctx context.Context, plan *taskprovider.AccountingPlanRef, input taskprovider.CreateAccountingTodoInput) (*taskprovider.AccountingTodoRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("basecamp provider not configured")
	}
	if plan == nil {
		return nil, errors.New("missing accounting plan reference")
	}

	projectID, err := strconv.Atoi(plan.BoardID)
	if err != nil {
		return nil, fmt.Errorf("invalid board id: %w", err)
	}

	groupIDStr, ok := plan.GroupLookup[input.Group]
	if !ok {
		return nil, fmt.Errorf("group %s not initialized", input.Group)
	}
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid group id: %w", err)
	}

	assigneeIDs := make([]int, 0, len(input.Assignees))
	for _, a := range input.Assignees {
		if a.ExternalID == "" {
			continue
		}
		id, err := strconv.Atoi(a.ExternalID)
		if err != nil {
			continue
		}
		assigneeIDs = append(assigneeIDs, id)
	}

	dueOn := input.DueDate
	if dueOn.IsZero() {
		dueOn = time.Now()
	}

	todo := bcModel.Todo{
		Content:     input.Title,
		Description: input.Description,
		DueOn:       fmt.Sprintf("%d-%d-%d", dueOn.Day(), dueOn.Month(), dueOn.Year()),
		AssigneeIDs: assigneeIDs,
	}

	created, err := p.svc.Todo.Create(projectID, groupID, todo)
	if err != nil {
		return nil, err
	}

	return &taskprovider.AccountingTodoRef{
		Provider:   taskprovider.ProviderBasecamp,
		ExternalID: strconv.Itoa(created.ID),
		Group:      input.Group,
	}, nil
}

func (p *Provider) ParseAccountingWebhook(ctx context.Context, req taskprovider.AccountingWebhookRequest) (*taskprovider.AccountingWebhookPayload, error) {
	if len(req.Body) == 0 {
		return nil, errors.New("empty webhook body")
	}
	var msg appmodel.BasecampWebhookMessage
	if err := json.Unmarshal(req.Body, &msg); err != nil {
		return nil, err
	}
	if msg.Recording.Title == "" {
		return nil, errors.New("missing recording title")
	}
	parts := invoiceRegex.FindStringSubmatch(msg.Recording.Title)
	if len(parts) != 4 {
		return nil, errors.New("unknown title format")
	}
	amountStr := strings.ReplaceAll(parts[2], ".", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return nil, err
	}
	return &taskprovider.AccountingWebhookPayload{
		Provider:  taskprovider.ProviderBasecamp,
		Group:     taskprovider.AccountingGroupOut,
		Title:     strings.TrimSpace(msg.Recording.Title),
		Amount:    float64(amount),
		Currency:  strings.ToUpper(strings.TrimSpace(parts[3])),
		TodoID:    strconv.Itoa(msg.Recording.ID),
		TodoRowID: strconv.Itoa(msg.Recording.ID),
		Actor:     msg.Creator.Name,
		Status:    "completed",
		Raw:       req.Body,
	}, nil
}

func (p *Provider) accountingProjectAndSet() (int, int) {
	projectID := defaultAccountingProjectID(p.cfg)
	todoSetID := defaultAccountingTodoSetID(p.cfg)
	if p.cfg != nil && p.cfg.Env != "prod" {
		projectID = defaultPlaygroundProjectID(p.cfg)
		todoSetID = defaultPlaygroundTodoSetID(p.cfg)
	}
	if cfgID := p.cfg.AccountingIntegration.Basecamp.ProjectID; cfgID != 0 {
		projectID = cfgID
	}
	if cfgSet := p.cfg.AccountingIntegration.Basecamp.TodoSetID; cfgSet != 0 {
		todoSetID = cfgSet
	}
	if p.cfg != nil && p.cfg.Env != "prod" {
		// In non-prod we still prefer playground overrides if provided
		if pid := p.cfg.Basecamp.PlaygroundProjectID; pid != 0 {
			projectID = pid
		}
		if tid := p.cfg.Basecamp.PlaygroundTodoSetID; tid != 0 {
			todoSetID = tid
		}
	}
	return projectID, todoSetID
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
