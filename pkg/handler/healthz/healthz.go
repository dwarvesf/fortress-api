package healthz_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type handler struct{}

func New() IHandler {
	return &handler{}
}

// Healthz handler
// Return "OK"
func (h *handler) Healthz(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, "OK")
}
