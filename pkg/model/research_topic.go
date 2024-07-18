package model

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
