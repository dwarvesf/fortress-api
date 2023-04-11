package model

type ProjectInfo struct {
	BaseModel

	ProjectID              *UUID `json:"project_id"`
	BasecampBucketID       int64 `json:"basecamp_bucket_id"`
	BasecampScheduleID     int64 `json:"basecamp_schedule_id"`
	BasecampCampfireID     int64 `json:"basecamp_campfire_id"`
	BasecampTodolistID     int64 `json:"basecamp_todolist_id"`
	BasecampMessageBoardID int64 `json:"basecamp_message_board_id"`
	BasecampSentryID       int64 `json:"basecamp_sentry_id"`
	GitlabID               int64 `json:"gitlab_id"`
	Repositories           JSON  `json:"repositories"`

	Project *Project `json:"project"`
}
