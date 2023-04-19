package model

import bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Action uint8

const (
	BasecampCommentMsg    string = "basecamp_comment"
	BasecampTodoMsg       string = "basecamp_todo"
	BasecampHiringTodoMsg string = "basecamp_todo_hiring"
)

type WorkerMessage struct {
	Type    string
	Payload interface{}
}

// BasecampCommentMessageModel is use for worker to create a basecamp comment
type BasecampCommentMessageModel struct {
	ProjectID   int
	RecordingID int
	Payload     *bcModel.Comment
}

type BasecampTodoMessageModel struct {
	ProjectID int
	ListID    int
	Payload   bcModel.Todo
}
