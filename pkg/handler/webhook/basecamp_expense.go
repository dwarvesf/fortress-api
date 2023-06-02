package webhook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
	bc "github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

func (h *handler) basecampExpenseValidate(msg model.BasecampWebhookMessage) error {
	recordingBucketName := consts.BucketNameWoodLand
	assigneeIDs := []int{consts.HanBasecampID}
	projectID := consts.WoodlandID

	if h.config.Env != "prod" {
		recordingBucketName = consts.BucketNamePlayGround
		assigneeIDs = []int{consts.NamNguyenBasecampID}
		projectID = consts.PlaygroundID
	}

	if msg.Recording.Bucket.Name != recordingBucketName {
		return nil
	}
	// Todo ref: https://3.basecamp.com/4108948/buckets/9410372/todos/3204666678
	// Assign HanNgo whenever expense todo was created
	if msg.Kind == consts.TodoCreate {
		todo, err := h.service.Basecamp.Todo.Get(msg.Recording.URL)
		if err != nil {
			return err
		}

		todo.AssigneeIDs = assigneeIDs
		_, err = h.service.Basecamp.Todo.Update(projectID, *todo)
		if err != nil {
			return err
		}
	}

	_, err := h.extractExpenseData(msg)
	if err != nil {
		m, err := h.service.Basecamp.BasecampMention(msg.Creator.ID)
		if err != nil {
			return err
		}

		errMsg := fmt.Sprintf(
			`Hi %v, I'm not smart enough to understand your expense submission. Please ensure the following format 😊

			Title: < Reason > | < Amount > | < VND/USD >
			Assign To: Han Ngo, < payee >

			Example:
			Title: Tiền mèo | 400.000 | VND
			Assign To: Han Ngo, < payee >`, m)

		h.worker.Enqueue(bcModel.BasecampCommentMsg, h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, errMsg, ""))
		return nil
	}

	commentMessage := h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID,
		msg.Recording.ID,
		"Your format looks good 👍",
		"",
	)

	h.worker.Enqueue(bcModel.BasecampCommentMsg, commentMessage)

	return nil
}

func (h *handler) createBasecampExpense(msg model.BasecampWebhookMessage, rawData []byte) error {
	commentMessage := h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, model.CommentCreateExpenseFailed, bcModel.CommentMsgTypeFailed)

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, commentMessage)
	}()

	obj, err := h.extractExpenseData(msg)
	if err != nil {
		return err
	}

	if obj == nil {
		return nil
	}

	err = obj.MetaData.UnmarshalJSON(rawData)
	if err != nil {
		return err
	}

	err = h.service.Basecamp.CreateBasecampExpense(*obj)
	if err != nil {
		return err
	}

	commentMessage = h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, model.CommentCreateExpenseSuccessfully, bcModel.CommentMsgTypeCompleted)

	return nil
}

func (h *handler) UncheckBasecampExpenseHandler(msg model.BasecampWebhookMessage, rawData []byte) error {
	commentMsg := h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, model.CommentDeleteExpenseFailed, bcModel.CommentMsgTypeFailed)

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, commentMsg)
	}()

	obj, err := h.extractExpenseData(msg)
	if err != nil {
		return err
	}
	err = obj.MetaData.UnmarshalJSON(rawData)
	if err != nil {
		return err
	}

	if obj == nil {
		return nil
	}

	err = h.service.Basecamp.UncheckBasecampExpenseHandler(*obj)
	if err != nil {
		return err
	}

	commentMsg = h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, model.CommentDeleteExpenseSuccessfully, bcModel.CommentMsgTypeCompleted)

	return nil
}

// extractExpenseData takes a webhook message and parse it into BasecampExpenseData structure
func (h *handler) extractExpenseData(msg model.BasecampWebhookMessage) (*bc.BasecampExpenseData, error) {
	res := &bc.BasecampExpenseData{BasecampID: msg.Recording.ID}

	parts := strings.Split(msg.Recording.Title, "|")
	if len(parts) < 3 {
		err := errors.New("invalid expense format")
		return nil, err
	}

	// extract reason
	datetime := fmt.Sprintf(" %s %v", msg.Recording.UpdatedAt.Month().String(), msg.Recording.UpdatedAt.Year())
	res.Reason = strings.TrimSpace(parts[0])
	res.Reason += datetime

	// extract amount
	amount := h.service.Basecamp.ExtractBasecampExpenseAmount(strings.TrimSpace(parts[1]))
	if amount == 0 {
		err := errors.New("invalid amount section of expense format")
		return nil, err
	}
	res.Amount = amount

	// extract currency type
	t := strings.ToLower(strings.TrimSpace(parts[2]))
	if t != "vnd" && t != "usd" {
		return nil, errors.New("invalid format in currency type section (VND or USD) of expense format")
	}
	url, err := h.service.Basecamp.Recording.TryToGetInvoiceImageURL(msg.Recording.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get image url by error %v", err)
	}

	res.CurrencyType = strings.ToUpper(t)

	list, err := h.service.Basecamp.Todo.GetList(msg.Recording.Parent.URL)
	if err != nil {
		return nil, err
	}
	msg.Recording.Parent.Title = list.Parent.Title
	if msg.IsExpenseComplete() {
		res.CreatorEmail = msg.Recording.Creator.Email
	}
	if msg.IsOperationComplete() {
		res.CreatorEmail = msg.Creator.Email
	}
	res.InvoiceImageURL = url

	return res, nil
}
