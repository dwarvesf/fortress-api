package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/mw"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	// auth
	v1.POST("/auth", h.Auth.Auth)

	// user profile
	v1.GET("/profile", mw.WithAuth, h.Employee.GetProfile)

	// employees
	v1.GET("/employees", h.Employee.List)
	v1.POST("/employees", h.Employee.Create)
	v1.GET("/employees/:id", mw.WithAuth, h.Employee.One)
	v1.PUT("/employees/:id/general-info", mw.WithAuth, h.Employee.UpdateGeneralInfo)
	v1.PUT("/employees/:id/skills", mw.WithAuth, h.Employee.UpdateSkills)
	v1.PUT("/employees/:id/employee-status", mw.WithAuth, h.Employee.UpdateEmployeeStatus)

	// auth

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatus)
	v1.GET("/metadata/stacks", h.Metadata.Stacks)
	v1.GET("/metadata/seniorities", h.Metadata.Seniorities)
	v1.GET("/metadata/chapters", h.Metadata.Chapters)
	v1.GET("/metadata/account-roles", h.Metadata.AccountRoles)
	v1.GET("/metadata/account-statuses", h.Metadata.AccountStatuses)
	v1.GET("/metadata/positions", h.Metadata.Positions)
	v1.GET("/metadata/countries", h.Metadata.GetCountries)
	v1.GET("/metadata/countries/:country_id/cities", h.Metadata.GetCities)
	v1.GET("/metadata/project-statuses", h.Metadata.ProjectStatuses)

	// projects
	v1.POST("/projects", h.Project.Create)
	v1.GET("/projects", h.Project.List)
	v1.PUT("/projects/:id/status", h.Project.UpdateProjectStatus)
}
