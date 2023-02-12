package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/client/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_Detail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			id:               "67f9f420-cdd5-4793-88c7-d2068bd17f61",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/detail/200.json",
		},
		{
			name:             "not_found",
			id:               "2655832e-f009-4b73-a535-64c3a22e558e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/detail/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/detail/detail.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				ctrl := controller.New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Detail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Client.Detail] response mismatched")
			})
		})
	}
}

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/list/list.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clients"), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				ctrl := controller.New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.List(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Client.List] response mismatched")
			})
		})
	}
}

func TestHandler_Create(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		body             request.CreateClientInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "67f9f420-cdd5-4793-88c7-d2068bd17f61",
			body: request.CreateClientInput{
				Name:               "John",
				Description:        "description",
				RegistrationNumber: "456",
				Address:            "CA",
				Country:            "USA",
				Industry:           "education",
				Website:            "b.com",
				Contacts: []*request.CreateClientContactInput{
					{
						Name:          "contact name 1",
						Role:          "manager",
						Emails:        []string{"john1@gmail.com", "john2@gmail.com"},
						IsMainContact: true,
					},
					{
						Name:          "contact name 2",
						Role:          "manager",
						Emails:        []string{"john3@gmail.com", "john4@gmail.com"},
						IsMainContact: false,
					},
				},
			},
			wantCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create/create.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clients"), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				ctrl := controller.New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Create(ctx)
				require.Equal(t, tt.wantCode, w.Code)
			})
		})
	}
}

func TestHandler_Update(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		body             request.UpdateClientInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "67f9f420-cdd5-4793-88c7-d2068bd17f61",
			body: request.UpdateClientInput{
				Name:               "John",
				Description:        "description",
				RegistrationNumber: "456",
				Address:            "CA",
				Country:            "USA",
				Industry:           "education",
				Website:            "b.com",
				Contacts: []*request.UpdateClientContactInput{
					{
						Name:          "contact name 1",
						Role:          "manager",
						Emails:        []string{"john1@gmail.com", "john2@gmail.com"},
						IsMainContact: true,
					},
					{
						Name:          "contact name 2",
						Role:          "manager",
						Emails:        []string{"john3@gmail.com", "john4@gmail.com"},
						IsMainContact: false,
					},
				},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_client/200.json",
		},
		{
			name:             "not_found",
			id:               "2655832e-f009-4b73-a535-64c3a22e558e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_client/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_client/update_client.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/clients/%s", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				ctrl := controller.New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Update(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Client.Update] response mismatched")
			})
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		body             request.UpdateClientInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			id:               "67f9f420-cdd5-4793-88c7-d2068bd17f61",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete/200.json",
		},
		{
			name:             "not_found",
			id:               "2655832e-f009-4b73-a535-64c3a22e558e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete/delete.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/clients/%s", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				ctrl := controller.New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Delete(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Client.Detele] response mismatched")
			})
		})
	}
}
