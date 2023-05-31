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

func (h *handler) BasecampExpenseValidateHandler(msg model.BasecampWebhookMessage) error {
	if msg.Recording.Bucket.Name != "Woodland" {
		return nil
	}

	// Todo ref: https://3.basecamp.com/4108948/buckets/9410372/todos/3204666678
	// Assign HanNgo whenever expense todo was created
	if msg.Kind == consts.TodoCreate {
		todo, err := h.service.Basecamp.Todo.Get(msg.Recording.URL)
		if err != nil {
			return err
		}
		assigneeIDs := []int{consts.HanBasecampID}
		projectID := consts.WoodlandID
		if h.config.Env != "prod" {
			assigneeIDs = []int{consts.KhanhTruongBasecampID}
			projectID = consts.PlaygroundID
		}
		todo.AssigneeIDs = assigneeIDs
		_, err = h.service.Basecamp.Todo.Update(projectID, *todo)
		if err != nil {
			return err
		}
	}

	_, err := h.ExtractExpenseData(msg)
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

		h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID,
			msg.Recording.ID,
			&bcModel.Comment{Content: errMsg},
		)
		return nil
	}
	h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID,
		msg.Recording.ID,
		&bcModel.Comment{Content: `Your format looks good 👍`})

	return nil
}

func (h *handler) BasecampExpenseHandler(msg model.BasecampWebhookMessage, rawData []byte) error {
	var comment func()
	defer func() {
		if comment == nil {
			h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildFailedComment(model.CommentCreateExpenseFailed))
			return
		}
		comment()
	}()

	obj, err := h.ExtractExpenseData(msg)
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

	err = h.service.Basecamp.BasecampExpenseHandler(*obj)
	if err != nil {
		return err
	}

	comment = func() {
		h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildCompletedComment(model.CommentCreateExpenseSuccessfully))
	}

	return nil
}

func (h *handler) UncheckBasecampExpenseHandler(msg model.BasecampWebhookMessage, rawData []byte) error {
	var comment func()
	defer func() {
		if comment == nil {
			h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildFailedComment(model.CommentDeleteExpenseFailed))
			return
		}
		comment()
	}()

	obj, err := h.ExtractExpenseData(msg)
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

	comment = func() {
		h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildCompletedComment(model.CommentDeleteExpenseSuccessfully))
	}

	return nil
}

// ExtractExpenseData takes a webhook message and parse it into BasecampExpenseData structure
func (h *handler) ExtractExpenseData(msg model.BasecampWebhookMessage) (*bc.BasecampExpenseData, error) {
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
