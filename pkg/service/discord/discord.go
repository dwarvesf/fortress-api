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
	"github.com/dwarvesf/fortress-api/pkg/service/discord/helpers"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

var (
	client = http.DefaultClient
)

const (
	memoCategoryFleeting   = "00_fleeting"
	memoCategoryLiterature = "01_literature"
	memoCategoryEarn       = "earn"
	memoCategoryOthers     = "others"
	discordEmbedMaxLen     = 4096
)

type discordClient struct {
	cfg                *config.Config
	session            *discordgo.Session
	breakdownDetector  helpers.BreakdownDetector
	leaderboardBuilder helpers.LeaderboardBuilder
	messageFormatter   helpers.MessageFormatter
	timeCalculator     helpers.TimeCalculator
}

func New(cfg *config.Config) IService {
	ses, _ := discordgo.New("Bot " + cfg.Discord.SecretToken)

	// Initialize helper functions with default configurations
	breakdownDetector := helpers.NewBreakdownDetector(helpers.BreakdownDetectionConfig{
		TitleKeywords:          []string{"breakdown"},
		TagKeywords:            []string{"breakdown"},
		CaseSensitive:          false,
		RequireBothTitleAndTag: false,
	})

	timeCalculator := helpers.NewTimeCalculator(helpers.TimeCalculatorConfig{
		WeekStartDay:  time.Monday,
		MonthStartDay: 1,
		DateFormat:    "2-Jan",
		RangeFormat:   "%s-%s",
		TimeZone:      "UTC",
	})

	messageFormatter := helpers.NewMessageFormatter(helpers.MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2 Jan 2006",
	})

	// For now, pass nil for AuthorResolver - will be updated when implementing remaining helpers
	leaderboardBuilder := helpers.NewLeaderboardBuilder(nil, helpers.LeaderboardConfig{
		MaxEntries:     10,
		ShowZeroScores: false,
		SortDescending: true,
	})

	return &discordClient{
		cfg:                cfg,
		session:            ses,
		breakdownDetector:  breakdownDetector,
		leaderboardBuilder: leaderboardBuilder,
		messageFormatter:   messageFormatter,
		timeCalculator:     timeCalculator,
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

func (d *discordClient) ListEvents() ([]*discordgo.GuildScheduledEvent, error) {
	return d.session.GuildScheduledEvents(d.cfg.Discord.IDs.DwarvesGuild, false)
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

func (d *discordClient) SendMessage(discordMsg model.DiscordMessage, webhookUrl string) (*model.DiscordMessage, error) {
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
	if username == "" {
		return nil, nil
	}

	guildMembers, err := d.SearchMember(username)
	if err != nil {
		return nil, err
	}

	var discordMember *discordgo.Member
	for _, m := range guildMembers {
		if m.User.Username == username {
			discordMember = m
		}
	}

	return discordMember, nil
}

func (d *discordClient) GetRoles() (Roles, error) {
	roles, err := d.session.GuildRoles(d.cfg.Discord.IDs.DwarvesGuild)
	if err != nil {
		return Roles{}, err
	}

	return Roles(roles), nil
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

func (r Roles) ByID(id string) *discordgo.Role {
	for _, dRole := range r {
		if dRole.ID == id {
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

func CreateDeliveryMetricWeeklyReportMessage(deliveryMetric *view.DeliveryMetricWeeklyReport, leaderBoard *view.WeeklyLeaderBoard) *discordgo.MessageEmbed {
	var messageEmbed []*discordgo.MessageEmbedField
	content := "*Track software team's performance. Encourages competition and collaboration. Optimizes project delivery. Promotes accountability.*\n\n"

	if leaderBoard != nil {
		leaderBoardStr := getLeaderBoardAsString(leaderBoard.Items)
		content += "**Leaderboard**\n"
		content += leaderBoardStr
		content += "\n\n"
	}

	previousWeek := fmt.Sprintf("**Previous Week - %v**\n", deliveryMetric.LastWeek.Date.Format("02 Jan 2006"))
	previousWeek += fmt.Sprintf("%v`Total Point.  %v pts`\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastWeek.TotalPoints)))
	previousWeek += fmt.Sprintf("%v`Effort.       %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastWeek.Effort)))
	previousWeek += fmt.Sprintf("%v`AVG W.Point.  %v pts`\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastWeek.AvgPoint)))
	previousWeek += fmt.Sprintf("%v`AVG W.Effort. %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastWeek.AvgEffort)))

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
	currentWeek += fmt.Sprintf("%v`Total Point.  %v pts` (%v)\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.TotalPoints)), pointChange)
	currentWeek += fmt.Sprintf("%v`Effort.       %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.Effort)), effortChange)
	currentWeek += fmt.Sprintf("%v`AVG W.Point.  %v pts` (%v)\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.AvgPoint)), avgPointChange)
	currentWeek += fmt.Sprintf("%v`AVG W.Effort. %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentWeek.AvgEffort)), avgEffortChange)

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

	return msg
}

func (d *discordClient) DeliveryMetricWeeklyReport(deliveryMetric *view.DeliveryMetricWeeklyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error) {
	msg := CreateDeliveryMetricWeeklyReportMessage(deliveryMetric, leaderBoard)
	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func CreateDeliveryMetricMonthlyReportMessage(deliveryMetric *view.DeliveryMetricMonthlyReport, leaderBoard *view.WeeklyLeaderBoard) *discordgo.MessageEmbed {
	content := "*Track software team's performance. Encourages competition and collaboration. Optimizes project delivery. Promotes accountability.*\n\n"

	if leaderBoard != nil {
		leaderBoardStr := getLeaderBoardAsString(leaderBoard.Items)
		content += "**Leaderboard**\n"
		content += leaderBoardStr
		content += "\n\n"
	}

	previousMonth := fmt.Sprintf("**Previous Month - %v**\n", deliveryMetric.LastMonth.Month.Format("Jan 2006"))
	previousMonth += fmt.Sprintf("%v`Total Point.  %v pts`\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastMonth.TotalWeight)))
	previousMonth += fmt.Sprintf("%v`Effort.       %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastMonth.Effort)))
	previousMonth += fmt.Sprintf("%v`AVG Point.    %v pts`\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastMonth.AvgWeight)))
	previousMonth += fmt.Sprintf("%v`AVG Effort.   %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastMonth.AvgEffort)))
	previousMonth += fmt.Sprintf("%v`AVG W.Point.  %v pts`\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.LastMonth.AvgWeeklyWeight)))
	previousMonth += fmt.Sprintf("%v`AVG W.Effort. %v hrs`\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.LastMonth.AvgWeeklyEffort)))

	content += previousMonth

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

	avgPointChange := fmt.Sprintf("%v%v%%", emojiUp, deliveryMetric.AvgWeightChangePercentage)
	if deliveryMetric.AvgWeightChangePercentage < 0 {
		avgPointChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgWeightChangePercentage)
	}

	avgEffortChange := fmt.Sprintf("%v %v%%", emojiUp, deliveryMetric.AvgEffortChangePercentage)
	if deliveryMetric.AvgEffortChangePercentage < 0 {
		avgEffortChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgEffortChangePercentage)
	}

	avgWeeklyPointChange := fmt.Sprintf("%v%v%%", emojiUp, deliveryMetric.AvgWeeklyPointChangePercentage)
	if deliveryMetric.AvgWeeklyPointChangePercentage < 0 {
		avgWeeklyPointChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgWeeklyPointChangePercentage)
	}

	avgWeeklyEffortChange := fmt.Sprintf("%v %v%%", emojiUp, deliveryMetric.AvgWeeklyEffortChangePercentage)
	if deliveryMetric.AvgWeeklyEffortChangePercentage < 0 {
		avgWeeklyEffortChange = fmt.Sprintf("%v%v%%", emojiDown, deliveryMetric.AvgWeeklyEffortChangePercentage)
	}

	currentMonth := fmt.Sprintf("\n**Current Month - %v**\n", deliveryMetric.CurrentMonth.Month.Format("Jan 2006"))
	currentMonth += fmt.Sprintf("%v`Total Point.  %v pts` (%v)\n", getEmoji("STAR_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.TotalWeight)), pointChange)
	currentMonth += fmt.Sprintf("%v`Effort.       %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.Effort)), effortChange)
	currentMonth += fmt.Sprintf("%v`AVG Point.    %v pts` (%v)\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.AvgWeight)), avgPointChange)
	currentMonth += fmt.Sprintf("%v`AVG Effort.   %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.AvgEffort)), avgEffortChange)
	currentMonth += fmt.Sprintf("%v`AVG W.Point.  %v pts` (%v)\n", getEmoji("INCREASING_ANIMATED"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.AvgWeeklyWeight)), avgWeeklyPointChange)
	currentMonth += fmt.Sprintf("%v`AVG W.Effort. %v hrs` (%v)\n", getEmoji("CLOCK_NEW"), utils.FloatToString(float64(deliveryMetric.CurrentMonth.AvgWeeklyEffort)), avgWeeklyEffortChange)

	content += currentMonth

	month := deliveryMetric.CurrentMonth.Month.Format("Jan 2006")
	msg := &discordgo.MessageEmbed{
		Title:       "**🏆 DELIVERY MONTHLY REPORT 🏆**" + " - " + strings.ToUpper(month),
		Description: content,
		Footer: &discordgo.MessageEmbedFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
	}

	return msg
}

func (d *discordClient) DeliveryMetricMonthlyReport(deliveryMetric *view.DeliveryMetricMonthlyReport, leaderBoard *view.WeeklyLeaderBoard, channelID string) (*discordgo.Message, error) {
	msg := CreateDeliveryMetricMonthlyReportMessage(deliveryMetric, leaderBoard)
	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func getLeaderBoardAsString(data []view.LeaderBoardItem) string {
	emojiMap := map[int]string{
		1: getEmoji("BADGE1"),
		2: getEmoji("BADGE2"),
		3: getEmoji("BADGE3"),
		4: getEmoji("BADGE5"),
		5: getEmoji("BADGE5"),
	}
	// Sort the data by rank in ascending order
	var currentRank int
	var leaderBoardString strings.Builder
	for _, employee := range data {
		rank := employee.Rank
		if rank > 5 {
			break
		}
		if rank == 5 {
			rank = 4
		}
		if rank != currentRank {
			if currentRank > 0 {
				leaderBoardString.WriteString("\n")
			}
			currentRank = employee.Rank
			leaderBoardString.WriteString(fmt.Sprintf("%v ", emojiMap[currentRank]))
		}

		leaderBoardString.WriteString(fmt.Sprintf("<@%v> ", employee.DiscordID))
	}

	return leaderBoardString.String()
}
func (d *discordClient) SendEmbeddedMessageWithChannel(original *model.OriginalDiscordMessage, embed *discordgo.MessageEmbed, channelId string) (*discordgo.Message, error) {
	msg, err := d.session.ChannelMessageSendEmbed(channelId, normalize(original, embed))
	return msg, err
}

func (d *discordClient) SendNewMemoMessage(
	guildID string,
	memos []model.MemoLog,
	channelID string,
	getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
) (*discordgo.Message, error) {
	for i, content := range memos {
		if i <= 10 {
			description := content.Description
			if len(description) > 300 {
				description = description[0:300] + "..."
			}
			// Create an embedded message with a gift-like format
			embed := &discordgo.MessageEmbed{
				Description: fmt.Sprintf("Mew memo is published! Check it out [%s](%s) \n\n%s",
					content.Title,
					content.URL,
					description,
				),
				Color: 0xFF69B4, // Pink color for the embed
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://media0.giphy.com/media/v1.Y2lkPTc5MGI3NjExNWJ0ZzA4aHJ1Nm02aDZhNXFpY3pnMGR3aDNibGNseTcyOG9xc2d1cCZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9cw/TjBUgaxEO98wFkLStc/giphy.gif",
				},
			}

			_, err := d.SendEmbeddedMessageWithChannel(nil, embed, channelID)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (d *discordClient) SendWeeklyMemosMessage(
	guildID string,
	memos []model.MemoLog,
	weekRangeStr,
	channelID string,
	getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
	newAuthors []string,
	resolveAuthorsByTitle func(title string) ([]string, error),
	getDiscordIDByUsername func(username string) (string, error),
) (*discordgo.Message, error) {
	// Helper function to format Discord mention from username
	formatDiscordMention := func(username string) string {
		if discordID, err := getDiscordIDByUsername(username); err == nil && discordID != "" {
			// Successfully found Discord ID - use proper mention format
			return fmt.Sprintf("<@%s>", discordID)
		}
		// Fallback to plain text if Discord ID not found (expected in dev environment)
		return "@" + username
	}

	// 1. Detect breakdown posts for count (leaderboard moved to separate function)
	breakdowns := d.breakdownDetector.DetectBreakdowns(memos)
	breakdownCount := len(breakdowns)

	// 2. Build content using new simplified format (without leaderboard)
	var content strings.Builder
	var memolistString strings.Builder

	content.WriteString("What is going on with our memo this week?\n\n")
	content.WriteString("**📊 Overview**\n\n")
	content.WriteString(fmt.Sprintf("- Total publication. %v posts\n", len(memos)))

	// Add new authors section if there are any
	if len(newAuthors) > 0 {
		content.WriteString("- New authors. ")
		for i, author := range newAuthors {
			if i > 0 {
				content.WriteString(", ")
			}
			content.WriteString(formatDiscordMention(author))
		}
		content.WriteString("\n")
	}

	content.WriteString(fmt.Sprintf("- New breakdowns. %v posts\n\n", breakdownCount))

	memolistString.WriteString("**📖 Publications**\n\n")

	// Simple numbered list format: "1. Title - @username"
	for idx, mem := range memos {
		authorField := ""
		for _, discordAccountID := range mem.DiscordAccountIDs {
			discordAccount, err := getDiscordAccountByID(discordAccountID)
			if err != nil {
				// If fetching fails, use the ID as fallback
				if authorField != "" {
					authorField += ", "
				}
				authorField += fmt.Sprintf("@%s", discordAccountID)
				continue
			}

			if authorField != "" {
				authorField += ", "
			}

			// Use Discord ID for proper mentions if available, otherwise fallback to username
			if discordAccount.DiscordID != "" {
				authorField += fmt.Sprintf("<@%s>", discordAccount.DiscordID)
			} else if discordAccount.DiscordUsername != "" {
				authorField += fmt.Sprintf("@%s", discordAccount.DiscordUsername)
			} else {
				authorField += fmt.Sprintf("@%s", discordAccountID)
			}
		}

		// If no valid authors found through Discord accounts, try parquet resolution by title
		if authorField == "" {
			if parquetAuthors, err := resolveAuthorsByTitle(mem.Title); err == nil && len(parquetAuthors) > 0 {
				for i, author := range parquetAuthors {
					if i > 0 {
						authorField += ", "
					}
					authorField += fmt.Sprintf("@%s", author)
				}
			}
		}

		// Final fallback
		if authorField == "" {
			authorField = "@unknown-user"
		}

		// Format with clickable link if URL exists, otherwise plain text
		if mem.URL != "" {
			memolistString.WriteString(fmt.Sprintf("%d. [%s](%s) - %s\n", idx+1, mem.Title, mem.URL, authorField))
		} else {
			memolistString.WriteString(fmt.Sprintf("%d. %s - %s\n", idx+1, mem.Title, authorField))
		}
	}

	content.WriteString(memolistString.String())

	// 5. Create and send Discord embed
	msg := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("**Weekly memo report (%v)**", weekRangeStr),
		Description: content.String(),
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func (d *discordClient) SendMonthlyMemosMessage(
	guildID string,
	memos []model.MemoLog,
	monthRangeStr,
	channelID string,
	getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
	newAuthors []string,
	getDiscordIDByUsername func(username string) (string, error),
) (*discordgo.Message, error) {
	// 1. Detect breakdown posts for count (ICY calculation moved to leaderboard message)
	breakdowns := d.breakdownDetector.DetectBreakdowns(memos)

	// 2. Extract unique authors from all posts
	uniqueAuthors := make(map[string]bool)
	for _, memo := range memos {
		for _, username := range memo.AuthorMemoUsernames {
			if username != "" {
				uniqueAuthors[username] = true
			}
		}
	}

	// Helper function to format Discord mention from username
	formatDiscordMention := func(username string) string {
		if discordID, err := getDiscordIDByUsername(username); err == nil && discordID != "" {
			// Successfully found Discord ID - use proper mention format
			return fmt.Sprintf("<@%s>", discordID)
		}
		// Fallback to plain text if Discord ID not found (expected in dev environment)
		return "@" + username
	}

	// 5. Build simple content format
	var content strings.Builder

	content.WriteString("What is going on with our memo this month?\n\n")
	content.WriteString("**📊 Overview**\n\n")
	content.WriteString(fmt.Sprintf("- Total publication. %d posts\n", len(memos)))

	// Add new authors if any (detected using simple logic: authors with exactly 1 memo total)
	if len(newAuthors) > 0 {
		newAuthorsStr := ""
		for i, author := range newAuthors {
			if i > 0 {
				newAuthorsStr += ", "
			}
			newAuthorsStr += formatDiscordMention(author)
		}
		content.WriteString(fmt.Sprintf("- New authors. %s\n", newAuthorsStr))
	}

	content.WriteString(fmt.Sprintf("- New breakdowns. %d posts\n\n", len(breakdowns)))

	// Publications list
	content.WriteString("**📖 Publications**\n\n")
	for idx, memo := range memos {
		// Get first author
		author := "unknown"
		if len(memo.AuthorMemoUsernames) > 0 {
			author = formatDiscordMention(memo.AuthorMemoUsernames[0])
		}
		// Format with clickable link if URL exists, otherwise plain text
		if memo.URL != "" {
			content.WriteString(fmt.Sprintf("%d. [%s](%s) - %s\n", idx+1, memo.Title, memo.URL, author))
		} else {
			content.WriteString(fmt.Sprintf("%d. %s - %s\n", idx+1, memo.Title, author))
		}
	}

	// 6. Create and send Discord embed
	msg := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Monthly memo report (%s)", monthRangeStr),
		Description: content.String(),
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func (d *discordClient) SendLeaderboardMessage(
	guildID string,
	period string, // "weekly" or "monthly"
	channelID string,
	getDiscordAccountByID func(discordAccountID string) (*model.DiscordAccount, error),
	getDiscordIDByUsername func(username string) (string, error),
	getAllTimeMemos func() ([]model.MemoLog, error), // Function to fetch all-time memos since July 2025
) (*discordgo.Message, error) {
	// Helper function to format Discord mention from username
	formatDiscordMention := func(username string) string {
		if discordID, err := getDiscordIDByUsername(username); err == nil && discordID != "" {
			return fmt.Sprintf("<@%s>", discordID)
		}
		return "@" + username
	}

	// 1. Fetch all-time memos since July 2025
	allTimeMemos, err := getAllTimeMemos()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all-time memos: %w", err)
	}

	// Note: AuthorMemoUsernames should already be populated by getMemosFromParquet method
	// No manual population needed as the memo fetching method handles this correctly

	// 3. Detect breakdown posts for leaderboard from all-time data
	breakdowns := d.breakdownDetector.DetectBreakdowns(allTimeMemos)

	// 2. Build breakdown leaderboard using helper
	leaderboard := d.leaderboardBuilder.BuildFromBreakdowns(breakdowns)

	// 4. Check if we have any leaderboard data
	if len(leaderboard) == 0 {
		// Send a message indicating leaderboard building issue
		msg := &discordgo.MessageEmbed{
			Title:       "🏆 Breakdown Leaderboard",
			Description: "Found breakdown posts, but unable to build leaderboard due to author attribution issues. Please check memo author information and Discord account linking. 📝",
			Color:       0xFFA500, // Orange color for warning
		}
		return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
	}

	// 4. Build leaderboard content
	var content strings.Builder
	content.WriteString("Here are the top contributors for breakdown posts:\n\n")

	maxEntries := 10 // Show more entries for all-time leaderboard

	for i, entry := range leaderboard {
		if i >= maxEntries {
			break
		}

		// Format with Discord mentions when available
		var displayName string
		if entry.DiscordID != "" {
			displayName = fmt.Sprintf("<@%s>", entry.DiscordID)
		} else if entry.Username != "" {
			// Handle organization names like "Dwarves,Foundation"
			if strings.Contains(entry.Username, ",") {
				// Clean up organization names for better display
				displayName = strings.ReplaceAll(entry.Username, ",", " ")
			} else {
				displayName = formatDiscordMention(entry.Username)
			}
		} else {
			displayName = "Unknown"
		}

		content.WriteString(fmt.Sprintf("%d. %s x%d\n",
			entry.Rank, displayName, entry.BreakdownCount))
	}

	// Add ICY reward information and encouragement message
	// Disable ICY reward display temporarily
	// 3. Calculate ICY rewards (only for breakdowns)
	// totalICY := len(breakdowns) * 25
	totalICY := 0
	if totalICY > 0 {
		content.WriteString(fmt.Sprintf("\n💰 **Total ICY reward: %d ICY**\n", totalICY))
	}
	content.WriteString("\nGreat work on sharing knowledge! 🚀")

	// 5. Create and send Discord embed
	msg := &discordgo.MessageEmbed{
		Title:       "🏆 Breakdown Leaderboard",
		Description: content.String(),
		Color:       0xFFD700, // Gold color for leaderboard
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Keep writing breakdowns to climb the leaderboard!",
		},
	}

	return d.SendEmbeddedMessageWithChannel(nil, msg, channelID)
}

func (d *discordClient) SendDiscordMessageWithChannel(ses *discordgo.Session, msg *discordgo.Message, channelId string) error {
	_, err := ses.ChannelMessageSend(channelId, msg.Content)
	return err
}

func (d *discordClient) GetChannelMessages(channelID, before, after string, limit int) ([]*discordgo.Message, error) {
	return d.session.ChannelMessages(channelID, limit, before, after, "")
}

func (d *discordClient) GetEventByID(eventID string) (*discordgo.GuildScheduledEvent, error) {
	return d.session.GuildScheduledEvent(d.cfg.Discord.IDs.DwarvesGuild, eventID, false)
}

func (d *discordClient) ListActiveThreadsByChannelID(guildID, channelID string) ([]discordgo.Channel, error) {
	threadsList, err := d.session.GuildThreadsActive(guildID)
	if err != nil {
		return nil, err
	}

	result := make([]discordgo.Channel, 0)
	for _, thread := range threadsList.Threads {
		if thread.ParentID == channelID {
			result = append(result, *thread)
		}
	}

	return result, nil
}
