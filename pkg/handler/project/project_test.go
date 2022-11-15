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
			ctx.Request = httptest.NewRequest("POST", fmt.Sprintf("%s", "/api/v1/projects/"+tt.id+"/status"), bodyReader)
			ctx.Request.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM")
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock)

			metadataHandler.UpdateProjectStatus(ctx)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.Equal(t, tt.wantCode, w.Code)
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
				Type:              string(model.ProjectTypeDwarves),
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				DeliveryManagerID: model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				CountryID:         "4ef64490-c906-4192-a7f9-d2221dadfe4c",
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create/200.json",
		},
		{
			name: "invalid_status",
			args: CreateProjectInput{
				Name:              "project1",
				Status:            "something",
				Type:              string(model.ProjectTypeDwarves),
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         "4ef64490-c906-4192-a7f9-d2221dadfe4c",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_status.json",
		},
		{
			name: "invalid_type",
			args: CreateProjectInput{
				Name:              "project1",
				Status:            string(model.ProjectStatusOnBoarding),
				Type:              "something",
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         "4ef64490-c906-4192-a7f9-d2221dadfe4c",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create/400_invalid_type.json",
		},
		{
			name: "missing_status",
			args: CreateProjectInput{
				Name:              "project1",
				Type:              string(model.ProjectTypeDwarves),
				StartDate:         "2022-11-14",
				AccountManagerID:  model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
				DeliveryManagerID: model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
				CountryID:         "4ef64490-c906-4192-a7f9-d2221dadfe4c",
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
