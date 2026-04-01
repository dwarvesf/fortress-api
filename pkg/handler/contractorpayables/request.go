package contractorpayables

// PreviewCommitRequest contains query parameters for preview endpoint
type PreviewCommitRequest struct {
	Month      string `form:"month" binding:"required"` // YYYY-MM format
	Batch      int    `form:"batch" binding:"required,oneof=1 15"`
	Contractor string `form:"contractor"` // Optional: contractor discord username to filter
}

// CommitRequest contains the request body for commit endpoint
type CommitRequest struct {
	Month      string `json:"month"`
	Batch      int    `json:"batch"`
	Contractor string `json:"contractor"`
	FileName   string `json:"file_name"`
	Year       int    `json:"year"`
}
