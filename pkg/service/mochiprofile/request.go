package mochiprofile

import "time"

type ProfilePlatform string

const (
	ProfilePlatformDiscord ProfilePlatform = "discord"
	ProfilePlatformGithub  ProfilePlatform = "github"
)

type Pagination struct {
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Size  int64 `json:"size"`
}

type GetMochiProfilesResponse struct {
	Data       []MochiProfile `json:"data"`
	Pagination Pagination     `json:"pagination"`
}

type MochiProfile struct {
	ID                 string               `json:"id"`
	ProfileName        string               `json:"profile_name"`
	Avatar             string               `json:"avatar"`
	AssociatedAccounts []AssociatedAccounts `json:"associated_accounts"`
	Type               string               `json:"type"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

type AssociatedAccounts struct {
	Id                 string                 `json:"id"`
	ProfileId          string                 `json:"profile_id"`
	Platform           ProfilePlatform        `json:"platform"`
	PlatformIdentifier string                 `json:"platform_identifier"`
	PlatformMetadata   map[string]interface{} `json:"platform_metadata"`
	IsGuildMember      bool                   `json:"is_guild_member"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

type ListProfilesRequest struct {
	Types []ProfileType `json:"type"`
	IDs   []ProfileID   `json:"ids"`
	Page  int64         `json:"page"`
	Size  int64         `json:"size"`
}

type ProfileID string

func (i ProfileID) String() string {
	return string(i)
}

type ProfileType string

const (
	ProfileTypeUser        ProfileType = "user"
	ProfileTypeApplication ProfileType = "application"
	ProfileTypeVault       ProfileType = "vault"
)

func (t ProfileType) String() string {
	return string(t)
}
