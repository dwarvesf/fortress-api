package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller/auth"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func ConvertControllerErr(c *gin.Context, err error) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError

	switch err {
	case auth.ErrRoleNotfound:
		status = http.StatusNotFound
	case auth.ErrUserNotFound:
		status = http.StatusNotFound
	case auth.ErrUserInactivated:
		status = http.StatusBadRequest
	default:
		status = http.StatusInternalServerError
	}
	c.JSON(status, view.CreateResponse[any](nil, nil, err, nil, ""))
}
