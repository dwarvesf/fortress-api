package mochiprofile

type ProfilePlatform string

const (
	ProfilePlatformDiscord ProfilePlatform = "discord"
)

type GetMochiProfilesResponse struct {
	Data []MochiProfile `json:"data"`
}

type MochiProfile struct {
	Id                 string               `json:"id"`
	CreatedAt          string               `json:"created_at"`
	UpdatedAt          string               `json:"updated_at"`
	ProfileName        string               `json:"profile_name"`
	Avatar             string               `json:"avatar"`
	AssociatedAccounts []AssociatedAccounts `json:"associated_accounts"`
	Type               string               `json:"type"`
}

type AssociatedAccounts struct {
	Id                 string          `json:"id"`
	ProfileId          string          `json:"profile_id"`
	Platform           ProfilePlatform `json:"platform"`
	PlatformIdentifier string          `json:"platform_identifier"`
	PlatformMetadata   interface{}     `json:"platform_metadata"`
	IsGuildMember      bool            `json:"is_guild_member"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
}
