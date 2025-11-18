package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

type fakeExpenseProvider struct {
	payload *taskprovider.ExpenseWebhookPayload
	err     error
	calls   []string
	result  *taskprovider.ExpenseValidationResult
	valErr  error
}

func (f *fakeExpenseProvider) Type() taskprovider.ProviderType { return taskprovider.ProviderNocoDB }
func (f *fakeExpenseProvider) ParseExpenseWebhook(ctx context.Context, req taskprovider.ExpenseWebhookRequest) (*taskprovider.ExpenseWebhookPayload, error) {
	return f.payload, f.err
}
func (f *fakeExpenseProvider) ValidateSubmission(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseValidationResult, error) {
	f.calls = append(f.calls, "validate")
	if f.result != nil {
		return f.result, f.valErr
	}
	return &taskprovider.ExpenseValidationResult{Valid: true}, nil
}
func (f *fakeExpenseProvider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
	f.calls = append(f.calls, "create")
	return nil, nil
}
func (f *fakeExpenseProvider) CompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	f.calls = append(f.calls, "complete")
	return nil
}
func (f *fakeExpenseProvider) UncompleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	f.calls = append(f.calls, "uncomplete")
	return nil
}
func (f *fakeExpenseProvider) DeleteExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) error {
	return nil
}
func (f *fakeExpenseProvider) PostFeedback(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload, input taskprovider.ExpenseFeedbackInput) error {
	f.calls = append(f.calls, "feedback")
	return nil
}

func TestHandleNocoExpense_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhooks/nocodb/expense", strings.NewReader(`{}`))

	h := handler{
		config: &config.Config{ExpenseIntegration: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{WebhookSecret: "secret"}}},
		logger: logger.NewLogrusLogger("error"),
	}

	h.HandleNocoExpense(c)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHandleNocoExpense_CreateFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	body := `{}`
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhooks/nocodb/expense", strings.NewReader(body))
	c.Request.Header.Set("X-NocoDB-Signature", computeTestSignature("secret", body))

	payload := &taskprovider.ExpenseWebhookPayload{EventType: taskprovider.ExpenseEventCreate}
	provider := &fakeExpenseProvider{payload: payload}
	ctrl := &controller.Controller{}

	h := handler{
		service:    &service.Service{ExpenseProvider: provider},
		controller: ctrl,
		config:     &config.Config{ExpenseIntegration: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{WebhookSecret: "secret"}}},
		logger:     logger.NewLogrusLogger("error"),
	}

	h.HandleNocoExpense(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, provider.calls, "create")
}

func TestHandleNocoExpense_ValidateFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	body := `{"event":"validate"}`
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhooks/nocodb/expense", strings.NewReader(body))
	c.Request.Header.Set("X-NocoDB-Signature", computeTestSignature("secret", body))

	provider := &fakeExpenseProvider{payload: &taskprovider.ExpenseWebhookPayload{EventType: taskprovider.ExpenseEventValidate}}
	h := handler{
		service: &service.Service{ExpenseProvider: provider},
		config:  &config.Config{ExpenseIntegration: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{WebhookSecret: "secret"}}},
		logger:  logger.NewLogrusLogger("error"),
	}

	h.HandleNocoExpense(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, provider.calls, "validate")
	require.NotContains(t, provider.calls, "create")
}

func TestHandleNocoExpense_UncompleteFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	body := `{}`
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/webhooks/nocodb/expense", strings.NewReader(body))
	c.Request.Header.Set("X-NocoDB-Signature", computeTestSignature("secret", body))

	provider := &fakeExpenseProvider{payload: &taskprovider.ExpenseWebhookPayload{EventType: taskprovider.ExpenseEventUncomplete}}
	h := handler{
		service: &service.Service{ExpenseProvider: provider},
		config:  &config.Config{ExpenseIntegration: config.ExpenseIntegration{Noco: config.ExpenseNocoIntegration{WebhookSecret: "secret"}}},
		logger:  logger.NewLogrusLogger("error"),
	}

	h.HandleNocoExpense(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, provider.calls, "uncomplete")
}

func computeTestSignature(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}
