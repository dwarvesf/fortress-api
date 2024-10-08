package errs

import "errors"

var (
	ErrEmptyReportView  = errors.New("view is empty")
	ErrEmptyChannelID   = errors.New("channelID is empty")
	ErrEmptyGuildID     = errors.New("guildID is empty")
	ErrEmptyCreatorID   = errors.New("creatorID is empty")
	ErrEmptyName        = errors.New("name is empty")
	ErrEmptyDate        = errors.New("date is nil")
	ErrEmptyID          = errors.New("discord user id is nil")
	ErrEmptyTopic       = errors.New("topic is nil")
	ErrInvalidDate      = errors.New("date is invalid")
	ErrInvalidTimeRange = errors.New("start time must be before end time")
)
