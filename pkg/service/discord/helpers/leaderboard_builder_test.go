package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

// MockAuthorResolver for testing leaderboard builder
type MockAuthorResolver struct {
	mock.Mock
}

func (m *MockAuthorResolver) ResolveAuthorsToDiscordIDs(usernames []string) ([]string, error) {
	args := m.Called(usernames)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthorResolver) ResolveUsernameToDiscordID(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthorResolver) ResolveAllAuthors(memos []model.MemoLog) (map[string]string, error) {
	args := m.Called(memos)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockAuthorResolver) WarmCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthorResolver) ClearCache() {
	m.Called()
}

func (m *MockAuthorResolver) GetMetrics() AuthorResolutionMetrics {
	args := m.Called()
	return args.Get(0).(AuthorResolutionMetrics)
}

func TestLeaderboardBuilder_BuildFromBreakdowns(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	config := LeaderboardConfig{
		MaxEntries:     10,
		ShowZeroScores: false,
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, config)
	
	t.Run("basic_breakdown_leaderboard", func(t *testing.T) {
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},
			{AuthorMemoUsernames: []string{"bob"}},
			{AuthorMemoUsernames: []string{"alice"}}, // Alice has 2 breakdowns
			{AuthorMemoUsernames: []string{"charlie"}},
			{AuthorMemoUsernames: []string{"bob"}}, // Bob has 2 breakdowns
		}
		
		// Mock Discord ID resolution
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob", "charlie"}).Return(
			[]string{"111", "222", "333"}, nil)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		assert.Len(t, leaderboard, 3)
		
		// Verify ranking (should be sorted by score descending)
		assert.Equal(t, 1, leaderboard[0].Rank)
		assert.Equal(t, 1, leaderboard[1].Rank) // Tie for first place
		assert.Equal(t, 3, leaderboard[2].Rank) // Third place (after the tie)
		
		// Verify scores
		assert.Equal(t, 2, leaderboard[0].BreakdownCount)
		assert.Equal(t, 2, leaderboard[1].BreakdownCount)
		assert.Equal(t, 1, leaderboard[2].BreakdownCount)
		
		// Verify Discord IDs are resolved
		discordIDs := []string{leaderboard[0].DiscordID, leaderboard[1].DiscordID, leaderboard[2].DiscordID}
		assert.Contains(t, discordIDs, "111")
		assert.Contains(t, discordIDs, "222")
		assert.Contains(t, discordIDs, "333")
		
		mockResolver.AssertExpectations(t)
	})
	
	t.Run("multiple_authors_per_post", func(t *testing.T) {
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice", "bob"}}, // Both get credit
			{AuthorMemoUsernames: []string{"alice"}},        // Alice gets another
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob"}).Return(
			[]string{"111", "222"}, nil)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		assert.Len(t, leaderboard, 2)
		
		// Alice should have 2 breakdowns, Bob should have 1
		aliceEntry := findLeaderboardEntry(leaderboard, "alice")
		bobEntry := findLeaderboardEntry(leaderboard, "bob")
		
		assert.NotNil(t, aliceEntry)
		assert.NotNil(t, bobEntry)
		assert.Equal(t, 2, aliceEntry.BreakdownCount)
		assert.Equal(t, 1, bobEntry.BreakdownCount)
		assert.Equal(t, 1, aliceEntry.Rank)
		assert.Equal(t, 2, bobEntry.Rank)
		
		mockResolver.AssertExpectations(t)
	})
	
	t.Run("empty_breakdown_list", func(t *testing.T) {
		leaderboard := builder.BuildFromBreakdowns([]model.MemoLog{})
		assert.Empty(t, leaderboard)
	})
}

func TestLeaderboardBuilder_BuildFromAllPosts(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	config := LeaderboardConfig{
		MaxEntries:     5,
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, config)
	
	t.Run("all_posts_leaderboard", func(t *testing.T) {
		posts := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},
			{AuthorMemoUsernames: []string{"bob"}},
			{AuthorMemoUsernames: []string{"alice"}}, // Alice has 2 posts
			{AuthorMemoUsernames: []string{"charlie"}},
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob", "charlie"}).Return(
			[]string{"111", "222", "333"}, nil)
		
		leaderboard := builder.BuildFromAllPosts(posts)
		
		assert.Len(t, leaderboard, 3)
		
		// Alice should be first with 2 posts
		aliceEntry := findLeaderboardEntry(leaderboard, "alice")
		assert.NotNil(t, aliceEntry)
		assert.Equal(t, 2, aliceEntry.TotalPosts)
		assert.Equal(t, 1, aliceEntry.Rank)
		
		mockResolver.AssertExpectations(t)
	})
}

func TestLeaderboardBuilder_BuildWithCustomScoring(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	config := LeaderboardConfig{
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, config)
	
	t.Run("custom_scoring_function", func(t *testing.T) {
		posts := []model.MemoLog{
			{
				AuthorMemoUsernames: []string{"alice"},
				Title:               "Short post", // Score: 1
			},
			{
				AuthorMemoUsernames: []string{"bob"},
				Title:               "This is a longer post with more content", // Score: 2
			},
			{
				AuthorMemoUsernames: []string{"alice"},
				Title:               "Another longer post for Alice", // Score: 2
			},
		}
		
		// Custom scoring based on title length
		customScoringFunc := func(memo model.MemoLog) int {
			if len(memo.Title) > 20 {
				return 2 // Longer posts get more points
			}
			return 1 // Shorter posts get fewer points
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob"}).Return(
			[]string{"111", "222"}, nil)
		
		leaderboard := builder.BuildWithCustomScoring(posts, customScoringFunc)
		
		assert.Len(t, leaderboard, 2)
		
		// Alice should have score 3 (1 + 2), Bob should have score 2
		aliceEntry := findLeaderboardEntry(leaderboard, "alice")
		bobEntry := findLeaderboardEntry(leaderboard, "bob")
		
		assert.NotNil(t, aliceEntry)
		assert.NotNil(t, bobEntry)
		assert.Equal(t, 3, aliceEntry.Score)
		assert.Equal(t, 2, bobEntry.Score)
		assert.Equal(t, 1, aliceEntry.Rank)
		assert.Equal(t, 2, bobEntry.Rank)
		
		mockResolver.AssertExpectations(t)
	})
}

func TestLeaderboardBuilder_ConfigurationOptions(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	
	t.Run("max_entries_limit", func(t *testing.T) {
		config := LeaderboardConfig{
			MaxEntries:     2, // Limit to top 2
			SortDescending: true,
		}
		
		builder := NewLeaderboardBuilder(mockResolver, config)
		
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},   // 1 breakdown
			{AuthorMemoUsernames: []string{"bob"}},     // 1 breakdown  
			{AuthorMemoUsernames: []string{"charlie"}}, // 1 breakdown
			{AuthorMemoUsernames: []string{"alice"}},   // Alice: 2 total
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", mock.AnythingOfType("[]string")).Return(
			[]string{"111", "222", "333"}, nil)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		// Should only return top 2 entries due to MaxEntries limit
		assert.LessOrEqual(t, len(leaderboard), 2)
		
		mockResolver.AssertExpectations(t)
	})
	
	t.Run("hide_zero_scores", func(t *testing.T) {
		config := LeaderboardConfig{
			ShowZeroScores: false,
			SortDescending: true,
		}
		
		builder := NewLeaderboardBuilder(mockResolver, config)
		
		// Use custom scoring that gives some authors zero score
		posts := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},   // Will get score 1
			{AuthorMemoUsernames: []string{"bob"}},     // Will get score 0
			{AuthorMemoUsernames: []string{"charlie"}}, // Will get score 1
		}
		
		customScoringFunc := func(memo model.MemoLog) int {
			if memo.AuthorMemoUsernames[0] == "bob" {
				return 0 // Bob gets no points
			}
			return 1
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "charlie"}).Return(
			[]string{"111", "333"}, nil)
		
		leaderboard := builder.BuildWithCustomScoring(posts, customScoringFunc)
		
		// Should only include alice and charlie (not bob with 0 score)
		assert.Len(t, leaderboard, 2)
		for _, entry := range leaderboard {
			assert.Greater(t, entry.Score, 0)
		}
		
		mockResolver.AssertExpectations(t)
	})
	
	t.Run("alphabetical_tie_breaking", func(t *testing.T) {
		config := LeaderboardConfig{
			SortDescending: true,
			TieBreaking: struct {
				UseAlphabetical bool `json:"use_alphabetical"`
				UseTimestamp    bool `json:"use_timestamp"`
			}{
				UseAlphabetical: true,
			},
		}
		
		builder := NewLeaderboardBuilder(mockResolver, config)
		
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"zoe"}},   // Same score as alice
			{AuthorMemoUsernames: []string{"alice"}}, // Same score as zoe
		}
		
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "zoe"}).Return(
			[]string{"111", "999"}, nil)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		assert.Len(t, leaderboard, 2)
		
		// With alphabetical tie breaking, alice should come before zoe
		assert.Equal(t, "alice", leaderboard[0].Username)
		assert.Equal(t, "zoe", leaderboard[1].Username)
		
		// Both should have rank 1 since they're tied in score
		assert.Equal(t, 1, leaderboard[0].Rank)
		assert.Equal(t, 1, leaderboard[1].Rank)
		
		mockResolver.AssertExpectations(t)
	})
}

func TestLeaderboardBuilder_AuthorResolutionErrors(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	config := LeaderboardConfig{
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, config)
	
	t.Run("author_resolution_fails_gracefully", func(t *testing.T) {
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},
			{AuthorMemoUsernames: []string{"bob"}},
		}
		
		// Mock resolution to fail
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob"}).Return(
			[]string{}, assert.AnError)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		// Should still return leaderboard even if Discord ID resolution fails
		assert.Len(t, leaderboard, 2)
		
		// Verify usernames are still present
		aliceEntry := findLeaderboardEntry(leaderboard, "alice")
		bobEntry := findLeaderboardEntry(leaderboard, "bob")
		assert.NotNil(t, aliceEntry)
		assert.NotNil(t, bobEntry)
		
		mockResolver.AssertExpectations(t)
	})
}

func TestLeaderboardBuilder_GetAndUpdateConfig(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	initialConfig := LeaderboardConfig{
		MaxEntries:     5,
		ShowZeroScores: false,
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, initialConfig)
	
	t.Run("get_config", func(t *testing.T) {
		config := builder.GetConfig()
		assert.Equal(t, initialConfig, config)
	})
	
	t.Run("update_config", func(t *testing.T) {
		newConfig := LeaderboardConfig{
			MaxEntries:     10,
			ShowZeroScores: true,
			SortDescending: false,
		}
		
		builder.UpdateConfig(newConfig)
		updatedConfig := builder.GetConfig()
		assert.Equal(t, newConfig, updatedConfig)
	})
}

func TestLeaderboardBuilder_EmptyAuthorHandling(t *testing.T) {
	mockResolver := new(MockAuthorResolver)
	config := LeaderboardConfig{
		SortDescending: true,
	}
	
	builder := NewLeaderboardBuilder(mockResolver, config)
	
	t.Run("skip_empty_authors", func(t *testing.T) {
		breakdowns := []model.MemoLog{
			{AuthorMemoUsernames: []string{"alice"}},
			{AuthorMemoUsernames: []string{"", "bob"}}, // Empty username should be skipped
			{AuthorMemoUsernames: []string{}},          // Completely empty should be skipped
		}
		
		// Should only resolve alice and bob (empty authors filtered out)
		mockResolver.On("ResolveAuthorsToDiscordIDs", []string{"alice", "bob"}).Return(
			[]string{"111", "222"}, nil)
		
		leaderboard := builder.BuildFromBreakdowns(breakdowns)
		
		assert.Len(t, leaderboard, 2)
		
		aliceEntry := findLeaderboardEntry(leaderboard, "alice")
		bobEntry := findLeaderboardEntry(leaderboard, "bob")
		assert.NotNil(t, aliceEntry)
		assert.NotNil(t, bobEntry)
		
		mockResolver.AssertExpectations(t)
	})
}

// Helper function to find leaderboard entry by username
func findLeaderboardEntry(leaderboard []LeaderboardEntry, username string) *LeaderboardEntry {
	for _, entry := range leaderboard {
		if entry.Username == username {
			return &entry
		}
	}
	return nil
}