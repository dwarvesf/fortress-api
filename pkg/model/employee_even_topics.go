package model

// Relationship relationships for table employee_event_topics
type Relationship string

// values for Relationship
const (
	RelationshipPeer        Relationship = "peer"
	RelationshipLineManager Relationship = "line-manager"
	RelationshipChapterLead Relationship = "chapter-lead"
	RelationshipSelf        Relationship = "self"
)

// IsValid validation for Relationship
func (e Relationship) IsValid() bool {
	switch e {
	case
		RelationshipPeer,
		RelationshipLineManager,
		RelationshipChapterLead,
		RelationshipSelf:
		return true
	}
	return false
}

// EmployeeEventTopic model for table employee_event_topics
type EmployeeEventTopic struct {
	BaseModel

	Title      string
	EventID    UUID
	EmployeeID UUID
	ProjectID  UUID
}
