package helpers

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// dataTransformer implements the DataTransformer interface for converting parquet data to MemoLog models
type dataTransformer struct {
	mu     sync.RWMutex
	config DataTransformationConfig
	stats  DataTransformationStats
}

// NewDataTransformer creates a new DataTransformer instance
func NewDataTransformer(config DataTransformationConfig) DataTransformer {
	// Set default configuration if not provided
	if len(config.DateFormats) == 0 {
		config.DateFormats = []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
			"01/02/2006",
			"Jan 2, 2006",
		}
	}
	if config.DefaultReward == "" {
		config.DefaultReward = "25"
	}
	if len(config.DefaultCategory) == 0 {
		config.DefaultCategory = []string{"others"}
	}
	if config.MaxContentLength == 0 {
		config.MaxContentLength = 10000 // 10KB limit
	}
	if config.AuthorResolutionRetries == 0 {
		config.AuthorResolutionRetries = 3
	}

	return &dataTransformer{
		config: config,
		stats: DataTransformationStats{
			TotalTransformed:     0,
			SuccessfulTransforms: 0,
			FailedTransforms:     0,
			LastTransformation:   time.Time{},
		},
	}
}

// TransformParquetRecord transforms a single parquet record to MemoLog
func (d *dataTransformer) TransformParquetRecord(record ParquetMemoRecord) (*model.MemoLog, error) {
	if err := d.validateRecord(record); err != nil {
		d.updateStats(false)
		return nil, fmt.Errorf("invalid record: %w", err)
	}

	// Parse date from record
	parsedDate, err := d.parseDate(record.Date)
	if err != nil {
		d.updateStats(false)
		return nil, fmt.Errorf("failed to parse date '%s': %w", record.Date, err)
	}

	// Clean and validate authors
	cleanAuthors := d.cleanAuthors(record.Authors)
	if len(cleanAuthors) == 0 && !d.config.SkipInvalidRecords {
		d.updateStats(false)
		return nil, fmt.Errorf("no valid authors found in record")
	}

	// Clean and validate tags
	cleanTags := d.cleanTags(record.Tags)

	// Truncate content if necessary
	content := d.truncateContent(record.Content)

	// Parse reward as decimal
	reward, err := decimal.NewFromString(d.config.DefaultReward)
	if err != nil {
		reward = decimal.NewFromInt(25) // Fallback to 25 if parsing fails
	}

	// Create MemoLog model
	memoLog := &model.MemoLog{
		Title:               strings.TrimSpace(record.Title),
		Description:         content,
		URL:                 strings.TrimSpace(record.URL),
		Tags:                model.JSONArrayString(cleanTags),
		PublishedAt:         &parsedDate,
		Reward:              reward,
		Category:            pq.StringArray(d.config.DefaultCategory),
		DiscordAccountIDs:   model.JSONArrayString(cleanAuthors), // Store authors as Discord account IDs
		AuthorMemoUsernames: cleanAuthors,                        // Store usernames for response
	}

	// Set BaseModel fields
	now := time.Now()
	memoLog.CreatedAt = now
	memoLog.UpdatedAt = &now

	d.updateStats(true)
	return memoLog, nil
}

// TransformParquetRecords transforms a batch of parquet records
func (d *dataTransformer) TransformParquetRecords(records []ParquetMemoRecord) ([]model.MemoLog, error) {
	if len(records) == 0 {
		return []model.MemoLog{}, nil
	}

	var results []model.MemoLog
	var errors []string

	for i, record := range records {
		memoLog, err := d.TransformParquetRecord(record)
		if err != nil {
			errorMsg := fmt.Sprintf("record %d: %v", i, err)
			errors = append(errors, errorMsg)
			
			if !d.config.SkipInvalidRecords {
				return nil, fmt.Errorf("batch transformation failed at record %d: %w", i, err)
			}
			continue
		}
		
		results = append(results, *memoLog)
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all records failed transformation: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// GetTransformationStats returns current transformation statistics
func (d *dataTransformer) GetTransformationStats() DataTransformationStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stats
}

// UpdateConfig updates the transformation configuration
func (d *dataTransformer) UpdateConfig(config DataTransformationConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config = config
}

// validateRecord validates a parquet record for transformation
func (d *dataTransformer) validateRecord(record ParquetMemoRecord) error {
	if strings.TrimSpace(record.Date) == "" {
		return fmt.Errorf("date field is required")
	}
	
	if strings.TrimSpace(record.Title) == "" {
		return fmt.Errorf("title field is required")
	}
	
	if d.config.EnableValidation {
		// Additional validation when enabled
		if len(record.Authors) == 0 {
			return fmt.Errorf("at least one author is required")
		}

		// Validate URL format if provided
		if record.URL != "" {
			url := strings.TrimSpace(record.URL)
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				return fmt.Errorf("invalid URL format: %s", url)
			}
		}
	}
	
	return nil
}

// parseDate attempts to parse date string using configured formats
func (d *dataTransformer) parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try each configured date format
	for _, format := range d.config.DateFormats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s' with any configured format", dateStr)
}

// cleanAuthors cleans and validates author list
func (d *dataTransformer) cleanAuthors(authors []string) []string {
	var cleanAuthors []string
	seen := make(map[string]bool)

	for _, author := range authors {
		// Clean author name
		cleanAuthor := strings.TrimSpace(author)
		if cleanAuthor == "" {
			continue
		}

		// Remove duplicates (case-insensitive)
		lowerAuthor := strings.ToLower(cleanAuthor)
		if seen[lowerAuthor] {
			continue
		}
		seen[lowerAuthor] = true

		// Validate author name format
		if d.isValidAuthorName(cleanAuthor) {
			cleanAuthors = append(cleanAuthors, cleanAuthor)
		}
	}

	return cleanAuthors
}

// cleanTags cleans and validates tag list
func (d *dataTransformer) cleanTags(tags []string) []string {
	var cleanTags []string
	seen := make(map[string]bool)

	for _, tag := range tags {
		// Clean tag
		cleanTag := strings.TrimSpace(tag)
		if cleanTag == "" {
			continue
		}

		// Convert to lowercase for consistency
		cleanTag = strings.ToLower(cleanTag)

		// Remove duplicates
		if seen[cleanTag] {
			continue
		}
		seen[cleanTag] = true

		// Validate tag format
		if d.isValidTag(cleanTag) {
			cleanTags = append(cleanTags, cleanTag)
		}
	}

	return cleanTags
}

// truncateContent truncates content to maximum allowed length
func (d *dataTransformer) truncateContent(content string) string {
	content = strings.TrimSpace(content)
	if len(content) <= d.config.MaxContentLength {
		return content
	}

	// If MaxContentLength is too small to accommodate truncation indicator, return empty
	if d.config.MaxContentLength < 15 {
		return ""
	}

	// Truncate and add indication
	truncated := content[:d.config.MaxContentLength-15] // Leave space for indicator
	truncated = strings.TrimSpace(truncated)
	truncated += "... [truncated]"

	return truncated
}

// isValidAuthorName validates author name format
func (d *dataTransformer) isValidAuthorName(author string) bool {
	// Basic validation rules for author names
	if len(author) < 1 || len(author) > 100 {
		return false
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"@everyone",
		"@here",
		"<script>",
		"javascript:",
	}

	lowerAuthor := strings.ToLower(author)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerAuthor, pattern) {
			return false
		}
	}

	return true
}

// isValidTag validates tag format
func (d *dataTransformer) isValidTag(tag string) bool {
	// Basic validation rules for tags
	if len(tag) < 1 || len(tag) > 50 {
		return false
	}

	// Tags should only contain alphanumeric characters, hyphens, and underscores
	for _, r := range tag {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}

	return true
}

// updateStats updates transformation statistics
func (d *dataTransformer) updateStats(success bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.stats.TotalTransformed++
	d.stats.LastTransformation = time.Now()
	
	if success {
		d.stats.SuccessfulTransforms++
	} else {
		d.stats.FailedTransforms++
	}
}

// GetSupportedDateFormats returns the list of supported date formats
func (d *dataTransformer) GetSupportedDateFormats() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return append([]string(nil), d.config.DateFormats...) // Return a copy
}

// ValidateSchema checks if parquet data matches expected schema
func (d *dataTransformer) ValidateSchema(sample []ParquetMemoRecord) error {
	if len(sample) == 0 {
		return fmt.Errorf("no sample data provided for schema validation")
	}

	// Check first few records for schema consistency
	checkCount := len(sample)
	if checkCount > 10 {
		checkCount = 10 // Check at most 10 records
	}

	_ = []string{"date", "title", "authors"} // Required fields for reference
	
	for i := 0; i < checkCount; i++ {
		record := sample[i]
		
		// Check required fields
		if record.Date == "" {
			return fmt.Errorf("record %d: missing required field 'date'", i)
		}
		if record.Title == "" {
			return fmt.Errorf("record %d: missing required field 'title'", i)
		}
		if len(record.Authors) == 0 {
			return fmt.Errorf("record %d: missing required field 'authors'", i)
		}
		
		// Validate date format
		_, err := d.parseDate(record.Date)
		if err != nil {
			return fmt.Errorf("record %d: invalid date format '%s'", i, record.Date)
		}
	}

	return nil
}

// TransformWithMetadata transforms records and returns additional metadata
func (d *dataTransformer) TransformWithMetadata(records []ParquetMemoRecord) ([]model.MemoLog, map[string]interface{}, error) {
	startTime := time.Now()
	
	results, err := d.TransformParquetRecords(records)
	
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	// Collect metadata
	metadata := map[string]interface{}{
		"input_records":      len(records),
		"output_records":     len(results),
		"processing_time_ms": duration.Milliseconds(),
		"success_rate":       0.0,
		"transformation_date": endTime,
	}
	
	if len(records) > 0 {
		metadata["success_rate"] = float64(len(results)) / float64(len(records)) * 100
	}
	
	// Add statistics
	stats := d.GetTransformationStats()
	metadata["total_transformed"] = stats.TotalTransformed
	metadata["successful_transforms"] = stats.SuccessfulTransforms
	metadata["failed_transforms"] = stats.FailedTransforms
	
	return results, metadata, err
}

// GetConfig returns the current transformation configuration
func (d *dataTransformer) GetConfig() DataTransformationConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// Reset resets transformation statistics
func (d *dataTransformer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stats = DataTransformationStats{
		TotalTransformed:     0,
		SuccessfulTransforms: 0,
		FailedTransforms:     0,
		LastTransformation:   time.Time{},
	}
}