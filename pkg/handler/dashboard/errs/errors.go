package errs

import "errors"

var (
	ErrInvalidProjectID                 = errors.New("invalid project ID")
	ErrProjectNotFound                  = errors.New("project not found")
	ErrEventNotFound                    = errors.New("event not found")
	ErrProjectNotionNotFound            = errors.New("project notion not found")
	ErrInvalidEngagementDashboardFilter = errors.New("invalid engagement dashboard filter")
	ErrInvalidStartDate                 = errors.New("invalid startDate")
	ErrInvalidWorkUnitDistributionType  = errors.New("invalid work unit distribution type")
	ErrInvalidWorkUnitDistributionSort  = errors.New("invalid sort value")
)
