package webhook

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func basecampWebhookMessageFromCtx(c *gin.Context, l logger.Logger) (model.BasecampWebhookMessage, []byte, error) {
	var msg model.BasecampWebhookMessage
	if c.Request.Body == nil {
		return msg, nil, errors.New("empty request body")
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read basecamp webhook body")
		return msg, nil, err
	}
	_ = c.Request.Body.Close()
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := msg.Decode(body); err != nil {
		l.Error(err, "failed to decode basecamp webhook message JSON")
		return msg, nil, err
	}
	return msg, body, nil
}

// ValidateBasecampExpense dry-run expense request for validation
func (h *handler) ValidateBasecampExpense(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "ValidateBasecampExpense",
	})

	msg, body, err := basecampWebhookMessageFromCtx(c, l)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	provider := h.service.ExpenseProvider
	if provider == nil {
		err := errors.New("expense provider not configured")
		l.Error(err, "missing expense provider")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload, err := provider.ParseExpenseWebhook(c.Request.Context(), taskprovider.ExpenseWebhookRequest{
		Headers:         headerMap(c.Request.Header),
		Body:            body,
		BasecampMessage: &msg,
	})
	if err != nil {
		l.Error(err, "failed to parse basecamp expense webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload == nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	result, err := provider.ValidateSubmission(c.Request.Context(), payload)
	if err != nil {
		l.Error(err, "expense validation failed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if result == nil || result.Skip {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}
	if result.Message != "" {
		if err := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
			Message: result.Message,
			Kind:    result.FeedbackKind,
		}); err != nil {
			l.Error(err, "failed to post expense validation feedback")
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// CreateBasecampExpense runs expense process in basecamp
func (h *handler) CreateBasecampExpense(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "CreateBasecampExpense",
	})

	msg, body, err := basecampWebhookMessageFromCtx(c, l)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	provider := h.service.ExpenseProvider
	if provider == nil {
		err := errors.New("expense provider not configured")
		l.Error(err, "missing expense provider")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload, err := provider.ParseExpenseWebhook(c.Request.Context(), taskprovider.ExpenseWebhookRequest{
		Headers:         headerMap(c.Request.Header),
		Body:            body,
		BasecampMessage: &msg,
	})
	if err != nil {
		l.Error(err, "failed to parse basecamp expense webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload == nil || payload.EventType != taskprovider.ExpenseEventCreate {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if _, err := provider.CreateExpense(c.Request.Context(), payload); err != nil {
		l.Error(err, "failed to create basecamp expense")
		if postErr := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
			Message: consts.CommentCreateExpenseFailed,
			Kind:    bcModel.CommentMsgTypeFailed,
		}); postErr != nil {
			l.Error(postErr, "failed to post expense failure feedback")
		}
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if err := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
		Message: consts.CommentCreateExpenseSuccessfully,
		Kind:    bcModel.CommentMsgTypeCompleted,
	}); err != nil {
		l.Error(err, "failed to post expense success feedback")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// UncheckBasecampExpense will remove expesne record after expense todo complete
func (h *handler) UncheckBasecampExpense(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "UncheckBasecampExpense",
	})

	msg, body, err := basecampWebhookMessageFromCtx(c, l)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	provider := h.service.ExpenseProvider
	if provider == nil {
		err := errors.New("expense provider not configured")
		l.Error(err, "missing expense provider")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload, err := provider.ParseExpenseWebhook(c.Request.Context(), taskprovider.ExpenseWebhookRequest{
		Headers:         headerMap(c.Request.Header),
		Body:            body,
		BasecampMessage: &msg,
	})
	if err != nil {
		l.Error(err, "failed to parse basecamp expense webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload == nil || payload.EventType != taskprovider.ExpenseEventUncomplete {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if err := provider.UncompleteExpense(c.Request.Context(), payload); err != nil {
		l.Error(err, "failed to uncomplete basecamp expense")
		if postErr := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
			Message: consts.CommentDeleteExpenseFailed,
			Kind:    bcModel.CommentMsgTypeFailed,
		}); postErr != nil {
			l.Error(postErr, "failed to post expense delete failure feedback")
		}
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if err := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
		Message: consts.CommentDeleteExpenseSuccessfully,
		Kind:    bcModel.CommentMsgTypeCompleted,
	}); err != nil {
		l.Error(err, "failed to post expense delete success feedback")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// StoreAccountingTransaction run commpany accouting expense process
func (h *handler) StoreAccountingTransaction(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "StoreAccountingTransaction",
	})

	body := readBody(c.Request.Body)
	provider := h.service.AccountingProvider
	if provider == nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errors.New("accounting provider missing"), nil, ""))
		return
	}
	payload, err := provider.ParseAccountingWebhook(c.Request.Context(), taskprovider.AccountingWebhookRequest{
		Headers: headerMap(c.Request.Header),
		Body:    body,
	})
	if err != nil {
		l.Error(err, "failed to parse basecamp accounting webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload == nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// MarkInvoiceAsPaidViaBasecamp --
func (h *handler) MarkInvoiceAsPaidViaBasecamp(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "MarkInvoiceAsPaidViaBasecamp",
	})

	msg, _, err := basecampWebhookMessageFromCtx(c, l)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if err := h.markInvoiceAsPaid(&msg); err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

func (h *handler) markInvoiceAsPaid(msg *model.BasecampWebhookMessage) error {
	invoice, ref, err := h.GetInvoiceViaBasecampTitle(msg)
	if err != nil {
		h.enqueueInvoiceComment(ref, msg.Recording.Bucket.ID, msg.Recording.ID, err.Error(), bcModel.CommentMsgTypeFailed)
		return err
	}

	if invoice == nil {
		return nil
	}

	if _, err := h.controller.Invoice.MarkInvoiceAsPaidWithTaskRef(invoice, ref, true); err != nil {
		return err
	}

	// log discord as audit log
	_ = h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_paid",
		Data: map[string]interface{}{
			"invoice_number": invoice.Number,
		},
	})

	return nil
}

// ValidateOnLeaveRequest validates on-leave request and give feedback comments
func (h *handler) ValidateOnLeaveRequest(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "ValidateOnLeaveRequest",
	})

	var msg model.BasecampWebhookMessage
	err := msg.Decode(msg.Read(c.Request.Body))
	if err != nil {
		l.Error(err, "decode Basecamp msg failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.handleOnLeaveValidation(msg)
	if err != nil {
		l.Error(err, "onleave validation failed")
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
}

// ApproveOnLeaveRequest saves on-leave request in a database
func (h *handler) ApproveOnLeaveRequest(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "basecamp",
		"method":  "ApproveOnLeaveRequest",
	})

	var msg model.BasecampWebhookMessage
	err := msg.Decode(msg.Read(c.Request.Body))
	if err != nil {
		l.Error(err, "failed to decode basecamp message")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.handleApproveOnLeaveRequest(msg)
	if err != nil {
		l.Error(err, "failed to handle approve on leave request")
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
}

func (h *handler) enqueueInvoiceComment(ref *taskprovider.InvoiceTaskRef, bucketID, todoID int, message, msgType string) {
	if ref != nil && h.service.TaskProvider != nil {
		h.worker.Enqueue(taskprovider.WorkerMessageInvoiceComment, taskprovider.InvoiceCommentJob{
			Ref: ref,
			Input: taskprovider.InvoiceCommentInput{
				Message: message,
				Type:    msgType,
			},
		})
		return
	}

	h.worker.Enqueue(bcModel.BasecampCommentMsg, h.service.Basecamp.BuildCommentMessage(bucketID, todoID, message, msgType))
}
func readBody(rc io.ReadCloser) []byte {
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	return b
}
