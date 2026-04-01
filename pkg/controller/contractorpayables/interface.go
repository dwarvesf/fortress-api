package contractorpayables

import (
	"context"
)

// IController defines the interface for contractor payables controller
type IController interface {
	PreviewCommit(ctx context.Context, month string, batch int, contractor string) (*PreviewCommitResponse, error)
	CommitPayables(ctx context.Context, month string, batch int, contractor string) (*CommitResponse, error)
	CommitPayablesByFile(ctx context.Context, fileName string, year int) (*CommitResponse, error)
	// GetCachedPreview retrieves a cached preview (from PreviewCommit call)
	GetCachedPreview(month string, batch int) (*PreviewCommitResponse, bool)
}

// PreviewCommitResponse contains the preview data
type PreviewCommitResponse struct {
	Month       string              `json:"month"`
	Batch       int                 `json:"batch"`
	Count       int                 `json:"count"`
	TotalAmount float64             `json:"total_amount"`
	Contractors []ContractorPreview `json:"contractors"`
}

// ContractorPreview contains preview data for a single contractor
type ContractorPreview struct {
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	PayableID string  `json:"payable_id"`
}

// CommitResponse contains the result of commit operation
type CommitResponse struct {
	Mode     string        `json:"mode,omitempty"`
	FileName string        `json:"file_name,omitempty"`
	Year     int           `json:"year,omitempty"`
	Month    string        `json:"month"`
	Batch    int           `json:"batch"`
	Updated  int           `json:"updated"`
	Failed   int           `json:"failed"`
	Errors   []CommitError `json:"errors,omitempty"`
}

// CommitError contains error details for failed updates
type CommitError struct {
	PayableID string `json:"payable_id"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Error     string `json:"error"`
}
