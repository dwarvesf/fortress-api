package discord

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type IController interface {
	Log(in model.LogDiscordInput) error
	PublicAdvanceSalaryLog(in model.LogDiscordInput) error
	ListDiscordResearchTopics(ctx context.Context, days, limit, offset int) ([]model.DiscordResearchTopic, int64, error)
	UserOgifStats(ctx context.Context, discordID string, after time.Time) (OgifStats, error)
	GetOgifLeaderboard(ctx context.Context, after time.Time, limit int) ([]model.OgifLeaderboardRecord, error)
	ListDiscordChannelMessageLogs(ctx context.Context, channelID string, startTime *time.Time, endTime *time.Time) ([]model.DiscordTextMessageLog, error)
}

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	config  *config.Config
	repo    store.DBRepo
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		service: service,
		logger:  logger,
		config:  cfg,
		repo:    repo,
	}
}

func (c *controller) Log(in model.LogDiscordInput) error {
	// Get discord template
	template, err := c.store.DiscordLogTemplate.GetTemplateByType(c.repo.DB(), in.Type)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Get Discord Template failed")
		return err
	}

	data := in.Data.(map[string]interface{})

	// get employee_id in discord format if any
	if employeeID, ok := data["employee_id"]; ok {
		employee, err := c.store.Employee.One(c.repo.DB(), employeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := employee.DisplayName
		if employee.DiscordAccount != nil && employee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", employee.DiscordAccount.DiscordID)
		}

		data["employee_id"] = accountID
	}

	if updatedEmployeeID, ok := data["updated_employee_id"]; ok {
		updatedEmployee, err := c.store.Employee.One(c.repo.DB(), updatedEmployeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := updatedEmployee.DisplayName
		if updatedEmployee.DiscordAccount != nil && updatedEmployee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", updatedEmployee.DiscordAccount.DiscordID)
		}

		data["updated_employee_id"] = accountID
	}

	// Replace template
	content := template.Content
	for k, v := range data {
		content = strings.ReplaceAll(content, fmt.Sprintf("{{ %s }}", k), fmt.Sprintf("%v", v))
	}

	// log discord
	_, err = c.service.Discord.SendMessage(model.DiscordMessage{
		Content: content,
	}, c.config.Discord.Webhooks.AuditLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}

func (c *controller) PublicAdvanceSalaryLog(in model.LogDiscordInput) error {
	data := in.Data.(map[string]interface{})

	icyAmount := data["icy_amount"]
	usdAmount := data["usd_amount"]

	desc := fmt.Sprintf("🧊 %v ICY (%v) has been sent to an anonymous peep as a salary advance.\n", icyAmount, usdAmount)
	desc += "\nFull-time peeps can use `?salary advance` to take a short-term credit benefit."

	embedMessage := model.DiscordMessageEmbed{
		Author:      model.DiscordMessageAuthor{},
		Title:       "💸 New ICY Payment 💸",
		URL:         "",
		Description: desc,
		Color:       3447003,
		Fields:      nil,
		Thumbnail:   model.DiscordMessageImage{},
		Image:       model.DiscordMessageImage{},
		Footer: model.DiscordMessageFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000+07:00"),
	}

	// log discord
	_, err := c.service.Discord.SendMessage(model.DiscordMessage{
		Embeds: []model.DiscordMessageEmbed{embedMessage},
	}, c.config.Discord.Webhooks.ICYPublicLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}

func (c *controller) ListDiscordResearchTopics(ctx context.Context, days, limit, offset int) ([]model.DiscordResearchTopic, int64, error) {
	topics, err := c.service.Discord.ListActiveThreadsByChannelID(c.config.Discord.IDs.DwarvesGuild, c.config.Discord.IDs.ResearchChannel)
	if err != nil {
		c.logger.Error(err, "Fetch list research topics failed")
		return nil, 0, err
	}

	type result struct {
		topic model.DiscordResearchTopic
		err   error
	}

	topicCh := make(chan string, len(topics))
	resultCh := make(chan result, len(topics))
	workers := 5

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for topicID := range topicCh {
				totalMsgCount, topUsers, lastActiveTime, err := c.topicPopularity(topicID, days)
				if err != nil {
					c.logger.Error(err, "Build research topic model failed")
					resultCh <- result{err: err}
					continue
				}
				if totalMsgCount == 0 {
					continue
				}

				resultCh <- result{
					topic: model.DiscordResearchTopic{
						Name:              topicID, // Assume topic.Name is topicID for this example
						URL:               fmt.Sprintf("https://discord.com/channels/%s/%s", c.config.Discord.IDs.DwarvesGuild, topicID),
						MsgCount:          totalMsgCount,
						SortedActiveUsers: topUsers,
						LastActiveTime:    lastActiveTime,
					},
				}
			}
		}()
	}

	for _, topic := range topics {
		topicCh <- topic.ID
	}
	close(topicCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	finalResults := make([]model.DiscordResearchTopic, 0)
	for res := range resultCh {
		if res.err != nil {
			return nil, 0, res.err
		}
		finalResults = append(finalResults, res.topic)
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].MsgCount > finalResults[j].MsgCount
	})

	total := int64(len(finalResults))

	// Apply pagination
	if int64(offset) >= total {
		return []model.DiscordResearchTopic{}, total, nil
	}
	end := offset + limit
	if end > len(finalResults) {
		end = len(finalResults)
	}

	finalResults = finalResults[offset:end]

	return finalResults, total, nil
}

func (c *controller) topicPopularity(topicID string, days int) (int64, []model.DiscordTopicActiveUser, time.Time, error) {
	var totalMessages int64
	var beforeID string
	var lastActiveTime time.Time

	lastNDays := time.Now().AddDate(0, 0, -days)
	userMessageCount := make(map[string]int64)
	limit := 100

	for {
		messages, err := c.service.Discord.GetChannelMessages(topicID, beforeID, "", limit)
		if err != nil {
			return 0, nil, time.Now(), err
		}
		if len(messages) == 0 {
			break
		}

		if beforeID == "" {
			lastActiveTime = messages[0].Timestamp
		}

		for _, msg := range messages {
			if days != 0 && msg.Timestamp.Before(lastNDays) {
				break
			}

			userMessageCount[msg.Author.ID]++
			totalMessages++
		}

		if len(messages) < limit {
			break
		}

		beforeID = messages[len(messages)-1].ID
	}

	var userCounts []model.DiscordTopicActiveUser
	for userID, count := range userMessageCount {
		userCounts = append(userCounts, model.DiscordTopicActiveUser{UserID: userID, MsgCount: count})
	}

	sort.Slice(userCounts, func(i, j int) bool {
		return userCounts[i].MsgCount > userCounts[j].MsgCount
	})

	var result []model.DiscordTopicActiveUser
	for i := 0; i < 3 && i < len(userCounts); i++ {
		result = append(result, userCounts[i])
	}

	return totalMessages, result, lastActiveTime, nil
}

// OgifStats contains list of ogif and some stats
type OgifStats struct {
	OgifList               []model.EventSpeaker `json:"ogifList"`
	UserAllTimeSpeaksCount int64                `json:"userAllTimeSpeaksCount"`
	UserAllTimeRank        int64                `json:"userAllTimeRank"`
	UserCurrentSpeaksCount int64                `json:"userCurrentSpeaksCount"`
	UserCurrentRank        int64                `json:"userCurrentRank"`
	TotalSpeakCount        int64                `json:"totalSpeakCount"`
	CurrentSpeakCount      int64                `json:"currentSpeakCount"`
}

// UserOgifStats returns list ogif with some stats
func (c *controller) UserOgifStats(ctx context.Context, discordID string, after time.Time) (OgifStats, error) {
	logger := c.logger.AddField("discordID", discordID).AddField("after", after)

	ogftList, err := c.store.EventSpeaker.List(c.repo.DB(), discordID, &after, "ogif")
	if err != nil {
		logger.Error(err, "error when retrieving list event speaker")
		return OgifStats{}, err
	}

	ogifStats, err := c.store.EventSpeaker.GetSpeakerStats(c.repo.DB(), discordID, &after, "ogif")
	if err != nil {
		logger.Error(err, "error when retrieving speaker stats")
	}

	allTimeOgifStats, err := c.store.EventSpeaker.GetSpeakerStats(c.repo.DB(), discordID, nil, "ogif")
	if err != nil {
		logger.Error(err, "error when retrieving all time speaker stats")
		return OgifStats{}, err
	}

	allTimeTotalCount, err := c.store.EventSpeaker.Count(c.repo.DB(), "", nil, "ogif")
	if err != nil {
		logger.Error(err, "error when counting all time total speak")
	}

	return OgifStats{
		OgifList:               ogftList,
		UserAllTimeSpeaksCount: allTimeOgifStats.TotalSpeakCount,
		UserAllTimeRank:        allTimeOgifStats.SpeakRank,
		UserCurrentSpeaksCount: ogifStats.TotalSpeakCount,
		UserCurrentRank:        ogifStats.SpeakRank,
		TotalSpeakCount:        allTimeTotalCount,
		CurrentSpeakCount:      int64(len(ogftList)),
	}, nil
}

// GetOgifLeaderboard returns the OGIF leaderboard
func (c *controller) GetOgifLeaderboard(ctx context.Context, after time.Time, limit int) ([]model.OgifLeaderboardRecord, error) {
	logger := c.logger.AddField("after", after).AddField("limit", limit)

	leaderboard, err := c.store.EventSpeaker.GetLeaderboard(c.repo.DB(), &after, limit, "ogif")
	if err != nil {
		logger.Error(err, "error when retrieving OGIF leaderboard")
		return nil, err
	}

	return leaderboard, nil
}

func (c *controller) ListDiscordChannelMessageLogs(ctx context.Context, channelID string, startTime *time.Time, endTime *time.Time) ([]model.DiscordTextMessageLog, error) {
	threads, err := c.service.Discord.ListActiveThreadsByChannelID(c.config.Discord.IDs.DwarvesGuild, channelID)
	if err != nil {
		return nil, err
	}
	channelIDs := make([]string, 0)
	channelIDs = append(channelIDs, channelID)
	for _, thread := range threads {
		channelIDs = append(channelIDs, thread.ID)
	}

	members, err := c.service.Discord.GetMembers()
	if err != nil {
		return nil, err
	}

	membersMap := make(map[string]string)
	for _, mem := range members {
		membersMap[fmt.Sprintf("<@%s>", mem.User.ID)] = fmt.Sprintf("@%s", mem.User.Username)
	}

	messageLogs := make([]model.DiscordTextMessageLog, 0)

	for _, channnelID := range channelIDs {
		messages, err := c.service.Discord.GetChannelMessagesInDateRange(channnelID, 100, startTime, endTime)
		if err != nil {
			return nil, err
		}

		for _, msg := range messages {
			messageLogs = append(messageLogs, model.DiscordTextMessageLog{
				ID:         msg.ID,
				Content:    msg.Content,
				AuthorName: msg.Author.Username,
				AuthorID:   msg.Author.ID,
				ChannelID:  msg.ChannelID,
				GuildID:    msg.GuildID,
				Timestamp:  msg.Timestamp,
			})
		}
	}

	for _, msg := range messageLogs {
		content := msg.Content
		for id, mem := range membersMap {
			if strings.Contains(content, id) {
				content = strings.ReplaceAll(content, id, mem)
			}
		}
	}

	sort.Slice(messageLogs, func(i, j int) bool {
		return messageLogs[i].Timestamp.Before(messageLogs[j].Timestamp)
	})
	return messageLogs, nil
}
