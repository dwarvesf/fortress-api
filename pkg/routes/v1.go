package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/mw"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, repo store.DBRepo, s *store.Store, cfg *config.Config) {
	v1 := r.Group("/api/v1")
	pmw := mw.NewPermissionMiddleware(s, repo)
	amw := mw.NewAuthMiddleware(cfg)

	// auth
	v1.POST("/auth", h.Auth.Auth)
	v1.GET("/auth/me", amw.WithAuth, pmw.WithPerm(model.PermissionAuthRead), h.Auth.Me)

	// user profile
	v1.GET("/profile", amw.WithAuth, h.Profile.GetProfile)
	v1.PUT("/profile", amw.WithAuth, h.Profile.UpdateInfo)
	v1.POST("/profile/upload-avatar", amw.WithAuth, h.Profile.UploadAvatar)

	// employees
	v1.POST("/employees", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesCreate), h.Employee.Create)
	v1.POST("/employees/search", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.List)
	v1.GET("/employees/:id", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesRead), h.Employee.One)
	v1.PUT("/employees/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateGeneralInfo)
	v1.PUT("/employees/:id/personal-info", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdatePersonalInfo)
	v1.PUT("/employees/:id/skills", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateSkills)
	v1.PUT("/employees/:id/employee-status", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UpdateEmployeeStatus)
	v1.POST("/employees/:id/upload-content", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UploadContent)
	v1.POST("/employees/:id/upload-avatar", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeesEdit), h.Employee.UploadAvatar)
	v1.POST("/employees/:id/mentees", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeMenteesCreate), h.Employee.AddMentee)
	v1.DELETE("/employees/:id/mentees/:menteeID", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeMenteesDelete), h.Employee.DeleteMentee)
	v1.PUT("/employees/:id/roles", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeRolesEdit), h.Employee.UpdateRole)

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatuses)
	v1.GET("/metadata/stacks", h.Metadata.Stacks)
	v1.GET("/metadata/seniorities", h.Metadata.Seniorities)
	v1.GET("/metadata/chapters", h.Metadata.Chapters)
	v1.GET("/metadata/account-roles", amw.WithAuth, h.Metadata.AccountRoles)
	v1.GET("/metadata/positions", h.Metadata.Positions)
	v1.GET("/metadata/countries", h.Metadata.GetCountries)
	v1.GET("/metadata/countries/:country_id/cities", h.Metadata.GetCities)
	v1.GET("/metadata/project-statuses", h.Metadata.ProjectStatuses)
	v1.GET("/metadata/questions", h.Metadata.GetQuestions)
	v1.PUT("/metadata/stacks/:id", pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdateStack)
	v1.POST("/metadata/stacks", pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreateStack)
	v1.DELETE("/metadata/stacks/:id", pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeleteStack)
	v1.PUT("/metadata/positions/:id", pmw.WithPerm(model.PermissionMetadataEdit), h.Metadata.UpdatePosition)
	v1.POST("/metadata/positions", pmw.WithPerm(model.PermissionMetadataCreate), h.Metadata.CreatePosition)
	v1.DELETE("/metadata/positions/:id", pmw.WithPerm(model.PermissionMetadataDelete), h.Metadata.DeletePosition)

	// projects
	projectGroup := v1.Group("projects")
	{
		projectGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsCreate), h.Project.Create)
		projectGroup.GET("", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsRead), h.Project.List)
		projectGroup.GET("/:id", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsRead), h.Project.Details)
		projectGroup.PUT("/:id/sending-survey-state", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateSendingSurveyState)
		projectGroup.POST("/:id/upload-avatar", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UploadAvatar)
		projectGroup.PUT("/:id/status", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateProjectStatus)
		projectGroup.POST("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersCreate), h.Project.AssignMember)
		projectGroup.GET("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersRead), h.Project.GetMembers)
		projectGroup.PUT("/:id/members", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersEdit), h.Project.UpdateMember)
		projectGroup.PUT("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersEdit), h.Project.UnassignMember)
		projectGroup.DELETE("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectMembersDelete), h.Project.DeleteMember)
		projectGroup.PUT("/:id/general-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateGeneralInfo)
		projectGroup.PUT("/:id/contact-info", amw.WithAuth, pmw.WithPerm(model.PermissionProjectsEdit), h.Project.UpdateContactInfo)
		projectGroup.GET("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsRead), h.Project.GetWorkUnits)
		projectGroup.POST("/:id/work-units", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsCreate), h.Project.CreateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UpdateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/archive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.ArchiveWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/unarchive", amw.WithAuth, pmw.WithPerm(model.PermissionProjectWorkUnitsEdit), h.Project.UnarchiveWorkUnit)
	}

	feedbackGroup := v1.Group("/feedbacks")
	{
		feedbackGroup.GET("", pmw.WithPerm(model.PermissionFeedbacksRead), h.Feedback.List)
		feedbackGroup.GET("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksRead), h.Feedback.Detail)
		feedbackGroup.POST("/:id/topics/:topicID/submit", amw.WithAuth, pmw.WithPerm(model.PermissionFeedbacksCreate), h.Feedback.Submit)
	}

	surveyGroup := v1.Group("/surveys")
	{
		surveyGroup.POST("", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysCreate), h.Survey.CreateSurvey)
		surveyGroup.GET("", pmw.WithPerm(model.PermissionSurveysRead), h.Survey.ListSurvey)
		surveyGroup.GET("/:id", pmw.WithPerm(model.PermissionSurveysRead), h.Survey.GetSurveyDetail)
		surveyGroup.DELETE("/:id", pmw.WithPerm(model.PermissionSurveysDelete), h.Survey.DeleteSurvey)
		surveyGroup.POST("/:id/send", pmw.WithPerm(model.PermissionSurveysCreate), h.Survey.SendSurvey)
		surveyGroup.GET("/:id/topics/:topicID/reviews/:reviewID", amw.WithAuth, pmw.WithPerm(model.PermissionEmployeeEventQuestionsRead), h.Survey.GetSurveyReviewDetail)
		surveyGroup.DELETE("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm(model.PermissionSurveysDelete), h.Survey.DeleteSurveyTopic)
		surveyGroup.GET("/:id/topics/:topicID", pmw.WithPerm(model.PermissionSurveysRead), h.Survey.GetSurveyTopicDetail)
		surveyGroup.PUT("/:id/topics/:topicID/employees", pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.UpdateTopicReviewers)
		surveyGroup.PUT("/:id/done", pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.MarkDone)
		surveyGroup.DELETE("/:id/topics/:topicID/employees", pmw.WithPerm(model.PermissionSurveysEdit), h.Survey.DeleteTopicReviewers)
	}

	valuation := v1.Group("/valuation")
	{
		valuation.GET("/:year", pmw.WithPerm(model.PermissionValuationRead), h.Valuation.One)
	}

	dashboard := v1.Group("/dashboards")
	{
		dashboard.GET("/projects/sizes", pmw.WithPerm("dashboards.read"), h.Dashboard.ProjectSizes)
		dashboard.GET("/work-surveys", pmw.WithPerm("dashboards.read"), h.Dashboard.WorkSurveys)
	}

}
