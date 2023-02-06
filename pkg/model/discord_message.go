package model

type DiscordMessage struct {
	Username   string                     `json:"username"`
	AvatarURL  string                     `json:"avatar_url"`
	Content    string                     `json:"content"`
	Embeds     []DiscordMessageEmbed      `json:"embeds"`
	Components []DiscordMessageComponents `json:"components"`
}

type DiscordMessageEmbed struct {
	Author      DiscordMessageAuthor  `json:"author"`
	Title       string                `json:"title"`
	URL         string                `json:"url"`
	Description string                `json:"description"`
	Color       int64                 `json:"color"`
	Fields      []DiscordMessageField `json:"fields"`
	Thumbnail   DiscordMessageImage   `json:"thumbnail"`
	Image       DiscordMessageImage   `json:"image"`
	Footer      DiscordMessageFooter  `json:"footer"`
}

type DiscordMessageAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type DiscordMessageField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline *bool  `json:"inline,omitempty"`
}

type DiscordMessageFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type DiscordMessageImage struct {
	URL string `json:"url"`
}

type DiscordMessageComponents struct {
	Components []DiscordMessageComponent `json:"components"`
	Type       int64                     `json:"type"`
}

type DiscordMessageComponent struct {
	CustomID string      `json:"custom_id"`
	Disabled bool        `json:"disabled"`
	Emoji    interface{} `json:"emoji"`
	Label    string      `json:"label"`
	Style    int64       `json:"style"`
	Type     int64       `json:"type"`
	URL      interface{} `json:"url"`
}
