package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

func TestHandleInvoiceCommentJob_Success(t *testing.T) {
	provider := &stubProvider{}
	svc := &service.Service{TaskProvider: provider}
	w := New(context.Background(), make(chan model.WorkerMessage, 1), svc, logger.NewLogrusLogger("info"))

	job := taskprovider.InvoiceCommentJob{
		Ref: &taskprovider.InvoiceTaskRef{ExternalID: "row-1"},
		Input: taskprovider.InvoiceCommentInput{
			Message: "hello",
		},
	}

	err := w.handleInvoiceCommentJob(logger.NewLogrusLogger("info"), job)
	require.NoError(t, err)
	require.Equal(t, job, provider.lastJob)
}

func TestHandleInvoiceCommentJob_Errors(t *testing.T) {
	t.Run("invalid payload", func(t *testing.T) {
		svc := &service.Service{}
		w := New(context.Background(), make(chan model.WorkerMessage, 1), svc, logger.NewLogrusLogger("info"))
		err := w.handleInvoiceCommentJob(logger.NewLogrusLogger("info"), "string")
		require.EqualError(t, err, "invalid invoice comment job payload")
	})

	t.Run("missing provider", func(t *testing.T) {
		svc := &service.Service{}
		w := New(context.Background(), make(chan model.WorkerMessage, 1), svc, logger.NewLogrusLogger("info"))
		job := taskprovider.InvoiceCommentJob{}
		err := w.handleInvoiceCommentJob(logger.NewLogrusLogger("info"), job)
		require.EqualError(t, err, "task provider not configured")
	})
}

type stubProvider struct {
	lastJob taskprovider.InvoiceCommentJob
}

func (s *stubProvider) Type() taskprovider.ProviderType { return taskprovider.ProviderNocoDB }
func (s *stubProvider) EnsureTask(context.Context, taskprovider.CreateInvoiceTaskInput) (*taskprovider.InvoiceTaskRef, error) {
	return nil, nil
}
func (s *stubProvider) UploadAttachment(context.Context, *taskprovider.InvoiceTaskRef, taskprovider.InvoiceAttachmentInput) (*taskprovider.InvoiceAttachmentRef, error) {
	return nil, nil
}
func (s *stubProvider) PostComment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceCommentInput) error {
	s.lastJob = taskprovider.InvoiceCommentJob{Ref: ref, Input: input}
	return nil
}
func (s *stubProvider) CompleteTask(context.Context, *taskprovider.InvoiceTaskRef) error { return nil }
