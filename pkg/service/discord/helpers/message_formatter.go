package helpers

import (
	"fmt"
	"strings"
	"sync"
)

type messageFormatter struct {
	config MessageFormattingConfig
	mutex  sync.RWMutex
}

// NewMessageFormatter creates a new message formatter with the given configuration.
// The formatter handles Discord embed generation with proper field limits, text truncation,
// and consistent styling for both weekly and monthly report types.
func NewMessageFormatter(config MessageFormattingConfig) MessageFormatter {
	return &messageFormatter{
		config: config,
	}
}

// FormatWeeklyReport formats weekly report data into a Discord embed structure.
// Includes post counts, breakdown statistics, leaderboard, and new author highlights.
// Handles field truncation and Discord's embed limits automatically.
func (f *messageFormatter) FormatWeeklyReport(data ReportData) (*DiscordEmbed, error) {
	f.mutex.RLock()
	config := f.config
	f.mutex.RUnlock()
	
	// Build title
	title := fmt.Sprintf("🏆 WEEKLY MEMO REPORT (%s) 🏆", data.TimeRange)
	
	// Build description with overview
	var description strings.Builder
	description.WriteString("*What's happening with our memos this week?*\n\n")
	description.WriteString("**OVERVIEW**\n")
	
	if len(data.AllPosts) == 0 {
		description.WriteString("No posts published this week.")
	} else {
		description.WriteString(fmt.Sprintf("📝 `Total posts:` **%d** posts\n", len(data.AllPosts)))
		description.WriteString(fmt.Sprintf("⚡ `Breakdowns:` **%d** posts\n", len(data.Breakdowns)))
		if len(data.NewAuthors) > 0 {
			description.WriteString(fmt.Sprintf("👋 `New authors:` **%d** contributors", len(data.NewAuthors)))
		}
	}
	
	embed := &DiscordEmbed{
		Title:       title,
		Description: description.String(),
		Fields:      []DiscordEmbedField{},
		Footer: &DiscordEmbedFooter{
			Text:    "📊 Weekly Memo Analytics",
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
		},
		Timestamp: &data.GeneratedAt,
	}
	
	// Add leaderboard field if we have breakdowns
	if len(data.Leaderboard) > 0 {
		leaderboardText := f.formatLeaderboard(data.Leaderboard)
		if len(leaderboardText) > 0 {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   "🏅 Breakdown Leaderboard",
				Value:  f.truncateText(leaderboardText, config.MaxFieldLength),
				Inline: false,
			})
		}
	}
	
	// Add new authors field
	if len(data.NewAuthors) > 0 {
		newAuthorsText := f.formatAuthorMentions(data.NewAuthors, data.AuthorMappings)
		if len(newAuthorsText) > 0 {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   "✨ New Contributors",
				Value:  f.truncateText(newAuthorsText, config.MaxFieldLength),
				Inline: false,
			})
		}
	}
	
	// Apply field limit
	if len(embed.Fields) > config.MaxEmbedFields {
		embed.Fields = embed.Fields[:config.MaxEmbedFields]
	}
	
	return embed, nil
}

// FormatMonthlyReport formats monthly report data into a Discord embed structure.
// Includes extended analytics with ICY calculations, comprehensive leaderboards,
// and enhanced contributor recognition for the monthly period.
func (f *messageFormatter) FormatMonthlyReport(data ReportData) (*DiscordEmbed, error) {
	f.mutex.RLock()
	config := f.config
	f.mutex.RUnlock()
	
	// Build title
	title := fmt.Sprintf("📅 MONTHLY MEMO REPORT - %s 📅", strings.ToUpper(data.TimeRange))
	
	// Build description with overview including ICY
	var description strings.Builder
	description.WriteString("*Monthly summary of memo activities and contributions*\n\n")
	description.WriteString("**MONTHLY OVERVIEW**\n")
	
	if len(data.AllPosts) == 0 {
		description.WriteString("No posts published this month.")
	} else {
		description.WriteString(fmt.Sprintf("📝 `Total posts:` **%d** posts\n", len(data.AllPosts)))
		description.WriteString(fmt.Sprintf("⚡ `Breakdowns:` **%d** posts\n", len(data.Breakdowns)))
		description.WriteString(fmt.Sprintf("🧊 `Total ICY earned:` **%d** ICY\n", data.TotalICY))
		if len(data.NewAuthors) > 0 {
			description.WriteString(fmt.Sprintf("👋 `New authors:` **%d** contributors", len(data.NewAuthors)))
		}
	}
	
	embed := &DiscordEmbed{
		Title:       title,
		Description: description.String(),
		Fields:      []DiscordEmbedField{},
		Footer: &DiscordEmbedFooter{
			Text:    "📈 Monthly Memo Analytics",
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
		},
		Timestamp: &data.GeneratedAt,
	}
	
	// Add leaderboard field
	if len(data.Leaderboard) > 0 {
		leaderboardText := f.formatLeaderboard(data.Leaderboard)
		if len(leaderboardText) > 0 {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   "🏆 Monthly Breakdown Leaders",
				Value:  f.truncateText(leaderboardText, config.MaxFieldLength),
				Inline: false,
			})
		}
	}
	
	// Add ICY distribution field if there are breakdowns
	if data.TotalICY > 0 {
		icyText := f.formatICYDistribution(data.Leaderboard, data.TotalICY)
		if len(icyText) > 0 {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   "💰 ICY Distribution",
				Value:  f.truncateText(icyText, config.MaxFieldLength),
				Inline: true,
			})
		}
	}
	
	// Add new authors field
	if len(data.NewAuthors) > 0 {
		newAuthorsText := f.formatAuthorMentions(data.NewAuthors, data.AuthorMappings)
		if len(newAuthorsText) > 0 {
			embed.Fields = append(embed.Fields, DiscordEmbedField{
				Name:   "🌟 New Contributors This Month",
				Value:  f.truncateText(newAuthorsText, config.MaxFieldLength),
				Inline: false,
			})
		}
	}
	
	// Apply field limit
	if len(embed.Fields) > config.MaxEmbedFields {
		embed.Fields = embed.Fields[:config.MaxEmbedFields]
	}
	
	return embed, nil
}

// formatLeaderboard converts leaderboard entries into a formatted string for Discord embeds.
// Uses rank emojis for top 3 positions and handles ties appropriately.
func (f *messageFormatter) formatLeaderboard(leaderboard []LeaderboardEntry) string {
	if len(leaderboard) == 0 {
		return "No breakdown authors this period."
	}
	
	var result strings.Builder
	
	// Rank emojis for top positions
	rankEmojis := map[int]string{
		1: "🥇",
		2: "🥈",
		3: "🥉",
	}
	
	currentRank := 0
	for _, entry := range leaderboard {
		if entry.Rank != currentRank {
			currentRank = entry.Rank
			
			// Add rank emoji or number
			if emoji, exists := rankEmojis[currentRank]; exists {
				result.WriteString(fmt.Sprintf("\n%s ", emoji))
			} else {
				result.WriteString(fmt.Sprintf("\n`%d.` ", currentRank))
			}
		} else {
			// Continue same rank (tied)
			result.WriteString(" ")
		}
		
		// Add Discord mention
		if entry.DiscordID != "" {
			result.WriteString(fmt.Sprintf("<@%s>", entry.DiscordID))
		} else {
			result.WriteString(fmt.Sprintf("@%s", entry.Username))
		}
		
		// Add breakdown count
		result.WriteString(fmt.Sprintf(" (x%d)", entry.BreakdownCount))
	}
	
	return strings.TrimSpace(result.String())
}

// formatAuthorMentions converts author usernames into Discord mentions where possible.
// Falls back to @username format when Discord ID mapping is unavailable.
func (f *messageFormatter) formatAuthorMentions(authors []string, authorMappings map[string]string) string {
	if len(authors) == 0 {
		return ""
	}
	
	var mentions []string
	for _, author := range authors {
		if discordID, exists := authorMappings[author]; exists && discordID != "" {
			mentions = append(mentions, fmt.Sprintf("<@%s>", discordID))
		} else {
			mentions = append(mentions, fmt.Sprintf("@%s", author))
		}
	}
	
	return strings.Join(mentions, " ")
}

// formatICYDistribution creates a formatted display of ICY token distribution.
// Shows total ICY earned and breakdown by top contributors for monthly reports.
func (f *messageFormatter) formatICYDistribution(leaderboard []LeaderboardEntry, totalICY int) string {
	if totalICY == 0 {
		return "No ICY distributed this month."
	}
	
	var result strings.Builder
	result.WriteString(fmt.Sprintf("**Total: %d ICY** 🧊\n", totalICY))
	result.WriteString("*Each breakdown = 25 ICY*\n\n")
	
	// Show top contributors
	topCount := 0
	for _, entry := range leaderboard {
		if topCount >= 5 { // Limit to top 5 for brevity
			break
		}
		
		icyEarned := entry.BreakdownCount * 25
		if icyEarned > 0 {
			result.WriteString(fmt.Sprintf("• %d ICY - ", icyEarned))
			if entry.DiscordID != "" {
				result.WriteString(fmt.Sprintf("<@%s>", entry.DiscordID))
			} else {
				result.WriteString(fmt.Sprintf("@%s", entry.Username))
			}
			result.WriteString("\n")
			topCount++
		}
	}
	
	return result.String()
}

// truncateText ensures text fits within Discord's field length limits.
// Adds configured truncation suffix when text needs to be shortened.
func (f *messageFormatter) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	f.mutex.RLock()
	suffix := f.config.TruncationSuffix
	f.mutex.RUnlock()
	
	if len(suffix) >= maxLength {
		return text[:maxLength] // Can't fit suffix, just truncate
	}
	
	truncateAt := maxLength - len(suffix)
	return text[:truncateAt] + suffix
}

// GetConfig returns the current message formatting configuration.
// This method is thread-safe and provides access to current settings.
func (f *messageFormatter) GetConfig() MessageFormattingConfig {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	
	return f.config
}

// UpdateConfig updates the formatter configuration with new settings.
// Changes take effect immediately for subsequent format operations.
// This method is thread-safe and can be called concurrently.
func (f *messageFormatter) UpdateConfig(config MessageFormattingConfig) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	f.config = config
}