package webhook

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func basecampWebhookMessageFromCtx(c *gin.Context) (model.BasecampWebhookMessage, error) {
	var msg model.BasecampWebhookMessage
	err := msg.Decode(msg.Read(c.Request.Body))
	if err != nil {
		return msg, err
	}
	return msg, nil
}

// BasecampExpenseValidate dry-run expense request for validation
func (h *handler) BasecampExpenseValidate(c *gin.Context) {
	msg, err := basecampWebhookMessageFromCtx(c)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	err = h.BasecampExpenseValidateHandler(msg)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// // BasecampExpense runs expense process in basecamp
func (h *handler) BasecampExpense(c *gin.Context) {
	msg, err := basecampWebhookMessageFromCtx(c)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	err = h.BasecampExpenseHandler(msg, msg.Read(c.Request.Body))
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// UncheckBasecampExpense will remove expesne record after expense todo complete
func (h *handler) UncheckBasecampExpense(c *gin.Context) {
	msg, err := basecampWebhookMessageFromCtx(c)
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	err = h.UncheckBasecampExpenseHandler(msg, msg.Read(c.Request.Body))
	if err != nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// StoreAccountingTransaction run commpany accouting expense process
func (h *handler) StoreAccountingTransaction(c *gin.Context) {
	msg, err := basecampWebhookMessageFromCtx(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.StoreAccountingTransactionFromBasecamp(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

// MarkInvoiceAsPaidViaBasecamp --
func (h *handler) MarkInvoiceAsPaidViaBasecamp(c *gin.Context) {
	msg, err := basecampWebhookMessageFromCtx(c)
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
	invoice, err := h.GetInvoiceViaBasecampTitle(msg)
	if err != nil {
		h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildFailedComment(err.Error()))
		return err
	}

	if invoice == nil {
		return nil
	}

	// if _, err := h.controller.Invoice.MarkInvoiceAsPaid(invoice, true); err != nil {
	if _, err := h.controller.Invoice.MarkInvoiceAsPaidByBasecampWebhookMessage(invoice, msg); err != nil {
		h.service.Basecamp.CommentResult(msg.Recording.Bucket.ID, msg.Recording.ID, h.service.Basecamp.BuildFailedComment(err.Error()))
		return err
	}

	return nil
}
