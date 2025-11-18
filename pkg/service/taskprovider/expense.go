package taskprovider

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// ExpenseEventType indicates which lifecycle event triggered the webhook.
type ExpenseEventType string

const (
	// ExpenseEventValidate fires when a submission requires validation/assignment.
	ExpenseEventValidate ExpenseEventType = "validate"
	// ExpenseEventCreate represents expense creation.
	ExpenseEventCreate ExpenseEventType = "create"
	// ExpenseEventComplete indicates completion/approval.
	ExpenseEventComplete ExpenseEventType = "complete"
	// ExpenseEventUncomplete indicates reversal/uncomplete/delete.
	ExpenseEventUncomplete ExpenseEventType = "uncomplete"
)

// ExpenseValidationResult captures the outcome of provider-side validation.
type ExpenseValidationResult struct {
	Valid        bool
	Message      string
	Err          error
	Skip         bool
	FeedbackKind string
}

// ExpenseFeedbackInput defines status/comment messages sent back to provider.
type ExpenseFeedbackInput struct {
	Message string
	Kind    string
}

// ExpenseTaskRef stores provider-specific identifiers for an expense row/todo.
type ExpenseTaskRef struct {
	Provider   ProviderType
	ExternalID string
	BucketID   int
	TodoID     int
	RowID      string
}

// ExpenseWebhookRequest abstracts incoming webhook data for each provider.
type ExpenseWebhookRequest struct {
	Headers         map[string]string
	Body            []byte
	BasecampMessage *model.BasecampWebhookMessage
}

// ExpenseWebhookPayload is the normalized structure returned by provider parsers.
type ExpenseWebhookPayload struct {
	Provider        ProviderType
	EventType       ExpenseEventType
	Title           string
	Reason          string
	AmountRaw       string
	Amount          int
	Currency        string
	CreatorName     string
	CreatorID       int
	CreatorEmail    string
	BucketName      string
	BucketID        int
	RecordingID     int
	RecordingURL    string
	TaskRef         string
	TaskBoard       string
	TaskAttachmentURL string
	TaskAttachments   []string
	Metadata        []byte
	BasecampMessage *model.BasecampWebhookMessage
	Raw             interface{}
}

// ExpenseProvider exposes operations needed by the expense flow.
type ExpenseProvider interface {
	Type() ProviderType
	ParseExpenseWebhook(ctx context.Context, req ExpenseWebhookRequest) (*ExpenseWebhookPayload, error)
	ValidateSubmission(ctx context.Context, payload *ExpenseWebhookPayload) (*ExpenseValidationResult, error)
	CreateExpense(ctx context.Context, payload *ExpenseWebhookPayload) (*ExpenseTaskRef, error)
	CompleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
	UncompleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
	DeleteExpense(ctx context.Context, payload *ExpenseWebhookPayload) error
	PostFeedback(ctx context.Context, payload *ExpenseWebhookPayload, input ExpenseFeedbackInput) error
}
