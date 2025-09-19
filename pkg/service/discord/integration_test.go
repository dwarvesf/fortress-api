package discord

import (
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/discord/helpers"
)

// Integration tests verify that helpers work together with real implementations
func TestDiscordService_IntegrationWithRealHelpers(t *testing.T) {
	// Create Discord service with real helper implementations
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Create realistic test data
	testMemos := []model.MemoLog{
		{
			Title:               "Weekly Breakdown - Engineering Team Updates",
			URL:                "https://example.com/breakdown1",
			DiscordAccountIDs:   model.JSONArrayString{"discord123"},
			Tags:               model.JSONArrayString{"breakdown", "weekly", "engineering"},
			Category:           pq.StringArray{"00_fleeting"},
			PublishedAt:        &[]time.Time{time.Now()}[0],
			AuthorMemoUsernames: []string{"eng_user1"},
		},
		{
			Title:               "Literature Review - System Architecture",
			URL:                "https://example.com/literature1", 
			DiscordAccountIDs:   model.JSONArrayString{"discord456"},
			Tags:               model.JSONArrayString{"literature", "architecture"},
			Category:           pq.StringArray{"01_literature"},
			PublishedAt:        &[]time.Time{time.Now()}[0],
			AuthorMemoUsernames: []string{"arch_user2"},
		},
		{
			Title:               "Breakdown Summary: Q4 Performance Analysis",
			URL:                "https://example.com/breakdown2",
			DiscordAccountIDs:   model.JSONArrayString{"discord123"},
			Tags:               model.JSONArrayString{"breakdown", "analysis", "q4"},
			Category:           pq.StringArray{"earn"},
			PublishedAt:        &[]time.Time{time.Now()}[0],
			AuthorMemoUsernames: []string{"eng_user1"},
		},
		{
			Title:               "Random Discussion Notes",
			URL:                "https://example.com/notes1",
			DiscordAccountIDs:   model.JSONArrayString{"discord789"},
			Tags:               model.JSONArrayString{"discussion", "notes"},
			Category:           pq.StringArray{"others"},
			PublishedAt:        &[]time.Time{time.Now()}[0],
			AuthorMemoUsernames: []string{"discuss_user3"},
		},
	}
	
	// Test breakdown detection with real implementation
	breakdowns := service.breakdownDetector.DetectBreakdowns(testMemos)
	
	// Should detect 2 breakdown posts based on title keywords
	assert.Len(t, breakdowns, 2, "Should detect 2 breakdown posts")
	assert.Contains(t, breakdowns[0].Title, "Breakdown")
	assert.Contains(t, breakdowns[1].Title, "Breakdown")
	
	// Test leaderboard building with real implementation
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	
	// Should have entries for breakdown authors
	require.NotEmpty(t, leaderboard, "Leaderboard should not be empty")
	
	// Find the top performer (discord123 has 2 breakdowns)
	topEntry := leaderboard[0]
	assert.Equal(t, 1, topEntry.Rank, "Top entry should have rank 1")
	assert.Equal(t, 2, topEntry.BreakdownCount, "Top entry should have 2 breakdowns")
	
	// Test time calculator with real implementation  
	weekStart, weekEnd, weekRange := service.timeCalculator.CalculateWeeklyRange()
	assert.False(t, weekStart.IsZero(), "Week start should not be zero")
	assert.False(t, weekEnd.IsZero(), "Week end should not be zero")
	assert.True(t, weekEnd.After(weekStart), "Week end should be after week start")
	assert.NotEmpty(t, weekRange, "Week range string should not be empty")
	
	monthStart, monthEnd, monthRange := service.timeCalculator.CalculateMonthlyRange()
	assert.False(t, monthStart.IsZero(), "Month start should not be zero")
	assert.False(t, monthEnd.IsZero(), "Month end should not be zero")
	assert.True(t, monthEnd.After(monthStart), "Month end should be after month start")
	assert.NotEmpty(t, monthRange, "Month range string should not be empty")
}

func TestDiscordService_WeeklyReportIntegration(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Create test data with breakdown posts
	testMemos := []model.MemoLog{
		{
			Title:               "Weekly Breakdown - Development Progress",
			URL:                "https://example.com/weekly1",
			DiscordAccountIDs:   model.JSONArrayString{"user1"},
			Tags:               model.JSONArrayString{"breakdown", "weekly"},
			Category:           pq.StringArray{"00_fleeting"},
			AuthorMemoUsernames: []string{"dev_user1"},
		},
		{
			Title:               "Performance Breakdown Report",
			URL:                "https://example.com/performance1",
			DiscordAccountIDs:   model.JSONArrayString{"user2"}, 
			Tags:               model.JSONArrayString{"breakdown", "performance"},
			Category:           pq.StringArray{"earn"},
			AuthorMemoUsernames: []string{"perf_user2"},
		},
		{
			Title:               "Regular Literature Notes",
			URL:                "https://example.com/lit1",
			DiscordAccountIDs:   model.JSONArrayString{"user3"},
			Tags:               model.JSONArrayString{"literature"},
			Category:           pq.StringArray{"01_literature"},
			AuthorMemoUsernames: []string{"lit_user3"},
		},
	}
	
	// Test the integrated workflow
	breakdowns := service.breakdownDetector.DetectBreakdowns(testMemos)
	assert.Len(t, breakdowns, 2, "Should detect 2 breakdown posts")
	
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	assert.NotEmpty(t, leaderboard, "Should have leaderboard entries")
	
	// Verify that both breakdown authors are included
	authors := make(map[string]bool)
	for _, entry := range leaderboard {
		authors[entry.Username] = true
	}
	assert.Len(t, authors, 2, "Should have 2 unique authors in leaderboard")
}

func TestDiscordService_MonthlyReportIntegration(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Create test data with multiple breakdown posts for ICY calculation
	testMemos := []model.MemoLog{
		{
			Title:               "Monthly Breakdown - Team Performance",
			URL:                "https://example.com/monthly1",
			DiscordAccountIDs:   model.JSONArrayString{"superuser"},
			Tags:               model.JSONArrayString{"breakdown", "monthly"},
			Category:           pq.StringArray{"00_fleeting"},
			AuthorMemoUsernames: []string{"super_user"},
		},
		{
			Title:               "Breakdown: Engineering Metrics",
			URL:                "https://example.com/metrics1",
			DiscordAccountIDs:   model.JSONArrayString{"superuser"},
			Tags:               model.JSONArrayString{"breakdown", "metrics"},
			Category:           pq.StringArray{"earn"},
			AuthorMemoUsernames: []string{"super_user"},
		},
		{
			Title:               "Weekly Summary Breakdown",
			URL:                "https://example.com/summary1",
			DiscordAccountIDs:   model.JSONArrayString{"superuser"},
			Tags:               model.JSONArrayString{"breakdown", "summary"},
			Category:           pq.StringArray{"earn"},
			AuthorMemoUsernames: []string{"super_user"},
		},
		{
			Title:               "Regular Discussion",
			URL:                "https://example.com/discussion1",
			DiscordAccountIDs:   model.JSONArrayString{"user4"},
			Tags:               model.JSONArrayString{"discussion"},
			Category:           pq.StringArray{"others"},
			AuthorMemoUsernames: []string{"discuss_user4"},
		},
	}
	
	// Test monthly workflow with ICY calculation
	breakdowns := service.breakdownDetector.DetectBreakdowns(testMemos)
	assert.Len(t, breakdowns, 3, "Should detect 3 breakdown posts")
	
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	assert.NotEmpty(t, leaderboard, "Should have leaderboard entries")
	
	// Test ICY calculation
	totalICY := len(breakdowns) * 25
	assert.Equal(t, 75, totalICY, "Total ICY should be 75 (3 breakdowns * 25)")
	
	// Verify top performer has all 3 breakdowns
	topEntry := leaderboard[0]
	assert.Equal(t, 3, topEntry.BreakdownCount, "Top performer should have 3 breakdowns")
	
	// Calculate ICY for top performer
	topPerformerICY := topEntry.BreakdownCount * 25
	assert.Equal(t, 75, topPerformerICY, "Top performer should earn 75 ICY")
}

func TestDiscordService_HelperConfiguration(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Test breakdown detector configuration
	stats := service.breakdownDetector.GetDetectionStats()
	assert.NotNil(t, stats, "Detection stats should be available")
	
	// Test leaderboard builder configuration 
	lbConfig := service.leaderboardBuilder.GetConfig()
	assert.Equal(t, 10, lbConfig.MaxEntries, "Leaderboard should be limited to 10 entries")
	assert.True(t, lbConfig.SortDescending, "Leaderboard should sort descending")
	
	// Test message formatter configuration
	msgConfig := service.messageFormatter.GetConfig()
	assert.Equal(t, 1024, msgConfig.MaxFieldLength, "Max field length should be 1024")
	assert.Equal(t, 25, msgConfig.MaxEmbedFields, "Max embed fields should be 25")
	
	// Test time calculator timezone
	timezone := service.timeCalculator.GetTimeZone()
	assert.Equal(t, "UTC", timezone, "Timezone should be UTC")
}

func TestDiscordService_ErrorHandling(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Test with empty memo list
	emptyMemos := []model.MemoLog{}
	breakdowns := service.breakdownDetector.DetectBreakdowns(emptyMemos)
	assert.Empty(t, breakdowns, "Should handle empty memo list")
	
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	assert.Empty(t, leaderboard, "Should handle empty breakdown list")
	
	// Test with nil memo data
	var nilMemos []model.MemoLog
	nilBreakdowns := service.breakdownDetector.DetectBreakdowns(nilMemos)
	assert.Empty(t, nilBreakdowns, "Should handle nil memo list gracefully")
}

func TestDiscordService_ReportDataStructure(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Create comprehensive test data
	testMemos := []model.MemoLog{
		{
			Title:               "Breakdown: Team Analysis",
			URL:                "https://example.com/team-analysis",
			DiscordAccountIDs:   model.JSONArrayString{"analyst1"},
			Tags:               model.JSONArrayString{"breakdown", "analysis"},
			Category:           pq.StringArray{"00_fleeting"},
			AuthorMemoUsernames: []string{"analyst1"},
		},
		{
			Title:               "Technical Literature Review",
			URL:                "https://example.com/tech-review",
			DiscordAccountIDs:   model.JSONArrayString{"reviewer1"},
			Tags:               model.JSONArrayString{"literature", "technical"},
			Category:           pq.StringArray{"01_literature"},
			AuthorMemoUsernames: []string{"reviewer1"},
		},
	}
	
	// Test complete workflow integration
	breakdowns := service.breakdownDetector.DetectBreakdowns(testMemos)
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	_, _, weekRange := service.timeCalculator.CalculateWeeklyRange()
	
	// Create report data structure (simulating what would be passed to message formatter)
	reportData := helpers.ReportData{
		Type:        helpers.ReportTypeWeekly,
		TimeRange:   weekRange,
		AllPosts:    testMemos,
		Breakdowns:  breakdowns,
		NewAuthors:  []string{},
		Leaderboard: leaderboard,
		AuthorMappings: map[string]string{
			"analyst1":  "123456789",
			"reviewer1": "987654321",
		},
		GeneratedAt: time.Now(),
	}
	
	// Verify report data structure
	assert.Equal(t, helpers.ReportTypeWeekly, reportData.Type)
	assert.Equal(t, 2, len(reportData.AllPosts))
	assert.Equal(t, 1, len(reportData.Breakdowns))
	assert.NotEmpty(t, reportData.TimeRange)
	assert.False(t, reportData.GeneratedAt.IsZero())
	
	// Test monthly report data with ICY calculation
	monthStart, monthEnd, monthRange := service.timeCalculator.CalculateMonthlyRange()
	_ = monthStart
	_ = monthEnd
	
	monthlyReportData := helpers.ReportData{
		Type:           helpers.ReportTypeMonthly,
		TimeRange:      monthRange,
		AllPosts:       testMemos,
		Breakdowns:     breakdowns,
		Leaderboard:    leaderboard,
		TotalICY:       len(breakdowns) * 25,
		AuthorMappings: reportData.AuthorMappings,
		GeneratedAt:    time.Now(),
	}
	
	assert.Equal(t, helpers.ReportTypeMonthly, monthlyReportData.Type)
	assert.Equal(t, 25, monthlyReportData.TotalICY)
}

func TestDiscordService_PerformanceWithLargeDataset(t *testing.T) {
	cfg := &config.Config{
		Discord: config.Discord{
			SecretToken: "test-token",
		},
	}
	
	service := New(cfg).(*discordClient)
	
	// Create large dataset for performance testing
	largeMemoSet := make([]model.MemoLog, 1000)
	for i := 0; i < 1000; i++ {
		title := "Regular Post " + string(rune(i))
		if i%10 == 0 {
			title = "Breakdown: Analysis " + string(rune(i))
		}
		
		largeMemoSet[i] = model.MemoLog{
			Title:               title,
			URL:                "https://example.com/memo" + string(rune(i)),
			DiscordAccountIDs:   model.JSONArrayString{"user" + string(rune(i%50))}, // 50 unique users
			Tags:               model.JSONArrayString{"test", "performance"},
			Category:           pq.StringArray{"00_fleeting"},
			AuthorMemoUsernames: []string{"user" + string(rune(i%50))}, // 50 unique users
		}
	}
	
	// Test performance with large dataset
	startTime := time.Now()
	breakdowns := service.breakdownDetector.DetectBreakdowns(largeMemoSet)
	detectionDuration := time.Since(startTime)
	
	startTime = time.Now()
	leaderboard := service.leaderboardBuilder.BuildFromBreakdowns(breakdowns)
	leaderboardDuration := time.Since(startTime)
	
	// Performance assertions (should complete quickly)
	assert.True(t, detectionDuration < time.Second, "Breakdown detection should complete within 1 second")
	assert.True(t, leaderboardDuration < time.Second, "Leaderboard building should complete within 1 second")
	
	// Correctness assertions
	assert.Equal(t, 100, len(breakdowns), "Should detect 100 breakdown posts (every 10th post)")
	assert.NotEmpty(t, leaderboard, "Should have leaderboard entries")
	assert.True(t, len(leaderboard) <= 10, "Leaderboard should be limited to 10 entries")
}