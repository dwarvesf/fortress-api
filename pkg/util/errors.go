package util

import (
	"errors"
)

var (
	ErrInvalidToken     = errors.New("token expired, please log out and log in again")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrBadToken         = errors.New("bad token")
)
