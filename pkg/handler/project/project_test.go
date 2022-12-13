package project

import (
	"bytes"
	"fmt"
	"github.com/dwarvesf/fortress-api/pkg/handler/project/request"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_UpdateProjectStatus(t *testing.T) {
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
		request          request.UpdateAccountStatusBody
		id               string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_project_status/200.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         404,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/404.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b11",
		},
		{
			name:             "failed_invalid_value_for_status",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/400.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "activee",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}

	for _, tt := range tests {
		byteReq, err := json.Marshal(tt.request)
		require.Nil(t, err)

		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			bodyReader := strings.NewReader(string(byteReq))
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%s/status", tt.id), bodyReader)
			ctx.Request.Header.Set("Authorization", testToken)
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock)

			metadataHandler.UpdateProjectStatus(ctx)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.Equal(t, tt.wantCode, w.Code)
			res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
			require.Nil(t, err)
			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateProjectStatus] response mismatched")
		})
	}
}

func TestHandler_Create(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		args             request.CreateProjectInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			args: request.CreateProjectInput{
				Name:              "project1",
				Status:            string(model.ProjectStatusOnBoarding),
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				DeliveryManagerID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail:      "a@gmail.com",
				ClientEmail:       "b@gmail.com",
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create/200.json",
		},
		{
			name: "invalid_status",
			args: request.CreateProjectInput{
				Name:              "project1",
				Status:            "something",
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail:      "a@gmail.com",
				ClientEmail:       "b@gmail.com",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_status.json",
		},
		{
			name: "missing_status",
			args: request.CreateProjectInput{
				Name:              "project1",
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail:      "a@gmail.com",
				ClientEmail:       "b@gmail.com",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_misssing_status.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Error(err)
				return
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewBuffer(body))
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.Header.Set("Content-Type", gin.MIMEJSON)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.Create(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "id")
			res, _ = utils.RemoveFieldInResponse(res, "createdAt")
			res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.Create] response mismatched")
		})
	}
}

func TestHandler_GetMembers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		id               string
		query            string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "status=active",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_members/200.json",
		},
		{
			name:             "happy_case_without_preload",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "preload=false",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_members/without_preload.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%v/members", tt.id), nil)
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.URL.RawQuery = tt.query
			ctx.AddParam("id", tt.id)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.GetMembers(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "id")
			res, _ = utils.RemoveFieldInResponse(res, "createdAt")
			res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.GetMembers] response mismatched")
		})
	}
}

func TestHandler_UpdateMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	repoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		id               string
		args             request.UpdateMemberInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				JoinedDate:     "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_member/200_success.json",
		},
		{
			name: "invalid_joined_date",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				JoinedDate:     "2022-13-01",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/400_invalid_joined_date.json",
		},
		{
			name: "project_not_found",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b16",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				JoinedDate:     "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_member/404_project_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Error(err)
				return
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/projects/%v/members", tt.id), bytes.NewBuffer(body))
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.Header.Set("Content-Type", gin.MIMEJSON)
			ctx.AddParam("id", tt.id)

			h := New(storeMock, repoMock, serviceMock, loggerMock)
			h.UpdateMember(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "projectSlotID")
			res, _ = utils.RemoveFieldInResponse(res, "projectMemberID")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.UpdateMember] response mismatched")
		})
	}
}

func TestHandler_AssignMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	repoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		id               string
		args             request.AssignMemberInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.AssignMemberInput{
				EmployeeID:  model.MustGetUUIDFromString("fae443f8-e8ff-4eec-b86c-98216d7662d8"),
				SeniorityID: model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				JoinedDate:     "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         false,
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/assign_member/200_success.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Error(err)
				return
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%v/members", tt.id), bytes.NewBuffer(body))
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.Header.Set("Content-Type", gin.MIMEJSON)
			ctx.AddParam("id", tt.id)

			h := New(storeMock, repoMock, serviceMock, loggerMock)
			h.AssignMember(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "projectSlotID")
			res, _ = utils.RemoveFieldInResponse(res, "projectMemberID")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.AssignMember] response mismatched")
		})
	}
}

func TestHandler_DeleteProjectMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		memberID         string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         200,
			wantResponsePath: "testdata/delete_member/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "fae443f8-e8ff-4eec-b86c-98216d7662d8",
		},
		{
			name:             "failed_invalid_member_id",
			wantCode:         404,
			wantResponsePath: "testdata/delete_member/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}, gin.Param{Key: "memberID", Value: tt.memberID}}
			ctx.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s/members/%s", tt.id, tt.memberID), nil)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock)

			metadataHandler.DeleteMember(ctx)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectStatus] response mismatched")
		})
	}
}

func TestHandler_Detail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		id               string
		query            string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_project/200.json",
		},
		{
			name:             "not_found",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b11",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_project/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%v", tt.id), nil)
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.URL.RawQuery = tt.query
			ctx.AddParam("id", tt.id)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.Details(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "createdAt")
			res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.Details] response mismatched")
		})
	}
}

func TestHandler_UpdateGeneralInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		input            request.UpdateProjectGeneralInfoInput
	}{
		{
			name:             "ok_update_project_general_infomation",
			wantCode:         200,
			wantResponsePath: "testdata/update_general_info/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
			},
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         404,
			wantResponsePath: "testdata/update_general_info/404.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d30715",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.input)
			require.Nil(t, err)

			bodyReader := strings.NewReader(string(byteReq))

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/general-info", tt.id), bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock)

			metadataHandler.UpdateGeneralInfo(ctx)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectGeneralInfo] response mismatched")
		})
	}
}

func TestHandler_UpdateContactInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		input            request.UpdateContactInfoInput
	}{
		{
			name:             "ok_update_project_contact_infomation",
			wantCode:         200,
			wantResponsePath: "testdata/update_contact_info/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateContactInfoInput{
				ClientEmail:       "fortress@gmai.com",
				ProjectEmail:      "fortress@d.foundation",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
			},
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         404,
			wantResponsePath: "testdata/update_contact_info/404.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d30714",
			input: request.UpdateContactInfoInput{
				ClientEmail:       "fortress@gmai.com",
				ProjectEmail:      "fortress@d.foundation",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byteReq, err := json.Marshal(tt.input)
			require.Nil(t, err)

			bodyReader := strings.NewReader(string(byteReq))

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
			ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/contact-info", tt.id), bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock)

			metadataHandler.UpdateContactInfo(ctx)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectContactInfo] response mismatched")
		})
	}
}

func TestHandler_GetListWorkUnit(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		query            string
	}{
		{
			name:             "ok_get_list_work_unit",
			wantCode:         200,
			wantResponsePath: "testdata/get_list_work_unit/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "status=active",
		},
		{
			name:             "err_invalid_status",
			wantCode:         400,
			wantResponsePath: "testdata/get_list_work_unit/400.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "status=activ",
		},
		{
			name:             "err_project_not_found",
			wantCode:         404,
			wantResponsePath: "testdata/get_list_work_unit/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b14",
			query:            "status=active",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.AddParam("id", tt.id)
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/work-units?%s", tt.id, tt.query), nil)
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.URL.RawQuery = tt.query

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.GetWorkUnits(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), string(w.Body.Bytes()), "[Handler.Project.GetListWorkUnit] response mismatched")
		})
	}
}

func TestHandler_UpdateWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		input            request.UpdateWorkUnitInput
		wantCode         int
		wantResponsePath string
	}{
		{
			name: "happy_case",
			input: request.UpdateWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "69b32f7e-0433-4566-a801-72909172940e",
				Body: request.UpdateWorkUnitBody{
					Name: "New Fortress Web",
					Type: model.WorkUnitTypeManagement,
					URL:  "https://github.com/dwarvesf/fortress-web",
					Members: []model.UUID{
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
					},
					Stacks: []model.UUID{
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					},
				},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_work_unit/200.json",
		},
		{
			name: "not_found_project",
			input: request.UpdateWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b16",
				WorkUnitID: "69b32f7e-0433-4566-a801-72909172940e",
				Body: request.UpdateWorkUnitBody{
					Name: "New Fortress Web",
					Type: model.WorkUnitTypeManagement,
					URL:  "https://github.com/dwarvesf/fortress-web",
					Members: []model.UUID{
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
					},
					Stacks: []model.UUID{
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					},
				},
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_work_unit/404.json",
		},
		{
			name: "invalid_type",
			input: request.UpdateWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "69b32f7e-0433-4566-a801-72909172940e",
				Body: request.UpdateWorkUnitBody{
					Name: "New Fortress Web",
					Type: "type",
					URL:  "https://github.com/dwarvesf/fortress-web",
					Members: []model.UUID{
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
					},
					Stacks: []model.UUID{
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_work_unit/400.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input.Body)
			if err != nil {
				t.Error(err)
				return
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/v1/projects/%s/work-units/%s", tt.input.ProjectID, tt.input.WorkUnitID),
				bytes.NewBuffer(body))
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.Header.Set("Content-Type", gin.MIMEJSON)
			ctx.AddParam("id", tt.input.ProjectID)
			ctx.AddParam("workUnitID", tt.input.WorkUnitID)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.UpdateWorkUnit(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Project.UpdateWorkUnit] response mismatched")
		})
	}
}

func TestHandler_CreateWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		input            request.CreateWorkUnitInput
		wantCode         int
		wantResponsePath string
	}{
		{
			name: "happy_case",
			input: request.CreateWorkUnitInput{
				ProjectID: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				Body: request.CreateWorkUnitBody{
					Name:   "workunit1",
					Type:   model.WorkUnitTypeDevelopment.String(),
					Status: model.WorkUnitStatusArchived.String(),
					URL:    "https://github.com/dwarvesf/fortress-api",
					Members: []model.UUID{
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
					},
					Stacks: []model.UUID{
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					},
				},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create_work_unit/200_success.json",
		},
		{
			name: "project_not_found",
			input: request.CreateWorkUnitInput{
				ProjectID: "8dc3be2e-19a4-4942-8a79-56db391a0b16",
				Body: request.CreateWorkUnitBody{
					Name:   "workunit1",
					Type:   model.WorkUnitTypeDevelopment.String(),
					Status: model.WorkUnitStatusArchived.String(),
					URL:    "https://github.com/dwarvesf/fortress-api",
					Members: []model.UUID{
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
					},
					Stacks: []model.UUID{
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					},
				},
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/create_work_unit/404_project_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.input.Body)
			if err != nil {
				t.Error(err)
				return
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/api/v1/projects/%s/work-units", tt.input.ProjectID),
				bytes.NewBuffer(body))
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.Header.Set("Content-Type", gin.MIMEJSON)
			ctx.AddParam("id", tt.input.ProjectID)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock)
			h.CreateWorkUnit(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res := w.Body.Bytes()
			res, _ = utils.RemoveFieldInResponse(res, "id")

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.CreateWorkUnit] response mismatched")
		})
	}
}

func TestHandler_ArchiveWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	repoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		input            request.ArchiveWorkUnitInput
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/archive_work_unit/200_ok.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "invalid_project_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/archive_work_unit/404_project_not_found.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b16",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPut,
				fmt.Sprintf("/api/v1/projects/%s/work-units/%s/archive", tt.input.ProjectID, tt.input.WorkUnitID),
				nil)

			ctx.Request.Header.Set("Authorization", testToken)
			ctx.AddParam("id", tt.input.ProjectID)
			ctx.AddParam("workUnitID", tt.input.WorkUnitID)

			h := New(storeMock, repoMock, serviceMock, loggerMock)
			h.ArchiveWorkUnit(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.ArchiveWorkUnit] response mismatched")
		})
	}
}

func TestHandler_UnarchiveWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	repoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		input            request.ArchiveWorkUnitInput
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/unarchive_work_unit/200_ok.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "invalid_project_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/unarchive_work_unit/404_project_not_found.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b16",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPut,
				fmt.Sprintf("/api/v1/projects/%s/work-units/%s/unarchive", tt.input.ProjectID, tt.input.WorkUnitID),
				nil)

			ctx.Request.Header.Set("Authorization", testToken)
			ctx.AddParam("id", tt.input.ProjectID)
			ctx.AddParam("workUnitID", tt.input.WorkUnitID)

			h := New(storeMock, repoMock, serviceMock, loggerMock)
			h.UnarchiveWorkUnit(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UnarchiveWorkUnit] response mismatched")
		})
	}
}
