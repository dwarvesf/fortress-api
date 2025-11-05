package webhook

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/logger"
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
	l := h.logger.Fields(logger.Fields{
		"handler":     "basecamp_expense",
		"method":      "createBasecampExpense",
		"basecampID":  msg.Recording.ID,
		"creatorID":   msg.Creator.ID,
		"title":       msg.Recording.Title,
	})

	commentMessage := h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, consts.CommentCreateExpenseFailed, bcModel.CommentMsgTypeFailed)

	defer func() {
		h.worker.Enqueue(bcModel.BasecampCommentMsg, commentMessage)
	}()

	obj, err := h.extractExpenseData(msg)
	if err != nil {
		l.Error(err, "failed to extract expense data from basecamp message")
		return err
	}

	if obj == nil {
		return nil
	}

	err = obj.MetaData.UnmarshalJSON(rawData)
	if err != nil {
		l.Error(err, "failed to unmarshal metadata JSON")
		return err
	}

	err = h.service.Basecamp.CreateBasecampExpense(*obj)
	if err != nil {
		l.Fields(logger.Fields{
			"amount":       obj.Amount,
			"currency":     obj.CurrencyType,
			"reason":       obj.Reason,
			"creatorEmail": obj.CreatorEmail,
		}).Error(err, "failed to create basecamp expense")
		return err
	}

	commentMessage = h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, consts.CommentCreateExpenseSuccessfully, bcModel.CommentMsgTypeCompleted)

	return nil
}

func (h *handler) UncheckBasecampExpenseHandler(msg model.BasecampWebhookMessage, rawData []byte) error {
	commentMsg := h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, consts.CommentDeleteExpenseFailed, bcModel.CommentMsgTypeFailed)

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

	commentMsg = h.service.Basecamp.BuildCommentMessage(msg.Recording.Bucket.ID, msg.Recording.ID, consts.CommentDeleteExpenseSuccessfully, bcModel.CommentMsgTypeCompleted)

	return nil
}

// extractExpenseData takes a webhook message and parse it into BasecampExpenseData structure
func (h *handler) extractExpenseData(msg model.BasecampWebhookMessage) (*bc.BasecampExpenseData, error) {
	l := h.logger.Fields(logger.Fields{
		"handler":    "basecamp_expense",
		"method":     "extractExpenseData",
		"basecampID": msg.Recording.ID,
		"title":      msg.Recording.Title,
	})

	res := &bc.BasecampExpenseData{BasecampID: msg.Recording.ID}

	parts := strings.Split(msg.Recording.Title, "|")
	if len(parts) < 3 {
		err := errors.New("invalid expense format")
		l.Fields(logger.Fields{
			"partsCount": len(parts),
			"parts":      parts,
		}).Error(err, "title must have 3 pipe-separated parts")
		return nil, err
	}

	// extract reason
	datetime := fmt.Sprintf(" %s %v", msg.Recording.UpdatedAt.Month().String(), msg.Recording.UpdatedAt.Year())
	res.Reason = strings.TrimSpace(parts[0])
	res.Reason += datetime

	// extract amount
	rawAmount := strings.TrimSpace(parts[1])
	amount := h.service.Basecamp.ExtractBasecampExpenseAmount(rawAmount)
	if amount == 0 {
		err := errors.New("invalid amount section of expense format")
		l.Fields(logger.Fields{
			"rawAmount": rawAmount,
			"parsed":    amount,
		}).Error(err, "amount is 0 or cannot be parsed")
		return nil, err
	}
	res.Amount = amount

	// extract currency type
	t := strings.ToLower(strings.TrimSpace(parts[2]))
	if t != "vnd" && t != "usd" {
		err := errors.New("invalid format in currency type section (VND or USD) of expense format")
		l.AddField("currency", t).Error(err, "currency must be VND or USD")
		return nil, err
	}
	url, err := h.service.Basecamp.Recording.TryToGetInvoiceImageURL(msg.Recording.URL)
	if err != nil {
		l.AddField("recordingURL", msg.Recording.URL).Error(err, "failed to get invoice image URL from basecamp")
		return nil, fmt.Errorf("failed to get image url by error %v", err)
	}

	res.CurrencyType = strings.ToUpper(t)

	list, err := h.service.Basecamp.Todo.GetList(msg.Recording.Parent.URL)
	if err != nil {
		l.AddField("parentURL", msg.Recording.Parent.URL).Error(err, "failed to get todo list from basecamp")
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
	res.InvoiceImageURL = url

	return res, nil
}
