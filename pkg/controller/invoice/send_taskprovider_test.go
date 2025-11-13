package invoice

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

func TestDispatchInvoiceTask_Success(t *testing.T) {
	provider := &fakeInvoiceProvider{}
	ctrl, queue := newTestInvoiceController(provider)

	invoice := &model.Invoice{
		Number:             "2025-INV-001",
		InvoiceFileContent: []byte("pdf-bytes"),
	}

	err := ctrl.dispatchInvoiceTask(invoice, "file.pdf")
	require.NoError(t, err)

	require.Equal(t, "[mock](https://drive.fake/file.pdf)", invoice.TodoAttachment)
	require.Len(t, provider.uploadCalls, 1)
	require.Equal(t, "file.pdf", provider.uploadCalls[0].FileName)

	select {
	case msg := <-queue:
		require.Equal(t, taskprovider.WorkerMessageInvoiceComment, msg.Type)
		job, ok := msg.Payload.(taskprovider.InvoiceCommentJob)
		require.True(t, ok)
		require.Equal(t, provider.ref, job.Ref)
		require.Contains(t, job.Input.Message, invoice.Number)
	default:
		t.Fatalf("expected invoice comment job to be enqueued")
	}
}

func TestDispatchInvoiceTask_Errors(t *testing.T) {
	t.Run("no provider", func(t *testing.T) {
		ctrl, _ := newTestInvoiceController(nil)
		err := ctrl.dispatchInvoiceTask(&model.Invoice{}, "file.pdf")
		require.EqualError(t, err, "task provider is not configured")
	})

	t.Run("upload failure", func(t *testing.T) {
		provider := &fakeInvoiceProvider{uploadErr: assertError("upload failed")}
		ctrl, _ := newTestInvoiceController(provider)
		err := ctrl.dispatchInvoiceTask(&model.Invoice{}, "file.pdf")
		require.ErrorContains(t, err, "create task attachment")
	})
}

type fakeInvoiceProvider struct {
	uploadCalls []taskprovider.InvoiceAttachmentInput
	ensureCalls int
	ref         *taskprovider.InvoiceTaskRef
	uploadErr   error
}

type assertError string

func (e assertError) Error() string { return string(e) }

func (f *fakeInvoiceProvider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderNocoDB
}

func (f *fakeInvoiceProvider) EnsureTask(ctx context.Context, input taskprovider.CreateInvoiceTaskInput) (*taskprovider.InvoiceTaskRef, error) {
	if f.ref == nil {
		f.ref = &taskprovider.InvoiceTaskRef{Provider: taskprovider.ProviderNocoDB, ExternalID: "row-1"}
	}
	f.ensureCalls++
	return f.ref, nil
}

func (f *fakeInvoiceProvider) UploadAttachment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceAttachmentInput) (*taskprovider.InvoiceAttachmentRef, error) {
	if f.uploadErr != nil {
		return nil, f.uploadErr
	}
	f.uploadCalls = append(f.uploadCalls, input)
	return &taskprovider.InvoiceAttachmentRef{
		ExternalID: "mock-external",
		Markup:     "[mock](https://drive.fake/file.pdf)",
	}, nil
}

func (f *fakeInvoiceProvider) PostComment(ctx context.Context, ref *taskprovider.InvoiceTaskRef, input taskprovider.InvoiceCommentInput) error {
	return nil
}

func (f *fakeInvoiceProvider) CompleteTask(ctx context.Context, ref *taskprovider.InvoiceTaskRef) error {
	return nil
}

func newTestInvoiceController(provider taskprovider.InvoiceProvider) (*controller, chan model.WorkerMessage) {
	svc := &service.Service{TaskProvider: provider}
	queue := make(chan model.WorkerMessage, 1)
	ctrl := &controller{
		service: svc,
		worker:  worker.New(context.Background(), queue, svc, logger.NewLogrusLogger()),
		logger:  logger.NewLogrusLogger(),
	}
	return ctrl, queue
}
