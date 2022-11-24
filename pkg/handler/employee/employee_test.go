package employee

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

const tokenTest = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_UpdateEmployeeStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		id               string
		body             UpdateWorkingStatusInput
		wantResponsePath string
	}{
		{
			name:     "ok_update_employee_status",
			wantCode: http.StatusOK,
			id:       "2655832e-f009-4b73-a535-64c3a22e558f",
			body: UpdateWorkingStatusInput{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.body)
			require.Nil(t, err)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(string(byteReq))
			ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/employees/%s/employee-status", tt.id), bodyReader)
			ctx.Request.Header.Set("Authorization", tokenTest)
			ctx.AddParam("id", tt.id)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)
			h.UpdateEmployeeStatus(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
			require.Nil(t, err)

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateEmployeeStatus] response mismatched")
		})
	}
}

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

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
		{
			name:             "with_preload_false",
			query:            "workingStatus=probation&preload=false",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_preload_false.json",
		},
		{
			name:             "without_preload",
			query:            "workingStatus=probation",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/without_preload.json",
		},
		{
			name:             "with_keyword",
			query:            "preload=false&keyword=thanh",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_keyword.json",
		},
		{
			name:             "with_stackid",
			query:            "preload=false&stackID=0ecf47c8-cca4-4c30-94bb-054b1124c44f",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_stackid.json",
		},
		{
			name:             "with_projectid",
			query:            "preload=false&projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_projectid.json",
		},
		{
			name:             "with_projectid_and_positionid",
			query:            "preload=false&projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15&positionID=01fb6322-d727-47e3-a242-5039ea4732fc",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_projectid_and_positionid.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/employees", nil)
			ctx.Request.Header.Set("Authorization", tokenTest)
			ctx.Request.URL.RawQuery = tt.query

			h := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)
			h.List(ctx)
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
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             UpdateGeneralInfoInput
		id               string
	}{
		{
			name:             "ok_update_general_info",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_general_info/200.json",
			body: UpdateGeneralInfoInput{
				FullName: "Phạm Đức Thành",
				Email:    "thanh@d.foundation",
				Phone:    "0123456788",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "fail_wrong_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_general_info/404.json",
			body: UpdateGeneralInfoInput{
				FullName: "Phạm Đức Thành",
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
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)

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

func Test_UpdateSkill(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             UpdateSkillsInput
		id               string
	}{
		{
			name:             "ok_update_skill",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_skills/200.json",
			body: UpdateSkillsInput{
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Chapter:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				Seniority: model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
				},
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "failed_invalid_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/404.json",
			body: UpdateSkillsInput{
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Chapter:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				Seniority: model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
				},
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.body)
			fmt.Println(err)
			require.Nil(t, err)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(string(byteReq))
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/employees/%s/skills", tt.id), bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)

			metadataHandler.UpdateSkills(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
			require.Nil(t, err)

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateSkills] response mismatched")
		})
	}
}

func Test_Create(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             CreateEmployeeInput
		id               string
	}{
		{
			name:             "exists_user_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/existed_user.json",
			body: CreateEmployeeInput{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				Status:        model.WorkingStatusOnBoarding.String(),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "validation_err_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/validation_err.json",
			body: CreateEmployeeInput{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoigmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				Status:        model.WorkingStatusOnBoarding.String(),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "invalid_uuid_create",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/create/invalid_uuid.json",
			body: CreateEmployeeInput{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("c44c987c-ad34-4745-be2b-942e8670d32a")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				RoleID:        model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				Status:        model.WorkingStatusOnBoarding.String(),
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
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)

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

func Test_UpdatePersonalInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	dob, err := time.Parse("2006-01-02", "1990-01-02")
	require.Nil(t, err)
	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             UpdatePersonalInfoInput
		id               string
	}{
		{
			name:             "ok_update_personal_info",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_personal_info/200.json",
			body: UpdatePersonalInfoInput{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham123@gmail.com",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "fail_wrong_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/404.json",
			body: UpdatePersonalInfoInput{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham123@gmail.com",
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
			ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/"+tt.id+"/personal-info", bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)

			metadataHandler.UpdatePersonalInfo(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			if !tt.wantErr {
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdatePersonalInfo] response mismatched")
			} else {
				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdatePersonalInfo] response mismatched")
			}
		})
	}
}
