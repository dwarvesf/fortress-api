package communitynft

import "errors"

var (
	ErrInvalidTokenID       = errors.New("invalid token id")
	ErrTokenNotFound        = errors.New("token not found")
	ErrMochiProfileNotFound = errors.New("mochi profile not found")
)
