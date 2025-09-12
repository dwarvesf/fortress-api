package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/discord/helpers"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

func main() {
	// Load configuration
	cfg := config.LoadTestConfig()
	
	// Initialize store and repository
	s := store.New()
	repo, err := store.NewRepo(cfg.DB)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Define the same time range as in the handler
	julyStart := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	now := time.Now()
	
	fmt.Printf("Querying memos from %v to %v\n", julyStart, now)
	
	// Query memos using the same method as the handler
	memos, err := s.MemoLog.GetLimitByTimeRange(repo.DB(), &julyStart, &now, 10000)
	if err != nil {
		log.Fatal("Failed to query memos:", err)
	}
	
	fmt.Printf("Found %d memos total\n", len(memos))
	
	// Print first few memos for verification
	fmt.Println("\nFirst 5 memos:")
	for i, memo := range memos {
		if i >= 5 {
			break
		}
		fmt.Printf("  %d. %s (tags: %v)\n", i+1, memo.Title, memo.Tags)
	}
	
	// Initialize breakdown detector with same config as Discord service
	detector := helpers.NewBreakdownDetector(helpers.BreakdownDetectionConfig{
		TitleKeywords:          []string{"breakdown", "weekly", "summary", "report"},
		TagKeywords:            []string{"breakdown", "weekly-report", "summary"},
		CaseSensitive:          false,
		RequireBothTitleAndTag: false,
	})
	
	// Detect breakdowns
	breakdowns := detector.DetectBreakdowns(memos)
	fmt.Printf("\nFound %d breakdown posts:\n", len(breakdowns))
	
	for i, breakdown := range breakdowns {
		fmt.Printf("  %d. %s (tags: %v)\n", i+1, breakdown.Title, breakdown.Tags)
	}
	
	// Show detection stats
	stats := detector.GetDetectionStats()
	fmt.Printf("\nDetection statistics:\n")
	fmt.Printf("  Total checked: %d\n", stats.TotalChecked)
	fmt.Printf("  Total breakdowns: %d\n", stats.TotalBreakdowns)
	fmt.Printf("  Detection rate: %.2f%%\n", stats.DetectionRate*100)
	fmt.Printf("  Title matches: %d\n", stats.TitleMatches)
	fmt.Printf("  Tag matches: %d\n", stats.TagMatches)
}
