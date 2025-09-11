package helpers

import (
	"fmt"
	"time"
)

type timeCalculator struct {
	location *time.Location
	config   TimeCalculatorConfig
}

// NewTimeCalculator creates a new time calculator with the given configuration.
// Handles timezone loading with automatic fallback to UTC for invalid timezones.
// The calculator provides consistent time range calculations across different periods.
func NewTimeCalculator(config TimeCalculatorConfig) TimeCalculator {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		location = time.UTC // Fallback to UTC
	}
	
	return &timeCalculator{
		location: location,
		config:   config,
	}
}

// CalculateWeeklyRange calculates the start and end times for the current week.
// Uses the configured week start day (typically Monday) and returns times in the configured timezone.
// Returns the calculated range along with a formatted string representation.
func (t *timeCalculator) CalculateWeeklyRange() (start, end time.Time, rangeStr string) {
	now := time.Now().In(t.location)
	
	// Find the start of the current week
	daysSinceWeekStart := int(now.Weekday() - t.config.WeekStartDay)
	if daysSinceWeekStart < 0 {
		daysSinceWeekStart += 7
	}
	
	start = now.AddDate(0, 0, -daysSinceWeekStart).Truncate(24 * time.Hour)
	end = start.AddDate(0, 0, 7).Add(-time.Second) // End of week
	
	rangeStr = t.FormatDateRange(start, end)
	return start, end, rangeStr
}

// CalculateMonthlyRange calculates the start and end times for the current month.
// Uses the configured month start day and handles month boundaries correctly.
// Returns the calculated range along with a formatted string representation.
func (t *timeCalculator) CalculateMonthlyRange() (start, end time.Time, rangeStr string) {
	now := time.Now().In(t.location)
	
	// Start of current month
	start = time.Date(now.Year(), now.Month(), t.config.MonthStartDay, 0, 0, 0, 0, t.location)
	
	// End of current month
	end = start.AddDate(0, 1, 0).Add(-time.Second)
	
	rangeStr = start.Format("January 2006")
	return start, end, rangeStr
}

// GetWeeklyComparisonPeriod returns a 30-day historical period for weekly comparisons.
// Used primarily for new author detection to compare current week against recent history.
func (t *timeCalculator) GetWeeklyComparisonPeriod() (start, end time.Time) {
	// Last 30 days for weekly comparison
	now := time.Now().In(t.location)
	start = now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	end = now
	return start, end
}

// GetMonthlyComparisonPeriod returns a 12-month historical period for monthly comparisons.
// Used primarily for new author detection to compare current month against yearly history.
func (t *timeCalculator) GetMonthlyComparisonPeriod() (start, end time.Time) {
	// Last 12 months for monthly comparison
	now := time.Now().In(t.location)
	start = now.AddDate(-1, 0, 0).Truncate(24 * time.Hour)
	end = now
	return start, end
}

// CalculateCustomRange calculates time ranges with custom periods and offsets.
// Supports "weekly" and "monthly" periods with positive or negative offsets.
// Positive offsets move forward in time, negative offsets move backward.
// Returns sanitized error messages to prevent information disclosure.
func (t *timeCalculator) CalculateCustomRange(period string, offset int) (start, end time.Time, rangeStr string, err error) {
	switch period {
	case "weekly":
		currentStart, currentEnd, _ := t.CalculateWeeklyRange()
		start = currentStart.AddDate(0, 0, offset*7)
		end = currentEnd.AddDate(0, 0, offset*7)
		rangeStr = t.FormatDateRange(start, end)
		return start, end, rangeStr, nil
	case "monthly":
		currentStart, currentEnd, _ := t.CalculateMonthlyRange()
		start = currentStart.AddDate(0, offset, 0)
		end = currentEnd.AddDate(0, offset, 0)
		rangeStr = start.Format("January 2006")
		return start, end, rangeStr, nil
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("invalid period type specified")
	}
}

// FormatDateRange formats a date range into a human-readable string.
// Uses smart formatting: same month shows "1-15 January", different months show "30 Jan - 5 Feb".
func (t *timeCalculator) FormatDateRange(start, end time.Time) string {
	if start.Month() == end.Month() {
		return fmt.Sprintf("%d-%d %s", 
			start.Day(), 
			end.Day(), 
			start.Month().String())
	}
	
	return fmt.Sprintf("%d %s - %d %s",
		start.Day(),
		start.Month().String(),
		end.Day(), 
		end.Month().String())
}

// GetCurrentWeekNumber returns the current ISO week number (1-53) in the configured timezone.
// Uses ISO 8601 week numbering standard for consistency across different systems.
func (t *timeCalculator) GetCurrentWeekNumber() int {
	now := time.Now().In(t.location)
	_, week := now.ISOWeek()
	return week
}

// GetCurrentMonth returns the current month name in the configured timezone.
// Returns the full month name (e.g., "January", "February") for display purposes.
func (t *timeCalculator) GetCurrentMonth() string {
	now := time.Now().In(t.location)
	return now.Month().String()
}

// GetTimeZone returns the configured timezone string.
// Useful for debugging and displaying current timezone settings to users.
func (t *timeCalculator) GetTimeZone() string {
	return t.config.TimeZone
}

// IsCurrentWeek checks if the given timestamp falls within the current week boundaries.
// Uses the configured week start day and timezone for accurate boundary determination.
func (t *timeCalculator) IsCurrentWeek(timestamp time.Time) bool {
	start, end, _ := t.CalculateWeeklyRange()
	return !timestamp.Before(start) && !timestamp.After(end)
}

// IsCurrentMonth checks if the given timestamp falls within the current month boundaries.
// Uses the configured month start day and timezone for accurate boundary determination.
func (t *timeCalculator) IsCurrentMonth(timestamp time.Time) bool {
	start, end, _ := t.CalculateMonthlyRange()
	return !timestamp.Before(start) && !timestamp.After(end)
}

// GetDayOfWeek returns the current day of the week in the configured timezone.
// Returns time.Weekday (Sunday=0, Monday=1, ..., Saturday=6) for consistent weekday handling.
func (t *timeCalculator) GetDayOfWeek() time.Weekday {
	now := time.Now().In(t.location)
	return now.Weekday()
}

// FormatTimestamp formats any timestamp using the configured date format.
// Falls back to "2-Jan" format if no custom format is configured. Respects the configured timezone.
func (t *timeCalculator) FormatTimestamp(timestamp time.Time) string {
	if t.config.DateFormat != "" {
		return timestamp.In(t.location).Format(t.config.DateFormat)
	}
	return timestamp.In(t.location).Format("2-Jan")
}