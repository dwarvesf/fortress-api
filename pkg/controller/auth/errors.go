package auth

import "errors"

var (
	ErrUserInactivated = errors.New("user is inactivated")
	ErrUserNotFound    = errors.New("user is not found")
)
