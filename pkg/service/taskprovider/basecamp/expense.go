package basecamp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/model"
	bc "github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

const expenseFormatSuccessMessage = "Your format looks good 👍"

func (p *Provider) ParseExpenseWebhook(ctx context.Context, req taskprovider.ExpenseWebhookRequest) (*taskprovider.ExpenseWebhookPayload, error) {
	if req.BasecampMessage == nil {
		return nil, errors.New("missing basecamp message")
	}
	msg := req.BasecampMessage
	eventType := mapExpenseEvent(msg.Kind)
	if eventType == taskprovider.ExpenseEventType("") {
		return nil, nil
	}
	return &taskprovider.ExpenseWebhookPayload{
		Provider:        taskprovider.ProviderBasecamp,
		EventType:       eventType,
		Title:           msg.Recording.Title,
		CreatorName:     msg.Creator.Name,
		CreatorID:       msg.Creator.ID,
		CreatorEmail:    msg.Creator.Email,
		BucketName:      msg.Recording.Bucket.Name,
		BucketID:        msg.Recording.Bucket.ID,
		RecordingID:     msg.Recording.ID,
		RecordingURL:    msg.Recording.URL,
		Amount:          0,
		TaskRef:         strconv.Itoa(msg.Recording.ID),
		TaskBoard:       msg.Recording.Bucket.Name,
		Metadata:        req.Body,
		BasecampMessage: msg,
		Raw:             msg,
	}, nil
}

func (p *Provider) ValidateSubmission(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseValidationResult, error) {
	if payload == nil {
		return nil, errors.New("missing expense payload")
	}
	msg := payload.BasecampMessage
	if msg == nil {
		return nil, errors.New("missing basecamp payload data")
	}
	if payload.EventType != taskprovider.ExpenseEventValidate {
		return &taskprovider.ExpenseValidationResult{Skip: true}, nil
	}

	bucketName, assignees, projectID := p.expenseBucketDefaults()
	if payload.BucketName != bucketName {
		return &taskprovider.ExpenseValidationResult{Skip: true}, nil
	}

	if err := p.assignDefaultExpenseAssignee(msg, assignees, projectID); err != nil {
		return nil, err
	}

	data, err := p.extractExpenseData(ctx, msg)
	if err != nil {
		mention, mentionErr := p.svc.BasecampMention(msg.Creator.ID)
		if mentionErr != nil {
			mention = msg.Creator.Name
		}
		return &taskprovider.ExpenseValidationResult{
			Valid:        false,
			Message:      formatExpenseErrorMessage(mention),
			FeedbackKind: bcModel.CommentMsgTypeFailed,
		}, nil
	}

	payload.Reason = data.Reason
	payload.Amount = data.Amount
	payload.Currency = data.CurrencyType
	payload.CreatorEmail = data.CreatorEmail
	return &taskprovider.ExpenseValidationResult{
		Valid:        true,
		Message:      expenseFormatSuccessMessage,
		FeedbackKind: bcModel.CommentMsgTypeCompleted,
	}, nil
}

func (p *Provider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
	if payload == nil {
		return nil, errors.New("missing expense payload")
	}
	msg := payload.BasecampMessage
	if msg == nil {
		return nil, errors.New("missing basecamp payload data")
	}
	data, err := p.extractExpenseData(ctx, msg)
	if err != nil {
		return nil, err
	}
	if len(payload.Metadata) > 0 {
		if err := data.MetaData.UnmarshalJSON(payload.Metadata); err != nil {
			return nil, err
		}
	}
	if err := p.svc.CreateBasecampExpense(*data); err != nil {
		return nil, err
	}
	return &taskprovider.ExpenseTaskRef{
		Provider:   taskprovider.ProviderBasecamp,
		ExternalID: strconv.Itoa(msg.Recording.ID),
		BucketID:   msg.Recording.Bucket.ID,
		TodoID:     msg.Recording.ID,
	}, nil
}

func (p *Provider) CompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	_, err := p.CreateExpense(ctx, payload)
	return err
}

func (p *Provider) UncompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	if payload == nil {
		return errors.New("missing expense payload")
	}
	msg := payload.BasecampMessage
	if msg == nil {
		return errors.New("missing basecamp payload data")
	}
	data, err := p.extractExpenseData(ctx, msg)
	if err != nil {
		return err
	}
	if len(payload.Metadata) > 0 {
		if err := data.MetaData.UnmarshalJSON(payload.Metadata); err != nil {
			return err
		}
	}
	return p.svc.UncheckBasecampExpenseHandler(*data)
}

func (p *Provider) DeleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	return p.UncompleteExpense(ctx, payload)
}

func (p *Provider) PostFeedback(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload, input taskprovider.ExpenseFeedbackInput) error {
	if payload == nil {
		return errors.New("missing expense payload")
	}
	msg := payload.BasecampMessage
	if msg == nil {
		return errors.New("missing basecamp payload data")
	}
	comment := p.svc.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, input.Message, input.Kind)
	return p.svc.Comment.Create(comment.ProjectID, comment.RecordingID, comment.Payload)
}

func mapExpenseEvent(kind string) taskprovider.ExpenseEventType {
	switch kind {
	case consts.TodoCreate:
		return taskprovider.ExpenseEventValidate
	case consts.TodoComplete:
		return taskprovider.ExpenseEventCreate
	case consts.TodoUncomplete:
		return taskprovider.ExpenseEventUncomplete
	default:
		return taskprovider.ExpenseEventType("")
	}
}

func (p *Provider) expenseBucketDefaults() (bucketName string, assignees []int, projectID int) {
	bucketName = consts.BucketNameWoodLand
	assignees = []int{consts.HanBasecampID}
	projectID = consts.WoodlandID
	if p.cfg != nil && p.cfg.Env != "prod" {
		bucketName = consts.BucketNamePlayGround
		assignees = []int{consts.NamNguyenBasecampID}
		projectID = consts.PlaygroundID
	}
	return
}

func (p *Provider) assignDefaultExpenseAssignee(msg *model.BasecampWebhookMessage, assigneeIDs []int, projectID int) error {
	if msg.Kind != consts.TodoCreate {
		return nil
	}
	todo, err := p.svc.Todo.Get(msg.Recording.URL)
	if err != nil {
		return err
	}
	todo.AssigneeIDs = assigneeIDs
	_, err = p.svc.Todo.Update(projectID, *todo)
	return err
}

func (p *Provider) extractExpenseData(ctx context.Context, msg *model.BasecampWebhookMessage) (*bc.BasecampExpenseData, error) {
	if msg == nil {
		return nil, errors.New("missing basecamp message")
	}
	res := &bc.BasecampExpenseData{BasecampID: msg.Recording.ID}

	parts := strings.Split(msg.Recording.Title, "|")
	if len(parts) < 3 {
		return nil, errors.New("invalid expense format")
	}

	datetime := fmt.Sprintf(" %s %v", msg.Recording.UpdatedAt.Month().String(), msg.Recording.UpdatedAt.Year())
	res.Reason = strings.TrimSpace(parts[0]) + datetime

	rawAmount := strings.TrimSpace(parts[1])
	amount := p.svc.ExtractBasecampExpenseAmount(rawAmount)
	if amount == 0 {
		return nil, errors.New("invalid amount section")
	}
	res.Amount = amount

	currency := strings.ToUpper(strings.TrimSpace(parts[2]))
	if currency != "VND" && currency != "USD" {
		return nil, errors.New("invalid currency type")
	}
	res.CurrencyType = currency

	url, err := p.svc.Recording.TryToGetInvoiceImageURL(msg.Recording.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice image url: %w", err)
	}
	res.InvoiceImageURL = url
	res.TaskAttachmentURL = url
	res.TaskProvider = "basecamp"
	res.TaskRef = strconv.Itoa(msg.Recording.ID)
	res.TaskBoard = msg.Recording.Bucket.Name
	res.TaskAttachments = bc.MarshalAttachmentArray([]string{url})

	list, err := p.svc.Todo.GetList(msg.Recording.Parent.URL)
	if err != nil {
		return nil, err
	}
	msg.Recording.Parent.Title = list.Parent.Title

	if msg.IsExpenseComplete() {
		res.CreatorEmail = msg.Recording.Creator.Email
		res.CreatorID = msg.Recording.Creator.ID
	}
	if msg.IsOperationComplete() {
		res.CreatorEmail = msg.Creator.Email
	}

	res.MetaData = datatypes.JSON([]byte("{}"))
	return res, nil
}

func formatExpenseErrorMessage(mention string) string {
	return fmt.Sprintf(`Hi %s, I'm not smart enough to understand your expense submission. Please ensure the following format 😊

Title: < Reason > | < Amount > | < VND/USD >
Assign To: Han Ngo, < payee >

Example:
Title: Tiền mèo | 400.000 | VND
Assign To: Han Ngo, < payee >`, mention)
}
