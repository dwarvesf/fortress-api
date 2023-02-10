package dashboard

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_GetProjectSizes(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		query            string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "ok",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/project_sizes/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/project_sizes/project_sizes.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboards/projects/sizes", nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetProjectSizes(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetProjectSizes] response mismatched")
			})
		})
	}
}

func TestHandler_GetResourceWorkSurveySummaries(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		query            string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_resource_work_survey_summaries/200.json",
		},
		{
			name:             "query_by_keyword",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_resource_work_survey_summaries/200_with_keyword.json",
			query:            "keyword=nam",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_resource_work_survey_summaries/seed.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/resources/work-survey-summaries?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetResourceWorkSurveySummaries(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetResourceWorkSurveySummaries] response mismatched")
			})
		})
	}
}

func TestHandler_GetWorkSurveys(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/work_surveys/200.json",
		},
		{
			name:             "happy_case_with_project_id",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/work_surveys/200_with_project_id.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/work_surveys/work_surveys.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/projects/work-surveys?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetWorkSurveys(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetWorkSurveys] response mismatched")
			})
		})
	}
}

func TestHandler_GetActionItemReports(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/action_items/200.json",
		},
		{
			name:             "happy_case_with_project_id",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/action_items/200_with_project_id.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/action_items/action_items.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/projects/action-items?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetActionItemReports(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetActionItemReports] response mismatched")
			})
		})
	}
}

func TestHandler_GetEngineeringHealth(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/engineering_healths/200_without_project_id.json",
		},
		{
			name:             "happy_case_with_project_id",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/engineering_healths/200_with_project_id.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/engineering_healths/engineering_healths.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/projects/engineering-healths?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetEngineeringHealth(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetEngineeringHealth] response mismatched")
			})
		})
	}
}

func TestHandler_GetAudits(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/audits/200_without_project_id.json",
		},
		{
			name:             "happy_case_with_project_id",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/audits/200_with_project_id.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/audits/audits.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/projects/audits?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetAudits(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetAudits] response mismatched")
			})
		})
	}
}

func TestHandler_GetActionItemSquashReports(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
		query            string
	}{
		{
			name:             "worse_case_with_project_id",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/action_item_squash/project_not_found.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b13",
		},
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/action_item_squash/200.json",
		},
		{
			name:             "happy_case_with_project_id",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/action_item_squash/200_with_project_id.json",
			query:            "projectID=8dc3be2e-19a4-4942-8a79-56db391a0b15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/action_item_squash/action_items.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/projects/action-item-squash?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetActionItemSquashReports(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetActionItemReports] response mismatched")
			})
		})
	}
}

func TestHandler_GetSummary(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		query            string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "ok",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/summary/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/summary/summary.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboards/projects/summary", nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetSummary(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetSummary] response mismatched")
			})
		})
	}
}

func TestHandler_GetResourcesAvailability(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_resources_availability/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_resources_availability/seed.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboards/resources/availabilities", nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetResourcesAvailability(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetResourcesAvailability] response mismatched")
			})
		})
	}
}

func TestHandler_GetEngagementInfo(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/engagement_info/200.json",
		},
		{
			name:             "no_record",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/engagement_info/no_record.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, fmt.Sprintf("./testdata/engagement_info/%s.sql", tt.name))
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboards/enagagement/info", nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetEngagementInfo(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetEngagementInfo] response mismatched")
			})
		})
	}
}

func TestHandler_GetEngagementInfoDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

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
			wantResponsePath: "testdata/engagement_info_detail/200.json",
			query:            "filter=seniority&startDate=2022-10-01",
		},
		{
			name:             "invalid_filter",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/engagement_info_detail/invalid_filter.json",
			query:            "filter=unknown&startDate=2022-10-01",
		},
		{
			name:             "invalid_start_date",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/engagement_info_detail/invalid_start_date.json",
			query:            "filter=chapter&startDate=unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				if tt.wantCode == http.StatusOK {
					testhelper.LoadTestSQLFile(t, txRepo, fmt.Sprintf("./testdata/engagement_info/%s.sql", tt.name))
				}
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/enagagement/detail?%s", tt.query), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetEngagementInfoDetail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetEngagementInfoDetail] response mismatched")
			})
		})
	}
}

func TestHandler_GetResourceUtilization(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_resource_utilization/200_ok.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_resource_utilization/seed.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/dashboards/resources/utilization", nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetResourceUtilization(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetResourceUtilization] response mismatched")
			})
		})
	}
}

func TestHandler_GetWorkUnitDistribution(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
		query            string
	}{
		{
			name:             "happy_case_with_name",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/work_unit_distribution/200_with_name.json",
			query:            "name=th",
		},
		{
			name:             "happy_case_with_name_and_sort",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/work_unit_distribution/200_with_name_and_sort.json",
			query:            "name=th&sort=asc",
		},
		{
			name:             "happy_case_with_name_and_sort",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/work_unit_distribution/200_with_name_sort_and_type.json",
			query:            "name=th&sort=asc&type=training",
		},
		{
			name:             "invalid_sort",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/work_unit_distribution/400_invalid_sort.json",
			query:            "name=th&sort=ascd&type=training",
		},
		{
			name:             "invalid_type",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/work_unit_distribution/400_invalid_type.json",
			query:            "name=th&sort=asc&type=trainding",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/work_unit_distribution/work_unit_distribution.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/resources/work-unit-distribution?%s", tt.query), nil)
				ctx.Request.URL.RawQuery = tt.query
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetWorkUnitDistribution(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetWorkUnitDistribution] response mismatched")
			})
		})
	}
}

func TestHandler_GetWorkUnitDistributionSummary(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case_with_name",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/work_unit_distribution_summary/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/work_unit_distribution_summary/work_unit_distribution_summary.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboards/resources/work-unit-distribution-summary"), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetWorkUnitDistributionSummary(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Dashboard.GetWorkUnitDistributionSummary] response mismatched")
			})
		})
	}
}
