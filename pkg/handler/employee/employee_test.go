package employee

import (
	"encoding/json"
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
	"github.com/dwarvesf/fortress-api/pkg/model"
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
			ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/2655832e-f009-4b73-a535-64c3a22e558f/employee-status", bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, serviceMock, loggerMock)

			metadataHandler.UpdateEmployeeStatus(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			var actualData view.UpdateEmployeeStatusResponse
			var expectedData view.UpdateEmployeeStatusResponse
			err = json.Unmarshal(w.Body.Bytes(), &actualData)
			require.NoError(t, err)
			err = json.Unmarshal(expRespRaw, &expectedData)
			require.NoError(t, err)

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
			ctx.Request = httptest.NewRequest("GET", "/api/v1/profile", nil)
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
			wantResponsePath: "testdata/list/without_working_status_and_pagination.json",
		},
		{
			name:             "have_workingStatus_and_no_pagination",
			query:            "workingStatus=contractor",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/have_working_and_no_pagination.json",
		},
		{
			name:             "have_workingStatus_and_pagination",
			query:            "workingStatus=contractor&page=1&size=5",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/have_working_and_pagination.json",
		},
		{
			name:             "out_of_content",
			query:            "workingStatus=contractor&page=5&size=5",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/out_of_content.json",
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

func Test_UpdateGeneralInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             EditGeneralInfo
		id               string
	}{
		{
			name:             "ok_edit_general_info",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_general_info/200.json",
			body: EditGeneralInfo{
				Fullname: "Phạm Đức Thành",
				Email:    "thanh@d.foundation",
				Phone:    "0123456788",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "fail_wrong_address",
			wantCode:         500,
			wantErr:          true,
			wantResponsePath: "testdata/update_general_info/500.json",
			body: EditGeneralInfo{
				Fullname: "Phạm Đức Thành",
				Email:    "thanh@d.foundation",
				Phone:    "0123456788",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.body)
			require.Nil(t, err)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(string(byteReq))
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/"+tt.id+"/general-info", bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, serviceMock, loggerMock)

			metadataHandler.UpdateGeneralInfo(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			if !tt.wantErr {
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateProjectStatus] response mismatched")
			} else {
				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectStatus] response mismatched")
			}
		})
	}
}

func Test_Create(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             CreateEmployee
		id               string
	}{
		{
			name:             "exists_user_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/existed_user.json",
			body: CreateEmployee{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				PositionID:    model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "validation_err_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/validation_err.json",
			body: CreateEmployee{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoigmail.com",
				PositionID:    model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "invalid_uuid_create",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/create/invalid_uuid.json",
			body: CreateEmployee{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				PositionID:    model.MustGetUUIDFromString("c44c987c-ad34-4745-be2b-942e8670d32a"),
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.body)
			require.Nil(t, err)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(string(byteReq))
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("POST", "/api/v1/employees/", bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, serviceMock, loggerMock)

			metadataHandler.Create(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			if !tt.wantErr {
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[employee.Handler.Create] response mismatched")
			} else {
				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[employee.Handler.Create] response mismatched")
			}
		})
	}
}
