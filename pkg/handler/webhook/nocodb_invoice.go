package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	sInvoice "github.com/dwarvesf/fortress-api/pkg/store/invoice"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type nocoInvoiceWebhook struct {
	Event   string                `json:"event"`
	Type    string                `json:"type"`
	Table   string                `json:"table"`
	Payload nocoInvoiceRowPayload `json:"payload"`
	Data    *nocoWebhookData      `json:"data"`
}

type nocoWebhookData struct {
	Table string           `json:"table_name"`
	Rows  []map[string]any `json:"rows"`
}

type nocoInvoiceRowPayload struct {
	ID                interface{} `json:"id"`
	InvoiceNumber     string      `json:"invoice_number"`
	FortressInvoiceID string      `json:"fortress_invoice_id"`
	Status            string      `json:"status"`
}

func (p nocoInvoiceRowPayload) recordID() string {
	switch v := p.ID.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (h *handler) MarkInvoiceAsPaidViaNoco(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "webhook",
		"method":  "MarkInvoiceAsPaidViaNoco",
	})

	secret := h.config.Noco.WebhookSecret
	if secret == "" {
		l.Error(errors.New("missing nocodb webhook secret"), "cannot verify webhook")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("signature verification disabled"), nil, ""))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		l.Error(err, "failed to read webhook body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	signature := c.GetHeader("X-NocoDB-Signature")
	authz := c.GetHeader("Authorization")
	if !verifyNocoSignature(secret, signature, authz, body) {
		l.Error(errors.New("invalid signature"), "nocodb signature mismatch")
		c.JSON(http.StatusUnauthorized, view.CreateResponse[any](nil, nil, errors.New("invalid signature"), nil, ""))
		return
	}

	var req nocoInvoiceWebhook
	if err := json.Unmarshal(body, &req); err != nil {
		l.Error(err, "failed to parse nocodb payload")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	payload := req.Payload
	table := req.Table
	if req.Data != nil {
		table = req.Data.Table
		if len(req.Data.Rows) > 0 {
			payload = extractRowPayload(req.Data.Rows[0])
		}
	}

	eventName := req.Event
	if eventName == "" {
		eventName = req.Type
	}

	logFields := logger.Fields{
		"event":             eventName,
		"table":             table,
		"invoiceNumber":     payload.InvoiceNumber,
		"fortressInvoiceID": payload.FortressInvoiceID,
		"status":            payload.Status,
		"recordID":          payload.recordID(),
		"rawPayload":        string(body),
	}
	if req.Data != nil {
		logFields["dataTable"] = req.Data.Table
		if len(req.Data.Rows) > 0 {
			logFields["dataRow"] = req.Data.Rows[0]
		}
	}
	l.Fields(logFields).Info("received nocodb invoice webhook")

	if !strings.EqualFold(table, "invoice_tasks") {
		l.Infof("ignore nocodb webhook - table %s", table)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if eventName != "row.updated" && eventName != "records.after.update" && eventName != "records.after.patch" {
		l.Infof("ignore nocodb event %s", eventName)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	if !strings.EqualFold(payload.Status, string(model.InvoiceStatusPaid)) {
		l.Infof("ignore nocodb status %s", payload.Status)
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	invoice, err := h.findInvoiceByNocoPayload(payload)
	if err != nil {
		l.Error(err, "failed to locate invoice for nocodb webhook")
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	if invoice == nil {
		c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
		return
	}

	ref := &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: payload.recordID(),
	}

	l.Fields(logger.Fields{
		"invoiceID": invoice.ID,
		"status":    invoice.Status,
		"recordID":  ref.ExternalID,
	}).Info("process nocodb invoice payment")

	if _, err := h.controller.Invoice.MarkInvoiceAsPaidWithTaskRef(invoice, ref, true); err != nil {
		l.Error(err, "failed to mark invoice as paid via nocodb")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l.Infof("invoice %s marked paid via nocodb", invoice.Number)

	// log discord as audit log
	if err := h.controller.Discord.Log(model.LogDiscordInput{
		Type: "invoice_paid",
		Data: map[string]interface{}{
			"invoice_number": invoice.Number,
		},
	}); err != nil {
		l.Error(err, "failed to log invoice paid to discord")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, ""))
}

func (h *handler) findInvoiceByNocoPayload(p nocoInvoiceRowPayload) (*model.Invoice, error) {
	if id := strings.TrimSpace(p.FortressInvoiceID); id != "" {
		invoice, err := h.store.Invoice.One(h.repo.DB(), &sInvoice.Query{ID: id})
		if err == nil {
			return invoice, nil
		}
	}
	if p.InvoiceNumber == "" {
		return nil, errors.New("missing invoice number")
	}
	return h.store.Invoice.One(h.repo.DB(), &sInvoice.Query{Number: p.InvoiceNumber})
}

func verifyNocoSignature(secret, signature, authorization string, body []byte) bool {
	if secret == "" {
		return false
	}

	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(authorization)), "bearer ") {
		token := strings.TrimSpace(authorization[len("Bearer "):])
		if token == secret {
			return true
		}
	}

	if signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expected)))
}

func extractRowPayload(row map[string]any) nocoInvoiceRowPayload {
	payload := nocoInvoiceRowPayload{}
	if row == nil {
		return payload
	}
	if v, ok := row["Id"]; ok {
		payload.ID = v
	} else if v, ok := row["id"]; ok {
		payload.ID = v
	}
	if v, ok := row["invoice_number"].(string); ok {
		payload.InvoiceNumber = v
	}
	if v, ok := row["fortress_invoice_id"].(string); ok {
		payload.FortressInvoiceID = v
	}
	if v, ok := row["status"].(string); ok {
		payload.Status = v
	}
	return payload
}
