package client

import "errors"

var (
	ErrInvalidClientID        = errors.New("invalid client id")
	ErrInvalidClientContactID = errors.New("invalid client contact id")
	ErrClientNotFound         = errors.New("client not found")
	ErrClientContactNotFound  = errors.New("client contact not found")
	ErrMismatchClientID       = errors.New("mismatch client id")
)
