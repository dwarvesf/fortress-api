package view

import "github.com/dwarvesf/fortress-api/pkg/model"

// DiscordResearchTopic represents discord research topic
type DiscordResearchTopic struct {
	Name              string
	URL               string
	MsgCount          int64
	SortedActiveUsers []DiscordTopicActiveUser
}

// DiscordTopicActiveUser represents active users who send most messages in topic
type DiscordTopicActiveUser struct {
	UserID   string
	MsgCount int64
}

// ListResearchTopicResponse represents list of research topic
type ListResearchTopicResponse struct {
	PaginationResponse
	Data []DiscordResearchTopic `json:"data"`
} // @name ListResearchTopicResponse

// ToListResearchTopicResponse returns list of research topic
func ToListResearchTopicResponse(rs []model.DiscordResearchTopic) []DiscordResearchTopic {
	data := make([]DiscordResearchTopic, 0)
	for _, r := range rs {
		data = append(data, DiscordResearchTopic{
			Name:              r.Name,
			URL:               r.URL,
			MsgCount:          r.MsgCount,
			SortedActiveUsers: ToDiscordTopicActiveUsers(r.SortedActiveUsers),
		})
	}
	return data
}

// ToDiscordTopicActiveUser returns list of active users who send most messages in topic
func ToDiscordTopicActiveUsers(rs []model.DiscordTopicActiveUser) []DiscordTopicActiveUser {
	data := make([]DiscordTopicActiveUser, 0)
	for _, r := range rs {
		data = append(data, DiscordTopicActiveUser{
			UserID:   r.UserID,
			MsgCount: r.MsgCount,
		})
	}
	return data
}
