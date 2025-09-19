package helpers

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestBreakdownDetector_DetectBreakdowns(t *testing.T) {
	tests := []struct {
		name           string
		config         BreakdownDetectionConfig
		memos          []model.MemoLog
		expected       int
		expectedTitles []string
	}{
		{
			name: "detect_by_title_case_insensitive",
			config: BreakdownDetectionConfig{
				TitleKeywords: []string{"breakdown", "deep dive"},
				CaseSensitive: false,
			},
			memos: []model.MemoLog{
				{Title: "Dify BREAKDOWN Analysis", Tags: model.JSONArrayString{}},
				{Title: "Deep Dive into AI Systems", Tags: model.JSONArrayString{}},
				{Title: "Regular post about tech", Tags: model.JSONArrayString{}},
				{Title: "breakdown of microservices", Tags: model.JSONArrayString{}},
			},
			expected:       3,
			expectedTitles: []string{"Dify BREAKDOWN Analysis", "Deep Dive into AI Systems", "breakdown of microservices"},
		},
		{
			name: "detect_by_tags_only",
			config: BreakdownDetectionConfig{
				TagKeywords:   []string{"breakdown"},
				CaseSensitive: false,
			},
			memos: []model.MemoLog{
				{Title: "Post 1", Tags: model.JSONArrayString{"breakdown", "tech"}},
				{Title: "Post 2", Tags: model.JSONArrayString{"general", "discussion"}},
				{Title: "Post 3", Tags: model.JSONArrayString{"BREAKDOWN", "analysis"}},
			},
			expected:       2,
			expectedTitles: []string{"Post 1", "Post 3"},
		},
		{
			name: "require_both_title_and_tag",
			config: BreakdownDetectionConfig{
				TitleKeywords:          []string{"breakdown"},
				TagKeywords:            []string{"breakdown"},
				RequireBothTitleAndTag: true,
				CaseSensitive:          false,
			},
			memos: []model.MemoLog{
				{Title: "breakdown post", Tags: model.JSONArrayString{"breakdown", "tech"}}, // Match both
				{Title: "breakdown post", Tags: model.JSONArrayString{"tech", "general"}},   // Title only
				{Title: "regular post", Tags: model.JSONArrayString{"breakdown", "analysis"}}, // Tag only
				{Title: "Regular post", Tags: model.JSONArrayString{"tech"}},                    // Neither
			},
			expected:       1,
			expectedTitles: []string{"breakdown post"},
		},
		{
			name: "case_sensitive_detection",
			config: BreakdownDetectionConfig{
				TitleKeywords: []string{"Breakdown"},
				CaseSensitive: true,
			},
			memos: []model.MemoLog{
				{Title: "Breakdown Analysis", Tags: model.JSONArrayString{}}, // Match
				{Title: "breakdown analysis", Tags: model.JSONArrayString{}}, // No match (case sensitive)
				{Title: "BREAKDOWN Study", Tags: model.JSONArrayString{}},    // No match
			},
			expected:       1,
			expectedTitles: []string{"Breakdown Analysis"},
		},
		{
			name: "empty_input_arrays",
			config: BreakdownDetectionConfig{
				TitleKeywords: []string{"breakdown"},
			},
			memos:          []model.MemoLog{}, // Empty array
			expected:       0,
			expectedTitles: []string{},
		},
		{
			name: "invalid_json_tags_handling",
			config: BreakdownDetectionConfig{
				TagKeywords: []string{"breakdown"},
			},
			memos: []model.MemoLog{
				{Title: "Post 1", Tags: model.JSONArrayString{"breakdown"}},  // Valid
				{Title: "Post 2", Tags: model.JSONArrayString{}},  // Empty tags
				{Title: "Post 3", Tags: model.JSONArrayString{"breakdown"}}, // Valid
			},
			expected:       2, // Should process valid records and skip invalid JSON
			expectedTitles: []string{"Post 1", "Post 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewBreakdownDetector(tt.config)
			breakdowns := detector.DetectBreakdowns(tt.memos)
			
			assert.Len(t, breakdowns, tt.expected)
			
			// Verify specific titles are detected
			actualTitles := make([]string, len(breakdowns))
			for i, breakdown := range breakdowns {
				actualTitles[i] = breakdown.Title
			}
			assert.ElementsMatch(t, tt.expectedTitles, actualTitles)
		})
	}
}

func TestBreakdownDetector_Performance(t *testing.T) {
	config := BreakdownDetectionConfig{
		TitleKeywords: []string{"breakdown", "deep dive", "analysis", "guide", "tutorial"},
		TagKeywords:   []string{"breakdown", "tech", "ai", "analysis"},
		CaseSensitive: false,
	}
	
	detector := NewBreakdownDetector(config)
	
	t.Run("large_dataset_performance", func(t *testing.T) {
		// Generate 10,000 test memos
		memos := make([]model.MemoLog, 10000)
		for i := range memos {
			if i%10 == 0 { // 10% are breakdowns
				memos[i] = model.MemoLog{
					Title: fmt.Sprintf("Breakdown analysis %d", i),
					Tags:  model.JSONArrayString{"breakdown", "tech"},
				}
			} else {
				memos[i] = model.MemoLog{
					Title: fmt.Sprintf("Regular post %d", i),
					Tags:  model.JSONArrayString{"general", "discussion"},
				}
			}
		}
		
		start := time.Now()
		breakdowns := detector.DetectBreakdowns(memos)
		duration := time.Since(start)
		
		// Verify results
		assert.Len(t, breakdowns, 1000) // Should detect 1000 breakdowns
		
		// Performance requirement: < 100ms for 10k records
		assert.Less(t, duration, 100*time.Millisecond, 
			"Detection should complete within 100ms for 10k records, took %v", duration)
		
		// Verify statistics are updated
		stats := detector.GetDetectionStats()
		assert.Equal(t, 10000, stats.TotalChecked)
		assert.Equal(t, 1000, stats.TotalBreakdowns)
		assert.Equal(t, 0.1, stats.DetectionRate)
	})
	
	t.Run("memory_usage_benchmark", func(t *testing.T) {
		var memBefore, memAfter runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		runtime.GC()
		
		// Process large dataset
		largeMemos := make([]model.MemoLog, 50000)
		for i := range largeMemos {
			largeMemos[i] = model.MemoLog{
				Title: fmt.Sprintf("Test post with breakdown analysis %d", i),
				Tags:  model.JSONArrayString{"tech", "breakdown", "analysis"},
			}
		}
		
		detector.DetectBreakdowns(largeMemos)
		
		runtime.ReadMemStats(&memAfter)
		memoryIncrease := memAfter.HeapAlloc - memBefore.HeapAlloc
		
		// Should not use more than 50MB for processing 50k records
		maxMemory := int64(50 * 1024 * 1024) // 50MB
		assert.Less(t, int64(memoryIncrease), maxMemory,
			"Memory usage too high: %d bytes for 50k records", memoryIncrease)
	})
}

func TestBreakdownDetector_EdgeCases(t *testing.T) {
	t.Run("nil_config_handling", func(t *testing.T) {
		config := BreakdownDetectionConfig{} // Empty config
		detector := NewBreakdownDetector(config)
		
		memos := []model.MemoLog{
			{Title: "breakdown post", Tags: model.JSONArrayString{"breakdown"}},
		}
		
		// Should handle empty keywords gracefully
		breakdowns := detector.DetectBreakdowns(memos)
		assert.Len(t, breakdowns, 0) // No keywords = no matches
	})
	
	t.Run("concurrent_access_safety", func(t *testing.T) {
		config := BreakdownDetectionConfig{
			TitleKeywords: []string{"breakdown"},
		}
		detector := NewBreakdownDetector(config)
		
		memo := []model.MemoLog{
			{Title: "breakdown post", Tags: model.JSONArrayString{}},
		}
		
		// Run concurrent detection operations
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				breakdowns := detector.DetectBreakdowns(memo)
				assert.Len(t, breakdowns, 1)
			}()
		}
		
		wg.Wait()
		
		// Verify final stats are consistent
		stats := detector.GetDetectionStats()
		assert.Equal(t, 100, stats.TotalChecked)
		assert.Equal(t, 100, stats.TotalBreakdowns)
	})
	
	t.Run("special_characters_in_keywords", func(t *testing.T) {
		config := BreakdownDetectionConfig{
			TitleKeywords: []string{"deep-dive", "break@down", "analysis(detailed)"},
			CaseSensitive: false,
		}
		detector := NewBreakdownDetector(config)
		
		memos := []model.MemoLog{
			{Title: "deep-dive into systems", Tags: model.JSONArrayString{}},
			{Title: "break@down of process", Tags: model.JSONArrayString{}},
			{Title: "analysis(detailed) report", Tags: model.JSONArrayString{}},
		}
		
		breakdowns := detector.DetectBreakdowns(memos)
		assert.Len(t, breakdowns, 3)
	})
}

func TestBreakdownDetector_IsBreakdown(t *testing.T) {
	config := BreakdownDetectionConfig{
		TitleKeywords: []string{"breakdown", "analysis"},
		TagKeywords:   []string{"breakdown", "deep-dive"},
		CaseSensitive: false,
	}
	
	detector := NewBreakdownDetector(config)
	
	tests := []struct {
		name     string
		title    string
		tags     []string
		expected bool
	}{
		{
			name:     "title_match",
			title:    "Breakdown of microservices",
			tags:     []string{"tech"},
			expected: true,
		},
		{
			name:     "tag_match",
			title:    "System architecture",
			tags:     []string{"breakdown", "tech"},
			expected: true,
		},
		{
			name:     "no_match",
			title:    "Regular post",
			tags:     []string{"general"},
			expected: false,
		},
		{
			name:     "empty_inputs",
			title:    "",
			tags:     []string{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsBreakdown(tt.title, tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBreakdownDetector_UpdateConfig(t *testing.T) {
	initialConfig := BreakdownDetectionConfig{
		TitleKeywords: []string{"breakdown"},
		CaseSensitive: false,
	}
	
	detector := NewBreakdownDetector(initialConfig)
	
	// Test with initial config
	isBreakdown := detector.IsBreakdown("analysis post", []string{})
	assert.False(t, isBreakdown)
	
	// Update config
	newConfig := BreakdownDetectionConfig{
		TitleKeywords: []string{"analysis"},
		CaseSensitive: false,
	}
	detector.UpdateConfig(newConfig)
	
	// Test with new config
	isBreakdown = detector.IsBreakdown("analysis post", []string{})
	assert.True(t, isBreakdown)
}

func TestBreakdownDetector_GetDetectionStats(t *testing.T) {
	config := BreakdownDetectionConfig{
		TitleKeywords: []string{"breakdown"},
	}
	
	detector := NewBreakdownDetector(config)
	
	// Initial stats should be zero
	stats := detector.GetDetectionStats()
	assert.Equal(t, 0, stats.TotalChecked)
	assert.Equal(t, 0, stats.TotalBreakdowns)
	assert.Equal(t, float64(0), stats.DetectionRate)
	
	// Process some memos
	memos := []model.MemoLog{
		{Title: "breakdown analysis", Tags: model.JSONArrayString{}},
		{Title: "regular post", Tags: model.JSONArrayString{}},
		{Title: "another breakdown", Tags: model.JSONArrayString{}},
	}
	
	detector.DetectBreakdowns(memos)
	
	// Stats should be updated
	stats = detector.GetDetectionStats()
	assert.Equal(t, 3, stats.TotalChecked)
	assert.Equal(t, 2, stats.TotalBreakdowns)
	assert.Equal(t, float64(2)/float64(3), stats.DetectionRate)
	assert.Equal(t, 2, stats.TitleMatches)
	assert.Equal(t, 0, stats.TagMatches)
	assert.True(t, time.Since(stats.LastUpdated) < time.Second)
}