package errs

import "errors"

var (
	ErrInvalidPublishedAt = errors.New("cannot parse publishedAt")
	ErrInvalidDateFormat  = errors.New("invalid date format")
)
