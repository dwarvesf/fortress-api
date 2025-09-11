package helpers

import (
	"sort"
	"sync"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type leaderboardBuilder struct {
	authorResolver AuthorResolver
	config         LeaderboardConfig
	mutex          sync.RWMutex
}

// NewLeaderboardBuilder creates a new leaderboard builder with the given author resolver and configuration.
// The author resolver is used to map usernames to Discord IDs for proper mention formatting.
// The configuration controls sorting behavior, entry limits, and display options.
func NewLeaderboardBuilder(authorResolver AuthorResolver, config LeaderboardConfig) LeaderboardBuilder {
	return &leaderboardBuilder{
		authorResolver: authorResolver,
		config:         config,
	}
}

// BuildFromBreakdowns creates a leaderboard specifically from breakdown posts.
// Each breakdown post contributes 1 point to the author's score. This is the primary
// method for generating weekly and monthly breakdown leaderboards.
func (b *leaderboardBuilder) BuildFromBreakdowns(breakdowns []model.MemoLog) []LeaderboardEntry {
	return b.BuildWithCustomScoring(breakdowns, func(memo model.MemoLog) int {
		return 1 // Each breakdown counts as 1 point
	})
}

// BuildFromAllPosts creates a leaderboard from all posts, treating each post equally.
// Each post contributes 1 point regardless of type. Useful for general activity rankings.
func (b *leaderboardBuilder) BuildFromAllPosts(memos []model.MemoLog) []LeaderboardEntry {
	return b.BuildWithCustomScoring(memos, func(memo model.MemoLog) int {
		return 1 // Each post counts as 1 point
	})
}

// BuildWithCustomScoring creates a leaderboard using a custom scoring function.
// The scoring function is applied to each memo to determine its contribution to the author's score.
// This provides maximum flexibility for different ranking algorithms and business logic.
func (b *leaderboardBuilder) BuildWithCustomScoring(memos []model.MemoLog, scoringFunc ScoringFunction) []LeaderboardEntry {
	b.mutex.RLock()
	config := b.config
	b.mutex.RUnlock()
	
	// Aggregate scores by author
	authorScores := make(map[string]*LeaderboardEntry)
	
	for _, memo := range memos {
		score := scoringFunc(memo)
		
		for _, username := range memo.AuthorMemoUsernames {
			if username == "" {
				continue // Skip empty usernames
			}
			
			if entry, exists := authorScores[username]; exists {
				entry.BreakdownCount += score
				entry.TotalPosts++
				entry.Score += score
			} else {
				authorScores[username] = &LeaderboardEntry{
					Username:       username,
					BreakdownCount: score,
					TotalPosts:     1,
					Score:          score,
				}
			}
		}
	}
	
	// Convert to slice and resolve Discord IDs
	entries := make([]LeaderboardEntry, 0, len(authorScores))
	usernames := make([]string, 0, len(authorScores))
	
	for username, entry := range authorScores {
		// Skip zero scores if configured
		if !config.ShowZeroScores && entry.Score == 0 {
			continue
		}
		
		entries = append(entries, *entry)
		usernames = append(usernames, username)
	}
	
	// Sort usernames for consistent ordering
	sort.Strings(usernames)
	
	// Resolve Discord IDs for all authors if we have any
	if len(usernames) > 0 && b.authorResolver != nil {
		discordIDs, err := b.authorResolver.ResolveAuthorsToDiscordIDs(usernames)
		if err != nil {
			// Log error but continue with usernames
			// (AuthorResolver handles fallback formatting)
		}
		
		// Create a mapping from username to Discord ID
		usernameToDiscordID := make(map[string]string)
		for i, username := range usernames {
			if i < len(discordIDs) {
				usernameToDiscordID[username] = discordIDs[i]
			}
		}
		
		// Update entries with Discord IDs
		for i := range entries {
			if discordID, exists := usernameToDiscordID[entries[i].Username]; exists {
				entries[i].DiscordID = discordID
			}
		}
	}
	
	// Sort entries
	b.sortEntries(entries, config)
	
	// Assign ranks and apply limit
	entries = b.assignRanksAndLimit(entries, config)
	
	return entries
}

// sortEntries sorts the leaderboard entries according to configuration settings.
// Primary sort is by score, with configurable tie-breaking using alphabetical order.
func (b *leaderboardBuilder) sortEntries(entries []LeaderboardEntry, config LeaderboardConfig) {
	sort.Slice(entries, func(i, j int) bool {
		// Primary sort by score
		if entries[i].Score != entries[j].Score {
			if config.SortDescending {
				return entries[i].Score > entries[j].Score
			}
			return entries[i].Score < entries[j].Score
		}
		
		// Tie breaking
		if config.TieBreaking.UseTimestamp {
			// Would need timestamp in LeaderboardEntry for this
			// For now, skip to alphabetical
		}
		
		if config.TieBreaking.UseAlphabetical {
			return entries[i].Username < entries[j].Username
		}
		
		return false // Maintain original order for ties
	})
}

// assignRanksAndLimit assigns rank numbers to entries and applies the maximum entries limit.
// Entries with the same score receive the same rank, with subsequent ranks adjusted appropriately.
func (b *leaderboardBuilder) assignRanksAndLimit(entries []LeaderboardEntry, config LeaderboardConfig) []LeaderboardEntry {
	// Assign ranks
	currentRank := 1
	for i := range entries {
		if i > 0 && entries[i].Score != entries[i-1].Score {
			currentRank = i + 1
		}
		entries[i].Rank = currentRank
	}
	
	// Apply limit
	if config.MaxEntries > 0 && len(entries) > config.MaxEntries {
		entries = entries[:config.MaxEntries]
	}
	
	return entries
}

// GetConfig returns the current leaderboard configuration.
// This method is thread-safe and returns a copy of the current settings.
func (b *leaderboardBuilder) GetConfig() LeaderboardConfig {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	return b.config
}

// UpdateConfig updates the leaderboard configuration with new settings.
// Changes take effect immediately for subsequent leaderboard generation.
// This method is thread-safe and can be called concurrently.
func (b *leaderboardBuilder) UpdateConfig(config LeaderboardConfig) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.config = config
}