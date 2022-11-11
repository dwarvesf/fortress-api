package routes

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// Test_loadV1Routes simply test we load route and handler correctly
func Test_loadV1Routes(t *testing.T) {
	expectedRoutes := map[string]gin.RouteInfo{
		"/api/v1/employees": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.List-fm",
		},
		"/api/v1/employees/:id": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.One-fm",
		},
		"/api/v1/metadata/working-status": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.WorkingStatus-fm",
		},
		"/api/v1/metadata/seniorities": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Seniorities-fm",
		},
		"/api/v1/metadata/chapters": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Chapters-fm",
		},
		"/api/v1/metadata/account-roles": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.AccountRoles-fm",
		},
		"/api/v1/metadata/account-statuses": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.AccountStatuses-fm",
		},
		"/api/v1/metadata/positions": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Positions-fm",
		},
		"/api/v1/auth": {
			Method:  "POST",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/auth.IHandler.Auth-fm",
		},
		"/api/v1/metadata/countries": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.GetCountries-fm",
		},
		"/api/v1/metadata/project-statuses": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.ProjectStatuses-fm",
		},
		"/api/v1/metadata/countries/:country_id/cities": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.GetCities-fm",
		},
		"/api/v1/employees/:id/employee-status": {
			Method:  "POST",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdateEmployeeStatus-fm",
		},
		"/api/v1/profile": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.GetProfile-fm",
		},
		"/api/v1/projects": {
			Method:  "GET",
			Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.List-fm",
		},
	}

	l := logger.NewLogrusLogger()
	h := handler.New(nil, nil, l)

	cfg := config.LoadConfig(config.DefaultConfigLoaders())

	router := gin.New()
	loadV1Routes(router, h, cfg)

	routeInfo := router.Routes()

	for _, r := range routeInfo {
		require.NotNil(t, r.HandlerFunc)
		expected, ok := expectedRoutes[r.Path]
		require.True(t, ok, fmt.Sprintf("unexpected path: %s", r.Path))
		ignoreFields := cmpopts.IgnoreFields(gin.RouteInfo{}, "HandlerFunc", "Path")
		if !cmp.Equal(expected, r, ignoreFields) {
			t.Errorf("route mismatched. \n route path: %v \n diff: %+v", r.Path,
				cmp.Diff(expected, r, ignoreFields))
			t.FailNow()
		}
	}
}
