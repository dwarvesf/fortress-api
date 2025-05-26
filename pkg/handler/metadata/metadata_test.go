package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/metadata/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk1ODMzMzA5NDUsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIn0.oIdlwWGBy4E1CbSoEX6r2B6NQLbew_J-RttpAcg6w8M"

func TestHandler_GetWorkingStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := &store.Store{}

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_all",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_working_status/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/working-status?%s", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.WorkingStatuses(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.GetWorkingStatus] response mismatched")
			})
		})
	}
}

func TestHandler_GetSeniority(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_seniorities",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_seniorities/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/seniorities", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Seniorities(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.seniorities] response mismatched")
			})
		})
	}
}

func TestHandler_GetChapters(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_chapters",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_chapters/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/chapters", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Chapters(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Chapters] response mismatched")
			})
		})
	}
}

func TestHandler_GetProjectStatuses(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_project_statuses",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_project_statuses/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/project-statuses", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.ProjectStatuses(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.ProjectStatuses] response mismatched")
			})
		})
	}
}

func TestHandler_GetPositions(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_positions",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_positions/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/positions", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Positions(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Positions] response mismatched")
			})
		})
	}
}

func TestHandler_GetTechStacks(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_stacks",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_stacks/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/stacks", nil)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Stacks(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Stacks] response mismatched")
			})
		})
	}
}

func TestHandler_GetQuestion(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		query            string
	}{
		{
			name:             "ok_get_questions",
			wantCode:         200,
			wantResponsePath: "testdata/get_questions/200.json",
			query:            "category=survey&subcategory=peer-review",
		},
		{
			name:             "invalid_subtype",
			wantCode:         400,
			wantResponsePath: "testdata/get_questions/400.json",
			query:            "category=feedback&subcategory=peer-revie",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest("GET", "/api/v1/metadata/questions", nil)
				ctx.Request.URL.RawQuery = tt.query
				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				h.GetQuestions(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.GetQuestions] response mismatched")
			})
		})
	}
}

func TestHandler_CreateStack(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		body             request.CreateStackInput
		wantResponsePath string
	}{
		{
			name:     "ok_create_stack",
			wantCode: http.StatusOK,
			body: request.CreateStackInput{
				Name:   "name",
				Code:   "code",
				Avatar: "avatar",
			},
			wantResponsePath: "testdata/create_stack/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/metadata/stacks", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.CreateStack(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.CreateStack] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateStack(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		id               string
		body             request.UpdateStackBody
		wantResponsePath string
	}{
		{
			name:     "ok_update_stack",
			wantCode: http.StatusOK,
			id:       "0ecf47c8-cca4-4c30-94bb-054b1124c44f",
			body: request.UpdateStackBody{
				Name:   "Newname",
				Code:   "code",
				Avatar: "avatar",
			},
			wantResponsePath: "testdata/update_stack/200.json",
		},
		{
			name:     "not_found_stack",
			wantCode: http.StatusNotFound,
			id:       "0ecf47c8-cca4-4c30-94bb-054b1124c44e",
			body: request.UpdateStackBody{
				Name:   "Newname",
				Code:   "code",
				Avatar: "avatar",
			},
			wantResponsePath: "testdata/update_stack/404.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/metadata/stacks/%s", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdateStack(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateStack] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteStack(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		id               string
		wantResponsePath string
	}{
		{
			name:             "ok_delete_stack",
			wantCode:         http.StatusOK,
			id:               "0ecf47c8-cca4-4c30-94bb-054b1124c44f",
			wantResponsePath: "testdata/delete_stack/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/metadata/stacks/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteStack(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.DeleteStack] response mismatched")
			})
		})
	}
}

func TestHandler_CreatePosition(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		body             request.CreatePositionInput
		wantResponsePath string
	}{
		{
			name:     "ok_create_position",
			wantCode: http.StatusOK,
			body: request.CreatePositionInput{
				Name: "name",
				Code: "code",
			},
			wantResponsePath: "testdata/create_position/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/metadata/positions", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.CreatePosition(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.CreatePosition] response mismatched")
			})
		})
	}
}

func TestHandler_UpdatePosition(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		id               string
		body             request.UpdatePositionBody
		wantResponsePath string
	}{
		{
			name:     "ok_update_position",
			wantCode: http.StatusOK,
			id:       "11ccffea-2cc9-4e98-9bef-3464dfe4dec8",
			body: request.UpdatePositionBody{
				Name: "Newname",
				Code: "code",
			},
			wantResponsePath: "testdata/update_position/200.json",
		},
		{
			name:     "not_found_position",
			wantCode: http.StatusNotFound,
			id:       "0ecf47c8-cca4-4c30-94bb-054b1124c44e",
			body: request.UpdatePositionBody{
				Name: "Newname",
				Code: "code",
			},
			wantResponsePath: "testdata/update_position/404.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/metadata/positions/%s", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdatePosition(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdatePosition] response mismatched")
			})
		})
	}
}

func TestHandler_DeletePosition(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		id               string
		wantResponsePath string
	}{
		{
			name:             "ok_delete_stack",
			wantCode:         http.StatusOK,
			id:               "11ccffea-2cc9-4e98-9bef-3464dfe4dec8",
			wantResponsePath: "testdata/delete_stack/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/metadata/stacks/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeletePosition(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.DeletePosition] response mismatched")
			})
		})
	}
}
