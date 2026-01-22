package view

// ContractorInvoiceResponse represents the response for contractor invoice generation
type ContractorInvoiceResponse struct {
	InvoiceNumber  string                      `json:"invoiceNumber"`
	ContractorName string                      `json:"contractorName"`
	Month          string                      `json:"month"`
	BillingType    string                      `json:"billingType"`
	Currency       string                      `json:"currency"`
	Total          float64                     `json:"total"`
	PDFFileURL     string                      `json:"pdfFileUrl,omitempty"` // Google Drive URL or local file path
	GeneratedAt    string                      `json:"generatedAt"`
	LineItems      []ContractorInvoiceLineItem `json:"lineItems"`
} // @name ContractorInvoiceResponse

// ContractorInvoiceLineItem represents a line item in contractor invoice
type ContractorInvoiceLineItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description"` // Proof of Work
	Hours       float64 `json:"hours,omitempty"`
	Rate        float64 `json:"rate,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
} // @name ContractorInvoiceLineItem

// BatchInvoiceResponse represents the response for batch invoice generation (async)
type BatchInvoiceResponse struct {
	Message string `json:"message"`
	Batch   int    `json:"batch"`
	Month   string `json:"month"`
} // @name BatchInvoiceResponse

// BatchInvoiceResult represents the result of processing a single contractor in batch mode
type BatchInvoiceResult struct {
	Contractor string  `json:"contractor"`
	Success    bool    `json:"success"`
	Skipped    bool    `json:"skipped,omitempty"`    // true if contractor was skipped (e.g., no pending payouts)
	SkipReason string  `json:"skipReason,omitempty"` // reason for skipping
	Total      float64 `json:"total,omitempty"`
	Currency   string  `json:"currency,omitempty"`
	Error      string  `json:"error,omitempty"`
}
