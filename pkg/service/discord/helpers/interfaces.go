package helpers

import (
	"context"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// BreakdownDetectionConfig defines configuration for breakdown detection
type BreakdownDetectionConfig struct {
	TitleKeywords          []string `json:"title_keywords"`
	TagKeywords            []string `json:"tag_keywords"`
	CaseSensitive          bool     `json:"case_sensitive"`
	RequireBothTitleAndTag bool     `json:"require_both_title_and_tag"`
}

// BreakdownDetectionStats tracks breakdown detection statistics
type BreakdownDetectionStats struct {
	TotalChecked    int       `json:"total_checked"`
	TotalBreakdowns int       `json:"total_breakdowns"`
	DetectionRate   float64   `json:"detection_rate"`
	TitleMatches    int       `json:"title_matches"`
	TagMatches      int       `json:"tag_matches"`
	LastUpdated     time.Time `json:"last_updated"`
}

// BreakdownDetector identifies breakdown posts from memo logs
type BreakdownDetector interface {
	// Detect breakdowns from a list of memos
	DetectBreakdowns(memos []model.MemoLog) []model.MemoLog
	
	// Check if a single memo is a breakdown
	IsBreakdown(title string, tags []string) bool
	
	// Get breakdown detection statistics
	GetDetectionStats() BreakdownDetectionStats
	
	// Update detection configuration
	UpdateConfig(config BreakdownDetectionConfig)
}

// AuthorCacheEntry represents a cached author resolution
type AuthorCacheEntry struct {
	DiscordID    string    `json:"discord_id"`
	ResolvedAt   time.Time `json:"resolved_at"`
	IsValid      bool      `json:"is_valid"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// AuthorResolutionConfig defines configuration for author resolution
type AuthorResolutionConfig struct {
	CacheTTL          time.Duration `json:"cache_ttl"`
	MaxCacheSize      int           `json:"max_cache_size"`
	WarmupEnabled     bool          `json:"warmup_enabled"`
	BatchSize         int           `json:"batch_size"`
	DatabaseTimeout   time.Duration `json:"database_timeout"`
	FallbackStrategy  string        `json:"fallback_strategy"` // "plain_username", "unverified_format", "mention_attempt"
	LogMissingAuthors bool          `json:"log_missing_authors"`
}

// AuthorResolutionMetrics tracks author resolution performance
type AuthorResolutionMetrics struct {
	CacheHitRate        float64   `json:"cache_hit_rate"`
	ResolutionErrorRate float64   `json:"resolution_error_rate"`
	CacheSize          int       `json:"cache_size"`
	DatabaseQueryCount  int64     `json:"database_query_count"`
	CacheEvictions     int64     `json:"cache_evictions"`
	LastWarmupTime     time.Time `json:"last_warmup_time"`
}

// AuthorResolver maps usernames to Discord IDs with caching
type AuthorResolver interface {
	// Resolve multiple usernames to Discord IDs in single call
	ResolveAuthorsToDiscordIDs(usernames []string) ([]string, error)
	
	// Resolve single username with detailed error information  
	ResolveUsernameToDiscordID(username string) (string, error)
	
	// Resolve all authors from memo list
	ResolveAllAuthors(memos []model.MemoLog) (map[string]string, error)
	
	// Pre-populate cache with known active authors
	WarmCache(ctx context.Context) error
	
	// Clear cache (for testing or memory management)
	ClearCache()
	
	// Get resolution metrics
	GetMetrics() AuthorResolutionMetrics
}

// LeaderboardEntry represents a single leaderboard entry
type LeaderboardEntry struct {
	Username       string `json:"username"`
	DiscordID      string `json:"discord_id"`
	BreakdownCount int    `json:"breakdown_count"`
	TotalPosts     int    `json:"total_posts"`
	Score          int    `json:"score"`
	Rank          int    `json:"rank"`
}

// ScoringFunction defines how to score a memo for leaderboard purposes
type ScoringFunction func(memo model.MemoLog) int

// LeaderboardConfig defines configuration for leaderboard generation
type LeaderboardConfig struct {
	MaxEntries     int  `json:"max_entries"`
	ShowZeroScores bool `json:"show_zero_scores"`
	SortDescending bool `json:"sort_descending"`
	TieBreaking    struct {
		UseAlphabetical bool `json:"use_alphabetical"`
		UseTimestamp    bool `json:"use_timestamp"`
	} `json:"tie_breaking"`
}

// LeaderboardBuilder generates author rankings and leaderboards
type LeaderboardBuilder interface {
	// Build leaderboard from breakdown posts only
	BuildFromBreakdowns(breakdowns []model.MemoLog) []LeaderboardEntry
	
	// Build leaderboard from all posts (for future use)
	BuildFromAllPosts(memos []model.MemoLog) []LeaderboardEntry
	
	// Build leaderboard with custom scoring function
	BuildWithCustomScoring(memos []model.MemoLog, scoringFunc ScoringFunction) []LeaderboardEntry
	
	// Get leaderboard configuration
	GetConfig() LeaderboardConfig
	
	// Update leaderboard configuration
	UpdateConfig(config LeaderboardConfig)
}

// TimeCalculatorConfig defines configuration for time calculations
type TimeCalculatorConfig struct {
	WeekStartDay  time.Weekday `json:"week_start_day"`    // Default: Monday
	MonthStartDay int          `json:"month_start_day"`   // Default: 1 (first day)
	DateFormat    string       `json:"date_format"`       // Default: "2-Jan"
	RangeFormat   string       `json:"range_format"`      // Default: "%s-%s"
	TimeZone      string       `json:"timezone"`          // Default: "UTC"
}

// TimeCalculator handles date range and period calculations
type TimeCalculator interface {
	// Calculate time ranges for reports
	CalculateWeeklyRange() (start, end time.Time, rangeStr string)
	CalculateMonthlyRange() (start, end time.Time, rangeStr string)
	
	// Get comparison periods for new author detection
	GetWeeklyComparisonPeriod() (start, end time.Time)
	GetMonthlyComparisonPeriod() (start, end time.Time)
	
	// Custom range calculation
	CalculateCustomRange(period string, offset int) (start, end time.Time, rangeStr string, err error)
	
	// Utility methods
	FormatDateRange(start, end time.Time) string
	GetCurrentWeekNumber() int
	GetCurrentMonth() string
	
	// Additional utility methods for enhanced functionality
	GetTimeZone() string
	IsCurrentWeek(timestamp time.Time) bool
	IsCurrentMonth(timestamp time.Time) bool
	GetDayOfWeek() time.Weekday
	FormatTimestamp(timestamp time.Time) string
}

// NewAuthorDetectionStats tracks new author detection statistics
type NewAuthorDetectionStats struct {
	LastDetectionTime   time.Time `json:"last_detection_time"`
	WeeklyNewAuthors    int       `json:"weekly_new_authors"`
	MonthlyNewAuthors   int       `json:"monthly_new_authors"`
	HistoricalCacheSize int       `json:"historical_cache_size"`
}

// NewAuthorDetector identifies first-time contributors in a given period
type NewAuthorDetector interface {
	// Detect new authors in current period vs historical period
	DetectNewAuthors(currentMemos []model.MemoLog, period string) ([]string, error)
	
	// Get historical author set for comparison
	GetHistoricalAuthors(period string) (map[string]bool, error)
	
	// Clear historical author cache
	ClearHistoricalCache()
	
	// Get detection statistics
	GetDetectionStats() NewAuthorDetectionStats
}

// ReportType defines the type of report being generated
type ReportType string

const (
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

// ReportData contains all data needed to format a report
type ReportData struct {
	Type           ReportType           `json:"type"`
	TimeRange      string               `json:"time_range"`
	AllPosts       []model.MemoLog      `json:"all_posts"`
	Breakdowns     []model.MemoLog      `json:"breakdowns"`
	NewAuthors     []string             `json:"new_authors"`
	Leaderboard    []LeaderboardEntry   `json:"leaderboard"`
	AuthorMappings map[string]string    `json:"author_mappings"`
	TotalICY       int                  `json:"total_icy,omitempty"`
	GeneratedAt    time.Time            `json:"generated_at"`
}

// MessageFormattingConfig defines configuration for Discord message formatting
type MessageFormattingConfig struct {
	MaxFieldLength   int    `json:"max_field_length"`
	MaxEmbedFields   int    `json:"max_embed_fields"`
	TruncationSuffix string `json:"truncation_suffix"`
	DateFormat       string `json:"date_format"`
}

// MessageFormatter formats Discord embed messages
type MessageFormatter interface {
	// Format weekly report into Discord embed
	FormatWeeklyReport(data ReportData) (*DiscordEmbed, error)
	
	// Format monthly report into Discord embed
	FormatMonthlyReport(data ReportData) (*DiscordEmbed, error)
	
	// Get formatter configuration
	GetConfig() MessageFormattingConfig
	
	// Update formatter configuration
	UpdateConfig(config MessageFormattingConfig)
}

// DiscordEmbed represents a simplified Discord embed structure
// We define our own to avoid importing discordgo in helpers
type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Fields      []DiscordEmbedField `json:"fields"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Timestamp   *time.Time          `json:"timestamp,omitempty"`
}

// DiscordEmbedField represents a Discord embed field
type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// DiscordEmbedFooter represents a Discord embed footer
type DiscordEmbedFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// ParquetMemoRecord represents a record from parquet data
type ParquetMemoRecord struct {
	Date    string   `json:"date"`
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	Tags    []string `json:"tags"`
	URL     string   `json:"url"`
	Content string   `json:"content"`
}

// DataTransformationConfig defines configuration for data transformation
type DataTransformationConfig struct {
	DateFormats                []string `json:"date_formats"`
	DefaultReward              string   `json:"default_reward"`
	DefaultCategory            []string `json:"default_category"`
	EnableValidation           bool     `json:"enable_validation"`
	SkipInvalidRecords         bool     `json:"skip_invalid_records"`
	MaxContentLength           int      `json:"max_content_length"`
	AuthorResolutionRetries    int      `json:"author_resolution_retries"`
}

// DataTransformer transforms parquet data to MemoLog models
type DataTransformer interface {
	// Transform single parquet record to MemoLog
	TransformParquetRecord(record ParquetMemoRecord) (*model.MemoLog, error)
	
	// Transform batch of parquet records
	TransformParquetRecords(records []ParquetMemoRecord) ([]model.MemoLog, error)
	
	// Get transformation statistics
	GetTransformationStats() DataTransformationStats
	
	// Update transformation configuration
	UpdateConfig(config DataTransformationConfig)
}

// DataTransformationStats tracks data transformation statistics
type DataTransformationStats struct {
	TotalTransformed    int       `json:"total_transformed"`
	SuccessfulTransforms int      `json:"successful_transforms"`
	FailedTransforms    int       `json:"failed_transforms"`
	LastTransformation  time.Time `json:"last_transformation"`
}

// Constructor functions for all helpers
// Note: Actual implementations are in their respective files

// Type aliases for constructor parameters
type Store = store.Store
type DBRepo = store.DBRepo