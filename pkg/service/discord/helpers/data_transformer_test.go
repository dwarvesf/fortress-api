package helpers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func setupDataTransformerTest() DataTransformer {
	config := DataTransformationConfig{
		DateFormats: []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z",
			"01/02/2006",
		},
		DefaultReward:           "25",
		DefaultCategory:         []string{"others"},
		EnableValidation:        true,
		SkipInvalidRecords:      false,
		MaxContentLength:        1000,
		AuthorResolutionRetries: 3,
	}
	
	return NewDataTransformer(config)
}

func TestNewDataTransformer(t *testing.T) {
	tests := []struct {
		name     string
		config   DataTransformationConfig
		expected DataTransformationConfig
	}{
		{
			name:   "default configuration",
			config: DataTransformationConfig{},
			expected: DataTransformationConfig{
				DateFormats:             []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.000Z", "2006-01-02 15:04:05", "01/02/2006", "Jan 2, 2006"},
				DefaultReward:           "25",
				DefaultCategory:         []string{"others"},
				MaxContentLength:        10000,
				AuthorResolutionRetries: 3,
			},
		},
		{
			name: "custom configuration",
			config: DataTransformationConfig{
				DateFormats:             []string{"2006-01-02"},
				DefaultReward:           "50",
				DefaultCategory:         []string{"custom"},
				MaxContentLength:        500,
				AuthorResolutionRetries: 5,
			},
			expected: DataTransformationConfig{
				DateFormats:             []string{"2006-01-02"},
				DefaultReward:           "50",
				DefaultCategory:         []string{"custom"},
				MaxContentLength:        500,
				AuthorResolutionRetries: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewDataTransformer(tt.config)
			
			assert.NotNil(t, transformer)
			
			// Test initial stats
			stats := transformer.GetTransformationStats()
			assert.Equal(t, 0, stats.TotalTransformed)
			assert.Equal(t, 0, stats.SuccessfulTransforms)
			assert.Equal(t, 0, stats.FailedTransforms)
			assert.True(t, stats.LastTransformation.IsZero())
		})
	}
}

func TestDataTransformer_TransformParquetRecord_Success(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	// Test valid record
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{"author1", "author2"},
		Tags:    []string{"test", "example"},
		URL:     "https://example.com",
		Content: "This is test content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Memo", result.Title)
	assert.Equal(t, "This is test content", result.Description)
	assert.Equal(t, []string{"author1", "author2"}, result.AuthorMemoUsernames)
	assert.Equal(t, model.JSONArrayString{"test", "example"}, result.Tags)
	assert.Equal(t, "https://example.com", result.URL)
	assert.Equal(t, "25", result.Reward.String())
	assert.Equal(t, pq.StringArray{"others"}, result.Category)
	
	// Verify date parsing
	expectedDate, _ := time.Parse("2006-01-02", "2024-01-15")
	assert.Equal(t, expectedDate, *result.PublishedAt)
}

func TestDataTransformer_TransformParquetRecord_InvalidDate(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	record := ParquetMemoRecord{
		Date:    "invalid-date",
		Title:   "Test Memo",
		Authors: []string{"author1"},
		Tags:    []string{"test"},
		URL:     "https://example.com",
		Content: "Content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse date")
	assert.Nil(t, result)
}

func TestDataTransformer_TransformParquetRecord_MissingRequiredFields(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	tests := []struct {
		name   string
		record ParquetMemoRecord
		error  string
	}{
		{
			name: "missing date",
			record: ParquetMemoRecord{
				Title:   "Test Memo",
				Authors: []string{"author1"},
			},
			error: "date field is required",
		},
		{
			name: "missing title",
			record: ParquetMemoRecord{
				Date:    "2024-01-15",
				Authors: []string{"author1"},
			},
			error: "title field is required",
		},
		{
			name: "missing authors with validation",
			record: ParquetMemoRecord{
				Date:    "2024-01-15",
				Title:   "Test Memo",
				Authors: []string{},
			},
			error: "at least one author is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.TransformParquetRecord(tt.record)
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
			assert.Nil(t, result)
		})
	}
}

func TestDataTransformer_TransformParquetRecord_ContentTruncation(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	// Create content longer than max length (1000 chars in test config)
	longContent := strings.Repeat("a", 1200)
	
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{"author1"},
		Tags:    []string{"test"},
		Content: longContent,
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, len(result.Description) <= 1000)
	assert.True(t, strings.HasSuffix(result.Description, "... [truncated]"))
}

func TestDataTransformer_TransformParquetRecord_AuthorCleaning(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{
			"  author1  ", // With whitespace
			"Author1",     // Duplicate (case different)
			"",           // Empty
			"author2",    // Valid
			"AUTHOR2",    // Duplicate (case different)
		},
		Tags:    []string{"test"},
		Content: "Content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Should have 2 unique authors: author1 and author2
	assert.Len(t, result.AuthorMemoUsernames, 2)
	assert.Contains(t, result.AuthorMemoUsernames, "author1")
	assert.Contains(t, result.AuthorMemoUsernames, "author2")
}

func TestDataTransformer_TransformParquetRecord_TagCleaning(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{"author1"},
		Tags: []string{
			"  TAG1  ", // With whitespace, should be lowercased
			"tag1",     // Duplicate
			"",        // Empty
			"Tag-2",   // Valid with hyphen
			"tag_3",   // Valid with underscore
			"Tag@4",   // Invalid character, should be filtered out
		},
		Content: "Content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Should have 3 unique valid tags: tag1, tag-2, tag_3
	assert.Len(t, result.Tags, 3)
	assert.Contains(t, []string(result.Tags), "tag1")
	assert.Contains(t, []string(result.Tags), "tag-2")
	assert.Contains(t, []string(result.Tags), "tag_3")
	assert.NotContains(t, []string(result.Tags), "tag@4")
}

func TestDataTransformer_TransformParquetRecord_SuspiciousAuthor(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{
			"normalauthor",
			"@everyone",      // Suspicious
			"<script>alert", // Suspicious
			"javascript:void", // Suspicious
		},
		Tags:    []string{"test"},
		Content: "Content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	
	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Should only have the normal author
	assert.Len(t, result.AuthorMemoUsernames, 1)
	assert.Equal(t, "normalauthor", result.AuthorMemoUsernames[0])
}

func TestDataTransformer_TransformParquetRecords_BatchSuccess(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	records := []ParquetMemoRecord{
		{
			Date:    "2024-01-15",
			Title:   "Memo 1",
			Authors: []string{"author1"},
			Tags:    []string{"tag1"},
			Content: "Content 1",
		},
		{
			Date:    "2024-01-16",
			Title:   "Memo 2",
			Authors: []string{"author2"},
			Tags:    []string{"tag2"},
			Content: "Content 2",
		},
	}
	
	results, err := transformer.TransformParquetRecords(records)
	
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Memo 1", results[0].Title)
	assert.Equal(t, "Memo 2", results[1].Title)
}

func TestDataTransformer_TransformParquetRecords_EmptyInput(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	results, err := transformer.TransformParquetRecords([]ParquetMemoRecord{})
	
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestDataTransformer_TransformParquetRecords_WithSkipInvalid(t *testing.T) {
	// Create transformer that skips invalid records
	config := DataTransformationConfig{
		DateFormats:        []string{"2006-01-02"},
		DefaultReward:      "25",
		DefaultCategory:    []string{"others"},
		EnableValidation:   true,
		SkipInvalidRecords: true,
		MaxContentLength:   1000,
	}
	transformer := NewDataTransformer(config)
	
	records := []ParquetMemoRecord{
		{
			Date:    "2024-01-15",
			Title:   "Valid Memo",
			Authors: []string{"author1"},
			Content: "Valid content",
		},
		{
			Date:    "invalid-date", // Invalid record
			Title:   "Invalid Memo",
			Authors: []string{"author2"},
			Content: "Invalid content",
		},
		{
			Date:    "2024-01-17",
			Title:   "Another Valid Memo",
			Authors: []string{"author3"},
			Content: "Another valid content",
		},
	}
	
	results, err := transformer.TransformParquetRecords(records)
	
	assert.NoError(t, err)
	assert.Len(t, results, 2) // Should skip the invalid record
	assert.Equal(t, "Valid Memo", results[0].Title)
	assert.Equal(t, "Another Valid Memo", results[1].Title)
}

func TestDataTransformer_TransformParquetRecords_FailOnInvalid(t *testing.T) {
	transformer := setupDataTransformerTest() // SkipInvalidRecords: false
	
	records := []ParquetMemoRecord{
		{
			Date:    "2024-01-15",
			Title:   "Valid Memo",
			Authors: []string{"author1"},
			Content: "Valid content",
		},
		{
			Date:    "invalid-date", // Invalid record
			Title:   "Invalid Memo",
			Authors: []string{"author2"},
			Content: "Invalid content",
		},
	}
	
	results, err := transformer.TransformParquetRecords(records)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch transformation failed")
	assert.Nil(t, results)
}

func TestDataTransformer_GetTransformationStats(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	// Initial stats
	stats := transformer.GetTransformationStats()
	assert.Equal(t, 0, stats.TotalTransformed)
	assert.Equal(t, 0, stats.SuccessfulTransforms)
	assert.Equal(t, 0, stats.FailedTransforms)
	
	// Transform some records
	validRecord := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Valid Memo",
		Authors: []string{"author1"},
		Content: "Content",
	}
	_, err := transformer.TransformParquetRecord(validRecord)
	assert.NoError(t, err)
	
	invalidRecord := ParquetMemoRecord{
		Date:    "invalid-date",
		Title:   "Invalid Memo",
		Authors: []string{"author1"},
		Content: "Content",
	}
	_, err = transformer.TransformParquetRecord(invalidRecord)
	assert.Error(t, err)
	
	// Check updated stats
	stats = transformer.GetTransformationStats()
	assert.Equal(t, 2, stats.TotalTransformed)
	assert.Equal(t, 1, stats.SuccessfulTransforms)
	assert.Equal(t, 1, stats.FailedTransforms)
	assert.False(t, stats.LastTransformation.IsZero())
}

func TestDataTransformer_UpdateConfig(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	newConfig := DataTransformationConfig{
		DateFormats:     []string{"01/02/2006"},
		DefaultReward:   "50",
		DefaultCategory: []string{"updated"},
	}
	
	transformer.UpdateConfig(newConfig)
	
	// Test that new config is used
	record := ParquetMemoRecord{
		Date:    "01/15/2024", // New format
		Title:   "Test Memo",
		Authors: []string{"author1"},
		Content: "Content",
	}
	
	result, err := transformer.TransformParquetRecord(record)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "50", result.Reward.String())
	assert.Equal(t, pq.StringArray{"updated"}, result.Category)
}

func TestDataTransformer_ValidateSchema(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	tests := []struct {
		name    string
		sample  []ParquetMemoRecord
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty sample",
			sample:  []ParquetMemoRecord{},
			wantErr: true,
			errMsg:  "no sample data provided",
		},
		{
			name: "valid schema",
			sample: []ParquetMemoRecord{
				{
					Date:    "2024-01-15",
					Title:   "Valid Memo",
					Authors: []string{"author1"},
					Content: "Content",
				},
			},
			wantErr: false,
		},
		{
			name: "missing date",
			sample: []ParquetMemoRecord{
				{
					Title:   "No Date Memo",
					Authors: []string{"author1"},
					Content: "Content",
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'date'",
		},
		{
			name: "missing title",
			sample: []ParquetMemoRecord{
				{
					Date:    "2024-01-15",
					Authors: []string{"author1"},
					Content: "Content",
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'title'",
		},
		{
			name: "missing authors",
			sample: []ParquetMemoRecord{
				{
					Date:    "2024-01-15",
					Title:   "No Authors Memo",
					Authors: []string{},
					Content: "Content",
				},
			},
			wantErr: true,
			errMsg:  "missing required field 'authors'",
		},
		{
			name: "invalid date format",
			sample: []ParquetMemoRecord{
				{
					Date:    "invalid-date",
					Title:   "Invalid Date Memo",
					Authors: []string{"author1"},
					Content: "Content",
				},
			},
			wantErr: true,
			errMsg:  "invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := transformer.(*dataTransformer).ValidateSchema(tt.sample)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataTransformer_TransformWithMetadata(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	records := []ParquetMemoRecord{
		{
			Date:    "2024-01-15",
			Title:   "Memo 1",
			Authors: []string{"author1"},
			Content: "Content 1",
		},
		{
			Date:    "2024-01-16",
			Title:   "Memo 2",
			Authors: []string{"author2"},
			Content: "Content 2",
		},
	}
	
	results, metadata, err := transformer.(*dataTransformer).TransformWithMetadata(records)
	
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NotNil(t, metadata)
	
	// Check metadata
	assert.Equal(t, 2, metadata["input_records"])
	assert.Equal(t, 2, metadata["output_records"])
	assert.Equal(t, 100.0, metadata["success_rate"])
	assert.NotNil(t, metadata["processing_time_ms"])
	assert.NotNil(t, metadata["transformation_date"])
	assert.Equal(t, 2, metadata["total_transformed"])
	assert.Equal(t, 2, metadata["successful_transforms"])
	assert.Equal(t, 0, metadata["failed_transforms"])
}

func TestDataTransformer_GetSupportedDateFormats(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	formats := transformer.(*dataTransformer).GetSupportedDateFormats()
	
	assert.NotEmpty(t, formats)
	assert.Contains(t, formats, "2006-01-02")
	assert.Contains(t, formats, "2006-01-02T15:04:05Z")
	assert.Contains(t, formats, "01/02/2006")
}

func TestDataTransformer_Reset(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	// Transform a record to create stats
	record := ParquetMemoRecord{
		Date:    "2024-01-15",
		Title:   "Test Memo",
		Authors: []string{"author1"},
		Content: "Content",
	}
	_, err := transformer.TransformParquetRecord(record)
	assert.NoError(t, err)
	
	// Verify stats exist
	stats := transformer.GetTransformationStats()
	assert.Greater(t, stats.TotalTransformed, 0)
	
	// Reset stats
	transformer.(*dataTransformer).Reset()
	
	// Verify stats are reset
	stats = transformer.GetTransformationStats()
	assert.Equal(t, 0, stats.TotalTransformed)
	assert.Equal(t, 0, stats.SuccessfulTransforms)
	assert.Equal(t, 0, stats.FailedTransforms)
	assert.True(t, stats.LastTransformation.IsZero())
}

func TestDataTransformer_GetConfig(t *testing.T) {
	config := DataTransformationConfig{
		DateFormats:     []string{"2006-01-02"},
		DefaultReward:   "30",
		DefaultCategory: []string{"test"},
	}
	transformer := NewDataTransformer(config)
	
	retrievedConfig := transformer.(*dataTransformer).GetConfig()
	
	assert.Equal(t, config.DefaultReward, retrievedConfig.DefaultReward)
	assert.Equal(t, config.DefaultCategory, retrievedConfig.DefaultCategory)
	assert.Contains(t, retrievedConfig.DateFormats, "2006-01-02")
}

func TestDataTransformer_ParseDate_AllFormats(t *testing.T) {
	transformer := setupDataTransformerTest()
	
	testCases := []struct {
		dateStr string
		format  string
		valid   bool
	}{
		{"2024-01-15", "2006-01-02", true},
		{"2024-01-15T10:30:45Z", "2006-01-02T15:04:05Z", true},
		{"01/15/2024", "01/02/2006", true},
		{"invalid-date", "", false},
		{"", "", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.dateStr, func(t *testing.T) {
			impl := transformer.(*dataTransformer)
			result, err := impl.parseDate(tc.dateStr)
			
			if tc.valid {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			} else {
				assert.Error(t, err)
				assert.True(t, result.IsZero())
			}
		})
	}
}