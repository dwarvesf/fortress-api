package request

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/handler/timesheet/errs"
)

// LogHoursRequest represents the request body for logging hours
type LogHoursRequest struct {
	DiscordID    string    `json:"discord_id" binding:"required"`
	ProjectID    string    `json:"project_id" binding:"required"`
	Date         time.Time `json:"date" binding:"required"`
	TaskType     string    `json:"task_type" binding:"required"` // Development, Design, Meeting
	Hours        float64   `json:"hours" binding:"required"`
	ProofOfWorks string    `json:"proof_of_works" binding:"required"`
	TaskOrderID  string    `json:"task_order_id"` // Optional
}

// Validate validates the request
func (r LogHoursRequest) Validate() error {
	if r.DiscordID == "" {
		return errs.ErrEmptyDiscordID
	}
	if r.ProjectID == "" {
		return errs.ErrEmptyProjectID
	}
	if r.Date.IsZero() {
		return errs.ErrEmptyDate
	}
	if r.TaskType == "" {
		return errs.ErrEmptyTaskType
	}
	if r.Hours <= 0 || r.Hours > 24 {
		return errs.ErrInvalidHours
	}
	if r.ProofOfWorks == "" {
		return errs.ErrEmptyProofOfWorks
	}

	// Validate TaskType is one of allowed values
	validTaskTypes := map[string]bool{
		"Development": true,
		"Design":      true,
		"Meeting":     true,
	}
	if !validTaskTypes[r.TaskType] {
		return errs.ErrInvalidTaskType
	}

	// Don't allow future dates
	if r.Date.After(time.Now().AddDate(0, 0, 1)) {
		return errs.ErrFutureDate
	}

	return nil
}

// GetEntriesRequest represents query parameters for getting entries
type GetEntriesRequest struct {
	DiscordID string `form:"discord_id" binding:"required"`
	StartDate string `form:"start_date"` // YYYY-MM-DD format
	EndDate   string `form:"end_date"`   // YYYY-MM-DD format
}

// GetWeeklySummaryRequest represents query parameters for weekly summary
type GetWeeklySummaryRequest struct {
	DiscordID  string `form:"discord_id" binding:"required"`
	WeekOffset int    `form:"week_offset"` // 0=current, -1=last week, etc.
}
