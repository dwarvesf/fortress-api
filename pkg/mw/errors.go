package mw

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized                    = errors.New("unauthorized")
	ErrInvalidUserID                   = errors.New("invalid user id")
	ErrInvalidToken                    = errors.New("invalid token")
	ErrAuthenticationTypeHeaderInvalid = errors.New("authentication type header is invalid")
	ErrUnexpectedAuthorizationHeader   = errors.New("unexpected authorization headers")
	ErrInvalidAPIKey                   = errors.New("invalid API key")
)

// errUnauthorized returns unauthorized custom error
func errUnauthorized(description string) error {
	return fmt.Errorf("unauthorized permission required: %v", description)
}
