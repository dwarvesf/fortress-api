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

	// Contractor invoice errors
	ErrContractorRatesNotFound = errors.New("contractor rates not found for the specified month")
	ErrTaskOrderLogNotFound    = errors.New("task order log not found for the specified month")
	ErrInvalidMonthFormat      = errors.New("invalid month format, expected YYYY-MM")
	ErrUnsupportedBillingType  = errors.New("unsupported billing type")
	ErrNoSubitemsFound         = errors.New("no line items found in task order log")
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
	case invoice.ErrProjectNotFound:
		status = http.StatusNotFound
	case invoice.ErrSenderNotFound:
		status = http.StatusNotFound
	case invoice.ErrBankAccountNotFound:
		status = http.StatusNotFound
	case invoice.ErrInvoiceStatusAlready:
		status = http.StatusInternalServerError

	default:
		status = http.StatusInternalServerError
	}

	c.JSON(status, view.CreateResponse[any](nil, nil, err, nil, ""))
}
