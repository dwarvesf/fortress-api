package auth

import "errors"

var (
	ErrUserInactivated   = errors.New("user is inactivated")
	ErrEmptyPrimaryEmail = errors.New("empty primary email")
	ErrUserNotFound      = errors.New("user is not found")
	ErrRoleNotfound      = errors.New("role is not found")
)
