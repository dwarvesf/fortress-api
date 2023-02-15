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
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.Create-fm",
			},
		},
		"/api/v1/employees/search": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.List-fm",
			},
		},
		"/api/v1/employees/:id": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.Details-fm",
			},
		},
		"/api/v1/employees/:id/upload-avatar": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UploadAvatar-fm",
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
		"/api/v1/line-managers": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.GetLineManagers-fm",
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
		"/api/v1/metadata/organizations": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.Organizations-fm",
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
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.CreatePosition-fm",
			},
		},
		"/api/v1/metadata/positions/:id": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.UpdatePosition-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.DeletePosition-fm",
			},
		},
		"/api/v1/auth": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/auth.IHandler.Auth-fm",
			},
		},
		"/api/v1/auth/me": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/auth.IHandler.Me-fm",
			},
		},
		"/api/v1/auth/api-key": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/auth.IHandler.CreateAPIKey-fm",
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
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.CreateStack-fm",
			},
		},
		"/api/v1/metadata/stacks/:id": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.UpdateStack-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.DeleteStack-fm",
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
		"/api/v1/projects/:id/upload-avatar": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UploadAvatar-fm",
			},
		},
		"/api/v1/projects/:id/sending-survey-state": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UpdateSendingSurveyState-fm",
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
		"/api/v1/projects/:id/slots/:slotID": {
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.DeleteSlot-fm",
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
		"/api/v1/projects/:id/work-units/:workUnitID/archive": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.ArchiveWorkUnit-fm",
			},
		},
		"/api/v1/projects/:id/work-units/:workUnitID/unarchive": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.UnarchiveWorkUnit-fm",
			},
		},
		"/api/v1/metadata/questions": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/metadata.IHandler.GetQuestions-fm",
			},
		},
		"/api/v1/feedbacks": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/feedback.IHandler.List-fm",
			},
		},
		"/api/v1/feedbacks/:id/topics/:topicID": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/feedback.IHandler.Detail-fm",
			},
		},
		"/api/v1/feedbacks/unreads": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/feedback.IHandler.CountUnreadFeedback-fm",
			},
		},
		"/api/v1/feedbacks/:id/topics/:topicID/submit": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/feedback.IHandler.Submit-fm",
			},
		},
		"/api/v1/surveys": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.CreateSurvey-fm",
			},
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.ListSurvey-fm",
			},
		},
		"/api/v1/surveys/:id/send": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.SendSurvey-fm",
			},
		},
		"/api/v1/surveys/:id": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.GetSurveyDetail-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.DeleteSurvey-fm",
			},
		},
		"/api/v1/surveys/:id/topics/:topicID/reviews/:reviewID": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.GetSurveyReviewDetail-fm",
			},
		},
		"/api/v1/surveys/:id/topics/:topicID/employees": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.UpdateTopicReviewers-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.DeleteTopicReviewers-fm",
			},
		},
		"/api/v1/surveys/:id/done": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.MarkDone-fm",
			},
		},
		"/api/v1/surveys/:id/topics/:topicID": {
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.DeleteSurveyTopic-fm",
			},
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/survey.IHandler.GetSurveyTopicDetail-fm",
			},
		},
		"/api/v1/employees/:id/roles": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/employee.IHandler.UpdateRole-fm",
			},
		},
		"/api/v1/valuation/:year": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/valuation.IHandler.One-fm",
			},
		},
		"/api/v1/dashboards/projects/sizes": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetProjectSizes-fm",
			},
		},
		"/api/v1/dashboards/projects/work-surveys": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetWorkSurveys-fm",
			},
		},
		"/api/v1/dashboards/projects/action-items": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetActionItemReports-fm",
			},
		},
		"/api/v1/dashboards/projects/engineering-healths": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetEngineeringHealth-fm",
			},
		},
		"/api/v1/dashboards/projects/audits": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetAudits-fm",
			},
		},
		"/api/v1/earn": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/earn.IHandler.List-fm",
			},
		},
		"/api/v1/tech-radar": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/techradar.IHandler.List-fm",
			},
		},
		"/api/v1/audiences": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/audience.IHandler.List-fm",
			},
		},
		"/api/v1/events": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/event.IHandler.List-fm",
			},
		},
		"/api/v1/digests": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/digest.IHandler.List-fm",
			},
		},
		"/api/v1/updates": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/update.IHandler.List-fm",
			},
		},
		"/api/v1/memos": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/memo.IHandler.List-fm",
			},
		},
		"/api/v1/hiring-positions": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/hiring.IHandler.List-fm",
			},
		},
		"/api/v1/staffing-demands": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/staffingdemand.IHandler.List-fm",
			},
		},
		"/api/v1/projects/milestones": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/project.IHandler.ListMilestones-fm",
			},
		},
		"/api/v1/dashboards/projects/action-item-squash": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetActionItemSquashReports-fm",
			},
		},
		"/api/v1/dashboards/projects/summary": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetSummary-fm",
			},
		},
		"/api/v1/dashboards/resources/availabilities": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetResourcesAvailability-fm",
			},
		},
		"/api/v1/dashboards/engagement/info": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetEngagementInfo-fm",
			},
		},
		"/api/v1/dashboards/engagement/detail": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetEngagementInfoDetail-fm",
			},
		},
		"/api/v1/dashboards/resources/utilization": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetResourceUtilization-fm",
			},
		},
		"/api/v1/dashboards/resources/work-unit-distribution": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetWorkUnitDistribution-fm",
			},
		},
		"/api/v1/dashboards/resources/work-unit-distribution-summary": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetWorkUnitDistributionSummary-fm",
			},
		},
		"/api/v1/dashboards/resources/work-survey-summaries": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/dashboard.IHandler.GetResourceWorkSurveySummaries-fm",
			},
		},
		"/cronjobs/birthday": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/birthday.IHandler.BirthdayDailyMessage-fm",
			},
		},
		"/cronjobs/audits": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/audit.IHandler.Sync-fm",
			},
		},
		"/cronjobs/sync-discord-info": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/discord.IHandler.SyncDiscordInfo-fm",
			},
		},
		"/webhooks/n8n": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/webhook.IHandler.N8n-fm",
			},
		},
		"/api/v1/bank-accounts": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/bankaccount.IHandler.List-fm",
			},
		},
		"/api/v1/invoices": {
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/invoice.IHandler.Create-fm",
			},
		},
		"/api/v1/invoices/:id/status": {
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/invoice.IHandler.Update-fm",
			},
		},
		"/api/v1/invoices/latest": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/invoice.IHandler.GetLatestInvoice-fm",
			},
		},
		"/api/v1/clients": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/client.IHandler.List-fm",
			},
			"POST": {
				Method:  "POST",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/client.IHandler.Create-fm",
			},
		},
		"/api/v1/clients/:id": {
			"GET": {
				Method:  "GET",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/client.IHandler.Detail-fm",
			},
			"PUT": {
				Method:  "PUT",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/client.IHandler.Update-fm",
			},
			"DELETE": {
				Method:  "DELETE",
				Handler: "github.com/dwarvesf/fortress-api/pkg/handler/client.IHandler.Delete-fm",
			},
		},
	}

	l := logger.NewLogrusLogger()
	cfg := config.LoadConfig(config.DefaultConfigLoaders())

	h := handler.New(nil, nil, nil, nil, l, cfg)

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
