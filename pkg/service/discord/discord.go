package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
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
		if len(messages) == 0 {
			// early break to avoid index out of bound error
			break
		}
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

func (d *discordClient) ReportBraineryMetrics(queryView string, braineryMetric *view.BraineryMetric, channelID string) (*discordgo.Message, error) {
	var messageEmbed []*discordgo.MessageEmbedField
	totalICY := decimal.NewFromInt(0)
	content := ""

	var newBraineryPost []view.Post
	newBraineryPost = append(newBraineryPost, braineryMetric.Contributors...)
	newBraineryPost = append(newBraineryPost, braineryMetric.NewContributors...)

	if len(newBraineryPost) == 0 {
		content += "There is no new brainery note in this period. This is where we keep track of our **top 10 latest** Brainery notes:\n\n"

		for _, itm := range braineryMetric.LatestPosts {
			content += fmt.Sprintf("• [%s](%s) <@%v>\n", itm.Title, itm.URL, itm.DiscordID)
		}
	} else {
		newBraineryPostStr := ""
		for _, itm := range newBraineryPost {
			totalICY = totalICY.Add(itm.Reward)
			newBraineryPostStr += fmt.Sprintf("• [%s](%s) <@%v>\n", itm.Title, itm.URL, itm.DiscordID)
		}

		if len(newBraineryPostStr) > 0 {
			content += "**Latest Notes** :fire::fire::fire:\n"
			content += newBraineryPostStr + "\n"
		}
	}

	if queryView == "monthly" {
		topContributor := calculateTopContributor(braineryMetric.TopContributors)
		content += topContributor + "\n"
	}

	newContributor := ""
	if len(braineryMetric.NewContributors) > 0 {
		ids := make(map[string]bool)
		for _, itm := range braineryMetric.NewContributors {
			v, ok := ids[itm.DiscordID]
			if ok && v {
				continue
			}
			ids[itm.DiscordID] = true
			newContributor += fmt.Sprintf("<@%v> ", itm.DiscordID)
		}
	}

	if newContributor != "" {
		content += "**New Contributors**\n"
		content += newContributor + "\n"
	}

	if totalICY.GreaterThan(decimal.NewFromInt(0)) {
		content += "\n**Total Reward Distributed**\n"
		content += totalICY.String() + " ICY 🧊"
	}

	tags := ""
	if len(braineryMetric.Tags) > 0 {
		for _, tag := range braineryMetric.Tags {
			tags += fmt.Sprintf("#%v ", tag)
		}
	}

	if len(tags) > 0 {
		embedField := &discordgo.MessageEmbedField{
			Name:   "Tags",
			Value:  tags,
			Inline: false,
		}

		messageEmbed = append(messageEmbed, embedField)
	}

	msg := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("BRAINERY %s REPORT", strings.ToTitle(queryView)),
		Fields:      messageEmbed,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func calculateTopContributor(topContributors []view.TopContributor) string {
	topContributorStr := ""
	if len(topContributors) == 0 {
		return ""
	}

	countMap := make(map[int][]string)
	var uniqueCounts []int

	for _, contributor := range topContributors {
		if contributor.Count > 1 {
			count := contributor.Count
			discordID := contributor.DiscordID
			countMap[count] = append(countMap[count], discordID)

			// Check if count is already in uniqueCounts
			found := false
			for _, uniqueCount := range uniqueCounts {
				if uniqueCount == count {
					found = true
					break
				}
			}

			// If count is not found, add it to uniqueCounts
			if !found {
				uniqueCounts = append(uniqueCounts, count)
			}
		}
	}

	emojiMap := map[int]string{
		0: ":first_place:",
		1: ":second_place:",
		2: ":third_place:",
	}

	// Iterate over uniqueCounts to access Discord IDs in order
	for idx, count := range uniqueCounts {
		discordIDs := countMap[count]
		discordIDStr := ""
		for i := 0; i < len(discordIDs); i++ {
			discordIDStr += "<@" + discordIDs[i] + ">, "
		}

		emojiIdx := idx
		if idx > 2 {
			emojiIdx = 2
		}

		topContributorStr += fmt.Sprintf("%v %v (x%v) \n", emojiMap[emojiIdx], strings.TrimSuffix(discordIDStr, ", "), count)
	}

	topContributor := ""
	if len(topContributorStr) > 0 {
		topContributor += "**Top Contributors**\n"
		topContributor += topContributorStr
	}

	return topContributor
}

func (d *discordClient) DeliveryMetricWeeklyReport(deliveryMetric *view.DeliveryMetricWeeklyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error) {
	var messageEmbed []*discordgo.MessageEmbedField
	content := "*Track software team's performance. Encourages competition and collaboration. Optimizes project delivery. Promotes accountability.*\n\n"

	if leaderBoard != nil {
		leaderBoardStr := getLeaderboardAsString(leaderBoard.Items)
		content += "**Leaderboard**\n"
		content += leaderBoardStr
		content += "\n\n"
	}

	previousWeek := fmt.Sprintf("**Previous Week - %v**\n", deliveryMetric.LastWeek.Date.Format("02 Jan 2006"))
	previousWeek += fmt.Sprintf("%v`Total Point.       %v pts`\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastWeek.TotalPoints)))
	previousWeek += fmt.Sprintf("%v`Effort.            %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastWeek.Effort)))
	previousWeek += fmt.Sprintf("%v`AVG Weekly Point.  %v pts`\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastWeek.AvgPoint)))
	previousWeek += fmt.Sprintf("%v`AVG Weekly Effort. %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastWeek.AvgEffort)))

	content += previousWeek

	emojiUp := getEmoji("ARROW_UP_ANIMATED")
	emojiDown := getEmoji("ARROW_DOWN_ANIMATED")

	pointChange := fmt.Sprintf("%v %v%%", emojiUp, deliveryMetric.TotalPointChangePercentage)
	if deliveryMetric.TotalPointChangePercentage < 0 {
		pointChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.TotalPointChangePercentage)
	}

	effortChange := fmt.Sprintf("%v%v%%", emojiUp, deliveryMetric.EffortChangePercentage)
	if deliveryMetric.EffortChangePercentage < 0 {
		effortChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.EffortChangePercentage)
	}

	avgPointChange := fmt.Sprintf("%v%v%%", emojiUp, deliveryMetric.AvgPointChangePercentage)
	if deliveryMetric.AvgPointChangePercentage < 0 {
		avgPointChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgPointChangePercentage)
	}

	avgEffortChange := fmt.Sprintf("%v %v%%", emojiUp, deliveryMetric.AvgEffortChangePercentage)
	if deliveryMetric.AvgEffortChangePercentage < 0 {
		avgEffortChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgEffortChangePercentage)
	}

	date := deliveryMetric.CurrentWeek.Date.Format("02 Jan 2006")
	currentWeek := fmt.Sprintf("\n**Current Week - %v**\n", deliveryMetric.CurrentWeek.Date.Format("02 Jan 2006"))
	currentWeek += fmt.Sprintf("%v`Total Point.       %v pts` (%v)\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.TotalPoints)), pointChange)
	currentWeek += fmt.Sprintf("%v`Effort.            %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.Effort)), effortChange)
	currentWeek += fmt.Sprintf("%v`AVG Weekly Point.  %v pts` (%v)\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.AvgPoint)), avgPointChange)
	currentWeek += fmt.Sprintf("%v`AVG Weekly Effort. %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.AvgEffort)), avgEffortChange)

	content += currentWeek

	msg := &discordgo.MessageEmbed{
		Title:       "**🏆 DELIVERY WEEKLY REPORT 🏆**" + " - " + strings.ToUpper(date),
		Fields:      messageEmbed,
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func (d *discordClient) DeliveryMetricMonthlyReport(deliveryMetric *view.DeliveryMetricMonthlyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error) {
	content := "*Track software team's performance. Encourages competition and collaboration. Optimizes project delivery. Promotes accountability.*\n\n"
	msg := &discordgo.MessageEmbed{
		Title:       "**🏆 DELIVERY MONTHLY REPORT 🏆**" + " - " + strings.ToUpper("DATE"),
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func getLeaderboardAsString(data []view.LeaderBoardItem) string {
	emojiMap := map[int]string{
		1: getEmoji("BADGE1"),
		2: getEmoji("BADGE2"),
		3: getEmoji("BADGE3"),
		4: getEmoji("BADGE5"),
		5: getEmoji("BADGE5"),
	}
	// Sort the data by rank in ascending order
	var currentRank int
	var leaderboardString strings.Builder
	for _, employee := range data {
		if employee.Rank != currentRank {
			if currentRank > 0 {
				leaderboardString.WriteString("\n")
			}
			currentRank = employee.Rank
			leaderboardString.WriteString(fmt.Sprintf("%v ", emojiMap[currentRank]))
		}

		leaderboardString.WriteString(fmt.Sprintf("<@%v> ", employee.DiscordID))
	}

	return leaderboardString.String()
}
func (d *discordClient) SendEmbeddedMessageWithChannel(original *model.OriginalDiscordMessage, embed *discordgo.MessageEmbed, channelId string) (*discordgo.Message, error) {
	msg, err := d.session.ChannelMessageSendEmbed(channelId, normalize(original, embed))
	return msg, err
}
