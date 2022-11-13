package utils

import (
	"errors"
)

var (
	ErrInvalidToken                    = errors.New("token expired, please log out and log in again")
	ErrInvalidSignature                = errors.New("invalid signature")
	ErrBadToken                        = errors.New("bad token")
	ErrAuthenticationTypeHeaderInvalid = errors.New("authentication type header is invalid")
	ErrUnexpectedAuthorizationHeader   = errors.New("unexpected authorization headers")
	ErrInvalidUUID                     = errors.New("invalid UUID")
)
