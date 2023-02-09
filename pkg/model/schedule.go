package model

import "time"

type Schedule struct {
	BaseModel

	Name         string
	Description  string
	ScheduleType string
	SyncedAt     *time.Time

	StartTime *time.Time
	EndTime   *time.Time

	GoogleCalendar *ScheduleGoogleCalendar
	DiscordEvent   *ScheduleDiscordEvent
	NotionPage     *ScheduleNotionPage
}

type ScheduleGoogleCalendar struct {
	BaseModel

	ScheduleID       UUID
	GoogleCalendarID string
	Description      string
	HangoutLink      string
}

type ScheduleDiscordEvent struct {
	BaseModel
	ScheduleID     UUID
	DiscordEventID string
	Description    string
	VoiceChannelID string
}

type ScheduleNotionPage struct {
	BaseModel
	ScheduleID   UUID
	NotionPageID string
	Description  string
}
