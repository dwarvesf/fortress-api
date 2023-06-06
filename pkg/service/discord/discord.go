package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	client = http.DefaultClient
)

type discordClient struct {
	cfg     *config.Config
	session *discordgo.Session
}

func New(cfg *config.Config) IService {
	ses, _ := discordgo.New("Bot " + cfg.Discord.SecretToken)
	return &discordClient{
		cfg:     cfg,
		session: ses,
	}
}

func (d *discordClient) PostBirthdayMsg(msg string) (model.DiscordMessage, error) {
	discordMsg := model.DiscordMessage{Content: msg}
	reqByte, err := json.Marshal(discordMsg)
	if err != nil {
		return discordMsg, err
	}

	payload := bytes.NewReader(reqByte)
	if _, err := d.newRequest(http.MethodPost, d.cfg.Discord.Webhooks.Campfire, payload); err != nil {
		return discordMsg, err
	}
	return discordMsg, nil
}

func (d *discordClient) CreateEvent(event *model.Schedule) (*discordgo.GuildScheduledEvent, error) {
	discordEvent := &discordgo.GuildScheduledEventParams{
		Name:               event.Name,
		Description:        event.Description,
		ScheduledStartTime: event.StartTime,
		ScheduledEndTime:   event.EndTime,
		PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
	}

	// by default, set channel to unknown
	discordEvent.EntityType = discordgo.GuildScheduledEventEntityTypeExternal
	discordEvent.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
		Location: "Unknown",
	}

	if event.GoogleCalendar.HangoutLink != "" {
		discordEvent.EntityType = discordgo.GuildScheduledEventEntityTypeExternal
		discordEvent.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
			Location: event.GoogleCalendar.HangoutLink,
		}
	}

	return d.session.GuildScheduledEventCreate(d.cfg.Discord.IDs.DwarvesGuild, discordEvent)
}

func (d *discordClient) UpdateEvent(event *model.Schedule) (*discordgo.GuildScheduledEvent, error) {
	discordEvent := &discordgo.GuildScheduledEventParams{
		Name:               event.Name,
		Description:        event.Description,
		ScheduledStartTime: event.StartTime,
		ScheduledEndTime:   event.EndTime,
	}

	return d.session.GuildScheduledEventEdit(d.cfg.Discord.IDs.DwarvesGuild, event.DiscordEvent.DiscordEventID, discordEvent)
}

func (d *discordClient) DeleteEvent(event *model.Schedule) error {
	return d.session.GuildScheduledEventDelete(d.cfg.Discord.IDs.DwarvesGuild, event.DiscordEvent.DiscordEventID)
}

func (d *discordClient) newRequest(method string, url string, payload io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (d *discordClient) GetMembers() ([]*discordgo.Member, error) {
	members := make([]*discordgo.Member, 0)

	after := ""
	limit := 1000
	for {
		guildMembers, err := d.session.GuildMembers(d.cfg.Discord.IDs.DwarvesGuild, after, limit)
		if err != nil {
			return nil, err
		}

		members = append(members, guildMembers...)

		if len(guildMembers) < limit {
			break
		}

		after = guildMembers[len(guildMembers)-1].User.ID
	}

	return members, nil
}

func (d *discordClient) SendMessage(msg, webhookUrl string) (*model.DiscordMessage, error) {
	discordMsg := model.DiscordMessage{Content: msg}
	reqByte, err := json.Marshal(discordMsg)
	if err != nil {
		return &discordMsg, err
	}

	payload := bytes.NewReader(reqByte)
	res, err := d.session.Client.Post(webhookUrl, "application/json", payload)
	if err != nil {
		return &discordMsg, err
	}
	defer res.Body.Close()

	return &discordMsg, nil
}

func (d *discordClient) SearchMember(discordName string) ([]*discordgo.Member, error) {
	members := make([]*discordgo.Member, 0)
	guildMembers, err := d.session.GuildMembersSearch(d.cfg.Discord.IDs.DwarvesGuild, discordName, 1000)
	if err != nil {
		return nil, err
	}

	members = append(members, guildMembers...)

	return members, nil
}

func (d *discordClient) GetMember(userID string) (*discordgo.Member, error) {
	member, err := d.session.GuildMember(d.cfg.Discord.IDs.DwarvesGuild, userID)
	if err != nil {
		return nil, err
	}

	return member, nil
}

func (d *discordClient) GetMemberByName(discordName string) ([]*discordgo.Member, error) {
	members := make([]*discordgo.Member, 0)
	guildMembers, err := d.session.GuildMembersSearch(d.cfg.Discord.IDs.DwarvesGuild, discordName, 1000)
	if err != nil {
		return nil, err
	}

	members = append(members, guildMembers...)

	return members, nil
}

func (d *discordClient) GetMemberByUsername(username string) (*discordgo.Member, error) {
	if len(username) == 0 {
		return nil, nil
	}

	discordNameParts := strings.Split(username, "#")

	guildMembers, err := d.SearchMember(discordNameParts[0])
	if err != nil {
		return nil, err
	}

	var discordMember *discordgo.Member
	for _, m := range guildMembers {
		if len(discordNameParts) == 1 {
			if m.User.Username == discordNameParts[0] {
				discordMember = m
			}
			break
		}
		if len(discordNameParts) > 1 {
			if m.User.Username == discordNameParts[0] && m.User.Discriminator == discordNameParts[1] {
				discordMember = m
			}
			break
		}
	}

	return discordMember, nil
}

func (d *discordClient) GetRoles() (Roles, error) {
	roles, err := d.session.GuildRoles(d.cfg.Discord.IDs.DwarvesGuild)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (d *discordClient) AddRole(userID, roleID string) error {
	return d.session.GuildMemberRoleAdd(d.cfg.Discord.IDs.DwarvesGuild, userID, roleID)
}

func (d *discordClient) RemoveRole(userID string, roleID string) error {
	return d.session.GuildMemberRoleRemove(d.cfg.Discord.IDs.DwarvesGuild, userID, roleID)
}

type Roles discordgo.Roles

func (r Roles) DwarvesRoles() []*discordgo.Role {
	roleMap := getDwarvesRolesMap()

	dwarvesRoles := make([]*discordgo.Role, 0)
	for _, dRole := range r {
		_, ok := roleMap[dRole.Name]
		if ok {
			dwarvesRoles = append(dwarvesRoles, dRole)
		}
	}

	return dwarvesRoles
}

func (r Roles) ByCode(code string) *discordgo.Role {
	for _, dRole := range r {
		if dRole.Name == code {
			return dRole
		}
	}

	return nil
}

func getDwarvesRolesMap() map[string]bool {
	return map[string]bool{
		"moderator":  true,
		"dwarf":      true,
		"booster":    true,
		"apprentice": true,
		"crafter":    true,
		"specialist": true,
		"principal":  true,
		"peeps":      true,
		"learning":   true,
		"engagement": true,
		"delivery":   true,
		"labs":       true,
		"baby dwarf": true,
		"ladies":     true,
		"sers":       true,
		"consultant": true,
		"chad":       true,
	}
}

func (d *discordClient) GetChannels() ([]*discordgo.Channel, error) {
	return d.session.GuildChannels(d.cfg.Discord.IDs.DwarvesGuild)
}

func (d *discordClient) GetMessagesAfterCursor(
	channelID string,
	cursorMessageID string,
	lastMessageID string,
) ([]*discordgo.Message, error) {
	cursorMessageIDUint, err := strconv.ParseUint(cursorMessageID, 10, 64)
	if err != nil {
		return nil, err
	}
	lastMessageIDUint, err := strconv.ParseUint(lastMessageID, 10, 64)
	if err != nil {
		return nil, err
	}

	allMessages := make([]*discordgo.Message, 0)
	for cursorMessageIDUint < lastMessageIDUint {
		messages, err := d.session.ChannelMessages(
			channelID,
			100, // 100 is the maximal number allowed
			"",
			cursorMessageID,
			"",
		)
		if err != nil {
			return nil, err
		}
		// reversal is needed since messages are sorted by newest first
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}

		allMessages = append(allMessages, messages...)
		newestMessage := messages[len(messages)-1]
		cursorMessageID = newestMessage.ID
		cursorMessageIDUint, err = strconv.ParseUint(cursorMessageID, 10, 64)
		if err != nil {
			return nil, err
		}
		// a pause is needed to avoid Discord's rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allMessages, nil
}
