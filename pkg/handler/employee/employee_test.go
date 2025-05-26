package employee

import (
	"context"
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
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/handler/employee/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk1ODMzMzA5NDUsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIn0.oIdlwWGBy4E1CbSoEX6r2B6NQLbew_J-RttpAcg6w8M"

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

	tests := []struct {
		name             string
		body             request.GetListEmployeeQuery
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
			body: request.GetListEmployeeQuery{
				WorkingStatuses: []string{"contractor"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/have_working_and_no_pagination.json",
		},
		{
			name: "have_workingStatuses_and_pagination",
			body: request.GetListEmployeeQuery{
				Pagination: view.Pagination{
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
			body: request.GetListEmployeeQuery{
				Pagination: view.Pagination{
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
			body: request.GetListEmployeeQuery{
				Preload:         false,
				WorkingStatuses: []string{"probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_preload_false.json",
		},
		{
			name: "without_preload",
			body: request.GetListEmployeeQuery{
				Preload:         false,
				WorkingStatuses: []string{"probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/without_preload.json",
		},
		{
			name: "with_keyword",
			body: request.GetListEmployeeQuery{
				Preload: false,
				Keyword: "thanh",
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_keyword.json",
		},
		{
			name: "with_stack_code",
			body: request.GetListEmployeeQuery{
				Preload: false,
				Stacks:  []string{"golang"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_stack_code.json",
		},
		{
			name: "with_project_code_and_position_code",
			body: request.GetListEmployeeQuery{
				Preload:   false,
				Projects:  []string{"fortress"},
				Positions: []string{"blockchain"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_project_code_and_position_code.json",
		},
		{
			name: "with_list_working_status",
			body: request.GetListEmployeeQuery{
				Preload:         false,
				WorkingStatuses: []string{"contractor", "probation"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_list_working_status.json",
		},
		{
			name: "invalid_position_code",
			body: request.GetListEmployeeQuery{
				Positions: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_position_code.json",
		},
		{
			name: "invalid_chapter_code",
			body: request.GetListEmployeeQuery{
				Chapters: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_chapter_code.json",
		},
		{
			name: "invalid_organization_code",
			body: request.GetListEmployeeQuery{
				Organizations: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_organization_code.json",
		},
		{
			name: "invalid_project_code",
			body: request.GetListEmployeeQuery{
				Projects: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_project_code.json",
		},
		{
			name: "invalid_seniority_code",
			body: request.GetListEmployeeQuery{
				Seniorities: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_seniority_code.json",
		},
		{
			name: "invalid_stack_code",
			body: request.GetListEmployeeQuery{
				Stacks: []string{""},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list/invalid_stack_code.json",
		},
		{
			name: "with_multiple_stacks",
			body: request.GetListEmployeeQuery{
				Stacks: []string{"golang", "react"},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/with_multiple_stack.json",
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

				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
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
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

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

				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Details(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Employee.Details] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateEmployeeStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

	tests := []struct {
		name             string
		wantCode         int
		id               string
		body             request.UpdateWorkingStatusRequest
		wantResponsePath string
	}{
		{
			name:     "ok_update_employee_status",
			wantCode: http.StatusOK,
			id:       "2655832e-f009-4b73-a535-64c3a22e558f",
			body: request.UpdateWorkingStatusRequest{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/200.json",
		},
		{
			name:     "empty_employee_id",
			wantCode: http.StatusBadRequest,
			id:       "",
			body: request.UpdateWorkingStatusRequest{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/400.json",
		},
		{
			name:     "wrong_format_employee_id",
			wantCode: http.StatusBadRequest,
			id:       "2655832e-f009-4b73-a535-64c3a22e558fa",
			body: request.UpdateWorkingStatusRequest{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/400.json",
		},
		{
			name:     "invalid_employee_status",
			wantCode: http.StatusBadRequest,
			id:       "2655832e-f009-4b73-a535-64c3a22e558f",
			body: request.UpdateWorkingStatusRequest{
				EmployeeStatus: "contractorr",
			},
			wantResponsePath: "testdata/update_employee_status/invalid_employee_status.json",
		},
		{
			name:     "employee_not_found",
			wantCode: http.StatusNotFound,
			id:       "2655832e-f009-4b73-a535-64c3a22e558d",
			body: request.UpdateWorkingStatusRequest{
				EmployeeStatus: "contractor",
			},
			wantResponsePath: "testdata/update_employee_status/404.json",
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

				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
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

//func Test_UpdateGeneralInfo(t *testing.T) {
//	cfg := config.LoadTestConfig()
//	loggerMock := logger.NewLogrusLogger()
//	serviceMock := service.New(&cfg, nil, nil)
//	storeMock := store.New()
//	queue := make(chan model.WorkerMessage, 1000)
//	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)
//
//	tests := []struct {
//		name             string
//		wantCode         int
//		wantErr          bool
//		wantResponsePath string
//		body             request.UpdateEmployeeGeneralInfoInput
//		id               string
//	}{
//		{
//			name:             "ok_update_general_info",
//			wantCode:         http.StatusOK,
//			wantErr:          false,
//			wantResponsePath: "testdata/update_general_info/200.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName:    "Phạm Đức Thành",
//				Email:       "thanh@d.foundation",
//				Phone:       "0123456788",
//				DisplayName: "new",
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558f",
//		},
//		{
//			name:             "fail_wrong_employee_id",
//			wantCode:         http.StatusNotFound,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/404.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName: "Phạm Đức Thành",
//				Email:    "thanh@d.foundation",
//				Phone:    "0123456788",
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558a",
//		},
//		{
//			name:             "empty_employee_id",
//			wantCode:         http.StatusBadRequest,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/400.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName: "Phạm Đức Thành",
//				Email:    "thanh@d.foundation",
//				Phone:    "0123456788",
//			},
//			id: "",
//		},
//		{
//			name:             "wrong_format_employee_id",
//			wantCode:         http.StatusBadRequest,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/400.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName: "Phạm Đức Thành",
//				Email:    "thanh@d.foundation",
//				Phone:    "0123456788",
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558aa",
//		},
//		{
//			name:             "not_found_line_manager",
//			wantCode:         http.StatusNotFound,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/line_manager_not_found.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName:      "Phạm Đức Thành",
//				Email:         "thanh@d.foundation",
//				Phone:         "0123456788",
//				LineManagerID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558b"),
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558a",
//		},
//		{
//			name:             "invalid_join_date",
//			wantCode:         http.StatusBadRequest,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/invalid_join_date.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName:   "Phạm Đức Thành",
//				Email:      "thanh@d.foundation",
//				Phone:      "0123456788",
//				JoinedDate: "2006-13-12",
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558f",
//		},
//		{
//			name:             "organization_not_found",
//			wantCode:         http.StatusNotFound,
//			wantErr:          true,
//			wantResponsePath: "testdata/update_general_info/organization_not_found.json",
//			body: request.UpdateEmployeeGeneralInfoInput{
//				FullName:        "Phạm Đức Thành",
//				Email:           "thanh@d.foundation",
//				Phone:           "0123456788",
//				OrganizationIDs: []model.UUID{model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f")},
//			},
//			id: "2655832e-f009-4b73-a535-64c3a22e558f",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
//				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_general_info/update_general_info.sql")
//				byteReq, err := json.Marshal(tt.body)
//				require.Nil(t, err)
//				w := httptest.NewRecorder()
//
//				ctx, _ := gin.CreateTestContext(w)
//				bodyReader := strings.NewReader(string(byteReq))
//				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
//				ctx.Request = httptest.NewRequest("PUT", "/api/v1/employees/"+tt.id+"/general-info", bodyReader)
//				ctx.Request.Header.Set("Authorization", testToken)
//				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
//				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
//
//				h.UpdateGeneralInfo(ctx)
//
//				require.Equal(t, tt.wantCode, w.Code)
//				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
//				require.NoError(t, err)
//
//				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
//				require.Nil(t, err)
//
//				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateGeneralInfo] response mismatched")
//			})
//		})
//	}
//}

func Test_UpdateSkill(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)
	// testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.UpdateSkillsRequest
		id               string
	}{
		{
			name:             "ok_update_skill",
			wantCode:         http.StatusOK,
			wantErr:          false,
			wantResponsePath: "testdata/update_skills/200.json",
			body: request.UpdateSkillsRequest{
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
			name:             "not_found_employee",
			wantCode:         http.StatusNotFound,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/404.json",
			body: request.UpdateSkillsRequest{
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
		{
			name:             "empty_employee_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/400.json",
			body: request.UpdateSkillsRequest{
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
			id: "",
		},
		{
			name:             "wrong_format_employee_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/400.json",
			body: request.UpdateSkillsRequest{
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
			id: "2655832e-f009-4b73-a535-64c3a22e558aa",
		},
		{
			name:             "invalid_stack",
			wantCode:         http.StatusNotFound,
			wantErr:          true,
			wantResponsePath: "testdata/update_skills/stack_not_found.json",
			body: request.UpdateSkillsRequest{
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Chapters:  []model.UUID{model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8")},
				Seniority: model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44d"),
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
				},
				LeadingChapters: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				},
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
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
				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				h.UpdateSkills(ctx)

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
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.CreateEmployeeRequest
		id               string
	}{
		{
			name:             "exists_user_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/existed_user.json",
			body: request.CreateEmployeeRequest{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				Roles: []model.UUID{
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Status:     model.WorkingStatusOnBoarding.String(),
				JoinedDate: time.Now().Format("2006-01-02"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "validation_err_create",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/create/validation_err.json",
			body: request.CreateEmployeeRequest{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoigmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				Roles: []model.UUID{
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Status:     model.WorkingStatusOnBoarding.String(),
				JoinedDate: time.Now().Format("2006-01-02"),
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "invalid_uuid_create",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/create/invalid_uuid.json",
			body: request.CreateEmployeeRequest{
				FullName:      "Lê Nguyễn Minh Khôi",
				DisplayName:   "Khoi Le",
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "khoi@gmail.com",
				Positions:     []model.UUID{model.MustGetUUIDFromString("c44c987c-ad34-4745-be2b-942e8670d32a")},
				Salary:        300,
				SeniorityID:   model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
				Roles: []model.UUID{
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				Status:     model.WorkingStatusOnBoarding.String(),
				JoinedDate: time.Now().Format("2006-01-02"),
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
				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				h.Create(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[employee.Handler.Create] response mismatched")
			})
		})
	}
}

func Test_UpdatePersonalInfo(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

	dob, err := time.Parse("2006-01-02", "1990-01-02")
	require.Nil(t, err)
	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		body             request.UpdatePersonalInfoRequest
		id               string
	}{
		{
			name:             "ok_update_personal_info",
			wantCode:         http.StatusOK,
			wantErr:          false,
			wantResponsePath: "testdata/update_personal_info/200.json",
			body: request.UpdatePersonalInfoRequest{
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
			name:             "employee_not_found",
			wantCode:         http.StatusNotFound,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/404.json",
			body: request.UpdatePersonalInfoRequest{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham124@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minh",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558a",
		},
		{
			name:             "empty_employee_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/400.json",
			body: request.UpdatePersonalInfoRequest{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham124@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minh",
			},
			id: "",
		},
		{
			name:             "wrong_employee_id_format",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/400.json",
			body: request.UpdatePersonalInfoRequest{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham124@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minh",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558aa",
		},
		{
			name:             "invalid_country",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_personal_info/invalid_country.json",
			body: request.UpdatePersonalInfoRequest{
				DoB:           &dob,
				Gender:        "Male",
				Address:       "Phan Huy Ich, Tan Binh District, Ho Chi Minh, Vietnam",
				PersonalEmail: "thanhpham124@gmail.com",
				Country:       "Vietnam",
				City:          "Hồ Chí Minhh",
			},
			id: "2655832e-f009-4b73-a535-64c3a22e558f",
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
				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				h.UpdatePersonalInfo(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, _ := utils.RemoveFieldInResponse(w.Body.Bytes(), "createdAt")
				res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdatePersonalInfo] response mismatched")
			})
		})
	}
}

func TestHandler_GetLineManagers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.NewForTest()
	storeMock := store.New()
	queue := make(chan model.WorkerMessage, 1000)
	workerMock := worker.New(context.Background(), queue, serviceMock, loggerMock)

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

				ctrl := controller.New(storeMock, txRepo, serviceMock, workerMock, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetLineManagers(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Employee.GetLineManagers] response mismatched")
			})
		})
	}
}
