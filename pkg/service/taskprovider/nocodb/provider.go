package nocodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	nocodbsvc "github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

// Provider implements taskprovider.InvoiceProvider on top of NocoDB.
type Provider struct {
	svc *nocodbsvc.Service
}

// New creates a NocoDB-backed invoice provider.
func New(svc *nocodbsvc.Service) *Provider {
	if svc == nil {
		return nil
	}
	return &Provider{svc: svc}
}

func (p *Provider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderNocoDB
}

func (p *Provider) EnsureTask(ctx context.Context, input taskprovider.CreateInvoiceTaskInput) (*taskprovider.InvoiceTaskRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if input.Invoice == nil {
		return nil, errors.New("missing invoice data")
	}
	payload := buildInvoicePayload(input.Invoice)
	id, err := p.svc.UpsertInvoiceRecord(ctx, input.Invoice.Number, payload)
	if err != nil {
		return nil, err
	}
	return &taskprovider.InvoiceTaskRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: id,
	}, nil
}

func (p *Provider) UploadAttachment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceAttachmentInput) (*taskprovider.InvoiceAttachmentRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if len(input.Content) == 0 {
		return nil, errors.New("missing attachment content")
	}
	fileName := strings.TrimSpace(input.FileName)
	if fileName == "" {
		fileName = fmt.Sprintf("invoice-%d.pdf", time.Now().Unix())
	}
	result, err := p.svc.UploadInvoiceAttachment(ctx, fileName, input.ContentType, input.Content)
	if err != nil {
		return nil, err
	}
	url := result.AccessibleURL(p.svc.BaseURL())
	if url == "" {
		url = input.URL
	}
	if url == "" {
		url = fmt.Sprintf("https://storage.googleapis.com/%s", fileName)
	}
	markup := fmt.Sprintf("[Invoice PDF](%s)", url)
	return &taskprovider.InvoiceAttachmentRef{
		ExternalID: url,
		Markup:     markup,
		Meta:       result.ToMap(),
	}, nil
}

func (p *Provider) PostComment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceCommentInput) error {
	if p == nil || p.svc == nil {
		return errors.New("nocodb provider is not configured")
	}
	if ref == nil || ref.ExternalID == "" {
		return errors.New("missing invoice reference")
	}
	return p.svc.CreateInvoiceComment(ctx, ref.ExternalID, "system", input.Message, input.Type)
}

func (p *Provider) CompleteTask(ctx context.Context, ref *taskprovider.InvoiceTaskRef) error {
	if p == nil || p.svc == nil {
		return errors.New("nocodb provider is not configured")
	}
	if ref == nil || ref.ExternalID == "" {
		return errors.New("missing invoice reference")
	}
	return p.svc.UpdateInvoiceStatus(ctx, ref.ExternalID, string(model.InvoiceStatusPaid))
}

func (p *Provider) CreateMonthlyPlan(ctx context.Context, input taskprovider.CreateAccountingPlanInput) (*taskprovider.AccountingPlanRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if p.svc.AccountingTodosTableID() == "" {
		return nil, errors.New("nocodb accounting todos table is not configured")
	}
	if input.Label == "" {
		input.Label = fmt.Sprintf("Accounting | %d-%d", input.Month, input.Year)
	}
	return &taskprovider.AccountingPlanRef{
		Provider: taskprovider.ProviderNocoDB,
		BoardID:  p.svc.AccountingTodosTableID(),
		ListID:   input.Label,
		Month:    input.Month,
		Year:     input.Year,
		GroupLookup: map[taskprovider.AccountingGroup]string{
			taskprovider.AccountingGroupIn:  string(taskprovider.AccountingGroupIn),
			taskprovider.AccountingGroupOut: string(taskprovider.AccountingGroupOut),
		},
	}, nil
}

func (p *Provider) CreateAccountingTodo(ctx context.Context, plan *taskprovider.AccountingPlanRef, input taskprovider.CreateAccountingTodoInput) (*taskprovider.AccountingTodoRef, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if plan == nil || plan.BoardID == "" {
		return nil, errors.New("missing accounting plan reference")
	}
	group := string(input.Group)
	if plan.GroupLookup != nil {
		if mapped, ok := plan.GroupLookup[input.Group]; ok {
			group = mapped
		}
	}
	due := input.DueDate
	if due.IsZero() {
		due = time.Now()
	}
	payload := map[string]interface{}{
		"board_label":  plan.ListID,
		"task_group":   group,
		"title":        input.Title,
		"description":  input.Description,
		"due_on":       due.Format("2006-01-02"),
		"assignee_ids": assigneeIDs(input.Assignees),
		"status":       "open",
	}
	if len(input.Metadata) > 0 {
		payload["metadata"] = input.Metadata
	}
	if input.Group == taskprovider.AccountingGroupIn {
		payload["type"] = "income"
	} else {
		payload["type"] = "expense"
	}
	id, err := p.svc.CreateAccountingTodo(ctx, payload)
	if err != nil {
		return nil, err
	}
	return &taskprovider.AccountingTodoRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: id,
		Group:      input.Group,
	}, nil
}

func (p *Provider) ParseAccountingWebhook(ctx context.Context, req taskprovider.AccountingWebhookRequest) (*taskprovider.AccountingWebhookPayload, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if len(req.Body) == 0 {
		return nil, errors.New("empty webhook body")
	}

	var hook nocoAccountingWebhook
	if err := json.Unmarshal(req.Body, &hook); err != nil {
		return nil, err
	}

	row := hook.rowData()
	if row == nil {
		return nil, errors.New("missing row data")
	}

	amountRaw := strings.TrimSpace(fmt.Sprint(row.Amount))
	currencyRaw := strings.TrimSpace(row.Currency)
	var amount float64
	if amountRaw != "" {
		var err error
		amount, err = parseAmount(row.Amount)
		if err != nil {
			return nil, err
		}
	}

	group := toAccountingGroup(row.groupValue())
	todoRowID := row.TodoRowID
	if todoRowID == "" {
		todoRowID = fmt.Sprintf("%v", row.ID)
	}
	if todoRowID == "" {
		todoRowID = hook.RowID
	}

	meta := map[string]any{}
	if row.Metadata != nil {
		meta = row.Metadata
	}
	if row.BoardLabel != "" {
		meta["board_label"] = row.BoardLabel
	}
	if row.InvoiceID != "" {
		meta["invoice_id"] = row.InvoiceID
	}
	if row.InvoiceNumber != "" {
		meta["invoice_number"] = row.InvoiceNumber
	}
	if row.InvoiceTaskID != nil {
		meta["invoice_task_id"] = row.InvoiceTaskID
	}

	return &taskprovider.AccountingWebhookPayload{
		Provider:  taskprovider.ProviderNocoDB,
		Group:     group,
		Title:     row.Title,
		Amount:    amount,
		Currency:  strings.ToUpper(currencyRaw),
		TodoID:    hook.RowID,
		TodoRowID: todoRowID,
		Actor:     hook.ActorName(),
		Status:    row.Status,
		Raw:       req.Body,
		Metadata:  meta,
	}, nil
}

func buildInvoicePayload(iv *model.Invoice) map[string]interface{} {
	currency := ""
	if iv.Bank != nil && iv.Bank.Currency != nil {
		currency = iv.Bank.Currency.Name
	} else if iv.Project != nil && iv.Project.BankAccount != nil && iv.Project.BankAccount.Currency != nil {
		currency = iv.Project.BankAccount.Currency.Name
	}
	payload := map[string]interface{}{
		"invoice_number":      iv.Number,
		"month":               iv.Month,
		"year":                iv.Year,
		"status":              string(iv.Status),
		"amount":              iv.Total,
		"currency":            currency,
		"fortress_invoice_id": iv.ID.String(),
	}
	if attachments := invoiceAttachmentPayload(iv); len(attachments) > 0 {
		payload["attachment_url"] = attachments
	}
	return payload
}

type nocoAccountingWebhook struct {
	TableID     string             `json:"tableId"`
	RowID       string             `json:"rowId"`
	Event       string             `json:"event"`
	Type        string             `json:"type"`
	TriggeredBy nocoTrigger        `json:"triggeredBy"`
	New         *nocoAccountingRow `json:"new"`
	Data        *nocoWebhookData   `json:"data"`
}

type nocoWebhookData struct {
	Table string           `json:"table_name"`
	Rows  []map[string]any `json:"rows"`
}

type nocoTrigger struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (n nocoAccountingWebhook) ActorName() string {
	if n.TriggeredBy.Name != "" {
		return n.TriggeredBy.Name
	}
	return n.TriggeredBy.Email
}

func (n nocoAccountingWebhook) rowData() *nocoAccountingRow {
	if n.New != nil {
		return n.New
	}
	if n.Data != nil && len(n.Data.Rows) > 0 {
		record := n.Data.Rows[0]
		row := &nocoAccountingRow{}
		bytes, err := json.Marshal(record)
		if err != nil {
			return row
		}
		if err := json.Unmarshal(bytes, row); err != nil {
			return row
		}
		return row
	}
	return nil
}

type nocoAccountingRow struct {
	ID            interface{}    `json:"id"`
	TodoRowID     string         `json:"todo_row_id"`
	BoardLabel    string         `json:"board_label"`
	Group         string         `json:"group"`
	TaskGroup     string         `json:"task_group"`
	Title         string         `json:"title"`
	Amount        interface{}    `json:"amount"`
	Currency      string         `json:"currency"`
	Status        string         `json:"status"`
	Metadata      map[string]any `json:"metadata"`
	InvoiceID     string         `json:"invoice_id"`
	InvoiceNumber string         `json:"invoice_number"`
	InvoiceTaskID interface{}    `json:"invoice_task_id"`
}

func (r *nocoAccountingRow) groupValue() string {
	if r == nil {
		return ""
	}
	if strings.TrimSpace(r.TaskGroup) != "" {
		return r.TaskGroup
	}
	return r.Group
}

func parseAmount(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return 0, nil
		}
		return strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported amount type %T", v)
	}
}

func toAccountingGroup(value string) taskprovider.AccountingGroup {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "in", "income":
		return taskprovider.AccountingGroupIn
	default:
		return taskprovider.AccountingGroupOut
	}
}

func assigneeIDs(items []taskprovider.AccountingAssignee) []string {
	out := make([]string, 0, len(items))
	for _, a := range items {
		if a.ExternalID != "" {
			out = append(out, a.ExternalID)
			continue
		}
		if a.Email != "" {
			out = append(out, a.Email)
		}
	}
	return out
}

func invoiceAttachmentPayload(iv *model.Invoice) []map[string]any {
	if iv == nil {
		return nil
	}
	if len(iv.InvoiceAttachmentMeta) > 0 {
		return []map[string]any{cloneAttachmentMeta(iv.InvoiceAttachmentMeta)}
	}
	if strings.TrimSpace(iv.InvoiceFileURL) == "" {
		return nil
	}
	return []map[string]any{
		{
			"url":   iv.InvoiceFileURL,
			"title": filepath.Base(iv.InvoiceFileURL),
		},
	}
}

func cloneAttachmentMeta(meta map[string]any) map[string]any {
	out := make(map[string]any, len(meta))
	for k, v := range meta {
		out[k] = v
	}
	return out
}
