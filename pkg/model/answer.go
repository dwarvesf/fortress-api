package model

// Answer model for answers table
type Answer struct {
	BaseModel

	QuestionID UUID
	AnswerBy   UUID
	Answer     string
	Note       string
}
