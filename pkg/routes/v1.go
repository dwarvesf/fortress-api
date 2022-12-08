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

	// user profile
	v1.GET("/profile", amw.WithAuth, h.Profile.GetProfile)
	v1.PUT("/profile", amw.WithAuth, h.Profile.UpdateInfo)
	v1.POST("/profile/upload-avatar", amw.WithAuth, h.Profile.UploadAvatar)

	// employees
	v1.GET("/employees", amw.WithAuth, pmw.WithPerm("employees.read"), h.Employee.List)
	v1.POST("/employees", amw.WithAuth, pmw.WithPerm("employees.create"), h.Employee.Create)
	v1.GET("/employees/:id", amw.WithAuth, pmw.WithPerm("employees.read"), h.Employee.One)
	v1.PUT("/employees/:id/general-info", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateGeneralInfo)
	v1.PUT("/employees/:id/personal-info", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdatePersonalInfo)
	v1.PUT("/employees/:id/skills", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateSkills)
	v1.PUT("/employees/:id/employee-status", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UpdateEmployeeStatus)
	v1.POST("/employees/:id/upload-content", amw.WithAuth, pmw.WithPerm("employees.edit"), h.Employee.UploadContent)

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

	// projects
	v1.POST("/projects", amw.WithAuth, pmw.WithPerm("projects.create"), h.Project.Create)
	v1.GET("/projects", amw.WithAuth, pmw.WithPerm("projects.read"), h.Project.List)
	v1.GET("/projects/:id", amw.WithAuth, pmw.WithPerm("projects.read"), h.Project.Details)
	v1.PUT("/projects/:id/status", amw.WithAuth, pmw.WithPerm("projects.read"), h.Project.UpdateProjectStatus)
	v1.POST("/projects/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.create"), h.Project.AssignMember)
	v1.GET("/projects/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.read"), h.Project.GetMembers)
	v1.PUT("/projects/:id/members", amw.WithAuth, pmw.WithPerm("projectMembers.edit"), h.Project.UpdateMember)
	v1.PUT("/projects/:id/members/:memberID", amw.WithAuth, pmw.WithPerm("projectMembers.edit"), h.Project.UnassignMember)
	v1.DELETE("/projects/:id/members/:memberID", amw.WithAuth, pmw.WithPerm("projectMembers.delete"), h.Project.DeleteMember)
	v1.PUT("/projects/:id/general-info", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateGeneralInfo)
	v1.PUT("/projects/:id/contact-info", amw.WithAuth, pmw.WithPerm("projects.edit"), h.Project.UpdateContactInfo)
	v1.GET("/projects/:id/work-units", amw.WithAuth, pmw.WithPerm("projectWorkUnits.read"), h.Project.GetWorkUnits)
	v1.POST("/projects/:id/work-units", amw.WithAuth, pmw.WithPerm("projectWorkUnits.create"), h.Project.CreateWorkUnit)
	v1.PUT("/projects/:id/work-units/:workUnitID", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.UpdateWorkUnit)
	v1.PUT("/projects/:id/work-units/:workUnitID/archive", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.ArchiveWorkUnit)
	v1.PUT("/projects/:id/work-units/:workUnitID/unarchive", amw.WithAuth, pmw.WithPerm("projectWorkUnits.edit"), h.Project.UnarchiveWorkUnit)

	feedbackGroup := v1.Group("/feedbacks", amw.WithAuth)
	{
		feedbackGroup.GET("", pmw.WithPerm("feedbacks.read"), h.Feedback.List)
		feedbackGroup.GET("/:id/topics/:topicID", amw.WithAuth, pmw.WithPerm("employeeEventQuestions.read"), h.Feedback.Detail)
	}
}
