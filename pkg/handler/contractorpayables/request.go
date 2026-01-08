package contractorpayables

// PreviewCommitRequest contains query parameters for preview endpoint
type PreviewCommitRequest struct {
	Month string `form:"month" binding:"required"` // YYYY-MM format
	Batch int    `form:"batch" binding:"required,oneof=1 15"`
}

// CommitRequest contains the request body for commit endpoint
type CommitRequest struct {
	Month string `json:"month" binding:"required"` // YYYY-MM format
	Batch int    `json:"batch" binding:"required,oneof=1 15"`
}
