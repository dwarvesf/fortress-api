package employee

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
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func TestHandler_UpdateEmployeeStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_update_employee_status",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/update_employee_status/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(`
			{
				"employeeStatus":"active"
			}
			`)
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: "2655832e-f009-4b73-a535-64c3a22e558f"}}
			ctx.Request = httptest.NewRequest("POST", fmt.Sprintf("%s", "/api/v1/employees/2655832e-f009-4b73-a535-64c3a22e558f/employee-status"), bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, serviceMock, loggerMock)

			metadataHandler.UpdateEmployeeStatus(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			var actualData view.UpdateEmployeeStatusResponse
			var expectedData view.UpdateEmployeeStatusResponse
			err = json.Unmarshal(w.Body.Bytes(), &actualData)
			err = json.Unmarshal(expRespRaw, &expectedData)

			actualData.Data.UpdatedAt = nil
			expectedData.Data.UpdatedAt = nil

			require.Equal(t, expectedData, actualData)
		})
	}
}

const tokenTest = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_GetProfile(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "ok_get_profile",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_profile/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("POST", fmt.Sprintf("%s", "/api/v1/profile"), nil)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, serviceMock, loggerMock)

			metadataHandler.GetProfile(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.GetProfile] response mismatched")
		})
	}
}

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New(&cfg)

	tests := []struct {
		name             string
		query            string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "without_working_status_and_pagination",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/employee/list/without_working_status_and_pagination.json",
		},
		{
			name:             "have_workingStatus_and_no_pagination",
			query:            "workingStatus=contractor",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/employee/list/have_working_and_no_pagination.json",
		},
		{
			name:             "have_workingStatus_and_pagination",
			query:            "workingStatus=contractor&page=1&size=5",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/employee/list/have_working_and_pagination.json",
		},
		{
			name:             "out_of_content",
			query:            "workingStatus=contractor&page=5&size=5",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/employee/list/out_of_content.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
			ctx.Request.Header.Set("Authorization", tokenTest)
			ctx.Request.URL.RawQuery = tt.query

			h := New(storeMock, serviceMock, loggerMock)
			h.List(ctx)
			tt.wantCode = http.StatusOK
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "updatedAt")
			if err != nil {
				t.Errorf("failed to remove updatedAt: %v", err)
			}

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Employee.List] response mismatched")
		})
	}
}
