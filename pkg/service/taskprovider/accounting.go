package taskprovider

import (
	"context"
	"time"
)

// AccountingGroup labels the logical bucket a todo belongs to.
type AccountingGroup string

const (
	// AccountingGroupIn corresponds to revenue/income items.
	AccountingGroupIn AccountingGroup = "in"
	// AccountingGroupOut corresponds to expense/payment items.
	AccountingGroupOut AccountingGroup = "out"
)

// CreateAccountingPlanInput describes the monthly board that needs to be created.
type CreateAccountingPlanInput struct {
	Month    int
	Year     int
	Label    string
	Metadata map[string]any
}

// AccountingPlanRef holds identifiers for the created monthly board and groups.
type AccountingPlanRef struct {
	Provider    ProviderType
	BoardID     string
	ListID      string
	Month       int
	Year        int
	GroupLookup map[AccountingGroup]string
}

// AccountingAssignee references an upstream user/contact record.
type AccountingAssignee struct {
	ExternalID string
	Email      string
}

// CreateAccountingTodoInput encapsulates metadata required for an accounting todo.
type CreateAccountingTodoInput struct {
	Group       AccountingGroup
	Title       string
	Description string
	DueDate     time.Time
	Assignees   []AccountingAssignee
	Metadata    map[string]any
}

// AccountingTodoRef stores provider-specific identifiers for created todos.
type AccountingTodoRef struct {
	Provider   ProviderType
	ExternalID string
	Group      AccountingGroup
}

// AccountingWebhookRequest captures the raw webhook payload for parsing.
type AccountingWebhookRequest struct {
	Headers map[string]string
	Body    []byte
}

// AccountingWebhookPayload is the normalized structure returned by providers.
type AccountingWebhookPayload struct {
	Provider  ProviderType
	Group     AccountingGroup
	Title     string
	Amount    float64
	Currency  string
	TodoID    string
	TodoRowID string
	Actor     string
	Status    string
	Raw       []byte
	Metadata  map[string]any
}

// AccountingProvider exposes methods required to manage accounting todos via providers.
type AccountingProvider interface {
	Type() ProviderType
	CreateMonthlyPlan(ctx context.Context, input CreateAccountingPlanInput) (*AccountingPlanRef, error)
	CreateAccountingTodo(ctx context.Context, plan *AccountingPlanRef, input CreateAccountingTodoInput) (*AccountingTodoRef, error)
	ParseAccountingWebhook(ctx context.Context, req AccountingWebhookRequest) (*AccountingWebhookPayload, error)
}
