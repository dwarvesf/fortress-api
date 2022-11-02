package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler"
)

func loadV1Routes(r *gin.Engine, h *handler.Handler, cfg *config.Config) {
	r.Group("/api/v1")
}
