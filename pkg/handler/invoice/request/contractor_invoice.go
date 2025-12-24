package request

import "regexp"

// GenerateContractorInvoiceRequest represents the request to generate a contractor invoice
type GenerateContractorInvoiceRequest struct {
	ContractorDiscord string `json:"contractorDiscord" binding:"required"`
	Month             string `json:"month" binding:"required"` // YYYY-MM format
} // @name GenerateContractorInvoiceRequest

// Validate validates the request
func (r *GenerateContractorInvoiceRequest) Validate() error {
	if !isValidMonthFormat(r.Month) {
		return ErrInvalidMonthFormat
	}
	return nil
}

// isValidMonthFormat validates month format is YYYY-MM
func isValidMonthFormat(month string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, month)
	return matched
}

// ErrInvalidMonthFormat is returned when month format is invalid
var ErrInvalidMonthFormat = newContractorInvoiceError("invalid month format, expected YYYY-MM")

// newContractorInvoiceError creates a new contractor invoice error
func newContractorInvoiceError(msg string) error {
	return &contractorInvoiceError{msg: msg}
}

type contractorInvoiceError struct {
	msg string
}

func (e *contractorInvoiceError) Error() string {
	return e.msg
}
