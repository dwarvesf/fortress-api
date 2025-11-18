package webhook

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) HandleNocoExpense(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "HandleNocoExpense",
	})

	secret := h.config.ExpenseIntegration.Noco.WebhookSecret
	if secret == "" {
		l.Error(errors.New("missing nocodb expense webhook secret"), "cannot verify expense webhook")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("signature verification disabled"), nil, ""))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read nocodb expense webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	signature := c.GetHeader("X-NocoDB-Signature")
	authHeader := c.GetHeader("Authorization")
	if !verifyNocoSignature(secret, signature, authHeader, body) {
		l.Error(errors.New("invalid signature"), "nocodb expense signature mismatch")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
		return
	}

	provider := h.service.ExpenseProvider
	if provider == nil || provider.Type() != taskprovider.ProviderNocoDB {
		err := errors.New("expense provider is not configured for nocodb")
		l.Error(err, "invalid expense provider setup")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload, err := provider.ParseExpenseWebhook(c.Request.Context(), taskprovider.ExpenseWebhookRequest{
		Headers: headerMap(c.Request.Header),
		Body:    body,
	})
	if err != nil {
		l.Error(err, "failed to parse nocodb expense webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload == nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	switch payload.EventType {
	case taskprovider.ExpenseEventValidate:
		result, err := provider.ValidateSubmission(c.Request.Context(), payload)
		if err != nil {
			l.Error(err, "expense validation failed")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		if result != nil && !result.Skip && result.Message != "" {
			if postErr := provider.PostFeedback(c.Request.Context(), payload, taskprovider.ExpenseFeedbackInput{
				Message: result.Message,
				Kind:    result.FeedbackKind,
			}); postErr != nil {
				l.Error(postErr, "failed to post nocodb expense feedback")
			}
		}
	case taskprovider.ExpenseEventCreate:
		if _, err := provider.CreateExpense(c.Request.Context(), payload); err != nil {
			l.Error(err, "failed to create nocodb expense record")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	case taskprovider.ExpenseEventComplete:
		if err := provider.CompleteExpense(c.Request.Context(), payload); err != nil {
			l.Error(err, "failed to complete nocodb expense record")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	case taskprovider.ExpenseEventUncomplete:
		if err := provider.UncompleteExpense(c.Request.Context(), payload); err != nil {
			l.Error(err, "failed to uncomplete nocodb expense record")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
	default:
		l.Infof("ignore nocodb expense event %s", payload.EventType)
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}
