package model

type LogDiscordInput struct {
	Type string
	Data interface{}
}

type DiscordLogTemplate struct {
	ID          string `json:"id"`
	Description string
	Content     string
}
