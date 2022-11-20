package project

import (
	"bytes"
	"fmt"
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

const tokenTest = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

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
		request          updateAccountStatusBody
		id               string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_project_status/200.json",
			request: updateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         500,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/500.json",
			request: updateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b11",
		},
		{
			name:             "failed_invalid_value_for_status",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/400.json",
			request: updateAccountStatusBody{
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
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
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
		args             CreateProjectInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			args: CreateProjectInput{
				Name:              "project1",
				Status:            string(model.ProjectStatusOnBoarding),
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				DeliveryManagerID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create/200.json",
		},
		{
			name: "invalid_status",
			args: CreateProjectInput{
				Name:              "project1",
				Status:            "something",
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_status.json",
		},
		{
			name: "missing_status",
			args: CreateProjectInput{
				Name:              "project1",
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
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
			ctx.Request.Header.Set("Authorization", tokenTest)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%v/members", tt.id), nil)
			ctx.Request.Header.Set("Authorization", tokenTest)
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
		args             UpdateMemberInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusPending.String(),
				JoinedDate:     "2022-11-15",
				LeftDate:       "2023-11-15",
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
			args: UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusPending.String(),
				JoinedDate:     "2022-13-01",
				LeftDate:       "2023-11-15",
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
			args: UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusPending.String(),
				JoinedDate:     "2022-11-15",
				LeftDate:       "2023-11-15",
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
			ctx.Request.Header.Set("Authorization", tokenTest)
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
		args             AssignMemberInput
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name: "happy_case",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: AssignMemberInput{
				EmployeeID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID: model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusPending.String(),
				JoinedDate:     "2022-11-15",
				LeftDate:       "2023-11-15",
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
			ctx.Request.Header.Set("Authorization", tokenTest)
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
		memberID         string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         200,
			wantResponsePath: "testdata/delete_member/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "2655832e-f009-4b73-a535-64c3a22e558f",
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

			require.JSONEq(t, string(expRespRaw), string(w.Body.Bytes()), "[Handler.UpdateProjectStatus] response mismatched")
		})
	}
}
func TestHandler_Details(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name string

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
			ctx.Request.Header.Set("Authorization", tokenTest)
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
