package contractorpayables

// PreviewCommitRequest contains query parameters for preview endpoint
type PreviewCommitRequest struct {
	Month      string `form:"month"`
	Batch      int    `form:"batch"`
	Contractor string `form:"contractor"` // Optional: contractor discord username to filter
	FileName   string `form:"file_name"`
	Year       int    `form:"year"`
}

// CommitRequest contains the request body for commit endpoint
type CommitRequest struct {
	Month      string `json:"month"`
	Batch      int    `json:"batch"`
	Contractor string `json:"contractor"`
	FileName   string `json:"file_name"`
	Year       int    `json:"year"`
}
