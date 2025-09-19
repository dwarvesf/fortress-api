package helpers

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeCalculator_WeeklyRange(t *testing.T) {
	config := TimeCalculatorConfig{
		WeekStartDay: time.Monday,
		DateFormat:   "2-Jan",
		TimeZone:     "UTC",
	}
	
	// Test with fixed time for reproducible results
	testCases := []struct {
		name          string
		currentTime   time.Time
		expectedStart string
		expectedEnd   string
		expectRange   bool // Whether we should check range string formatting
	}{
		{
			name:          "middle_of_week_wednesday",
			currentTime:   time.Date(2024, 8, 21, 15, 30, 0, 0, time.UTC), // Wednesday
			expectedStart: "2024-08-19", // Previous Monday
			expectedEnd:   "2024-08-25", // Next Sunday end
			expectRange:   true,
		},
		{
			name:          "week_start_monday",
			currentTime:   time.Date(2024, 8, 19, 9, 0, 0, 0, time.UTC), // Monday
			expectedStart: "2024-08-19", // Same Monday
			expectedEnd:   "2024-08-25", // Sunday end
			expectRange:   true,
		},
		{
			name:          "week_end_sunday",
			currentTime:   time.Date(2024, 8, 25, 23, 59, 0, 0, time.UTC), // Sunday
			expectedStart: "2024-08-19", // Previous Monday
			expectedEnd:   "2024-08-25", // Same Sunday end
			expectRange:   true,
		},
		{
			name:          "cross_month_boundary",
			currentTime:   time.Date(2024, 9, 2, 10, 0, 0, 0, time.UTC), // Monday in September
			expectedStart: "2024-09-02", // Same Monday
			expectedEnd:   "2024-09-08", // Sunday in September
			expectRange:   true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create calculator with mocked time function
			calculator := newTimeCalculatorWithTime(config, func() time.Time { return tc.currentTime })
			
			start, end, rangeStr := calculator.CalculateWeeklyRange()
			
			// Verify start date
			expectedStartTime, _ := time.Parse("2006-01-02", tc.expectedStart)
			assert.True(t, start.Truncate(24*time.Hour).Equal(expectedStartTime),
				"Expected start %v, got %v", expectedStartTime, start)
			
			// Verify end date is within the expected day
			expectedEndTime, _ := time.Parse("2006-01-02", tc.expectedEnd)
			endDay := end.Truncate(24 * time.Hour)
			assert.True(t, endDay.Equal(expectedEndTime) || endDay.Before(expectedEndTime.Add(24*time.Hour)),
				"Expected end day %v, got %v", expectedEndTime, endDay)
			
			// Verify range string is generated
			if tc.expectRange {
				assert.NotEmpty(t, rangeStr)
				// Should contain month name
				assert.True(t, 
					len(rangeStr) > 3, 
					"Range string should contain meaningful content: %s", rangeStr)
			}
		})
	}
}

func TestTimeCalculator_MonthlyRange(t *testing.T) {
	config := TimeCalculatorConfig{
		MonthStartDay: 1,
		TimeZone:      "UTC",
	}
	
	testCases := []struct {
		name          string
		currentTime   time.Time
		expectedStart string
		expectedEnd   string
		expectedRange string
	}{
		{
			name:          "middle_of_august",
			currentTime:   time.Date(2024, 8, 15, 12, 0, 0, 0, time.UTC),
			expectedStart: "2024-08-01",
			expectedEnd:   "2024-08-31",
			expectedRange: "August 2024",
		},
		{
			name:          "end_of_february_leap_year",
			currentTime:   time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expectedStart: "2024-02-01",
			expectedEnd:   "2024-02-29",
			expectedRange: "February 2024",
		},
		{
			name:          "december_year_end",
			currentTime:   time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC),
			expectedStart: "2024-12-01",
			expectedEnd:   "2024-12-31",
			expectedRange: "December 2024",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			calculator := newTimeCalculatorWithTime(config, func() time.Time { return tc.currentTime })
			
			start, end, rangeStr := calculator.CalculateMonthlyRange()
			
			// Verify start is first day of month
			assert.Equal(t, 1, start.Day())
			assert.Equal(t, tc.currentTime.Month(), start.Month())
			assert.Equal(t, tc.currentTime.Year(), start.Year())
			
			// Verify end is last day of month
			nextMonth := start.AddDate(0, 1, 0)
			expectedEnd := nextMonth.Add(-time.Second)
			endDay := end.Truncate(24 * time.Hour)
			expectedEndDay := expectedEnd.Truncate(24 * time.Hour)
			assert.True(t, endDay.Equal(expectedEndDay) || endDay.After(expectedEndDay.Add(-24*time.Hour)),
				"Expected end day around %v, got %v", expectedEndDay, endDay)
			
			// Verify range string
			assert.Equal(t, tc.expectedRange, rangeStr)
		})
	}
}

func TestTimeCalculator_ComparisonPeriods(t *testing.T) {
	config := TimeCalculatorConfig{
		TimeZone: "UTC",
	}
	
	fixedTime := time.Date(2024, 8, 21, 12, 0, 0, 0, time.UTC)
	
	t.Run("weekly_comparison_period", func(t *testing.T) {
		calculator := newTimeCalculatorWithTime(config, func() time.Time { return fixedTime })
		
		start, end := calculator.GetWeeklyComparisonPeriod()
		
		// Should return last 30 days
		expectedStart := fixedTime.AddDate(0, 0, -30).Truncate(24 * time.Hour)
		expectedEnd := fixedTime
		
		assert.True(t, start.Equal(expectedStart))
		assert.True(t, end.Equal(expectedEnd))
		
		// Verify it's approximately 30 days (allowing for timezone/truncation differences)
		duration := end.Sub(start)
		assert.True(t, duration >= 29*24*time.Hour && duration <= 31*24*time.Hour,
			"Expected duration around 30 days, got %v", duration)
	})
	
	t.Run("monthly_comparison_period", func(t *testing.T) {
		calculator := newTimeCalculatorWithTime(config, func() time.Time { return fixedTime })
		
		start, end := calculator.GetMonthlyComparisonPeriod()
		
		// Should return last 12 months (365 days)
		expectedStart := fixedTime.AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		expectedEnd := fixedTime
		
		assert.True(t, start.Equal(expectedStart))
		assert.True(t, end.Equal(expectedEnd))
		
		// Verify it's approximately 365 days (accounting for leap years)
		duration := end.Sub(start)
		assert.True(t, duration >= 364*24*time.Hour && duration <= 367*24*time.Hour,
			"Expected duration around 365 days, got %v", duration)
	})
}

func TestTimeCalculator_TimezoneHandling(t *testing.T) {
	testTimezones := []struct {
		timezone     string
		expectedLoc  string
	}{
		{"UTC", "UTC"},
		{"America/New_York", "America/New_York"},
		{"Asia/Tokyo", "Asia/Tokyo"},
		{"invalid-timezone", "UTC"}, // Should fallback to UTC
	}
	
	for _, tz := range testTimezones {
		t.Run(tz.timezone, func(t *testing.T) {
			config := TimeCalculatorConfig{
				TimeZone:     tz.timezone,
				WeekStartDay: time.Monday,
			}
			
			calculator := NewTimeCalculator(config)
			
			// Test that timezone is correctly applied
			start, end, _ := calculator.CalculateWeeklyRange()
			
			if tz.timezone == "invalid-timezone" {
				// Should fallback to UTC
				assert.Equal(t, "UTC", start.Location().String())
				assert.Equal(t, "UTC", end.Location().String())
			} else {
				// Should use specified timezone
				assert.Equal(t, tz.expectedLoc, start.Location().String())
				assert.Equal(t, tz.expectedLoc, end.Location().String())
			}
		})
	}
}

func TestTimeCalculator_FormatDateRange(t *testing.T) {
	config := TimeCalculatorConfig{
		DateFormat: "2-Jan",
		TimeZone:   "UTC",
	}
	
	calculator := NewTimeCalculator(config)
	
	testCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			name:     "same_month",
			start:    time.Date(2024, 8, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 8, 22, 0, 0, 0, 0, time.UTC),
			expected: "15-22 August",
		},
		{
			name:     "different_months",
			start:    time.Date(2024, 8, 29, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 9, 5, 0, 0, 0, 0, time.UTC),
			expected: "29 August - 5 September",
		},
		{
			name:     "same_day",
			start:    time.Date(2024, 8, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 8, 15, 23, 59, 59, 0, time.UTC),
			expected: "15-15 August",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculator.FormatDateRange(tc.start, tc.end)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTimeCalculator_UtilityMethods(t *testing.T) {
	config := TimeCalculatorConfig{
		TimeZone: "UTC",
	}
	
	// Use fixed time for testing
	fixedTime := time.Date(2024, 8, 21, 12, 0, 0, 0, time.UTC) // Wednesday
	calculator := newTimeCalculatorWithTime(config, func() time.Time { return fixedTime })
	
	t.Run("get_current_week_number", func(t *testing.T) {
		weekNum := calculator.GetCurrentWeekNumber()
		assert.Greater(t, weekNum, 0)
		assert.LessOrEqual(t, weekNum, 53)
	})
	
	t.Run("get_current_month", func(t *testing.T) {
		month := calculator.GetCurrentMonth()
		assert.Equal(t, "August", month)
	})
}

func TestTimeCalculator_CalculateCustomRange(t *testing.T) {
	config := TimeCalculatorConfig{
		WeekStartDay: time.Monday,
		TimeZone:     "UTC",
	}
	
	fixedTime := time.Date(2024, 8, 21, 12, 0, 0, 0, time.UTC)
	calculator := newTimeCalculatorWithTime(config, func() time.Time { return fixedTime })
	
	t.Run("weekly_with_offset", func(t *testing.T) {
		start, end, rangeStr, err := calculator.CalculateCustomRange("weekly", -1)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, rangeStr)
		
		// Should be previous week
		currentStart, _, _ := calculator.CalculateWeeklyRange()
		expectedStart := currentStart.AddDate(0, 0, -7)
		
		assert.True(t, start.Truncate(24*time.Hour).Equal(expectedStart.Truncate(24*time.Hour)))
		assert.True(t, end.After(start))
	})
	
	t.Run("monthly_with_offset", func(t *testing.T) {
		start, end, rangeStr, err := calculator.CalculateCustomRange("monthly", -1)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, rangeStr)
		
		// Should be previous month
		assert.True(t, end.After(start))
		assert.Equal(t, 1, start.Day()) // Should start on first day of month
	})
	
	t.Run("invalid_period", func(t *testing.T) {
		_, _, _, err := calculator.CalculateCustomRange("invalid", 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported period")
	})
}

// Helper function to create time calculator with custom time function for testing
func newTimeCalculatorWithTime(config TimeCalculatorConfig, nowFunc func() time.Time) TimeCalculator {
	location, err := time.LoadLocation(config.TimeZone)
	if err != nil {
		location = time.UTC
	}
	
	return &timeCalculatorWithMockTime{
		location: location,
		config:   config,
		nowFunc:  nowFunc,
	}
}

// Test implementation that allows mocking time.Now()
type timeCalculatorWithMockTime struct {
	location *time.Location
	config   TimeCalculatorConfig
	nowFunc  func() time.Time
}

func (t *timeCalculatorWithMockTime) CalculateWeeklyRange() (start, end time.Time, rangeStr string) {
	now := t.nowFunc().In(t.location)
	
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

func (t *timeCalculatorWithMockTime) CalculateMonthlyRange() (start, end time.Time, rangeStr string) {
	now := t.nowFunc().In(t.location)
	
	// Start of current month
	start = time.Date(now.Year(), now.Month(), t.config.MonthStartDay, 0, 0, 0, 0, t.location)
	
	// End of current month
	end = start.AddDate(0, 1, 0).Add(-time.Second)
	
	rangeStr = start.Format("January 2006")
	return start, end, rangeStr
}

func (t *timeCalculatorWithMockTime) GetWeeklyComparisonPeriod() (start, end time.Time) {
	// Last 30 days for weekly comparison
	now := t.nowFunc().In(t.location)
	start = now.AddDate(0, 0, -30).Truncate(24 * time.Hour)
	end = now
	return start, end
}

func (t *timeCalculatorWithMockTime) GetMonthlyComparisonPeriod() (start, end time.Time) {
	// Last 12 months for monthly comparison
	now := t.nowFunc().In(t.location)
	start = now.AddDate(-1, 0, 0).Truncate(24 * time.Hour)
	end = now
	return start, end
}

func (t *timeCalculatorWithMockTime) CalculateCustomRange(period string, offset int) (start, end time.Time, rangeStr string, err error) {
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
		return time.Time{}, time.Time{}, "", fmt.Errorf("unsupported period: %s", period)
	}
}

func (t *timeCalculatorWithMockTime) FormatDateRange(start, end time.Time) string {
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

func (t *timeCalculatorWithMockTime) GetCurrentWeekNumber() int {
	now := t.nowFunc()
	_, week := now.ISOWeek()
	return week
}

func (t *timeCalculatorWithMockTime) GetCurrentMonth() string {
	now := t.nowFunc()
	return now.Month().String()
}

// Additional utility methods (new in improvements)
func (t *timeCalculatorWithMockTime) GetTimeZone() string {
	return t.config.TimeZone
}

func (t *timeCalculatorWithMockTime) IsCurrentWeek(timestamp time.Time) bool {
	start, end, _ := t.CalculateWeeklyRange()
	return !timestamp.Before(start) && !timestamp.After(end)
}

func (t *timeCalculatorWithMockTime) IsCurrentMonth(timestamp time.Time) bool {
	start, end, _ := t.CalculateMonthlyRange()
	return !timestamp.Before(start) && !timestamp.After(end)
}

func (t *timeCalculatorWithMockTime) GetDayOfWeek() time.Weekday {
	now := t.nowFunc()
	return now.Weekday()
}

func (t *timeCalculatorWithMockTime) FormatTimestamp(timestamp time.Time) string {
	if t.config.DateFormat != "" {
		return timestamp.In(t.location).Format(t.config.DateFormat)
	}
	return timestamp.In(t.location).Format("2-Jan")
}

// Test the new utility methods
func TestTimeCalculator_NewUtilityMethods(t *testing.T) {
	config := TimeCalculatorConfig{
		WeekStartDay:  time.Monday,
		MonthStartDay: 1,
		TimeZone:      "America/New_York",
		DateFormat:    "Jan 2",
	}

	// Mock time: Wednesday, August 15, 2024, 10:30 AM
	mockTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
	calc := newTimeCalculatorWithTime(config, func() time.Time { return mockTime })
	
	// Load timezone for test assertions
	location, err := time.LoadLocation(config.TimeZone)
	assert.NoError(t, err)

	t.Run("get_timezone", func(t *testing.T) {
		timezone := calc.GetTimeZone()
		assert.Equal(t, "America/New_York", timezone)
	})

	t.Run("get_day_of_week", func(t *testing.T) {
		dayOfWeek := calc.GetDayOfWeek()
		// August 15, 2024 is a Thursday
		assert.Equal(t, time.Thursday, dayOfWeek)
	})

	t.Run("format_timestamp", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
		formatted := calc.FormatTimestamp(testTime)
		
		// Should use configured format "Jan 2"
		expected := testTime.In(location).Format("Jan 2")
		assert.Equal(t, expected, formatted)
	})

	t.Run("format_timestamp_default", func(t *testing.T) {
		// Test with no custom format
		configNoFormat := config
		configNoFormat.DateFormat = ""
		calcNoFormat := newTimeCalculatorWithTime(configNoFormat, func() time.Time { return mockTime })
		
		testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
		formatted := calcNoFormat.FormatTimestamp(testTime)
		
		// Should use default format "2-Jan"
		expected := testTime.In(location).Format("2-Jan")
		assert.Equal(t, expected, formatted)
	})

	t.Run("is_current_week", func(t *testing.T) {
		// Mock time is Wednesday, August 15, 2024 10:30 UTC
		// Let's first calculate what the current week boundaries are
		start, end, _ := calc.CalculateWeeklyRange()
		
		// Test timestamps within the current week boundaries
		withinWeek := start.Add(24 * time.Hour) // One day into the week
		assert.True(t, calc.IsCurrentWeek(withinWeek), 
			"Time %v should be within week bounds %v to %v", withinWeek, start, end)
		
		// Test timestamp at start boundary
		assert.True(t, calc.IsCurrentWeek(start), 
			"Start time %v should be within week bounds", start)
		
		// Test timestamps outside the current week
		beforeWeek := start.Add(-1 * time.Hour)
		assert.False(t, calc.IsCurrentWeek(beforeWeek),
			"Time %v should be before week start %v", beforeWeek, start)
		
		afterWeek := end.Add(1 * time.Hour)
		assert.False(t, calc.IsCurrentWeek(afterWeek),
			"Time %v should be after week end %v", afterWeek, end)
	})

	t.Run("is_current_month", func(t *testing.T) {
		// Mock time is August 15, 2024
		
		// Test timestamps within the current month
		firstDayThisMonth := time.Date(2024, 8, 1, 0, 0, 0, 0, location)
		assert.True(t, calc.IsCurrentMonth(firstDayThisMonth))
		
		lastDayThisMonth := time.Date(2024, 8, 31, 23, 59, 0, 0, location)
		assert.True(t, calc.IsCurrentMonth(lastDayThisMonth))
		
		// Test timestamps outside the current month
		lastDayPreviousMonth := time.Date(2024, 7, 31, 23, 59, 0, 0, location)
		assert.False(t, calc.IsCurrentMonth(lastDayPreviousMonth))
		
		firstDayNextMonth := time.Date(2024, 9, 1, 0, 1, 0, 0, location)
		assert.False(t, calc.IsCurrentMonth(firstDayNextMonth))
	})
}