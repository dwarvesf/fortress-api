package errs

import "errors"

var (
	ErrEmptyReportView = errors.New("view is empty")
	ErrEmptyChannelID  = errors.New("channelID is empty")
)
