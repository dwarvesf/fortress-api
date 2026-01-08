package contractorpayables

import (
	"context"
)

// IController defines the interface for contractor payables controller
type IController interface {
	PreviewCommit(ctx context.Context, month string, batch int) (*PreviewCommitResponse, error)
	CommitPayables(ctx context.Context, month string, batch int) (*CommitResponse, error)
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
	Month   string        `json:"month"`
	Batch   int           `json:"batch"`
	Updated int           `json:"updated"`
	Failed  int           `json:"failed"`
	Errors  []CommitError `json:"errors,omitempty"`
}

// CommitError contains error details for failed updates
type CommitError struct {
	PayableID string `json:"payable_id"`
	Error     string `json:"error"`
}
