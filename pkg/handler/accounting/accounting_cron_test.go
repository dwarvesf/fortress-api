package accounting

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func TestCreateTodoInOutGroup_RecordsTemplates(t *testing.T) {
	ctx := context.Background()
	taskRefStore := &fakeAccountingTaskRefStore{}
	h := handler{
		store:  &store.Store{AccountingTaskRef: taskRefStore},
		logger: stubLogger{},
		repo:   &fakeRepo{},
	}
	provider := &fakeAccountingProvider{}
	plan := &taskprovider.AccountingPlanRef{
		Provider: taskprovider.ProviderNocoDB,
		ListID:   "Accounting | February 2026",
		Month:    2,
		Year:     2026,
	}
	templates := []*model.OperationalService{
		{
			BaseModel: model.BaseModel{ID: model.NewUUID()},
			Name:      "Office Rental HN",
			Amount:    1_500_000,
			Currency:  &model.Currency{Name: "VND"},
		},
		{
			BaseModel: model.BaseModel{ID: model.NewUUID()},
			Name:      "IT Support",
			Amount:    500_000,
			Currency:  &model.Currency{Name: "USD"},
		},
	}

	err := h.createTodoInOutGroup(ctx, provider, plan, templates, 2, 2026)
	require.NoError(t, err)
	require.Len(t, provider.calls, 4, "office rental variant should expand to 3 todos (+1 other template)")
	require.Len(t, taskRefStore.entries, 4)

	meta := metadataToMap(t, taskRefStore.entries[0].Metadata)
	require.Equal(t, "Tiền điện", meta["variant"])

	metaGeneral := metadataToMap(t, taskRefStore.entries[2].Metadata)
	require.Equal(t, "VND", metaGeneral["currency"])
	require.EqualValues(t, 1_500_000, metaGeneral["amount"])
}

func TestCreateSalaryTodo_RecordsTwoEntries(t *testing.T) {
	ctx := context.Background()
	taskRefStore := &fakeAccountingTaskRefStore{}
	h := handler{
		store:  &store.Store{AccountingTaskRef: taskRefStore},
		logger: stubLogger{},
		repo:   &fakeRepo{},
	}
	provider := &fakeAccountingProvider{}
	plan := &taskprovider.AccountingPlanRef{
		Provider: taskprovider.ProviderNocoDB,
		ListID:   "Accounting | March 2026",
		Month:    3,
		Year:     2026,
	}

	err := h.createSalaryTodo(ctx, provider, plan, 3, 2026)
	require.NoError(t, err)
	require.Len(t, provider.calls, 2)
	require.Len(t, taskRefStore.entries, 2)

	first := metadataToMap(t, taskRefStore.entries[0].Metadata)
	require.Equal(t, "salary", first["type"])
	require.Equal(t, "15th", first["cycle"])
}

type fakeAccountingProvider struct {
	calls []taskprovider.CreateAccountingTodoInput
}

func (f *fakeAccountingProvider) Type() taskprovider.ProviderType {
	return taskprovider.ProviderNocoDB
}

func (f *fakeAccountingProvider) CreateMonthlyPlan(context.Context, taskprovider.CreateAccountingPlanInput) (*taskprovider.AccountingPlanRef, error) {
	return nil, nil
}

func (f *fakeAccountingProvider) CreateAccountingTodo(_ context.Context, _ *taskprovider.AccountingPlanRef, input taskprovider.CreateAccountingTodoInput) (*taskprovider.AccountingTodoRef, error) {
	f.calls = append(f.calls, input)
	return &taskprovider.AccountingTodoRef{
		Provider:   taskprovider.ProviderNocoDB,
		ExternalID: fmt.Sprintf("row-%d", len(f.calls)),
		Group:      input.Group,
	}, nil
}

func (f *fakeAccountingProvider) ParseAccountingWebhook(context.Context, taskprovider.AccountingWebhookRequest) (*taskprovider.AccountingWebhookPayload, error) {
	return nil, nil
}

type fakeAccountingTaskRefStore struct {
	entries []*model.AccountingTaskRef
}

func (s *fakeAccountingTaskRefStore) Create(_ *gorm.DB, ref *model.AccountingTaskRef) error {
	copy := *ref
	if len(ref.Metadata) > 0 {
		copy.Metadata = append(datatypes.JSON{}, ref.Metadata...)
	}
	s.entries = append(s.entries, &copy)
	return nil
}

type fakeRepo struct{}

func (fakeRepo) DB() *gorm.DB { return nil }

func (fakeRepo) NewTransaction() (store.DBRepo, store.FinallyFunc) {
	return fakeRepo{}, func(error) error { return nil }
}

func (fakeRepo) SetNewDB(*gorm.DB) {}

type stubLogger struct{}

func (stubLogger) Fields(logger.Fields) logger.Logger { return stubLogger{} }
func (stubLogger) Field(string, string) logger.Logger { return stubLogger{} }
func (stubLogger) AddField(string, any) logger.Logger { return stubLogger{} }
func (stubLogger) Debug(string)                       {}
func (stubLogger) Debugf(string, ...interface{})      {}
func (stubLogger) Info(string)                        {}
func (stubLogger) Infof(string, ...interface{})       {}
func (stubLogger) Warn(string)                        {}
func (stubLogger) Warnf(string, ...interface{})       {}
func (stubLogger) Error(error, string)                {}
func (stubLogger) Errorf(error, string, ...interface{}) {
}
func (stubLogger) Fatal(error, string)                       {}
func (stubLogger) Fatalf(error, string, ...interface{})      {}

func metadataToMap(t *testing.T, data datatypes.JSON) map[string]any {
	t.Helper()
	if len(data) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	return out
}
