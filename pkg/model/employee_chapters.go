package model

type EmployeeChapter struct {
	BaseModel

	EmployeeID UUID
	ChapterID  UUID

	Chapter Chapter
}
