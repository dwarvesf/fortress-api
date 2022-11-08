package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
	"github.com/dwarvesf/fortress-api/pkg/mw"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	// employees
	v1.GET("/employees", mw.WithAuth, mw.WithPerm(cfg, "employees.read"), h.Employee.List)
	v1.GET("/employees/:id", h.Employee.One)

	// auth
	v1.POST("/auth", h.Auth.Auth)

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatus)
	v1.GET("/metadata/positions", h.Metadata.Positions)
	v1.GET("/metadata/countries", h.Metadata.GetCountries)
	v1.GET("/metadata/countries/:country_id/cities", h.Metadata.GetCities)
}
