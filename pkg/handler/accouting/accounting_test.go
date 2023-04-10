package accounting

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
	"testing"
	"time"
)

func Test_handler_CreateAccountingTodo(t *testing.T) {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(cfg, nil, nil)
	storeMock := store.New()

	type fields struct {
		store   *store.Store
		service *service.Service
		logger  logger.Logger
		config  *config.Config
	}
	type args struct {
		month int
		year  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test create accounting todo",
			fields: fields{
				config:  cfg,
				store:   storeMock,
				service: serviceMock,
				logger:  loggerMock,
			},
			args: args{
				month: int(time.Now().Month()),
				year:  time.Now().Year(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create_accounting_todo/create_accounting_todo.sql")
				h := handler{
					store:   tt.fields.store,
					service: tt.fields.service,
					logger:  tt.fields.logger,
					repo:    txRepo,
					config:  tt.fields.config,
				}
				if err := h.CreateAccountingTodo(tt.args.month, tt.args.year); (err != nil) != tt.wantErr {
					t.Errorf("CreateAccountingTodo() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		})
	}
}
