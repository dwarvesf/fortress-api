package errs

import "errors"

var (
	ErrBadRequest                    = errors.New("bad request")
	ErrInvalidYear                   = errors.New("invalid year, must be current year")
	ErrCannotReadProjectBonusExplain = errors.New("cannot read project bonus explain")
	ErrPayrollNotSnapshotted         = errors.New("payroll not snapshotted")
)
