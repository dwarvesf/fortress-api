package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

var DwarvesAuthor = BasisEmployeeInfo{
	FullName:    "Dwarves' team",
	DisplayName: "Dwarves' team",
}

type Feedback struct {
	TopicID     string             `json:"topicID"`
	Title       string             `json:"title"`
	EventID     string             `json:"eventID"`
	EmployeeID  string             `json:"employeeID"`
	ProjectID   string             `json:"projectID"`
	ReviewerID  string             `json:"reviewerID"`
	Type        string             `json:"type"`
	Subtype     string             `json:"subtype"`
	Status      string             `json:"status"`
	IsRead      bool               `json:"isRead"`
	LastUpdated *time.Time         `json:"lastUpdated"`
	Author      *BasisEmployeeInfo `json:"author"`
}

func ToListFeedback(eTopics []*model.EmployeeEventTopic) []Feedback {
	var results = make([]Feedback, 0, len(eTopics))

	for _, topic := range eTopics {
		if len(topic.EmployeeEventReviewers) == 0 {
			continue
		}

		feedback := Feedback{
			TopicID:     topic.ID.String(),
			Title:       topic.Title,
			EventID:     topic.EventID.String(),
			EmployeeID:  topic.EmployeeID.String(),
			ProjectID:   topic.ProjectID.String(),
			Type:        topic.Event.Type.String(),
			Subtype:     topic.Event.Subtype.String(),
			Status:      topic.EmployeeEventReviewers[0].Status.String(),
			IsRead:      topic.EmployeeEventReviewers[0].IsRead,
			ReviewerID:  topic.EmployeeEventReviewers[0].ReviewerID.String(),
			LastUpdated: topic.EmployeeEventReviewers[0].UpdatedAt,
			Author:      ToBasisEmployeeInfo(topic.Event.Employee),
		}

		if topic.Event.Type == model.EventTypeSurvey {
			feedback.Author = &DwarvesAuthor
		}

		results = append(results, feedback)
	}

	return results
}

type ListFeedbackResponse struct {
	Data []Feedback `json:"data"`
}
