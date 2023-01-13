package errs

import "errors"

var (
	ErrFailedToGetFlag       = errors.New("failed to get flag")
	ErrMissingAuditorInAudit = errors.New("missing auditor in audit")
	ErrMissingProjectInAudit = errors.New("missing project in audit")
)
