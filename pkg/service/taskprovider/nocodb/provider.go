package nocodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	nocodbsvc "github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/expense"
)

// Provider implements taskprovider.InvoiceProvider on top of NocoDB.
type Provider struct {
	svc        *nocodbsvc.Service
	expenseCfg config.ExpenseIntegration

	store *store.Store
	repo  store.DBRepo
	cfg   *config.Config
	wise  wise.IService
	l     logger.Logger
}

// New creates a NocoDB-backed invoice provider.
func New(svc *nocodbsvc.Service, cfg *config.Config, s *store.Store, repo store.DBRepo, wiseService wise.IService, l logger.Logger) *Provider {
	if svc == nil {
		return nil
	}
	var expCfg config.ExpenseIntegration
	if cfg != nil {
		expCfg = cfg.ExpenseIntegration
	}
	return &Provider{
		svc:        svc,
		expenseCfg: expCfg,
		store:      s,
		repo:       repo,
		cfg:        cfg,
		wise:       wiseService,
		l:          l,
	}
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

func (p *Provider) ParseExpenseWebhook(ctx context.Context, req taskprovider.ExpenseWebhookRequest) (*taskprovider.ExpenseWebhookPayload, error) {
	if p == nil || p.svc == nil {
		return nil, errors.New("nocodb provider is not configured")
	}
	if len(req.Body) == 0 {
		return nil, errors.New("empty webhook body")
	}

	var hook nocoExpenseWebhook
	if err := json.Unmarshal(req.Body, &hook); err != nil {
		return nil, err
	}

	if !p.shouldProcessExpenseTable(hook.tableName()) {
		return nil, nil
	}

	row := hook.row()
	if row == nil {
		return nil, nil
	}

	payload := buildExpensePayload(row)
	if payload == nil {
		return nil, nil
	}

	event := hook.eventName()
	status := row.status()
	mappedEvent := mapNocoExpenseEvent(event, status)
	if mappedEvent == "" {
		return nil, nil
	}

	payload.Provider = taskprovider.ProviderNocoDB
	payload.EventType = mappedEvent
	payload.Raw = row.raw
	return payload, nil
}

func (p *Provider) ValidateSubmission(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseValidationResult, error) {
	if payload == nil {
		return nil, errors.New("missing expense payload")
	}
	// NocoDB forms already enforce formatting; no-op validation.
	return &taskprovider.ExpenseValidationResult{Skip: true}, nil
}

func (p *Provider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
	if payload == nil {
		return nil, errors.New("missing expense payload")
	}

	l := p.l.Fields(logger.Fields{
		"method":   "CreateExpense",
		"provider": "nocodb",
		"taskRef":  payload.TaskRef,
	})
	l.Debug("processing expense approval webhook from NocoDB (validation only, no DB persistence)")

	// Validate employee exists by email
	employee, err := p.store.Employee.OneByEmail(p.repo.DB(), strings.TrimSpace(payload.CreatorEmail))
	if err != nil || employee == nil {
		l.Fields(logger.Fields{
			"creatorEmail": payload.CreatorEmail,
		}).Error(err, "validation failed: employee not found for expense")
		if err == nil {
			return nil, fmt.Errorf("employee not found: %s", payload.CreatorEmail)
		}
		return nil, err
	}
	l.Debugf("validation passed: found employee %s for email %s", employee.ID, payload.CreatorEmail)

	// Validate currency (optional - just check it exists)
	currencyName := strings.ToUpper(payload.Currency)
	if currencyName == "" {
		currencyName = "VND"
	}
	curr, err := p.store.Currency.GetByName(p.repo.DB(), currencyName)
	if err != nil {
		l.AddField("currency", currencyName).Error(err, "validation failed: currency not found")
		return nil, fmt.Errorf("currency not found: %s", currencyName)
	}
	l.Debugf("validation passed: found currency %s", curr.Name)

	// Log approval event for debugging
	l.Infof("expense approved in NocoDB - employee: %s, amount: %d %s, reason: %s, task_ref: %s (NO DB PERSISTENCE - will be created on payroll commit)",
		employee.FullName,
		payload.Amount,
		currencyName,
		payload.Reason,
		payload.TaskRef,
	)

	return &taskprovider.ExpenseTaskRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: payload.TaskRef,
		RowID:      payload.TaskRef,
	}, nil
}

func (p *Provider) CompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	_, err := p.CreateExpense(ctx, payload)
	return err
}

func (p *Provider) UncompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	if payload == nil {
		return errors.New("missing expense payload")
	}

	l := p.l.Fields(logger.Fields{
		"method":   "UncompleteExpense",
		"provider": "nocodb",
		"taskRef":  payload.TaskRef,
	})
	l.Debug("uncompleting/deleting expense from nocodb webhook")

	query := &expense.ExpenseQuery{
		TaskProvider: string(taskprovider.ProviderNocoDB),
		TaskRef:      payload.TaskRef,
	}
	e, err := p.store.Expense.GetByQuery(p.repo.DB(), query)
	if err != nil {
		l.Error(err, "failed to find expense for uncomplete")
		return err
	}

	if _, err = p.store.Expense.Delete(p.repo.DB(), e); err != nil {
		l.AddField("expenseID", e.ID).Error(err, "failed to delete expense")
		return err
	}

	l.Infof("successfully deleted expense %s from nocodb task %s", e.ID, payload.TaskRef)
	return nil
}

func (p *Provider) DeleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	return p.UncompleteExpense(ctx, payload)
}

func (p *Provider) PostFeedback(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload, input taskprovider.ExpenseFeedbackInput) error {
	// NocoDB webhooks currently do not support inline comment responses; no-op.
	return nil
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

type nocoExpenseWebhook struct {
	Event   string                  `json:"event"`
	Type    string                  `json:"type"`
	Table   string                  `json:"table"`
	Payload map[string]any          `json:"payload"`
	Data    *nocoExpenseWebhookData `json:"data"`
}

type nocoExpenseWebhookData struct {
	Table string           `json:"table_name"`
	Rows  []map[string]any `json:"rows"`
}

type expenseRow struct {
	raw map[string]any
}

func (w nocoExpenseWebhook) eventName() string {
	if strings.TrimSpace(w.Event) != "" {
		return w.Event
	}
	return w.Type
}

func (w nocoExpenseWebhook) tableName() string {
	if strings.TrimSpace(w.Table) != "" {
		return w.Table
	}
	if w.Data != nil {
		return w.Data.Table
	}
	return ""
}

func (w nocoExpenseWebhook) row() *expenseRow {
	if len(w.Payload) > 0 {
		return &expenseRow{raw: w.Payload}
	}
	if w.Data != nil && len(w.Data.Rows) > 0 {
		return &expenseRow{raw: w.Data.Rows[0]}
	}
	return nil
}

func (r *expenseRow) string(keys ...string) string {
	if r == nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := r.raw[key]; ok {
			switch val := v.(type) {
			case string:
				if strings.TrimSpace(val) != "" {
					return strings.TrimSpace(val)
				}
			case fmt.Stringer:
				return strings.TrimSpace(val.String())
			default:
				s := strings.TrimSpace(fmt.Sprint(val))
				if s != "" && s != "<nil>" {
					return s
				}
			}
		}
	}
	return ""
}

func (r *expenseRow) amount() (int, string) {
	if r == nil {
		return 0, ""
	}
	val, ok := r.raw["amount"]
	if !ok {
		return 0, ""
	}
	raw := strings.TrimSpace(fmt.Sprint(val))
	parsed, err := parseAmount(val)
	if err != nil {
		return 0, raw
	}
	return int(math.Round(parsed)), raw
}

func (r *expenseRow) id() string {
	return r.string("Id", "id", "row_id", "rowId", "record_id")
}

func (r *expenseRow) status() string {
	return strings.ToLower(r.string("status"))
}

func (r *expenseRow) metadata() map[string]any {
	if r == nil {
		return nil
	}
	if v, ok := r.raw["metadata"].(map[string]any); ok {
		return v
	}
	return nil
}

func (r *expenseRow) attachmentURLs() []string {
	urls := []string{}
	add := func(values ...string) {
		for _, val := range values {
			trimmed := strings.TrimSpace(val)
			if trimmed != "" {
				urls = append(urls, trimmed)
			}
		}
	}
	add(r.string("attachment_url", "receipt_url", "file_url"))
	if v, ok := r.raw["attachments"].([]interface{}); ok {
		for _, item := range v {
			switch data := item.(type) {
			case map[string]any:
				add(firstString(data, "url", "signedUrl", "signed_url"))
			case string:
				add(data)
			}
		}
	}
	if v, ok := r.raw["attachment_files"].([]interface{}); ok {
		for _, item := range v {
			switch data := item.(type) {
			case map[string]any:
				add(firstString(data, "url", "signedUrl", "signed_url"))
			case string:
				add(data)
			}
		}
	}
	if v, ok := r.raw["attachment"].(map[string]any); ok {
		add(firstString(v, "url", "signedUrl", "signed_url"))
	}
	return dedupStrings(urls)
}

func (p *Provider) shouldProcessExpenseTable(table string) bool {
	table = strings.ToLower(strings.TrimSpace(table))
	if table == "" {
		return false
	}
	targets := []string{
		strings.ToLower(strings.TrimSpace(p.expenseCfg.Noco.TableID)),
	}
	for _, target := range targets {
		if target == "" {
			continue
		}
		if strings.Contains(table, target) || table == target {
			return true
		}
	}
	return strings.Contains(table, "expense")
}

func buildExpensePayload(row *expenseRow) *taskprovider.ExpenseWebhookPayload {
	if row == nil {
		return nil
	}
	amount, amountRaw := row.amount()
	attachments := row.attachmentURLs()
	primaryAttachment := ""
	if len(attachments) > 0 {
		primaryAttachment = attachments[0]
	}
	reason := firstNonEmpty(row.string("reason"), row.string("title"), row.string("description"))
	if reason == "" {
		reason = fmt.Sprintf("Expense %s", row.id())
	}
	return &taskprovider.ExpenseWebhookPayload{
		Title:             row.string("title"),
		Reason:            reason,
		AmountRaw:         amountRaw,
		Amount:            amount,
		Currency:          strings.ToUpper(row.string("currency")),
		CreatorEmail:      row.string("requester_team_email", "requester_email", "creator_email"),
		CreatorName:       firstNonEmpty(row.string("requester_name"), row.string("creator_name")),
		TaskRef:           row.id(),
		TaskBoard:         firstNonEmpty(row.string("board_label"), row.string("view_name"), row.string("view_id")),
		TaskAttachmentURL: primaryAttachment,
		TaskAttachments:   attachments,
		BucketName:        row.string("task_group", "group"),
		RecordingURL:      row.string("record_url", "share_url"),
		Metadata:          marshalMap(row.metadata()),
	}
}

func mapNocoExpenseEvent(event, status string) taskprovider.ExpenseEventType {
	evt := strings.ToLower(strings.TrimSpace(event))
	status = strings.ToLower(strings.TrimSpace(status))
	switch evt {
	case "row.created", "records.after.insert", "records.before.insert", "row.added":
		return taskprovider.ExpenseEventValidate
	case "row.deleted", "records.after.delete":
		return taskprovider.ExpenseEventUncomplete
	case "row.updated", "records.after.update", "records.after.patch":
		if isClosedExpenseStatus(status) {
			return taskprovider.ExpenseEventCreate
		}
		return taskprovider.ExpenseEventValidate
	default:
		return taskprovider.ExpenseEventType("")
	}
}

func isClosedExpenseStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "completed", "paid", "resolved", "done", "closed":
		return true
	default:
		return false
	}
}

func marshalMap(val interface{}) []byte {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case []byte:
		return v
	case map[string]any:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return b
	default:
		return nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			s := strings.TrimSpace(fmt.Sprint(val))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func dedupStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		unique = append(unique, trimmed)
	}
	return unique
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
