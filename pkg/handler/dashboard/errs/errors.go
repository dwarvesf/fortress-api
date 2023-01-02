package errs

import "errors"

var (
	ErrInvalidProjectID                 = errors.New("invalid project ID")
	ErrProjectNotFound                  = errors.New("project not found")
	ErrEventNotFound                    = errors.New("event not found")
	ErrInvalidEngagementDashboardFilter = errors.New("invalid engagement dashboard filter")
	ErrInvalidStartDate                 = errors.New("invalid startDate")
)
