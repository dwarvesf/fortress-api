package errs

import "errors"

var (
	ErrEventNotFound                    = errors.New("event not found")
	ErrInvalidEngagementDashboardFilter = errors.New("invalid engagement dashboard filter")
	ErrInvalidStartDate                 = errors.New("invalid startDate")
)
