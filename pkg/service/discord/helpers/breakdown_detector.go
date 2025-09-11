package helpers

import (
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type breakdownDetector struct {
	config BreakdownDetectionConfig
	stats  BreakdownDetectionStats
	mutex  sync.RWMutex
}

// NewBreakdownDetector creates a new breakdown detector with the given configuration.
// The detector will use the provided configuration to identify breakdown posts based on
// title keywords and/or tag keywords, with configurable case sensitivity and matching logic.
func NewBreakdownDetector(config BreakdownDetectionConfig) BreakdownDetector {
	return &breakdownDetector{
		config: config,
		stats: BreakdownDetectionStats{
			LastUpdated: time.Now(),
		},
	}
}

// DetectBreakdowns analyzes a list of memo logs and identifies which ones are breakdown posts.
// This method processes all memos in the input slice and returns only those that match the
// configured breakdown detection criteria. It also updates internal statistics.
// Returns a new slice containing only the breakdown posts, preserving their original order.
func (d *breakdownDetector) DetectBreakdowns(memos []model.MemoLog) []model.MemoLog {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	breakdowns := make([]model.MemoLog, 0)
	totalChecked := 0
	titleMatches := 0
	tagMatches := 0
	
	for _, memo := range memos {
		totalChecked++
		
		// Convert tags from JSONArrayString to []string
		tags := []string(memo.Tags)
		
		isBreakdown, matchedTitle, matchedTag := d.isBreakdownWithDetails(memo.Title, tags)
		if isBreakdown {
			breakdowns = append(breakdowns, memo)
			if matchedTitle {
				titleMatches++
			}
			if matchedTag {
				tagMatches++
			}
		}
	}
	
	// Update statistics
	d.stats.TotalChecked += totalChecked
	d.stats.TotalBreakdowns += len(breakdowns)
	d.stats.TitleMatches += titleMatches
	d.stats.TagMatches += tagMatches
	if d.stats.TotalChecked > 0 {
		d.stats.DetectionRate = float64(d.stats.TotalBreakdowns) / float64(d.stats.TotalChecked)
	}
	d.stats.LastUpdated = time.Now()
	
	return breakdowns
}

// IsBreakdown checks whether a single post is a breakdown based on its title and tags.
// This method provides a simple yes/no determination without updating statistics or processing
// full memo objects. It's thread-safe and can be called concurrently.
// Returns true if the title and/or tags match the configured breakdown criteria.
func (d *breakdownDetector) IsBreakdown(title string, tags []string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	isBreakdown, _, _ := d.isBreakdownWithDetails(title, tags)
	return isBreakdown
}

// isBreakdownWithDetails performs the actual breakdown detection logic and returns detailed results.
// Returns: (isBreakdown, titleMatched, tagMatched) for internal use and statistics tracking.
func (d *breakdownDetector) isBreakdownWithDetails(title string, tags []string) (bool, bool, bool) {
	titleMatch := d.matchesKeywords(title, d.config.TitleKeywords)
	tagMatch := d.matchesTagKeywords(tags, d.config.TagKeywords)
	
	if d.config.RequireBothTitleAndTag {
		return titleMatch && tagMatch, titleMatch, tagMatch
	}
	
	return titleMatch || tagMatch, titleMatch, tagMatch
}

// matchesKeywords checks if the given text contains any of the specified keywords.
// Respects the case sensitivity configuration and returns true on first match.
func (d *breakdownDetector) matchesKeywords(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}
	
	searchText := text
	if !d.config.CaseSensitive {
		searchText = strings.ToLower(text)
	}
	
	for _, keyword := range keywords {
		searchKeyword := keyword
		if !d.config.CaseSensitive {
			searchKeyword = strings.ToLower(keyword)
		}
		
		if strings.Contains(searchText, searchKeyword) {
			return true
		}
	}
	
	return false
}

// matchesTagKeywords checks if any of the provided tags match the configured tag keywords.
// Each tag is individually checked against all keywords using the standard keyword matching logic.
func (d *breakdownDetector) matchesTagKeywords(tags []string, keywords []string) bool {
	if len(keywords) == 0 {
		return false
	}
	
	for _, tag := range tags {
		if d.matchesKeywords(tag, keywords) {
			return true
		}
	}
	return false
}

// GetDetectionStats returns the current detection statistics including total posts checked,
// total breakdowns found, detection rate, and separate counts for title vs tag matches.
// This method is thread-safe and provides a snapshot of the detector's performance metrics.
func (d *breakdownDetector) GetDetectionStats() BreakdownDetectionStats {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	return d.stats
}

// UpdateConfig updates the detector's configuration with new settings.
// This operation is thread-safe and takes effect immediately for subsequent detections.
// Changing configuration does not reset accumulated statistics.
func (d *breakdownDetector) UpdateConfig(config BreakdownDetectionConfig) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	d.config = config
}