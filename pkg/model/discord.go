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

type DiscordRole string

func (r DiscordRole) String() string {
	return string(r)
}

const (
	DiscordRolePeeps DiscordRole = "peeps"
)
