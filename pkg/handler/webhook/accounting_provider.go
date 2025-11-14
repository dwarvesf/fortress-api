package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func (h *handler) StoreNocoAccountingTransaction(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "StoreNocoAccountingTransaction",
	})

	secret := h.config.Noco.WebhookSecret
	if secret == "" {
		l.Error(errors.New("missing nocodb webhook secret"), "cannot verify accounting webhook")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("webhook disabled"), nil, ""))
		return
	}

	auth := c.GetHeader("Authorization")
	if !validateBearer(auth, secret) {
		l.Error(errors.New("invalid nocodb bearer token"), "unauthorized")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid token"), nil, ""))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	// allow downstream readers if needed
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	provider := h.service.AccountingProvider
	if provider == nil || provider.Type() != taskprovider.ProviderNocoDB {
		err := errors.New("accounting provider is not configured for NocoDB")
		l.Error(err, "invalid provider setup")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload, err := provider.ParseAccountingWebhook(
		c.Request.Context(),
		taskprovider.AccountingWebhookRequest{
			Headers: headerMap(c.Request.Header),
			Body:    body,
		},
	)
	if err != nil {
		l.Error(err, "failed to parse nocodb accounting webhook")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if payload != nil {
		l.Fields(logger.Fields{
			"todoRowID": payload.TodoRowID,
			"group":     payload.Group,
			"status":    payload.Status,
		}).Info("parsed nocodb accounting webhook")
	}
	if payload == nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if !strings.EqualFold(payload.Status, "completed") {
		l.Infof("ignore nocodb accounting event with status %s", payload.Status)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if payload.Group == taskprovider.AccountingGroupIn {
		if err := h.syncInvoiceStateFromAccounting(c.Request.Context(), payload); err != nil {
			l.Error(err, "failed to sync invoice state from accounting event")
		} else {
			l.Info("synced invoice state from accounting event")
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

func validateBearer(header, secret string) bool {
	if header == "" || secret == "" {
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	return strings.TrimSpace(header[len(prefix):]) == secret
}

func headerMap(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, values := range h {
		if len(values) > 0 {
			out[k] = values[0]
		}
	}
	return out
}

func (h *handler) syncInvoiceStateFromAccounting(ctx context.Context, payload *taskprovider.AccountingWebhookPayload) error {
	if payload == nil {
		return errors.New("missing accounting payload")
	}
	meta := payload.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	todoIDs := []string{}
	if payload.TodoRowID != "" {
		todoIDs = append(todoIDs, payload.TodoRowID)
	}
	if payload.TodoID != "" {
		todoIDs = append(todoIDs, payload.TodoID)
	}
	for _, todoID := range todoIDs {
		if todoID == "" || h.service == nil || h.service.NocoDB == nil {
			continue
		}
		row, err := h.service.NocoDB.GetAccountingTodo(ctx, todoID)
		if err == nil {
			h.logger.Fields(logger.Fields{"todoRowID": todoID}).Info("loaded accounting todo for invoice sync")
			meta = mergeMetadata(meta, row)
			break
		}
		h.logger.Errorf(err, "failed to load accounting todo", "todoRowID", todoID)
	}
	invoiceTaskID := metaString(meta, "invoice_task_id")
	if invoiceTaskID == "" {
		return errors.New("no invoice linkage in accounting metadata")
	}
	if h.service == nil || h.service.NocoDB == nil {
		return errors.New("nocodb service is not configured")
	}
	if err := h.service.NocoDB.UpdateInvoiceStatus(ctx, invoiceTaskID, string(model.InvoiceStatusPaid)); err != nil {
		return fmt.Errorf("update nocodb invoice status: %w", err)
	}
	return nil
}

func metaString(meta map[string]any, key string) string {
	if meta == nil {
		return ""
	}
	val, ok := meta[key]
	if !ok || val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func mergeMetadata(dst map[string]any, row map[string]interface{}) map[string]any {
	if dst == nil {
		dst = map[string]any{}
	}
	if row == nil {
		return dst
	}
	raw, ok := row["metadata"]
	if ok && raw != nil {
		switch val := raw.(type) {
		case map[string]any:
			for k, v := range val {
				dst[k] = v
			}
		case string:
			if strings.TrimSpace(val) != "" {
				var decoded map[string]any
				if err := json.Unmarshal([]byte(val), &decoded); err == nil {
					for k, v := range decoded {
						dst[k] = v
					}
				}
			}
		}
	}
	if _, ok := dst["invoice_number"]; !ok {
		if title, ok := row["title"].(string); ok {
			dst["title"] = title
		}
	}
	return dst
}
