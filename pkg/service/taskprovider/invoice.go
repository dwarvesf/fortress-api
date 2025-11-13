package taskprovider

import (
	"context"
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// ProviderType identifies which backend powers the task provider.
type ProviderType string

const (
	ProviderBasecamp            ProviderType = "basecamp"
	ProviderNocoDB              ProviderType = "nocodb"
	WorkerMessageInvoiceComment             = "taskprovider_invoice_comment"
)

// ErrNotImplemented marks unimplemented provider operations (placeholder during migration).
var ErrNotImplemented = errors.New("task provider operation not implemented")

// InvoiceTaskRef captures identifiers needed to reference an invoice task in the upstream system.
type InvoiceTaskRef struct {
	Provider   ProviderType
	ExternalID string
	BucketID   int
	TodoID     int
}

// CreateInvoiceTaskInput describes the invoice metadata required to open a task.
type CreateInvoiceTaskInput struct {
	Invoice *model.Invoice
}

// InvoiceAttachmentInput wraps attachment metadata for uploads.
type InvoiceAttachmentInput struct {
	FileName    string
	ContentType string
	Content     []byte
	URL         string
}

// InvoiceAttachmentRef holds provider-specific attachment identifiers.
type InvoiceAttachmentRef struct {
	ExternalID string
	Markup     string
}

// InvoiceCommentInput defines the payload for posting status updates.
type InvoiceCommentInput struct {
	Message string
	Type    string
}

// InvoiceCommentJob is enqueued to the worker to process provider-specific comments.
type InvoiceCommentJob struct {
	Ref   *InvoiceTaskRef
	Input InvoiceCommentInput
}

// InvoiceProvider exposes operations necessary for invoice task management.
type InvoiceProvider interface {
	Type() ProviderType
	EnsureTask(ctx context.Context, input CreateInvoiceTaskInput) (*InvoiceTaskRef, error)
	UploadAttachment(ctx context.Context, ref *InvoiceTaskRef, input InvoiceAttachmentInput) (*InvoiceAttachmentRef, error)
	PostComment(ctx context.Context, ref *InvoiceTaskRef, input InvoiceCommentInput) error
	CompleteTask(ctx context.Context, ref *InvoiceTaskRef) error
}
