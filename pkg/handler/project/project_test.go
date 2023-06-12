package project

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dwarvesf/fortress-api/pkg/controller"

	"github.com/dwarvesf/fortress-api/pkg/handler/project/request"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_Detail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		id               string
		query            string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case_slug",
			id:               "fortress",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_project/200.json",
		},
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
		{
			name:             "not_found_slug",
			id:               "a",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_project/404.json",
		},
		{
			name:             "empty_project_id",
			id:               "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_project/invalid_project_id.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_project/get_project.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%v", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query
				ctx.AddParam("id", tt.id)

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Details(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "createdAt")
				res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.Details] response mismatched")
			})
		})
	}
}

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
		query            string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_projects/200.json",
			query:            "sort=-updatedAt",
		},
		{
			name:             "happy_case_with_pagination",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_projects/200_with_paging.json",
			query:            "page=1&size=1&sort=-updatedAt",
		},
		{
			name:             "invalid_project_type",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_projects/invalid_project_type.json",
			query:            "type=a",
		},
		{
			name:             "invalid_project_status",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_projects/invalid_project_status.json",
			query:            "status=a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_projects/get_projects.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.List(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "createdAt")
				res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.List] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateProjectStatus(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
			wantCode:         http.StatusOK,
			wantErr:          false,
			wantResponsePath: "testdata/update_project_status/200.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         http.StatusNotFound,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/404.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "active",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b11",
		},
		{
			name:             "failed_invalid_value_for_status",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/400.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "activee",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/invalid_project_id.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "activee",
			},
			id: "",
		},
		{
			name:             "wrong_format_project_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          true,
			wantResponsePath: "testdata/update_project_status/invalid_project_id.json",
			request: request.UpdateAccountStatusBody{
				ProjectStatus: "activee",
			},
			id: "8dc3be2e-19a4-4942-8a79-56db391a0b15a",
		},
	}

	for _, tt := range tests {
		testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
			testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_project_status/update_project_status.sql")
			byteReq, err := json.Marshal(tt.request)
			require.Nil(t, err)

			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/projects/%s/status", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateProjectStatus(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.Equal(t, tt.wantCode, w.Code)
				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.Nil(t, err)
				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.UpdateProjectStatus] response mismatched")
			})
		})
	}
}

func TestHandler_Create(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
				Name:      "Project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:      model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail:   "a@gmail.com",
				ClientEmail:    []string{"b@gmail.com", "c@gmail.com"},
				Function:       model.ProjectFunctionLearning.String(),
				BankAccountID:  model.MustGetUUIDFromString("e79eb5b3-e2cb-4d7f-9273-46f4be88cb20"),
				ClientID:       model.MustGetUUIDFromString("afb9cf05-9517-4fb9-a4f2-66e6d90ad215"),
				OrganizationID: model.MustGetUUIDFromString("31fdf38f-77c0-4c06-b530-e2be8bc297e0"),
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create/200.json",
		},
		{
			name: "duplicate_slug",
			args: request.CreateProjectInput{
				Name:      "Lorem Ipsum",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Code:         "lorem-ipsum",
				Function:     model.ProjectFunctionLearning.String(),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_duplicate_slug.json",
		},
		{
			name: "invalid_status",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    "something",
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionLearning.String(),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_status.json",
		},
		{
			name: "missing_status",
			args: request.CreateProjectInput{
				Name:      "project1",
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionLearning.String(),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_misssing_status.json",
		},
		{
			name: "missing_account_manager",
			args: request.CreateProjectInput{
				Name:      "Project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com", "c@gmail.com"},
				Function:     model.ProjectFunctionLearning.String(),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_missing_account_manager.json",
		},
		{
			name: "invalid_email_domain",
			args: request.CreateProjectInput{
				Name:      "Project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"bgmail.com", "c@gmail.com"},
				Function:     model.ProjectFunctionLearning.String(),
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_email.json",
		},
		{
			name: "invalid_function",
			args: request.CreateProjectInput{
				Name:      "Project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com", "c@gmail.com"},
				Function:     "a",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_function.json",
		},
		{
			name: "400_invalid_deployment_type",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionDevelopment.String(),
				Members: []request.AssignMemberInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						SeniorityID:    model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
						Positions:      []model.UUID{model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5")},
						DeploymentType: "a",
						Status:         model.ProjectMemberStatusOnBoarding.String(),
						Rate:           decimal.NewFromInt(10),
						StartDate:      "2022-11-14",
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_deployment_type.json",
		},
		{
			name: "invalid_project_member_status",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionDevelopment.String(),
				Members: []request.AssignMemberInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						SeniorityID:    model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
						Positions:      []model.UUID{model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5")},
						DeploymentType: model.MemberDeploymentTypeOfficial.String(),
						Status:         "a",
						Rate:           decimal.NewFromInt(10),
						StartDate:      "2022-11-14",
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_project_member_status.json",
		},
		{
			name: "empty_project_position",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionDevelopment.String(),
				Members: []request.AssignMemberInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						SeniorityID:    model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
						Positions:      []model.UUID{},
						DeploymentType: model.MemberDeploymentTypeOfficial.String(),
						Status:         model.ProjectMemberStatusOnBoarding.String(),
						Rate:           decimal.NewFromInt(10),
						StartDate:      "2022-11-14",
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_project_position_empty.json",
		},
		{
			name: "invalid_member_start_date",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionDevelopment.String(),
				Members: []request.AssignMemberInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						SeniorityID:    model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
						Positions:      []model.UUID{model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5")},
						DeploymentType: model.MemberDeploymentTypeOfficial.String(),
						Status:         model.ProjectMemberStatusOnBoarding.String(),
						Rate:           decimal.NewFromInt(10),
						StartDate:      "2022-13-14",
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_start_date_member.json",
		},
		{
			name: "invalid_member_end_date",
			args: request.CreateProjectInput{
				Name:      "project1",
				Status:    string(model.ProjectStatusOnBoarding),
				StartDate: "2022-11-14",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				CountryID:    model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				ProjectEmail: "a@gmail.com",
				ClientEmail:  []string{"b@gmail.com"},
				Function:     model.ProjectFunctionDevelopment.String(),
				Members: []request.AssignMemberInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						SeniorityID:    model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5"),
						Positions:      []model.UUID{model.MustGetUUIDFromString("39735742-829b-47f3-8f9d-daf0983914e5")},
						DeploymentType: model.MemberDeploymentTypeOfficial.String(),
						Status:         model.ProjectMemberStatusOnBoarding.String(),
						Rate:           decimal.NewFromInt(10),
						StartDate:      "2022-11-14",
						EndDate:        "2022-13-13",
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_end_date_member.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create/create.sql")
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

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Create(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "id")
				res, _ = utils.RemoveFieldInResponse(res, "createdAt")
				res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.Create] response mismatched")
			})
		})
	}
}

func TestHandler_GetMembers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
		{
			name:             "happy_case_with_unique",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "preload=false&distinct=true",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_members/distinct.json",
		},
		{
			name:             "invalid_project_id_format",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b156",
			query:            "preload=false",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_members/invalid_project_id.json",
		},
		{
			name:             "invalid_project_member_status",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "preload=false&status=invalid",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_members/invalid_project_member_status.json",
		},
		{
			name:             "project_not_found",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b16",
			query:            "preload=false",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_members/project_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_members/get_members.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/projects/%v/members", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query
				ctx.AddParam("id", tt.id)

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetMembers(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "id")
				res, _ = utils.RemoveFieldInResponse(res, "createdAt")
				res, _ = utils.RemoveFieldInResponse(res, "updatedAt")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.GetMembers] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
				Note:           "",
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_member/200_success.json",
		},
		{
			name: "invalid_start_date",
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
				StartDate:      "2022-13-01",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/400_invalid_start_date.json",
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
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_member/404_project_not_found.json",
		},
		{
			name: "empty_project_id",
			id:   "",
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
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/invalid_project_id.json",
		},
		{
			name: "invalid_project_id_format",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b168",
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
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/invalid_project_id.json",
		},
		{
			name: "invalid_deployment_type",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: "invalid",
				Status:         model.ProjectMemberStatusActive.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/invalid_deployment_type.json",
		},
		{
			name: "invalid_status",
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
				Status:         "invalid",
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_member/invalid_status.json",
		},
		{
			name: "seniority_not_found",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732f9"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusOnBoarding.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_member/seniority_not_found.json",
		},
		{
			name: "position_not_found",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb7"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec9"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusOnBoarding.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_member/position_not_found.json",
		},
		{
			name: "project_slot_not_found",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			args: request.UpdateMemberInput{
				ProjectSlotID: model.MustGetUUIDFromString("f32d08ca-8863-4ab3-8c84-a11849451eb8"),
				EmployeeID:    model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				SeniorityID:   model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusOnBoarding.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         true,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_member/project_slot_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_member/update_member.sql")
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

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdateMember(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "projectSlotID")
				res, _ = utils.RemoveFieldInResponse(res, "projectMemberID")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.UpdateMember] response mismatched")
			})
		})
	}
}

func TestHandler_AssignMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()

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
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         false,
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/assign_member/200_success.json",
		},
		{
			name: "empty_project_id",
			id:   "",
			args: request.AssignMemberInput{
				EmployeeID:  model.MustGetUUIDFromString("fae443f8-e8ff-4eec-b86c-98216d7662d8"),
				SeniorityID: model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         false,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/assign_member/invalid_project_id.json",
		},
		{
			name: "invalid_project_id_format",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b156",
			args: request.AssignMemberInput{
				EmployeeID:  model.MustGetUUIDFromString("fae443f8-e8ff-4eec-b86c-98216d7662d8"),
				SeniorityID: model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         false,
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/assign_member/invalid_project_id.json",
		},
		{
			name: "project_not_found",
			id:   "8dc3be2e-19a4-4942-8a79-56db391a0b16",
			args: request.AssignMemberInput{
				EmployeeID:  model.MustGetUUIDFromString("fae443f8-e8ff-4eec-b86c-98216d7662d8"),
				SeniorityID: model.MustGetUUIDFromString("01fb6322-d727-47e3-a242-5039ea4732fc"),
				Positions: []model.UUID{
					model.MustGetUUIDFromString("11ccffea-2cc9-4e98-9bef-3464dfe4dec8"),
					model.MustGetUUIDFromString("d796884d-a8c4-4525-81e7-54a3b6099eac"),
				},
				DeploymentType: model.MemberDeploymentTypeOfficial.String(),
				Status:         model.ProjectMemberStatusActive.String(),
				StartDate:      "2022-11-15",
				Rate:           decimal.NewFromInt(10),
				Discount:       decimal.NewFromInt(1),
				IsLead:         false,
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/assign_member/404_project_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/assign_member/assign_member.sql")
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
				ctrl := controller.New(storeMock, txRepo, serviceMock, nil, loggerMock, &cfg)
				h := New(ctrl, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.AssignMember(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "projectSlotID")
				res, _ = utils.RemoveFieldInResponse(res, "projectMemberID")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.AssignMember] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteProjectMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		memberID         string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete_member/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721aca",
		},
		{
			name:             "failed_invalid_member_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_member/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_member/empty_project_id.json",
			id:               "",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
		{
			name:             "empty_member_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_member/empty_member_id.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "",
		},
		{
			name:             "project_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_member/project_not_found.json",
			id:               "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_member/delete_member.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}, gin.Param{Key: "memberID", Value: tt.memberID}}
				ctx.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s/members/%s", tt.id, tt.memberID), nil)
				ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.DeleteMember(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.DeleteSlot] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteSlot(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		slotID           string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete_slot/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			slotID:           "f32d08ca-8863-4ab3-8c84-a11849451eb7",
		},
		{
			name:             "failed_invalid_member_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_slot/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			slotID:           "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_slot/empty_project_id.json",
			id:               "",
			slotID:           "f32d08ca-8863-4ab3-8c84-a11849451eb7",
		},
		{
			name:             "empty_member_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_slot/empty_slot_id.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			slotID:           "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_slot/delete_slot.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}, gin.Param{Key: "slotID", Value: tt.slotID}}
				ctx.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/projects/%s/slots/%s", tt.id, tt.slotID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.DeleteSlot(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.DeleteSlot] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateGeneralInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		input            request.UpdateProjectGeneralInfoInput
	}{
		{
			name:             "ok_update_project_general_information",
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
				Function:       model.ProjectFunctionManagement.String(),
				BankAccountID:  model.MustGetUUIDFromString("e79eb5b3-e2cb-4d7f-9273-46f4be88cb20"),
				ClientID:       model.MustGetUUIDFromString("afb9cf05-9517-4fb9-a4f2-66e6d90ad215"),
				OrganizationID: model.MustGetUUIDFromString("31fdf38f-77c0-4c06-b530-e2be8bc297e0"),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
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
				Function:       model.ProjectFunctionManagement.String(),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_general_info/invalid_project_id.json",
			id:               "",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
				Function:       model.ProjectFunctionManagement.String(),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_general_info/invalid_project_id.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d307156",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
				Function:       model.ProjectFunctionManagement.String(),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
		{
			name:             "invalid_project_function",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_general_info/invalid_project_function.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d30715",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
				Function:       "a",
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
		{
			name:             "country_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_general_info/country_not_found.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4d"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2e"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
				Function:       model.ProjectFunctionDevelopment.String(),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
		{
			name:             "stack_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_general_info/stack_not_found.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateProjectGeneralInfoInput{
				Name:      "Fortress",
				StartDate: "1990-01-02",
				CountryID: model.MustGetUUIDFromString("4ef64490-c906-4192-a7f9-d2221dadfe4c"),
				Stacks: []model.UUID{
					model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2d"),
					model.MustGetUUIDFromString("b403ef95-4269-4830-bbb6-8e56e5ec0af4"),
				},
				Function:       model.ProjectFunctionDevelopment.String(),
				AccountRating:  1,
				DeliveryRating: 1,
				LeadRating:     1,
				ImportantLevel: model.ProjectImportantLevelMedium.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_general_info/update_general_info.sql")
				byteReq, err := json.Marshal(tt.input)
				require.Nil(t, err)

				bodyReader := strings.NewReader(string(byteReq))

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/general-info", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateGeneralInfo(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectGeneralInfo] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateContactInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		input            request.UpdateContactInfoInput
	}{
		{
			name:             "ok_update_project_contact_information",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_contact_info/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "project_id_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_contact_info/404.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d30714",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_contact_info/invalid_project_id.json",
			id:               "",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_contact_info/invalid_project_id.json",
			id:               "d100efd1-bfce-4cd6-885c-1e4ac3d307149",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "invalid_email_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_contact_info/invalid_email_format.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortressgmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "account_manager_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_contact_info/account_manager_not_found.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558d"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
		{
			name:             "delivery_manager_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_contact_info/delivery_manager_not_found.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			input: request.UpdateContactInfoInput{
				ClientEmail:  []string{"fortress@gmail.com"},
				ProjectEmail: "fortress@d.foundation",
				AccountManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				DeliveryManagers: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98cd"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
				SalePersons: []request.ProjectHeadInput{
					{
						EmployeeID:     model.MustGetUUIDFromString("608ea227-45a5-4c8a-af43-6c7280d96340"),
						CommissionRate: decimal.NewFromInt(100),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_contact_info/update_contact_info.sql")
				byteReq, err := json.Marshal(tt.input)
				require.Nil(t, err)

				bodyReader := strings.NewReader(string(byteReq))

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/contact-info", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateContactInfo(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateProjectContactInfo] response mismatched")
			})
		})
	}
}

func TestHandler_GetListWorkUnit(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		query            string
	}{
		{
			name:             "ok_get_list_work_unit",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_list_work_unit/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "status=active",
		},
		{
			name:             "err_invalid_status",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_list_work_unit/400.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			query:            "status=activ",
		},
		{
			name:             "err_project_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_list_work_unit/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b14",
			query:            "status=active",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_list_work_unit/invalid_project_id.json",
			id:               "",
			query:            "status=active",
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_list_work_unit/invalid_project_id.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b156",
			query:            "status=active",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_list_work_unit/get_list_work_unit.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.AddParam("id", tt.id)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/projects/%s/work-units?%s", tt.id, tt.query), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetWorkUnits(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Project.GetListWorkUnit] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
		{
			name: "empty_project_id",
			input: request.UpdateWorkUnitInput{
				ProjectID:  "",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_work_unit/invalid_project_id.json",
		},
		{
			name: "invalid_format_project_id",
			input: request.UpdateWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b156",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_work_unit/invalid_project_id.json",
		},
		{
			name: "invalid_stack",
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
						model.MustGetUUIDFromString("fa0f4e46-7eab-4e5c-9d31-30489e69fe2d"),
						model.MustGetUUIDFromString("0ecf47c8-cca4-4c30-94bb-054b1124c44f"),
					}},
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/update_work_unit/stack_not_found.json",
		},
		{
			name: "invalid_stack",
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
					Stacks: []model.UUID{},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_work_unit/invalid_stack.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_work_unit/update_work_unit.sql")
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

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdateWorkUnit(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Project.UpdateWorkUnit] response mismatched")
			})
		})
	}
}

func TestHandler_CreateWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
		{
			name: "empty_project_id",
			input: request.CreateWorkUnitInput{
				ProjectID: "",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_work_unit/invalid_project_id.json",
		},
		{
			name: "invalid_project_id_format",
			input: request.CreateWorkUnitInput{
				ProjectID: "8dc3be2e-19a4-4942-8a79-56db391a0b156",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_work_unit/invalid_project_id.json",
		},
		{
			name: "invalid_work_unit_type",
			input: request.CreateWorkUnitInput{
				ProjectID: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				Body: request.CreateWorkUnitBody{
					Name:   "workunit1",
					Type:   "a",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_work_unit/invalid_work_unit_type.json",
		},
		{
			name: "invalid_work_unit_status",
			input: request.CreateWorkUnitInput{
				ProjectID: "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				Body: request.CreateWorkUnitBody{
					Name:   "workunit1",
					Type:   model.WorkUnitTypeDevelopment.String(),
					Status: "a",
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_work_unit/invalid_work_unit_status.json",
		},
		{
			name: "invalid_work_unit_stack",
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
					Stacks: []model.UUID{},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_work_unit/invalid_work_unit_stack.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create_work_unit/create_work_unit.sql")
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

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.CreateWorkUnit(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res := w.Body.Bytes()
				res, _ = utils.RemoveFieldInResponse(res, "id")

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Project.CreateWorkUnit] response mismatched")
			})
		})
	}
}

func TestHandler_ArchiveWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/archive_work_unit/invalid_project_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/archive_work_unit/invalid_project_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b156",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "empty_work_unit_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/archive_work_unit/invalid_work_unit_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "",
			},
		},
		{
			name:             "invalid_work_unit_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/archive_work_unit/invalid_work_unit_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27cd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/archive_work_unit/archive_work_unit.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut,
					fmt.Sprintf("/api/v1/projects/%s/work-units/%s/archive", tt.input.ProjectID, tt.input.WorkUnitID),
					nil)

				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.input.ProjectID)
				ctx.AddParam("workUnitID", tt.input.WorkUnitID)

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.ArchiveWorkUnit(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.ArchiveWorkUnit] response mismatched")
			})
		})
	}
}

func TestHandler_UnarchiveWorkUnit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

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
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unarchive_work_unit/invalid_project_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unarchive_work_unit/invalid_project_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b156",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27c",
			},
		},
		{
			name:             "empty_work_unit_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unarchive_work_unit/invalid_work_unit_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "",
			},
		},
		{
			name:             "invalid_work_unit_id_format",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unarchive_work_unit/invalid_work_unit_id.json",
			input: request.ArchiveWorkUnitInput{
				ProjectID:  "8dc3be2e-19a4-4942-8a79-56db391a0b15",
				WorkUnitID: "4797347d-21e0-4dac-a6c7-c98bf2d6b27cd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/unarchive_work_unit/unarchive_work_unit.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut,
					fmt.Sprintf("/api/v1/projects/%s/work-units/%s/unarchive", tt.input.ProjectID, tt.input.WorkUnitID),
					nil)

				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.input.ProjectID)
				ctx.AddParam("workUnitID", tt.input.WorkUnitID)

				h := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UnarchiveWorkUnit(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UnarchiveWorkUnit] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateSendingSurveyState(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		query            string
		id               string
	}{
		{
			name:             "ok_update_sending_survey",
			wantCode:         http.StatusOK,
			wantErr:          false,
			wantResponsePath: "testdata/update_sending_survey/200.json",
			query:            "allowsSendingSurvey=true",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
		{
			name:             "failed_invalid_project_id",
			wantCode:         http.StatusNotFound,
			wantErr:          true,
			wantResponsePath: "testdata/update_sending_survey/404.json",
			query:            "allowsSendingSurvey=true",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b11",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantErr:          false,
			wantResponsePath: "testdata/update_sending_survey/invalid_project_id.json",
			query:            "allowsSendingSurvey=true",
			id:               "",
		},
		{
			name:             "invalid_project_id_format",
			wantCode:         http.StatusBadRequest,
			wantErr:          false,
			wantResponsePath: "testdata/update_sending_survey/invalid_project_id.json",
			query:            "allowsSendingSurvey=true",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b156",
		},
	}

	for _, tt := range tests {
		testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
			testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_sending_survey/update_sending_survey.sql")

			t.Run(tt.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/sending-survey-state?%s", tt.id, tt.query), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateSendingSurveyState(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.Equal(t, tt.wantCode, w.Code)
				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UpdateSendingSurveyState] response mismatched")
			})
		})
	}
}

func TestHandler_UnassignMember(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()
	controllerMock := controller.New(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		id               string
		memberID         string
	}{
		{
			name:             "ok_update_project_status",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/unassign_member/200.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "2655832e-f009-4b73-a535-64c3a22e558f",
		},
		{
			name:             "failed_invalid_member_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/unassign_member/404.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
		{
			name:             "empty_project_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unassign_member/empty_project_id.json",
			id:               "",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
		{
			name:             "empty_member_id",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/unassign_member/empty_member_id.json",
			id:               "8dc3be2e-19a4-4942-8a79-56db391a0b15",
			memberID:         "",
		},
		{
			name:             "project_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/unassign_member/project_not_found.json",
			id:               "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
			memberID:         "cb889a9c-b20c-47ee-83b8-44b6d1721acb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/unassign_member/unassign_member.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.id}, gin.Param{Key: "memberID", Value: tt.memberID}}
				ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/projects/%s/members/%s", tt.id, tt.memberID), nil)
				ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
				metadataHandler := New(controllerMock, storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UnassignMember(ctx)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.UnassignMember] response mismatched")
			})
		})
	}
}
