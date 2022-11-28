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
	expectedRoutes := map[string]map[string]gin.RouteInfo{
		"/api/v1/employees": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.List-fm",
			},
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.Create-fm",
			},
		},
		"/api/v1/employees/:id": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.One-fm",
			},
		},
		"/api/v1/employees/:id/general-info": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdateGeneralInfo-fm",
			},
		},
		"/api/v1/employees/:id/personal-info": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdatePersonalInfo-fm",
			},
		},
		"/api/v1/employees/:id/skills": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdateSkills-fm",
			},
		},
		"/api/v1/employees/:id/upload-content": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UploadContent-fm",
			},
		},
		"/api/v1/metadata/working-status": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.WorkingStatuses-fm",
			},
		},
		"/api/v1/metadata/seniorities": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Seniorities-fm",
			},
		},
		"/api/v1/metadata/chapters": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Chapters-fm",
			},
		},
		"/api/v1/metadata/account-roles": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.AccountRoles-fm",
			},
		},
		"/api/v1/metadata/positions": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Positions-fm",
			},
		},
		"/api/v1/auth": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/auth.IHandler.Auth-fm",
			},
		},
		"/api/v1/metadata/countries": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.GetCountries-fm",
			},
		},
		"/api/v1/metadata/project-statuses": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.ProjectStatuses-fm",
			},
		},
		"/api/v1/metadata/stacks": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Stacks-fm",
			},
		},
		"/api/v1/metadata/countries/:country_id/cities": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.GetCities-fm",
			},
		},
		"/api/v1/employees/:id/employee-status": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdateEmployeeStatus-fm",
			},
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.Create-fm",
			},
		},
		"/api/v1/profile": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/profile.IHandler.GetProfile-fm",
			},
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/profile.IHandler.UpdateInfo-fm",
			},
		},
		"/api/v1/projects": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.List-fm",
			},
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.Create-fm",
			},
		},
		"/api/v1/projects/:id": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.Details-fm",
			},
		},
		"/api/v1/projects/:id/status": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateProjectStatus-fm",
			},
		},
		"/api/v1/projects/:id/members": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.AssignMember-fm",
			},
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.GetMembers-fm",
			},
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateMember-fm",
			},
		},
		"/api/v1/projects/:id/members/:memberID": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UnassignMember-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.DeleteMember-fm",
			},
		},
		"/api/v1/projects/:id/general-info": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateGeneralInfo-fm",
			},
		},
		"/api/v1/projects/:id/contact-info": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateContactInfo-fm",
			},
		},
		"/api/v1/profile/upload-avatar": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/profile.IHandler.UploadAvatar-fm",
			},
		},
		"/api/v1/projects/:id/work-units": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.GetWorkUnits-fm",
			},
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.CreateWorkUnit-fm",
			},
		},
		"/api/v1/projects/:id/work-units/:workUnitID": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateWorkUnit-fm",
			},
		},
	}

	l := logger.NewLogrusLogger()
	cfg := config.LoadConfig(config.DefaultConfigLoaders())

	h := handler.New(nil, nil, nil, l, cfg)

	router := gin.New()
	loadV1Routes(router, h, nil, nil, nil)

	routeInfo := router.Routes()

	for _, r := range routeInfo {
		require.NotNil(t, r.HandlerFunc)
		expected, ok := expectedRoutes[r.Path][r.Method]
		require.True(t, ok, fmt.Sprintf("unexpected path: %s", r.Path))
		ignoreFields := cmpopts.IgnoreFields(gin.RouteInfo{}, "HandlerFunc", "Path")
		if !cmp.Equal(expected, r, ignoreFields) {
			t.Errorf("route mismatched. \n route path: %v \n diff: %+v", r.Path,
				cmp.Diff(expected, r, ignoreFields))
			t.FailNow()
		}
	}
}
