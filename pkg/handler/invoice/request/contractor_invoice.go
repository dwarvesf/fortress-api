package request

import "regexp"

// GenerateContractorInvoiceRequest represents the request to generate a contractor invoice
type GenerateContractorInvoiceRequest struct {
	Contractor        string `json:"contractor" binding:"required"`
	Month             string `json:"month"`             // YYYY-MM format (optional, if empty fetches all pending)
	GroupFeeByProject *bool  `json:"groupFeeByProject"` // Group Fee items by project (optional, default true)
	InvoiceType       string `json:"invoiceType"`       // "service_and_refund" | "extra_payment" | "" (full invoice)
	SkipUpload        bool   `json:"skipUpload"`        // Skip Google Drive upload, save PDF locally
	DryRun            bool   `json:"dryRun"`            // Skip Notion payable record creation (implies skipUpload)
} // @name GenerateContractorInvoiceRequest

// Validate validates the request
func (r *GenerateContractorInvoiceRequest) Validate() error {
	// Month is optional - if provided, validate format
	if r.Month != "" && !isValidMonthFormat(r.Month) {
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
