package helpers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// newAuthorDetector implements the NewAuthorDetector interface with historical comparison
type newAuthorDetector struct {
	mu             sync.RWMutex
	store          *store.Store
	repo           store.DBRepo
	historicalCache map[string]map[string]bool // period -> authors set
	cacheTimestamp map[string]time.Time        // period -> last updated time
	stats          NewAuthorDetectionStats
}

// NewNewAuthorDetector creates a new NewAuthorDetector instance
func NewNewAuthorDetector(store *store.Store, repo store.DBRepo) NewAuthorDetector {
	return &newAuthorDetector{
		store:           store,
		repo:            repo,
		historicalCache: make(map[string]map[string]bool),
		cacheTimestamp:  make(map[string]time.Time),
		stats: NewAuthorDetectionStats{
			LastDetectionTime:   time.Time{},
			WeeklyNewAuthors:    0,
			MonthlyNewAuthors:   0,
			HistoricalCacheSize: 0,
		},
	}
}

// DetectNewAuthors identifies first-time contributors in current period vs historical period
func (n *newAuthorDetector) DetectNewAuthors(currentMemos []model.MemoLog, period string) ([]string, error) {
	if len(currentMemos) == 0 {
		return []string{}, nil
	}

	// Validate period
	if period != "weekly" && period != "monthly" {
		return nil, fmt.Errorf("invalid period '%s', must be 'weekly' or 'monthly'", period)
	}

	// Extract unique authors from current memos
	currentAuthors := n.extractUniqueAuthors(currentMemos)
	if len(currentAuthors) == 0 {
		return []string{}, nil
	}

	// Get historical authors for comparison
	historicalAuthors, err := n.GetHistoricalAuthors(period)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical authors for period '%s': %w", period, err)
	}

	// Find new authors (present in current but not in historical)
	var newAuthors []string
	for author := range currentAuthors {
		if !historicalAuthors[author] {
			newAuthors = append(newAuthors, author)
		}
	}

	// Update statistics
	n.mu.Lock()
	n.stats.LastDetectionTime = time.Now()
	switch period {
	case "weekly":
		n.stats.WeeklyNewAuthors = len(newAuthors)
	case "monthly":
		n.stats.MonthlyNewAuthors = len(newAuthors)
	}
	n.stats.HistoricalCacheSize = len(n.historicalCache)
	n.mu.Unlock()

	return newAuthors, nil
}

// GetHistoricalAuthors retrieves historical author set for comparison
func (n *newAuthorDetector) GetHistoricalAuthors(period string) (map[string]bool, error) {
	// Check cache first
	n.mu.RLock()
	if cached, exists := n.historicalCache[period]; exists {
		if timestamp, timeExists := n.cacheTimestamp[period]; timeExists {
			// Cache is valid for 1 hour to avoid excessive database queries
			if time.Since(timestamp) < 1*time.Hour {
				n.mu.RUnlock()
				return n.copyAuthorSet(cached), nil
			}
		}
	}
	n.mu.RUnlock()

	// Calculate historical period based on type
	historicalStart, historicalEnd, err := n.calculateHistoricalPeriod(period)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate historical period: %w", err)
	}

	// Query database for historical memos
	historicalMemos, err := n.store.MemoLog.GetLimitByTimeRange(n.repo.DB(), &historicalStart, &historicalEnd, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical memos: %w", err)
	}

	// Extract unique authors from historical memos
	historicalAuthors := n.extractUniqueAuthors(historicalMemos)

	// Cache the result
	n.mu.Lock()
	n.historicalCache[period] = historicalAuthors
	n.cacheTimestamp[period] = time.Now()
	n.mu.Unlock()

	return n.copyAuthorSet(historicalAuthors), nil
}

// ClearHistoricalCache clears the historical author cache
func (n *newAuthorDetector) ClearHistoricalCache() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.historicalCache = make(map[string]map[string]bool)
	n.cacheTimestamp = make(map[string]time.Time)
	n.stats.HistoricalCacheSize = 0
}

// GetDetectionStats returns current detection statistics
func (n *newAuthorDetector) GetDetectionStats() NewAuthorDetectionStats {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stats
}

// extractUniqueAuthors extracts unique author names from memo list
func (n *newAuthorDetector) extractUniqueAuthors(memos []model.MemoLog) map[string]bool {
	authors := make(map[string]bool)
	
	for _, memo := range memos {
		for _, author := range memo.AuthorMemoUsernames {
			// Clean and normalize author names
			cleanAuthor := strings.TrimSpace(author)
			if cleanAuthor != "" {
				// Normalize to lowercase for case-insensitive comparison
				authors[strings.ToLower(cleanAuthor)] = true
			}
		}
	}
	
	return authors
}

// calculateHistoricalPeriod calculates the historical comparison period
func (n *newAuthorDetector) calculateHistoricalPeriod(period string) (time.Time, time.Time, error) {
	now := time.Now()
	
	switch period {
	case "weekly":
		// For weekly: compare against last 30 days (excluding current week)
		// This gives us roughly 4 weeks of historical data
		currentWeekStart := now.AddDate(0, 0, -int(now.Weekday())+1) // Monday of current week
		historicalEnd := currentWeekStart.Add(-1 * time.Second)      // End of previous week
		historicalStart := historicalEnd.AddDate(0, 0, -30)         // 30 days before
		
		return historicalStart, historicalEnd, nil
		
	case "monthly":
		// For monthly: compare against last 12 months (excluding current month)
		currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		historicalEnd := currentMonthStart.Add(-1 * time.Second) // End of previous month
		historicalStart := historicalEnd.AddDate(-12, 0, 0)      // 12 months before
		
		return historicalStart, historicalEnd, nil
		
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported period: %s", period)
	}
}

// copyAuthorSet creates a deep copy of author set to prevent concurrent access issues
func (n *newAuthorDetector) copyAuthorSet(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for author, exists := range original {
		copy[author] = exists
	}
	return copy
}


// GetNewAuthorsByTimeRange is a utility method to get new authors for a specific time range
func (n *newAuthorDetector) GetNewAuthorsByTimeRange(currentStart, currentEnd time.Time, period string) ([]string, error) {
	// Get current period memos
	currentMemos, err := n.store.MemoLog.GetLimitByTimeRange(n.repo.DB(), &currentStart, &currentEnd, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get current period memos: %w", err)
	}
	
	// Use standard detection logic
	return n.DetectNewAuthors(currentMemos, period)
}

// GetAuthorFirstAppearance finds when an author first appeared in the system
func (n *newAuthorDetector) GetAuthorFirstAppearance(authorName string) (*time.Time, error) {
	if strings.TrimSpace(authorName) == "" {
		return nil, fmt.Errorf("author name cannot be empty")
	}
	
	// Query for the earliest memo by this author
	// Note: This is a simplified implementation. In production, you might want
	// to add a specific store method for this query for better performance
	
	// Get memos from a very early date (e.g., 5 years ago) to now
	earlyDate := time.Now().AddDate(-5, 0, 0)
	now := time.Now()
	
	memos, err := n.store.MemoLog.GetLimitByTimeRange(n.repo.DB(), &earlyDate, &now, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to query memos for author history: %w", err)
	}
	
	var firstAppearance *time.Time
	normalizedAuthor := strings.ToLower(strings.TrimSpace(authorName))
	
	for _, memo := range memos {
		for _, author := range memo.AuthorMemoUsernames {
			if strings.ToLower(strings.TrimSpace(author)) == normalizedAuthor {
				// Use PublishedAt or CreatedAt as the memo date
				var memoTime time.Time
				if memo.PublishedAt != nil {
					memoTime = *memo.PublishedAt
				} else {
					memoTime = memo.CreatedAt
				}
				if firstAppearance == nil || memoTime.Before(*firstAppearance) {
					firstAppearance = &memoTime
				}
				break // Found author in this memo, move to next memo
			}
		}
	}
	
	return firstAppearance, nil
}

// GetAuthorStatistics provides statistics about author activity
func (n *newAuthorDetector) GetAuthorStatistics(period string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Get current period range
	var start, end time.Time
	now := time.Now()
	
	switch period {
	case "weekly":
		start = now.AddDate(0, 0, -7)
		end = now
	case "monthly":
		start = now.AddDate(0, -1, 0)
		end = now
	default:
		return nil, fmt.Errorf("invalid period: %s", period)
	}
	
	// Get current memos
	currentMemos, err := n.store.MemoLog.GetLimitByTimeRange(n.repo.DB(), &start, &end, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get current memos: %w", err)
	}
	
	// Get historical authors
	historicalAuthors, err := n.GetHistoricalAuthors(period)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical authors: %w", err)
	}
	
	currentAuthors := n.extractUniqueAuthors(currentMemos)
	
	// Calculate statistics
	stats["total_current_authors"] = len(currentAuthors)
	stats["total_historical_authors"] = len(historicalAuthors)
	
	// Find returning authors
	returningCount := 0
	for author := range currentAuthors {
		if historicalAuthors[author] {
			returningCount++
		}
	}
	stats["returning_authors"] = returningCount
	
	// New authors
	newAuthors, _ := n.DetectNewAuthors(currentMemos, period)
	stats["new_authors_count"] = len(newAuthors)
	stats["new_authors_list"] = newAuthors
	
	// Calculate percentages
	if len(currentAuthors) > 0 {
		stats["new_author_percentage"] = float64(len(newAuthors)) / float64(len(currentAuthors)) * 100
		stats["returning_author_percentage"] = float64(returningCount) / float64(len(currentAuthors)) * 100
	} else {
		stats["new_author_percentage"] = 0.0
		stats["returning_author_percentage"] = 0.0
	}
	
	stats["period"] = period
	stats["start_date"] = start.Format("2006-01-02")
	stats["end_date"] = end.Format("2006-01-02")
	stats["generated_at"] = time.Now()
	
	return stats, nil
}