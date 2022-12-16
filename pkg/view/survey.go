package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type Survey struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Type      string        `json:"type"`
	Subtype   string        `json:"subtype"`
	Status    string        `json:"status"`
	StartDate *time.Time    `json:"startDate"`
	EndDate   *time.Time    `json:"endDate"`
	Count     FeedbackCount `json:"count"`
}

type FeedbackCount struct {
	Total int `json:"total"`
	Sent  int `json:"sent"`
	Done  int `json:"done"`
}

func ToListSurvey(events []*model.FeedbackEvent) []Survey {
	var results = make([]Survey, 0, len(events))

	for _, e := range events {
		var sent, done int

		for _, topic := range e.Topics {
			var topicSent, topicDone int

			for _, reviewer := range topic.EmployeeEventReviewers {
				if reviewer.AuthorStatus != model.EventAuthorStatusDraft {
					topicSent++
				}
				if reviewer.AuthorStatus == model.EventAuthorStatusDone {
					topicDone++
				}
			}

			if topicSent == len(topic.EmployeeEventReviewers) {
				sent++
			}

			if topicDone == len(topic.EmployeeEventReviewers) {
				done++
			}
		}

		results = append(results, Survey{
			ID:        e.ID.String(),
			Title:     e.Title,
			Type:      e.Type.String(),
			Subtype:   e.Subtype.String(),
			Status:    e.Status.String(),
			StartDate: e.StartDate,
			EndDate:   e.EndDate,
			Count: FeedbackCount{
				Total: len(e.Topics),
				Sent:  sent,
				Done:  done,
			},
		})
	}

	return results
}

type ListSurveyResponse struct {
	Data []Survey `json:"data"`
}

type SurveyDetail struct {
	EventID   string             `json:"eventID"`
	Title     string             `json:"title"`
	Type      string             `json:"type"`
	Subtype   string             `json:"subtype"`
	Status    string             `json:"status"`
	StartDate *time.Time         `json:"startDate"`
	EndDate   *time.Time         `json:"endDate"`
	Author    *BasicEmployeeInfo `json:"author"`
	Topics    []Topic            `json:"topics"`
}

type Topic struct {
	ID           string              `json:"id"`
	EventID      string              `json:"eventID"`
	ReviewID     string              `json:"reviewID,omitempty"`
	Title        string              `json:"title"`
	Type         string              `json:"type"`
	Subtype      string              `json:"subtype"`
	Status       string              `json:"status,omitempty"`
	Employee     BasicEmployeeInfo   `json:"employee"`
	Participants []BasicEmployeeInfo `json:"participants"`
	Count        *FeedbackCount      `json:"count"`
}

func ToSurveyDetail(event *model.FeedbackEvent) SurveyDetail {
	result := SurveyDetail{
		EventID:   event.ID.String(),
		Title:     event.Title,
		Type:      event.Type.String(),
		Subtype:   event.Subtype.String(),
		Status:    event.Status.String(),
		StartDate: event.StartDate,
		EndDate:   event.EndDate,
		Author:    toBasicEmployeeInfo(event.Employee),
	}

	var topics = make([]Topic, 0, len(event.Topics))

	for _, topic := range event.Topics {
		newTopic := Topic{
			ID:       topic.ID.String(),
			EventID:  topic.EventID.String(),
			Title:    topic.Title,
			Type:     topic.Event.Type.String(),
			Subtype:  topic.Event.Subtype.String(),
			Employee: *toBasicEmployeeInfo(*topic.Employee),
		}

		// just use for peer-review survey
		if topic.Event.Subtype == model.EventSubtypePeerReview {
			participants := make([]BasicEmployeeInfo, 0, len(topic.EmployeeEventReviewers))
			for _, reviewer := range topic.EmployeeEventReviewers {
				employee := toBasicEmployeeInfo(*reviewer.Reviewer)
				participants = append(participants, *employee)
			}

			var sent, done int
			for _, reviewer := range topic.EmployeeEventReviewers {
				if reviewer.AuthorStatus != model.EventAuthorStatusDraft {
					sent++
				}
				if reviewer.AuthorStatus == model.EventAuthorStatusDone {
					done++
				}
			}

			newTopic.Participants = participants
			newTopic.Count = &FeedbackCount{
				Total: len(topic.EmployeeEventReviewers),
				Sent:  sent,
				Done:  done,
			}
		}

		// just use for engagement survey
		if event.Subtype == model.EventSubtypeEngagement && len(topic.EmployeeEventReviewers) > 0 {
			newTopic.ReviewID = topic.EmployeeEventReviewers[0].ID.String()
			newTopic.Status = topic.EmployeeEventReviewers[0].AuthorStatus.String()
		}

		topics = append(topics, newTopic)
	}

	result.Topics = topics
	return result
}

type PeerReviewer struct {
	EventReviewerID string                    `json:"eventReviewerID"`
	Reviewer        *BasicEmployeeInfo        `json:"reviewer"`
	Status          model.EventReviewerStatus `json:"status"`
	Relationship    model.Relationship        `json:"relationship"`
}

type SurveyTopicDetail struct {
	TopicID      string             `json:"topicID"`
	Title        string             `json:"title"`
	Employee     *BasicEmployeeInfo `json:"employee"`
	Participants []PeerReviewer     `json:"participants"`
}

type SurveyTopicDetailResponse struct {
	Data SurveyTopicDetail `json:"data"`
}

func ToPeerReviewDetail(topic *model.EmployeeEventTopic) SurveyTopicDetail {
	rs := SurveyTopicDetail{
		TopicID:  topic.ID.String(),
		Employee: toBasicEmployeeInfo(*topic.Employee),
		Title:    topic.Title,
	}

	for _, eventReviewer := range topic.EmployeeEventReviewers {
		participant := PeerReviewer{
			EventReviewerID: eventReviewer.ID.String(),
			Status:          model.EventReviewerStatus(eventReviewer.AuthorStatus),
			Relationship:    eventReviewer.Relationship,
		}

		if eventReviewer.Reviewer != nil {
			participant.Reviewer = toBasicEmployeeInfo(*eventReviewer.Reviewer)
		}

		rs.Participants = append(rs.Participants, participant)

	}

	return rs
}

type ListSurveyDetailResponse struct {
	Data SurveyDetail `json:"data"`
}