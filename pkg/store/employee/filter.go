package employee

type SearchFilter struct {
	WorkingStatus string
}

type EditGeneralInfo struct {
	Fullname      string
	Email         string
	Phone         string
	LineManagerID string
	DiscordID     string
	GithubID      string
	NotionID      string
}
