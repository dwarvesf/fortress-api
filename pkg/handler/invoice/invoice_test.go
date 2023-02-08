package invoice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_Create(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		request          request.CreateInput
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create/200.json",
			request: request.CreateInput{
				Number:           "123",
				DueAt:            "2022-01-02",
				PaidAt:           "2022-01-02",
				Status:           "sent",
				Email:            "a@gmail.com",
				Description:      "description",
				Note:             "note",
				SubTotal:         1,
				Tax:              1,
				Discount:         1,
				Total:            1,
				ConversionAmount: 1,
				Month:            2,
				Year:             2023,
				ThreadID:         "123",
				ScheduledDate:    "2023-02-01",
				ConversionRate:   1.2,
				BankID:           model.MustGetUUIDFromString("e79eb5b3-e2cb-4d7f-9273-46f4be88cb20"),
				ProjectID:        model.MustGetUUIDFromString("8dc3be2e-19a4-4942-8a79-56db391a0b15"),
			},
		},
		{
			name:             "invalid_mail_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400.json",
			request: request.CreateInput{
				Number:           "123",
				DueAt:            "2022-01-02",
				PaidAt:           "2022-01-02",
				Status:           "sent",
				Email:            "agmail.com",
				Description:      "description",
				Note:             "note",
				SubTotal:         1,
				Tax:              1,
				Discount:         1,
				Total:            1,
				ConversionAmount: 1,
				Month:            2,
				Year:             2023,
				ThreadID:         "123",
				ScheduledDate:    "2023-02-01",
				ConversionRate:   1.2,
				BankID:           model.MustGetUUIDFromString("e79eb5b3-e2cb-4d7f-9273-46f4be88cb20"),
				ProjectID:        model.MustGetUUIDFromString("8dc3be2e-19a4-4942-8a79-56db391a0b15"),
			},
		},
		{
			name:             "invalid_mail_format",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/create/404.json",
			request: request.CreateInput{
				Number:           "123",
				DueAt:            "2022-01-02",
				PaidAt:           "2022-01-02",
				Status:           "sent",
				Email:            "a@gmail.com",
				Description:      "description",
				Note:             "note",
				SubTotal:         1,
				Tax:              1,
				Discount:         1,
				Total:            1,
				ConversionAmount: 1,
				Month:            2,
				Year:             2023,
				ThreadID:         "123",
				ScheduledDate:    "2023-02-01",
				ConversionRate:   1.2,
				BankID:           model.MustGetUUIDFromString("e79eb5b3-e2cb-4d7f-9273-46f4be88cb21"),
				ProjectID:        model.MustGetUUIDFromString("8dc3be2e-19a4-4942-8a79-56db391a0b15"),
			},
		},
	}

	for _, tt := range tests {
		testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
			testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create/create.sql")
			byteReq, err := json.Marshal(tt.request)
			require.Nil(t, err)

			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest("POST", "/api/v1/invoices", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Create(ctx)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.Equal(t, tt.wantCode, w.Code)
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "createdAt")
				res, err = utils.RemoveFieldInResponse(res, "updatedAt")
				res, err = utils.RemoveFieldInResponse(res, "id")
				require.Nil(t, err)
				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Create] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		request          request.UpdateInput
		id               string
	}{
		{
			name:             "ok_update_status",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_status/200.json",
			request: request.UpdateInput{
				Status: "draft",
			},
			id: "bf724631-300f-4b01-bd40-ab20c8c5c74c",
		},
		{
			name:             "invalid_status",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_status/400_invalid_status.json",
			request: request.UpdateInput{
				Status: "draftt",
			},
			id: "bf724631-300f-4b01-bd40-ab20c8c5c74c",
		},
		{
			name:             "invoice_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_status/404.json",
			request: request.UpdateInput{
				Status: "draft",
			},
			id: "bf724631-300f-4b01-bd40-ab20c8c5c74d",
		},
	}

	for _, tt := range tests {
		testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
			testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_status/update_status.sql")
			byteReq, err := json.Marshal(tt.request)
			require.Nil(t, err)

			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/invoices/%s/status", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Update(ctx)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.Equal(t, tt.wantCode, w.Code)
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)
				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Update] response mismatched")
			})
		})
	}
}

func TestHandler_GetLatest(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		query            string
	}{
		{
			name:             "ok_get_latest",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_latest/200.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}

	for _, tt := range tests {
		testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
			testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_latest/get_latest.sql")

			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/invoices/latest?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.GetLatestInvoice(ctx)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.Equal(t, tt.wantCode, w.Code)
				require.JSONEq(t, string(expRespRaw), string(w.Body.Bytes()), "[Handler.GetLatestInvoice] response mismatched")
			})
		})
	}
}
