package model

// EmployeeEventReviewer model for table employee_event_reviewer
type EmployeeEventReviewer struct {
	BaseModel

	EmployeeEventTopicID UUID
	ReviewerID           UUID
	Status               string
	Relationship         Relationship
	IsShared             bool
	IsRead               bool
}
