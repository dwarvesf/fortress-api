package errs

import "errors"

var (
	ErrInvalidTokenID        = errors.New("invalid token id")
	ErrTokenNotFound         = errors.New("token not found")
)
