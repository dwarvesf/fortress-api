package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

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

func New(cfg *config.Config) DiscordService {
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

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}
