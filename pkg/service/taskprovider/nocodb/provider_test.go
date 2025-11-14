package nocodb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/model"
	nocodbsvc "github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

func TestParseAccountingWebhook_Noco(t *testing.T) {
	body := []byte(`{
		"tableId": "tbl_accounting_transactions",
		"rowId": "row_123",
		"event": "row.updated",
		"triggeredBy": {"name": "Giang Than", "email": "giang@example.com"},
	"new": {
		"id": "row_tx_1",
		"todo_row_id": "row_todo_456",
		"board_label": "Accounting | January 2025",
		"task_group": "out",
			"title": "Office Rental | 1.500.000 | VND",
			"amount": 1500000,
			"currency": "VND",
			"status": "completed",
			"metadata": {"month": 1, "year": 2025}
		}
	}`)

	provider := &Provider{svc: &nocodbsvc.Service{}}
	payload, err := provider.ParseAccountingWebhook(
		nil,
		taskprovider.AccountingWebhookRequest{Body: body},
	)
	require.NoError(t, err)
	require.Equal(t, taskprovider.ProviderNocoDB, payload.Provider)
	require.Equal(t, taskprovider.AccountingGroupOut, payload.Group)
	require.Equal(t, float64(1500000), payload.Amount)
	require.Equal(t, "row_todo_456", payload.TodoRowID)
	require.Equal(t, "completed", payload.Status)

	meta := map[string]any{}
	require.NoError(t, json.Unmarshal(mustJSON(payload.Metadata), &meta))
	require.Equal(t, float64(1), meta["month"])
	require.Equal(t, float64(2025), meta["year"])
}

func TestParseAccountingWebhook_Noco_NoAmount(t *testing.T) {
	body := []byte(`{
		"tableId": "tbl_accounting_todos",
		"rowId": "row_789",
		"event": "row.updated",
		"triggeredBy": {"name": "Ops", "email": "ops@example.com"},
	"new": {
		"id": "row_todo_789",
		"todo_row_id": "row_todo_789",
		"board_label": "Accounting | December 2025",
		"task_group": "in",
			"title": "Yolo Lab 11/2025",
			"status": "completed",
			"metadata": {"invoice_number": "2025111-YOLO-015"}
		}
	}`)

	provider := &Provider{svc: &nocodbsvc.Service{}}
	payload, err := provider.ParseAccountingWebhook(
		nil,
		taskprovider.AccountingWebhookRequest{Body: body},
	)
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, taskprovider.AccountingGroupIn, payload.Group)
	require.Equal(t, float64(0), payload.Amount)
	require.Equal(t, "", payload.Currency)
}

func mustJSON(data map[string]any) []byte {
	buf, _ := json.Marshal(data)
	return buf
}

func TestBuildInvoicePayload_Attachment(t *testing.T) {
	baseInvoice := &model.Invoice{
		BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("11111111-1111-1111-1111-111111111111")},
		Number:    "2025-INV-01",
		Month:     11,
		Year:      2025,
		Status:    model.InvoiceStatusSent,
	}
	baseInvoice.InvoiceAttachmentMeta = map[string]any{"url": "https://nocodb/att.pdf", "title": "att.pdf"}
	payload := buildInvoicePayload(baseInvoice)
	attachments, ok := payload["attachment_url"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, attachments, 1)
	require.Equal(t, "https://nocodb/att.pdf", attachments[0]["url"])

	t.Run("fallback to url", func(t *testing.T) {
		iv := &model.Invoice{
			BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("22222222-2222-2222-2222-222222222222")},
			Number:    "2025-INV-02",
			Month:     11,
			Year:      2025,
			Status:    model.InvoiceStatusSent,
		}
		iv.InvoiceFileURL = "https://storage.googleapis.com/file.pdf"
		payload := buildInvoicePayload(iv)
		attachments, ok := payload["attachment_url"].([]map[string]any)
		require.True(t, ok)
		require.Equal(t, "https://storage.googleapis.com/file.pdf", attachments[0]["url"])
	})
}
