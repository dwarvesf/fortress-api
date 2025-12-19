package errs

import "errors"

var (
	ErrEmptyDiscordID      = errors.New("discord_id is required")
	ErrEmptyProjectID      = errors.New("project_id is required")
	ErrEmptyDate           = errors.New("date is required")
	ErrEmptyTaskType       = errors.New("task_type is required")
	ErrInvalidHours        = errors.New("hours must be between 0 and 24")
	ErrEmptyProofOfWorks   = errors.New("proof_of_works is required")
	ErrInvalidTaskType     = errors.New("task_type must be Development, Design, or Meeting")
	ErrFutureDate          = errors.New("date cannot be in the future")
	ErrContractorNotFound  = errors.New("contractor not found for discord ID")
	ErrTimesheetDBNotFound = errors.New("timesheet database not configured")
)
