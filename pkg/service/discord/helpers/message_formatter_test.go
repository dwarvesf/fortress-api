package helpers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestMessageFormatter_FormatWeeklyReport(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("basic_weekly_report", func(t *testing.T) {
		data := ReportData{
			Type:      ReportTypeWeekly,
			TimeRange: "15-21 August",
			AllPosts: []model.MemoLog{
				{Title: "Post 1", AuthorMemoUsernames: []string{"alice"}},
				{Title: "Post 2", AuthorMemoUsernames: []string{"bob"}},
				{Title: "Post 3", AuthorMemoUsernames: []string{"alice"}},
			},
			Breakdowns: []model.MemoLog{
				{Title: "Breakdown Analysis", AuthorMemoUsernames: []string{"alice"}},
			},
			NewAuthors: []string{"charlie"},
			Leaderboard: []LeaderboardEntry{
				{Username: "alice", DiscordID: "111", BreakdownCount: 1, Rank: 1},
			},
			AuthorMappings: map[string]string{
				"alice":   "111",
				"bob":     "222",
				"charlie": "333",
			},
			GeneratedAt: time.Date(2024, 8, 21, 12, 0, 0, 0, time.UTC),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Verify title contains week range
		assert.Contains(t, embed.Title, "WEEKLY MEMO REPORT")
		assert.Contains(t, embed.Title, "15-21 August")
		
		// Verify description contains overview
		assert.Contains(t, embed.Description, "Total posts")
		assert.Contains(t, embed.Description, "3")
		assert.Contains(t, embed.Description, "Breakdowns")
		assert.Contains(t, embed.Description, "1")
		
		// Verify fields are present
		assert.True(t, len(embed.Fields) > 0)
		
		// Check for leaderboard field
		leaderboardFound := false
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "Leaderboard") {
				leaderboardFound = true
				assert.Contains(t, field.Value, "<@111>") // Discord mention for alice
				break
			}
		}
		assert.True(t, leaderboardFound, "Leaderboard field should be present")
		
		// Check for new authors field
		newAuthorsFound := false
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "New Contributors") {
				newAuthorsFound = true
				assert.Contains(t, field.Value, "<@333>") // Discord mention for charlie
				break
			}
		}
		assert.True(t, newAuthorsFound, "New contributors field should be present")
	})
	
	t.Run("empty_weekly_report", func(t *testing.T) {
		data := ReportData{
			Type:           ReportTypeWeekly,
			TimeRange:      "22-28 August", 
			AllPosts:       []model.MemoLog{},
			Breakdowns:     []model.MemoLog{},
			NewAuthors:     []string{},
			Leaderboard:    []LeaderboardEntry{},
			AuthorMappings: map[string]string{},
			GeneratedAt:    time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		assert.Contains(t, embed.Description, "No posts")
	})
	
	t.Run("truncated_content", func(t *testing.T) {
		// Create very long leaderboard to test truncation
		longLeaderboard := make([]LeaderboardEntry, 100)
		for i := 0; i < 100; i++ {
			longLeaderboard[i] = LeaderboardEntry{
				Username:       "user" + strings.Repeat("x", 10),
				DiscordID:      "123456789",
				BreakdownCount: 1,
				Rank:           i + 1,
			}
		}
		
		data := ReportData{
			Type:        ReportTypeWeekly,
			TimeRange:   "15-21 August",
			AllPosts:    make([]model.MemoLog, 10),
			Breakdowns:  make([]model.MemoLog, 5),
			Leaderboard: longLeaderboard,
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Check that content is truncated
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "Leaderboard") {
				assert.True(t, len(field.Value) <= config.MaxFieldLength,
					"Field value should not exceed max length")
				if len(field.Value) == config.MaxFieldLength {
					assert.Contains(t, field.Value, config.TruncationSuffix)
				}
				break
			}
		}
	})
}

func TestMessageFormatter_FormatMonthlyReport(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("basic_monthly_report", func(t *testing.T) {
		data := ReportData{
			Type:      ReportTypeMonthly,
			TimeRange: "August 2024",
			AllPosts: []model.MemoLog{
				{Title: "Post 1", AuthorMemoUsernames: []string{"alice"}},
				{Title: "Post 2", AuthorMemoUsernames: []string{"bob"}},
				{Title: "Post 3", AuthorMemoUsernames: []string{"alice"}},
			},
			Breakdowns: []model.MemoLog{
				{Title: "Breakdown 1", AuthorMemoUsernames: []string{"alice"}},
				{Title: "Breakdown 2", AuthorMemoUsernames: []string{"bob"}},
			},
			NewAuthors: []string{"charlie", "dave"},
			Leaderboard: []LeaderboardEntry{
				{Username: "alice", DiscordID: "111", BreakdownCount: 1, Rank: 1},
				{Username: "bob", DiscordID: "222", BreakdownCount: 1, Rank: 1},
			},
			TotalICY: 50, // 2 breakdowns × 25 ICY each
			AuthorMappings: map[string]string{
				"alice":   "111",
				"bob":     "222", 
				"charlie": "333",
				"dave":    "444",
			},
			GeneratedAt: time.Date(2024, 8, 31, 23, 59, 0, 0, time.UTC),
		}
		
		embed, err := formatter.FormatMonthlyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Verify title contains month and year (case insensitive)
		assert.Contains(t, embed.Title, "MONTHLY MEMO REPORT")
		assert.Contains(t, strings.ToUpper(embed.Title), "AUGUST 2024")
		
		// Verify description contains overview with ICY
		assert.Contains(t, embed.Description, "Total posts")
		assert.Contains(t, embed.Description, "3")
		assert.Contains(t, embed.Description, "Breakdowns")
		assert.Contains(t, embed.Description, "2")
		assert.Contains(t, embed.Description, "ICY")
		assert.Contains(t, embed.Description, "50")
		
		// Verify fields are present
		assert.True(t, len(embed.Fields) > 0)
		
		// Check for ICY distribution field
		icyFound := false
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "ICY") || strings.Contains(field.Value, "50") {
				icyFound = true
				break
			}
		}
		assert.True(t, icyFound, "ICY information should be present")
	})
	
	t.Run("zero_icy_monthly_report", func(t *testing.T) {
		data := ReportData{
			Type:        ReportTypeMonthly,
			TimeRange:   "September 2024",
			AllPosts:    []model.MemoLog{{Title: "Regular post"}},
			Breakdowns:  []model.MemoLog{}, // No breakdowns = 0 ICY
			TotalICY:    0,
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatMonthlyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Should still show ICY info even if zero
		assert.Contains(t, embed.Description, "ICY")
		assert.Contains(t, embed.Description, "0")
	})
}

func TestMessageFormatter_FieldTruncation(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   50, // Very short for testing
		TruncationSuffix: "...",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("truncation_in_real_report", func(t *testing.T) {
		// Create report with very long content that will be truncated
		longLeaderboard := make([]LeaderboardEntry, 50)
		for i := 0; i < 50; i++ {
			longLeaderboard[i] = LeaderboardEntry{
				Username:       fmt.Sprintf("very_long_username_%d", i),
				DiscordID:      fmt.Sprintf("12345678%d", i),
				BreakdownCount: 1,
				Rank:           i + 1,
			}
		}
		
		data := ReportData{
			Type:        ReportTypeWeekly,
			TimeRange:   "15-21 August",
			Leaderboard: longLeaderboard,
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Check that some field was truncated due to short limit
		for _, field := range embed.Fields {
			if len(field.Value) > 0 {
				assert.LessOrEqual(t, len(field.Value), config.MaxFieldLength)
			}
		}
	})
}

func TestMessageFormatter_ConfigManagement(t *testing.T) {
	initialConfig := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(initialConfig)
	
	t.Run("get_config", func(t *testing.T) {
		config := formatter.GetConfig()
		assert.Equal(t, initialConfig, config)
	})
	
	t.Run("update_config", func(t *testing.T) {
		newConfig := MessageFormattingConfig{
			MaxFieldLength:   2048,
			MaxEmbedFields:   30,
			TruncationSuffix: " [truncated]",
			DateFormat:       "Jan 2",
		}
		
		formatter.UpdateConfig(newConfig)
		updatedConfig := formatter.GetConfig()
		assert.Equal(t, newConfig, updatedConfig)
	})
}

func TestMessageFormatter_ErrorHandling(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("invalid_report_data", func(t *testing.T) {
		// Test with nil time
		data := ReportData{
			Type:      ReportTypeWeekly,
			TimeRange: "",
			// Missing required fields
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		// Should handle gracefully
		assert.NoError(t, err)
		assert.NotNil(t, embed)
	})
}

func TestMessageFormatter_AuthorMentions(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("author_mentions_in_report", func(t *testing.T) {
		data := ReportData{
			Type:      ReportTypeWeekly,
			TimeRange: "15-21 August",
			NewAuthors: []string{"alice", "bob", "unknown_user"},
			AuthorMappings: map[string]string{
				"alice": "111",
				"bob":   "222",
				// unknown_user has no mapping
			},
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Find the new contributors field
		found := false
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "New Contributors") {
				found = true
				assert.Contains(t, field.Value, "<@111>")  // Alice mention
				assert.Contains(t, field.Value, "<@222>")  // Bob mention
				assert.Contains(t, field.Value, "@unknown_user") // Fallback
				break
			}
		}
		assert.True(t, found, "New contributors field should be present")
	})
}

func TestMessageFormatter_LeaderboardFormatting(t *testing.T) {
	config := MessageFormattingConfig{
		MaxFieldLength:   1024,
		MaxEmbedFields:   25,
		TruncationSuffix: "...",
		DateFormat:       "2-Jan",
	}
	
	formatter := NewMessageFormatter(config)
	
	t.Run("leaderboard_in_report", func(t *testing.T) {
		data := ReportData{
			Type:      ReportTypeWeekly,
			TimeRange: "15-21 August",
			Leaderboard: []LeaderboardEntry{
				{Username: "alice", DiscordID: "111", BreakdownCount: 3, Rank: 1},
				{Username: "bob", DiscordID: "222", BreakdownCount: 2, Rank: 2},
				{Username: "charlie", DiscordID: "333", BreakdownCount: 2, Rank: 2}, // Tied for 2nd
				{Username: "dave", DiscordID: "444", BreakdownCount: 1, Rank: 4},
			},
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Find the leaderboard field
		found := false
		for _, field := range embed.Fields {
			if strings.Contains(field.Name, "Leaderboard") {
				found = true
				assert.Contains(t, field.Value, "🥇") // First place emoji
				assert.Contains(t, field.Value, "🥈") // Second place emoji
				assert.Contains(t, field.Value, "<@111>") // Alice mention
				assert.Contains(t, field.Value, "<@222>") // Bob mention
				assert.Contains(t, field.Value, "x3")     // Alice's count
				break
			}
		}
		assert.True(t, found, "Leaderboard field should be present")
	})
	
	t.Run("empty_leaderboard_report", func(t *testing.T) {
		data := ReportData{
			Type:        ReportTypeWeekly,
			TimeRange:   "15-21 August",
			Leaderboard: []LeaderboardEntry{}, // Empty leaderboard
			GeneratedAt: time.Now(),
		}
		
		embed, err := formatter.FormatWeeklyReport(data)
		
		assert.NoError(t, err)
		assert.NotNil(t, embed)
		
		// Should not have a leaderboard field when empty
		for _, field := range embed.Fields {
			assert.NotContains(t, field.Name, "Leaderboard")
		}
	})
}