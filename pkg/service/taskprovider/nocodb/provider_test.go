package nocodb

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
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
		context.TODO(),
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
		context.TODO(),
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
func TestParseExpenseWebhook_RowCreated(t *testing.T) {
	provider := &Provider{
		svc: &nocodbsvc.Service{},
		expenseCfg: config.ExpenseIntegration{
			Noco: config.ExpenseNocoIntegration{TableID: "tbl_expenses"},
		},
	}
	body := []byte(`{
		"table": "tbl_expenses",
		"event": "row.created",
		"payload": {
			"id": "row_1",
			"title": "Laptop | 1,200 | USD",
			"amount": 1200,
		"currency": "usd",
		"status": "pending",
		"requester_team_email": "alice@example.com",
			"attachments": [{"url": "https://nocodb/files/row_1.pdf"}],
			"metadata": {"project": "apollo"}
		}
	}`)

	payload, err := provider.ParseExpenseWebhook(context.Background(), taskprovider.ExpenseWebhookRequest{Body: body})
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, taskprovider.ProviderNocoDB, payload.Provider)
	require.Equal(t, taskprovider.ExpenseEventValidate, payload.EventType)
	require.Equal(t, "Laptop | 1,200 | USD", payload.Title)
	require.Equal(t, 1200, payload.Amount)
	require.Equal(t, "USD", payload.Currency)
	require.Equal(t, "https://nocodb/files/row_1.pdf", payload.TaskAttachmentURL)
	require.Equal(t, []string{"https://nocodb/files/row_1.pdf"}, payload.TaskAttachments)
	require.Equal(t, "alice@example.com", payload.CreatorEmail)
	var meta map[string]any
	require.NoError(t, json.Unmarshal(payload.Metadata, &meta))
	require.Equal(t, "apollo", meta["project"])
}

func TestParseExpenseWebhook_RowUpdatedApproved(t *testing.T) {
	provider := &Provider{
		svc:        &nocodbsvc.Service{},
		expenseCfg: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{TableID: "tbl_expenses"}},
	}
	body := []byte(`{
		"table": "tbl_expenses",
		"event": "row.updated",
		"payload": {
			"id": "row_2",
			"title": "Taxi | 200 | USD",
			"amount": 200,
			"currency": "usd",
			"status": "approved"
		}
	}`)

	payload, err := provider.ParseExpenseWebhook(context.Background(), taskprovider.ExpenseWebhookRequest{Body: body})
	require.NoError(t, err)
	require.NotNil(t, payload)
	require.Equal(t, taskprovider.ExpenseEventCreate, payload.EventType)
	require.Equal(t, 200, payload.Amount)
	require.Equal(t, "row_2", payload.TaskRef)
}

func TestParseExpenseWebhook_FixturePayloads(t *testing.T) {
	provider := &Provider{
		svc:        &nocodbsvc.Service{},
		expenseCfg: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{TableID: "expense_submissions"}},
	}

	tests := []struct {
		name      string
		file      string
		eventType taskprovider.ExpenseEventType
		assertFn  func(t *testing.T, payload *taskprovider.ExpenseWebhookPayload)
	}{
		{
			name:      "row created validates data",
			file:      "expense_created.json",
			eventType: taskprovider.ExpenseEventValidate,
			assertFn: func(t *testing.T, payload *taskprovider.ExpenseWebhookPayload) {
				require.Equal(t, "Monitor Purchase", payload.Title)
				require.Equal(t, 250, payload.Amount)
				require.Equal(t, "USD", payload.Currency)
				require.Equal(t, "alice@example.com", payload.CreatorEmail)
				require.Equal(t, "row_created_1", payload.TaskRef)
				require.Equal(t, "https://nocodb.example.com/files/monitor.pdf", payload.TaskAttachmentURL)
				require.Equal(t, []string{"https://nocodb.example.com/files/monitor.pdf"}, payload.TaskAttachments)
			},
		},
		{
			name:      "status approved creates expense",
			file:      "expense_updated_completed.json",
			eventType: taskprovider.ExpenseEventCreate,
			assertFn: func(t *testing.T, payload *taskprovider.ExpenseWebhookPayload) {
				require.Equal(t, "row_update_2", payload.TaskRef)
				require.Equal(t, 200, payload.Amount)
				require.Equal(t, "VND", payload.Currency)
			},
		},
		{
			name:      "row deleted triggers uncomplete",
			file:      "expense_deleted.json",
			eventType: taskprovider.ExpenseEventUncomplete,
			assertFn: func(t *testing.T, payload *taskprovider.ExpenseWebhookPayload) {
				require.Equal(t, "row_deleted_3", payload.TaskRef)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := mustLoadExpenseFixture(t, tt.file)
			payload, err := provider.ParseExpenseWebhook(context.Background(), taskprovider.ExpenseWebhookRequest{Body: body})
			require.NoError(t, err)
			require.NotNil(t, payload)
			require.Equal(t, tt.eventType, payload.EventType)
			require.Equal(t, taskprovider.ProviderNocoDB, payload.Provider)
			tt := tt
			if tt.assertFn != nil {
				tt.assertFn(t, payload)
			}
		})
	}
}

func TestBuildExpenseDataFromPayload(t *testing.T) {
	payload := &taskprovider.ExpenseWebhookPayload{
		Reason:            "Laptop Purchase",
		Amount:            1500,
		Currency:          "USD",
		CreatorEmail:      "alice@example.com",
		TaskRef:           "row_3",
		TaskAttachmentURL: "https://nocodb/files/row_3.pdf",
		TaskBoard:         "Expenses | Nov",
		Metadata:          []byte(`{"project":"apollo"}`),
	}
	// buildExpenseDataFromPayload function was removed - test no longer valid
	_ = payload
}

func mustLoadExpenseFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}
