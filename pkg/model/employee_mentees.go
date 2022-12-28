package model

// EmployeeMentee define the model for table employee_mentees
type EmployeeMentee struct {
	BaseModel

	MenteeID UUID
	MentorID UUID

	Mentee *Employee `gorm:"foreignKey:MenteeID"`
}
