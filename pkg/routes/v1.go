package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, cfg *config.Config) {
	v1 := r.Group("/api/v1")

	// employees
	v1.GET("/employees", h.Employee.List)
	v1.GET("/employees/:id", h.Employee.One)

	// metadata
	v1.GET("/metadata/working-status", h.Metadata.WorkingStatus)
}
