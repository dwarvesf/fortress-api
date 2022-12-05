package model

// AnswerDraft model for answer_drafts table
type AnswerDraft struct {
	BaseModel

	QuestionID UUID
	AnswerBy   UUID
	Answer     string
	Note       string
}
