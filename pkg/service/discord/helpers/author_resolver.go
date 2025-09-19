package helpers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// authorCacheItem represents an item in the resolution cache
type authorCacheItem struct {
	entry     AuthorCacheEntry
	expiresAt time.Time
}

// authorResolver implements the AuthorResolver interface with caching
type authorResolver struct {
	mu          sync.RWMutex
	config      AuthorResolutionConfig
	cache       map[string]authorCacheItem
	store       *store.Store
	repo        store.DBRepo
	metrics     AuthorResolutionMetrics
	lastWarmup  time.Time
	evictions   int64
}

// NewAuthorResolver creates a new AuthorResolver with caching capabilities
func NewAuthorResolver(config AuthorResolutionConfig, store *store.Store, repo store.DBRepo) AuthorResolver {
	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour // Default 1 hour cache
	}
	if config.MaxCacheSize == 0 {
		config.MaxCacheSize = 1000 // Default max cache size
	}
	if config.BatchSize == 0 {
		config.BatchSize = 50 // Default batch size
	}
	if config.DatabaseTimeout == 0 {
		config.DatabaseTimeout = 10 * time.Second // Default timeout
	}
	if config.FallbackStrategy == "" {
		config.FallbackStrategy = "mention_attempt" // Default fallback
	}

	resolver := &authorResolver{
		config: config,
		cache:  make(map[string]authorCacheItem),
		store:  store,
		repo:   repo,
		metrics: AuthorResolutionMetrics{
			CacheHitRate:        0.0,
			ResolutionErrorRate: 0.0,
			CacheSize:          0,
			DatabaseQueryCount:  0,
			CacheEvictions:     0,
			LastWarmupTime:     time.Time{},
		},
	}

	// Warm up cache if enabled
	if config.WarmupEnabled {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), config.DatabaseTimeout)
			defer cancel()
			_ = resolver.WarmCache(ctx)
		}()
	}

	return resolver
}

// ResolveAuthorsToDiscordIDs resolves multiple usernames to Discord IDs in batch
func (a *authorResolver) ResolveAuthorsToDiscordIDs(usernames []string) ([]string, error) {
	if len(usernames) == 0 {
		return []string{}, nil
	}

	var results []string
	var uncachedUsernames []string

	// First pass: check cache for all usernames
	a.mu.RLock()
	for _, username := range usernames {
		if cached, exists := a.getCachedEntry(username); exists {
			results = append(results, cached.DiscordID)
			a.metrics.DatabaseQueryCount-- // Adjust since we're not querying
		} else {
			uncachedUsernames = append(uncachedUsernames, username)
			results = append(results, "") // Placeholder
		}
	}
	a.mu.RUnlock()

	// Second pass: resolve uncached usernames
	if len(uncachedUsernames) > 0 {
		resolvedMap, err := a.resolveUsernamesFromDatabase(uncachedUsernames)
		if err != nil {
			a.updateErrorRate(true)
			return results, fmt.Errorf("failed to resolve usernames from database: %w", err)
		}

		// Update results with resolved values
		for i, username := range usernames {
			if results[i] == "" { // This was uncached
				if discordID, found := resolvedMap[username]; found {
					results[i] = discordID
				} else {
					results[i] = a.getFallbackValue(username)
				}
			}
		}
	}

	a.updateHitRate(len(usernames)-len(uncachedUsernames), len(usernames))
	a.updateErrorRate(false)

	return results, nil
}

// ResolveUsernameToDiscordID resolves a single username to Discord ID with detailed error information
func (a *authorResolver) ResolveUsernameToDiscordID(username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	// Check cache first
	a.mu.RLock()
	if cached, exists := a.getCachedEntry(username); exists {
		a.mu.RUnlock()
		a.updateHitRate(1, 1)
		if !cached.IsValid {
			return "", fmt.Errorf("cached resolution failed: %s", cached.ErrorMessage)
		}
		return cached.DiscordID, nil
	}
	a.mu.RUnlock()

	// Resolve from database
	discordAccounts, err := a.store.DiscordAccount.ListByMemoUsername(a.repo.DB(), []string{username})
	a.metrics.DatabaseQueryCount++

	if err != nil {
		a.updateErrorRate(true)
		// Cache the error
		a.cacheEntry(username, AuthorCacheEntry{
			DiscordID:    "",
			ResolvedAt:   time.Now(),
			IsValid:      false,
			ErrorMessage: err.Error(),
		})
		return "", fmt.Errorf("database query failed for username '%s': %w", username, err)
	}

	// Process results
	if len(discordAccounts) == 0 {
		fallbackValue := a.getFallbackValue(username)
		// Cache the fallback result
		a.cacheEntry(username, AuthorCacheEntry{
			DiscordID:    fallbackValue,
			ResolvedAt:   time.Now(),
			IsValid:      true,
			ErrorMessage: "",
		})
		if a.config.LogMissingAuthors {
			fmt.Printf("No Discord account found for username: %s, using fallback: %s\n", username, fallbackValue)
		}
		a.updateHitRate(0, 1)
		return fallbackValue, nil
	}

	// Found exactly one match or multiple matches
	discordID := discordAccounts[0].DiscordID
	a.cacheEntry(username, AuthorCacheEntry{
		DiscordID:  discordID,
		ResolvedAt: time.Now(),
		IsValid:    true,
	})

	a.updateHitRate(0, 1)
	a.updateErrorRate(false)

	return discordID, nil
}

// ResolveAllAuthors resolves all unique authors from a list of memos
func (a *authorResolver) ResolveAllAuthors(memos []model.MemoLog) (map[string]string, error) {
	// Extract unique usernames
	usernameSet := make(map[string]bool)
	for _, memo := range memos {
		for _, author := range memo.AuthorMemoUsernames {
			if author != "" {
				usernameSet[strings.TrimSpace(author)] = true
			}
		}
	}

	// Convert to slice
	var usernames []string
	for username := range usernameSet {
		usernames = append(usernames, username)
	}

	if len(usernames) == 0 {
		return make(map[string]string), nil
	}

	// Resolve all usernames
	discordIDs, err := a.ResolveAuthorsToDiscordIDs(usernames)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve authors: %w", err)
	}

	// Build result map
	result := make(map[string]string)
	for i, username := range usernames {
		if i < len(discordIDs) {
			result[username] = discordIDs[i]
		}
	}

	return result, nil
}

// WarmCache pre-populates cache with known active authors from recent memos
func (a *authorResolver) WarmCache(ctx context.Context) error {
	// Get recent memo authors (last 30 days)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	memos, err := a.store.MemoLog.GetLimitByTimeRange(a.repo.DB(), &startTime, &endTime, 1000)
	if err != nil {
		return fmt.Errorf("failed to get recent memos for cache warmup: %w", err)
	}

	// Extract unique authors
	usernameSet := make(map[string]bool)
	for _, memo := range memos {
		for _, author := range memo.AuthorMemoUsernames {
			if author != "" {
				usernameSet[strings.TrimSpace(author)] = true
			}
		}
	}

	var usernames []string
	for username := range usernameSet {
		usernames = append(usernames, username)
	}

	if len(usernames) == 0 {
		return nil
	}

	// Batch resolve authors
	for i := 0; i < len(usernames); i += a.config.BatchSize {
		end := i + a.config.BatchSize
		if end > len(usernames) {
			end = len(usernames)
		}

		batch := usernames[i:end]
		_, _ = a.ResolveAuthorsToDiscordIDs(batch) // Ignore errors during warmup

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	a.mu.Lock()
	a.lastWarmup = time.Now()
	a.metrics.LastWarmupTime = a.lastWarmup
	a.mu.Unlock()

	return nil
}

// ClearCache clears the author resolution cache
func (a *authorResolver) ClearCache() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache = make(map[string]authorCacheItem)
	a.metrics.CacheSize = 0
}

// GetMetrics returns current resolution metrics
func (a *authorResolver) GetMetrics() AuthorResolutionMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	a.metrics.CacheSize = len(a.cache)
	a.metrics.CacheEvictions = a.evictions
	return a.metrics
}

// getCachedEntry retrieves an entry from cache if valid (internal method)
func (a *authorResolver) getCachedEntry(username string) (AuthorCacheEntry, bool) {
	item, exists := a.cache[username]
	if !exists {
		return AuthorCacheEntry{}, false
	}

	// Check expiration
	if time.Now().After(item.expiresAt) {
		delete(a.cache, username)
		a.evictions++
		return AuthorCacheEntry{}, false
	}

	return item.entry, true
}

// cacheEntry stores an entry in the cache (internal method)
func (a *authorResolver) cacheEntry(username string, entry AuthorCacheEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Evict if cache is full
	if len(a.cache) >= a.config.MaxCacheSize {
		a.evictOldestEntry()
	}

	a.cache[username] = authorCacheItem{
		entry:     entry,
		expiresAt: time.Now().Add(a.config.CacheTTL),
	}
}

// evictOldestEntry removes the oldest entry from cache (internal method)
func (a *authorResolver) evictOldestEntry() {
	var oldestKey string
	var oldestTime time.Time = time.Now().Add(24 * time.Hour) // Future time

	for key, item := range a.cache {
		if item.entry.ResolvedAt.Before(oldestTime) {
			oldestTime = item.entry.ResolvedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(a.cache, oldestKey)
		a.evictions++
	}
}

// resolveUsernamesFromDatabase queries database for multiple usernames
func (a *authorResolver) resolveUsernamesFromDatabase(usernames []string) (map[string]string, error) {
	discordAccounts, err := a.store.DiscordAccount.ListByMemoUsername(a.repo.DB(), usernames)
	a.metrics.DatabaseQueryCount++

	if err != nil {
		return nil, err
	}

	// Build resolution map
	result := make(map[string]string)
	
	// Create lookup maps for memo_username, discord_username, and github_username (fallback)
	for _, account := range discordAccounts {
		if account.MemoUsername != "" {
			result[account.MemoUsername] = account.DiscordID
		}
		if account.DiscordUsername != "" {
			result[account.DiscordUsername] = account.DiscordID
		}
		// Add GitHub username as fallback option
		if account.GithubUsername != "" {
			result[account.GithubUsername] = account.DiscordID
		}
	}

	// Cache all resolved entries
	for _, username := range usernames {
		if discordID, found := result[username]; found {
			a.cacheEntry(username, AuthorCacheEntry{
				DiscordID:  discordID,
				ResolvedAt: time.Now(),
				IsValid:    true,
			})
		} else {
			fallbackValue := a.getFallbackValue(username)
			result[username] = fallbackValue
			a.cacheEntry(username, AuthorCacheEntry{
				DiscordID:  fallbackValue,
				ResolvedAt: time.Now(),
				IsValid:    true,
			})
		}
	}

	return result, nil
}

// getFallbackValue returns appropriate fallback based on strategy
func (a *authorResolver) getFallbackValue(username string) string {
	switch a.config.FallbackStrategy {
	case "plain_username":
		return username
	case "unverified_format":
		return fmt.Sprintf("%s (unverified)", username)
	case "mention_attempt":
		// Try to format as Discord mention but it won't work without proper Discord ID
		return fmt.Sprintf("@%s", username)
	default:
		return username
	}
}

// updateHitRate updates cache hit rate metrics
func (a *authorResolver) updateHitRate(hits, total int) {
	if total <= 0 {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Simple moving average approach
	currentRate := float64(hits) / float64(total)
	a.metrics.CacheHitRate = (a.metrics.CacheHitRate + currentRate) / 2
}

// updateErrorRate updates resolution error rate metrics
func (a *authorResolver) updateErrorRate(hasError bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Simple moving average approach
	errorValue := 0.0
	if hasError {
		errorValue = 1.0
	}
	a.metrics.ResolutionErrorRate = (a.metrics.ResolutionErrorRate + errorValue) / 2
}