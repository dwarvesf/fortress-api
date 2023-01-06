package routes

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/mw"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, repo store.DBRepo, s *store.Store, cfg *config.Config) {
	v1 := r.Group("/api/v1")
	pmw := mw.NewPermissionMiddleware(s, repo)
	amw := mw.NewAuthMiddleware(cfg)

	// auth
	v1.POST("/auth", h.Auth.Auth)
	v1.GET("/auth/me", amw.WithAuth, pmw.WithPerm("auth.read"), h.Auth.Me)

	// user profile
	v1.GET("/profile", amw.WithAuth, h.Profile.GetProfile)
	v1.PUT("/profile", amw.WithAuth, h.Profile.UpdateInfo)
	v1.POST("/profile/upload-avatar", amw.WithAuth, h.Profile.UploadAvatar)

	// employees
	v1.POST("/employees", amw.WithAuth, pmw.WithPerm("employees.create"), h.Employee.Create)
	v1.POST("/employees/search", amw.WithAuth, pmw.WithPerm("employees.read"), h.Employee.List)
	v1.GET("/employees/:id", amw.WithAuth, pmw.WithPerm("employees.read"), h.Employee.One)
	v1.PUT("/employees/:id/general-info", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateGeneralInfo)
	v1.PUT("/employees/:id/personal-info", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdatePersonalInfo)
	v1.PUT("/employees/:id/skills", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateSkills)
	v1.PUT("/employees/:id/employee-status", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateEmployeeStatus)
	v1.POST("/employees/:id/upload-content", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UploadContent)
	v1.POST("/employees/:id/upload-avatar", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UploadAvatar)
	v1.POST("/employees/:id/mentees", amw.WithAuth, pmw.WithPerm("employeeMentees.create"), h.Employee.AddMentee)
	v1.DELETE("/employees/:id/mentees/:menteeID", amw.WithAuth, pmw.WithPerm("employeeMentees.delete"), h.Employee.DeleteMentee)
	v1.PUT("/employees/:id/roles", amw.WithAuth, pmw.WithPerm("employeeRoles.edit"), h.Employee.UpdateRole)

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatuses)
	v1.GET("/metadata/stacks", h.Metadata.Stacks)
	v1.GET("/metadata/seniorities", h.Metadata.Seniorities)
	v1.GET("/metadata/chapters", h.Metadata.Chapters)
	v1.GET("/metadata/account-roles", h.Metadata.AccountRoles)
	v1.GET("/metadata/positions", h.Metadata.Positions)
	v1.GET("/metadata/countries", h.Metadata.GetCountries)
	v1.GET("/metadata/countries/:country_id/cities", h.Metadata.GetCities)
	v1.GET("/metadata/project-statuses", h.Metadata.ProjectStatuses)
	v1.GET("/metadata/questions", h.Metadata.GetQuestions)
	v1.PUT("/metadata/stacks/:id", pmw.WithPerm("metadata.edit"), h.Metadata.UpdateStack)
	v1.POST("/metadata/stacks", pmw.WithPerm("metadata.create"), h.Metadata.CreateStack)
	v1.DELETE("/metadata/stacks/:id", pmw.WithPerm("metadata.delete"), h.Metadata.DeleteStack)
	v1.PUT("/metadata/positions/:id", pmw.WithPerm("metadata.edit"), h.Metadata.UpdatePosition)
	v1.POST("/metadata/positions", pmw.WithPerm("metadata.create"), h.Metadata.CreatePosition)
	v1.DELETE("/metadata/positions/:id", pmw.WithPerm("metadata.delete"), h.Metadata.DeletePosition)

	// projects
	projectGroup := v1.Group("projects")
	{
		projectGroup.POST("", amw.WithAuth, pmw.WithPerm("projects.create"), h.Project.Create)
		projectGroup.GET("", amw.WithAuth, pmw.WithPerm("projects.read"), h.Project.List)
		projectGroup.GET("/:id", amw.WithAuth, pmw.WithPerm("projects.read"), h.Project.Details)
		projectGroup.PUT("/:id/sending-survey-state", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateSendingSurveyState)
		projectGroup.POST("/:id/upload-avatar", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UploadAvatar)
		projectGroup.PUT("/:id/status", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateProjectStatus)
		projectGroup.POST("/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.create"), h.Project.AssignMember)
		projectGroup.GET("/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.read"), h.Project.GetMembers)
		projectGroup.PUT("/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.edit"), h.Project.UpdateMember)
		projectGroup.PUT("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm("projectMembers.edit"), h.Project.UnassignMember)
		projectGroup.DELETE("/:id/members/:memberID", amw.WithAuth, pmw.WithPerm("projectMembers.delete"), h.Project.DeleteMember)
		projectGroup.PUT("/:id/general-info", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateGeneralInfo)
		projectGroup.PUT("/:id/contact-info", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateContactInfo)
		projectGroup.GET("/:id/work-units", amw.WithAuth, pmw.WithPerm("projectWorkUnits.read"), h.Project.GetWorkUnits)
		projectGroup.POST("/:id/work-units", amw.WithAuth, pmw.WithPerm("projectWorkUnits.create"), h.Project.CreateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.UpdateWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/archive", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.ArchiveWorkUnit)
		projectGroup.PUT("/:id/work-units/:workUnitID/unarchive", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.UnarchiveWorkUnit)
	}

	feedbackGroup := v1.Group("/feedbacks")
	{
		feedbackGroup.GET("", pmw.WithPerm("feedbacks.read"), h.Feedback.List)
		feedbackGroup.GET("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm("feedbacks.read"), h.Feedback.Detail)
		feedbackGroup.POST("/:id/topics/:topicID/submit", amw.WithAuth, pmw.WithPerm("feedbacks.create"), h.Feedback.Submit)
	}

	surveyGroup := v1.Group("/surveys")
	{
		surveyGroup.POST("", amw.WithAuth, pmw.WithPerm("surveys.create"), h.Survey.CreateSurvey)
		surveyGroup.GET("", pmw.WithPerm("surveys.read"), h.Survey.ListSurvey)
		surveyGroup.GET("/:id", pmw.WithPerm("surveys.read"), h.Survey.GetSurveyDetail)
		surveyGroup.DELETE("/:id", pmw.WithPerm("surveys.delete"), h.Survey.DeleteSurvey)
		surveyGroup.POST("/:id/send", pmw.WithPerm("surveys.create"), h.Survey.SendSurvey)
		surveyGroup.GET("/:id/topics/:topicID/reviews/:reviewID", amw.WithAuth, pmw.WithPerm("employeeEventQuestions.read"), h.Survey.GetSurveyReviewDetail)
		surveyGroup.DELETE("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm("surveys.delete"), h.Survey.DeleteSurveyTopic)
		surveyGroup.GET("/:id/topics/:topicID", pmw.WithPerm("surveys.read"), h.Survey.GetSurveyTopicDetail)
		surveyGroup.PUT("/:id/topics/:topicID/employees", pmw.WithPerm("surveys.edit"), h.Survey.UpdateTopicReviewers)
		surveyGroup.PUT("/:id/done", pmw.WithPerm("surveys.edit"), h.Survey.MarkDone)
		surveyGroup.DELETE("/:id/topics/:topicID/employees", pmw.WithPerm("surveys.edit"), h.Survey.DeleteTopicReviewers)
	}

	valuation := v1.Group("/valuation")
	{
		valuation.GET("/:year", pmw.WithPerm("valuations.read"), h.Valuation.One)
	}

	dashboard := v1.Group("/dashboards")
	{
		dashboard.GET("/projects/sizes", pmw.WithPerm("dashboards.read"), h.Dashboard.ProjectSizes)
	}

}
