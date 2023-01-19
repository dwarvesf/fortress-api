package employee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		body             request.GetListEmployeeInput
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
			name: "have_workingStatus_and_no_pagination",
			body: request.GetListEmployeeInput{
				WorkingStatuses: []string{"contractor"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/have_working_and_no_pagination.json",
		},
		{
			name: "have_workingStatuses_and_pagination",
			body: request.GetListEmployeeInput{
				Pagination: model.Pagination{
					Page: 1,
					Size: 5,
				},
				WorkingStatuses: []string{"contractor"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/have_working_and_pagination.json",
		},
		{
			name: "out_of_content",
			body: request.GetListEmployeeInput{
				Pagination: model.Pagination{
					Page: 5,
					Size: 5,
				},
				WorkingStatuses: []string{"contractor"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/out_of_content.json",
		},
		{
			name: "with_preload_false",
			body: request.GetListEmployeeInput{
				Preload:         false,
				WorkingStatuses: []string{"probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_preload_false.json",
		},
		{
			name: "without_preload",
			body: request.GetListEmployeeInput{
				Preload:         false,
				WorkingStatuses: []string{"probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/without_preload.json",
		},
		{
			name: "with_keyword",
			body: request.GetListEmployeeInput{
				Preload: false,
				Keyword: "thanh",
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_keyword.json",
		},
		{
			name: "with_stack_code",
			body: request.GetListEmployeeInput{
				Preload: false,
				Stacks:  []string{"golang"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_stack_code.json",
		},
		{
			name: "with_project_code_and_position_code",
			body: request.GetListEmployeeInput{
				Preload:   false,
				Projects:  []string{"fortress"},
				Positions: []string{"blockchain"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_project_code_and_position_code.json",
		},
		{
			name: "with_list_working_status",
			body: request.GetListEmployeeInput{
				Preload:         false,
				WorkingStatuses: []string{"contractor", "probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_list_working_status.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/list/list.sql")
				w := httptest.NewRecorder()
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/employees/search", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.List(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "updatedAt")
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Employee.List] response mismatched")
			})
		})
	}
}

func TestHandler_One(t *testing.T) {
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
			id:               "2655832e-f009-4b73-a535-64c3a22e558f",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/one/200.json",
		},
		{
			name:             "not_found",
			id:               "2655832e-f009-4b73-a535-64c3a22e558e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/one/404.json",
		},
		{
			name:             "happy_case_username",
			id:               "thanh",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/one/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/one/one.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/employees/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.One(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Employee.One] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateEmployeeStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		id               string
		body             request.UpdateWorkingStatusInput
		wantResponsePath string
	}{
		{
			name:     "ok_update_employee_status",
			wantCode: http.StatusOK,
			id:       "2655832e-f009-4b73-a535-64c3a22e558f",
			body: request.UpdateWorkingStatusInput{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/200.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_employee_status/update_employee_status.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/employees/%s/employee-status", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdateEmployeeStatus(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateEmployeeStatus] response mismatched")
			})
		})
	}
}

func Test_UpdateGeneralInfo(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.UpdateEmployeeGeneralInfoInput
		id               string
	}{
		{
			name:             "ok_update_general_info",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_general_info/200.json",
			body: request.UpdateEmployeeGeneralInfoInput{
				FullName:    "Phạm Đức Thành",
				Email:       "thanh@d.foundation",
				Phone:       "0123456788",
				DisplayName: "new",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "fail_wrong_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_general_info/404.json",
			body: request.UpdateEmployeeGeneralInfoInput{
				FullName: "Phạm Đức Thành",
				Email:    "thanh@d.foundation",
				Phone:    "0123456788",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_general_info/update_general_info.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/"+tt.id+"/general-info", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateGeneralInfo(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateGeneralInfo] response mismatched")
			})
		})
	}
}

func Test_UpdateSkill(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	// testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.UpdateSkillsInput
		id               string
	}{
		{
			name:             "ok_update_skill",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_skills/200.json",
			body: request.UpdateSkillsInput{
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Chapters:  []model.UUID{model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8")},
				Seniority: model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
				},
				LeadingChapters: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				},
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "failed_invalid_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/404.json",
			body: request.UpdateSkillsInput{
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Chapters:  []model.UUID{model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8")},
				Seniority: model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
				},
				LeadingChapters: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				},
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_skills/update_skills.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/employees/%s/skills", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateSkills(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateSkills] response mismatched")
			})
		})
	}
}

func Test_Create(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.CreateEmployeeInput
		id               string
	}{
		{
			name:             "exists_user_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/existed_user.json",
			body: request.CreateEmployeeInput{
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
			body: request.CreateEmployeeInput{
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
			body: request.CreateEmployeeInput{
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
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create/create.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("POST", "/api/v1/employees/", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.Create(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				if !tt.wantErr {
					res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
					require.Nil(t, err)

					require.JSONEq(t, string(expRespRaw), string(res), "[employee.Handler.Create] response mismatched")
				} else {
					require.JSONEq(t, string(expRespRaw), w.Body.String(), "[employee.Handler.Create] response mismatched")
				}
			})
		})

	}
}

func Test_UpdatePersonalInfo(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	dob, err := time.Parse("2006-01-02", "1990-01-02")
	require.Nil(t, err)
	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.UpdatePersonalInfoInput
		id               string
	}{
		{
			name:             "ok_update_personal_info",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_personal_info/200.json",
			body: request.UpdatePersonalInfoInput{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham124@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minh",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "fail_wrong_employee_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/404.json",
			body: request.UpdatePersonalInfoInput{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham123@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minh",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_personal_info/update_personal_info.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/"+tt.id+"/personal-info", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdatePersonalInfo(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdatePersonalInfo] response mismatched")
			})
		})
	}
}

func Test_AddMentee(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.AddMenteeInput
		id               string
	}{
		{
			name:             "ok_add_mentee",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/add_mentee/200.json",
			body: request.AddMenteeInput{
				MenteeID: model.MustGetUUIDFromString("f7c6016b-85b5-47f7-8027-23c2db482197"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "employee_not_found",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/add_mentee/404.json",
			body: request.AddMenteeInput{
				MenteeID: model.MustGetUUIDFromString("f7c6016b-85b5-47f7-8027-23c2db482197"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558e",
		},
		{
			name:             "mentee_not_found",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/add_mentee/404_mentee_not_found.json",
			body: request.AddMenteeInput{
				MenteeID: model.MustGetUUIDFromString("f7c6016b-85b5-47f7-8027-23c2db482196"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "failed_mentor_themselves",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/add_mentee/400_mentor_themselves.json",
			body: request.AddMenteeInput{
				MenteeID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "failed_mentee_their_mentor",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/add_mentee/400_mentee_their_mentor.json",
			body: request.AddMenteeInput{
				MenteeID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
			},
			id: "d389d35e-c548-42cf-9f29-2a599969a8f2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/add_mentee/add_mentee.sql")
				byteReq, err := json.Marshal(tt.body)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("POST", "/api/v1/employees/"+tt.id+"/mentees", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.AddMentee(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.AddMentee] response mismatched")
			})
		})
	}
}

func Test_DeleteMentee(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		menteeID         string
		id               string
	}{
		{
			name:             "ok_delete_mentee",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/delete_mentee/200.json",
			id:               "2655832e-f009-4b73-a535-64c3a22e558f",
			menteeID:         "fae443f8-e8ff-4eec-b86c-98216d7662d8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_mentee/delete_mentee.sql")
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}, gin.Param{Key: "menteeID", Value: tt.menteeID}}
				ctx.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/employees/%s/mentees/%s", tt.id, tt.menteeID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.DeleteMentee(ctx)

				// require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.DeleteMentee] response mismatched")
			})
		})
	}
}

func TestHandler_GetLineManagers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		menteeID         string
		id               string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_line_managers/200_ok.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_line_managers/seed.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/line-managers", nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetLineManagers(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Employee.GetLineManagers] response mismatched")
			})
		})
	}
}
