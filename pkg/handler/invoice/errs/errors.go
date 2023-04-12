package errs

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

var (
	ErrInvalidDueAt          = errors.New("invalid due at")
	ErrInvalidPaidAt         = errors.New("invalid paid at")
	ErrInvalidInvoiceStatus  = errors.New("invalid invoice status")
	ErrInvalidInvoiceID      = errors.New("invalid invoice id")
	ErrInvalidProjectID      = errors.New("invalid project id")
	ErrInvalidDeveloperEmail = errors.New("invalid developer email in dev mode")
	ErrSenderNotFound        = errors.New("sender not found")
	ErrBankAccountNotFound   = errors.New("bank account not found")
	ErrProjectNotFound       = errors.New("project not found")
)

func ConvertControllerErr(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var status int

	switch err {
	case
		invoice.ErrInvoiceNotFound:
		status = http.StatusNotFound
	case invoice.ErrInvoiceStatusAlready:
		status = http.StatusInternalServerError

	default:
		status = http.StatusInternalServerError
	}

	c.JSON(status, view.CreateResponse[any](nil, nil, err, nil, ""))
}
