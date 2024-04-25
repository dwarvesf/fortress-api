package communitynft

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
)

type role struct {
	Name string
}
type nftEmbedData struct {
	TokenId string

	// Profile Info
	DisplayName  string
	Username     string
	AvatarBase64 string
	JoinedDate   string

	// User stats
	GlobalXP     int
	ChatActivity int
	Level        int

	// User discord roles
	Roles []role
}

func (c *controller) GetNftMetadata(tokenId int) (*model.NftMetadata, error) {
	// 1 Get owner of an NFT
	addr, err := c.service.CommunityNft.OwnerOf(tokenId)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	// 2 Get nft needed data
	// 2.1 get mochi profile of nft owner
	profile, err := c.service.MochiProfile.GetProfileByEvmAddress(addr)
	if err != nil {
		return nil, ErrMochiProfileNotFound
	}
	var discordProfile mochiprofile.AssociatedAccounts
	for _, acc := range profile.AssociatedAccounts {
		if acc.Platform == "discord" {
			discordProfile = acc
			break
		}
	}
	discordUsername, ok := discordProfile.PlatformMetadata["username"]
	if !ok {
		discordUsername = "unknown"
	}
	displayName := profile.ProfileName
	if displayName == "" {
		displayName = strings.ToTitle(fmt.Sprintf("%s", discordUsername))
	}
	joinedDate := profile.CreatedAt.Format("May 02, 2006")
	userAvatar, err := c.userAvatarBase64(profile.Avatar)
	if err != nil {
		return nil, err
	}

	// 2.2 get tono guild user stats of mochi profile
	dfGuildId := "462663954813157376"
	guildProfile, err := c.service.Tono.GetGuildUserProfile(profile.ID, dfGuildId)
	if err != nil {
		return nil, err
	}

	// 2.3 get user discord roles
	// TODO: get user discord roles
	roles := []role{{Name: "peeps"}}

	// 2.4 compose nft embed data
	embedData := nftEmbedData{
		TokenId:      fmt.Sprintf("#%d", tokenId),
		DisplayName:  displayName,
		Username:     fmt.Sprintf("@%s", discordUsername),
		JoinedDate:   joinedDate,
		GlobalXP:     guildProfile.GuildXP,
		ChatActivity: guildProfile.NrOfActions,
		Level:        guildProfile.CurrentLevel.Level,
		Roles:        roles,
		AvatarBase64: userAvatar,
	}

	// 3. Generate image from embed data
	img, err := c.nftImage(embedData)
	if err != nil {
		return nil, err
	}

	// 4. Compose metadata
	metadata := &model.NftMetadata{
		Name:            embedData.DisplayName,
		Description:     "Dwarves Foundation NFT",
		Image:           img,
		BackgroundColor: "F0F4FC",
		Attributes: []model.NftAttribute{
			{
				TraitType: "Global XP",
				Value:     fmt.Sprint(embedData.GlobalXP),
			},
			{
				TraitType: "Chat Activity",
				Value:     fmt.Sprint(embedData.ChatActivity),
			},
			{
				TraitType: "Level",
				Value:     fmt.Sprint(embedData.Level),
			},
		},
	}

	return metadata, nil
}

func (c *controller) nftImage(data nftEmbedData) (string, error) {
	templatePath := c.config.Invoice.TemplatePath
	if c.config.Env == "local" || templatePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			pwd = os.Getenv("GOPATH") + "/src/github.com/dwarvesf/fortress-api"
		}
		templatePath = filepath.Join(pwd, "pkg/templates")
	}
	tmplFileName := "community_nft.tpl"
	tmpl, err := template.New("nft").ParseFiles(filepath.Join(templatePath, tmplFileName))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, tmplFileName, data); err != nil {
		return "", err
	}
	img := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/svg+xml;base64,%s", img), nil
}

func (c *controller) userAvatarBase64(avatarURL string) (string, error) {
	if avatarURL == "" {
		return "", nil
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(avatarURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(data)
	imgB64Str := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, imgB64Str), nil
}
