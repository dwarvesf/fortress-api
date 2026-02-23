package timeutil

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
)

const dataFormat = "2006-01-02"

var months = [...]string{
	"january",
	"february",
	"march",
	"april",
	"may",
	"june",
	"july",
	"august",
	"september",
	"october",
	"november",
	"december",
}

// WeekdayDuration returns duration (in day) between 2 weekdays.
// mon - tue = 2 - 1 = 1
// mon - fri = 5 - 1 = 4
// mon - mon = 0 - 0 + 7 = 7
// mon - sun = 0 - 1 + 7 = 6
func WeekdayDuration(from, to time.Weekday) time.Duration {
	offset := to - from
	if from >= to {
		offset += 7
	}
	return time.Duration(offset) * 24 * time.Hour
}

// GetQuarterFromMonth get quarter from month
func GetQuarterFromMonth(m time.Month) int {
	switch {
	case m < 4:
		return 1
	case m < 7:
		return 2
	case m < 10:
		return 3
	default:
		return 4
	}
}

// IsCurrentMonth is a function checking whether a given month and year is current month
func IsCurrentMonth(m, y int) bool {
	if int(time.Now().Month()) != m {
		return false
	}
	if time.Now().Year() != y {
		return false
	}

	return true
}

// BeginningOfYear return first day of the year
func BeginningOfYear(year int) time.Time {
	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.Now().Location())
}

// EndOfYear return last day of the year
func EndOfYear(year int) time.Time {
	return BeginningOfYear(year).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// ParseStringToDate parse string input as date format to time.Time
func ParseStringToDate(s string) (*time.Time, error) {
	t, err := time.Parse(dataFormat, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ParseStringToDateWithFormat parse string input as date format to time.Time with input format
func ParseStringToDateWithFormat(s, format string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	t, err := time.Parse(format, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// IsSameDay indicate same day (ignore hour, minutes, seconds, ...)
func IsSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}

// ParseTimeToDateFormat parse time input to yyyy-mm-dd format
func ParseTimeToDateFormat(t *time.Time) string {
	str := t.String()
	results := strings.Split(str, " ")
	return results[0]
}

// LastDayOfMonth return value type time.Time
// of the last day from input month and year
func LastDayOfMonth(month, year int) time.Time {
	return now.New(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Now().Location())).EndOfMonth()
}

func FirstDayOfMonth(month, year int) time.Time {
	return now.New(time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Now().Location())).Time
}

func LastFridayOfMonth(month, year int) int {
	d := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	d = d.Add(-time.Duration((d.Weekday()+7-time.Friday)%7) * 24 * time.Hour)
	return d.Day()
}

func FormatDatetime(t time.Time) string {
	date := strconv.Itoa(t.Day())
	switch t.Day() % 10 {
	case 1:
		date += "st"
	case 2:
		date += "nd"
	case 3:
		date += "rd"
	default:
		date += "th"
	}
	return fmt.Sprintf("%v %v, %v", t.Month().String(), date, t.Year())
}

func LastMonthYear(month, year int) (int, int) {
	if month == 1 {
		return 12, year - 1
	}
	return month - 1, year
}

func ParseWithMultipleFormats(s string) (*time.Time, error) {
	formats := []string{time.RFC3339, "2006-01-02", "02-01-2006", "2-1-2006", "02/01/2006", "2/1/2006"}
	for i := range formats {
		t, err := time.Parse(formats[i], s)
		if err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf(`parsing time %s does not match any remaining time format`, s)
}

func GetMonthFromString(s string) (int, error) {
	for i := range months {
		if strings.ToLower(s) == months[i] {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf(`%s is not a month`, s)
}

func GetMonthAndYearOfNextMonth() (month, year int) {
	nextMonth := time.Now().AddDate(0, 1, 0)
	return int(nextMonth.Month()), nextMonth.Year()
}

func CountWeekendDays(from, to time.Time) int {
	if from.After(to) {
		temp := from
		from = to
		to = temp
	}
	weekend := 0
	for i := from; i.Before(to); i = i.AddDate(0, 0, 1) {
		if i.Weekday() == time.Saturday || i.Weekday() == time.Sunday {
			weekend++
		}
	}
	return weekend
}

// GetStartDayOfWeek get monday 00:00:00
// Example: today: 2023-05-12 07:34:21
// return 2023-05-08 00:00:00
// weekday value
// sunday = 0
// moday = 1
// tuesday = 2
// ...
func GetStartDayOfWeek(tm time.Time) time.Time {
	weekday := time.Duration(tm.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := tm.Date()
	currentStartDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local) // return 2023-05-12 00:00:00

	return currentStartDay.Add(-1 * (weekday - 1) * 24 * time.Hour)
}

// GetEndDayOfWeek get sunday 23:59:59
// Example: today: 2023-05-12 07:34:21
// return 2023-05-14 23:59:59
// weekday value
// sunday = 0
// moday = 1
// tuesday = 2
// ...
func GetEndDayOfWeek(tm time.Time) time.Time {
	weekday := time.Duration(tm.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := tm.Date()
	currentEndDay := time.Date(year, month, day, 23, 59, 59, 0, time.Local) // return 2023-05-12 00:00:00
	return currentEndDay.Add((7 - weekday) * 24 * time.Hour)
}

// in curl special character need to be convert to ascii
//
//	Example: curl --request GET \
//	  --url 'http://localhost:8200/api/v1/vault/55/transaction?start_time=2023-05-08T00%3A00%3A00%2B07%3A00&end_time=2023-05-14T23%3A59%3A59%2B07%3A00'
func FormatDateForCurl(isoTime string) string {
	isoTime = strings.ReplaceAll(isoTime, ":", "%3A")
	isoTime = strings.ReplaceAll(isoTime, "+", "%2B")
	return isoTime
}

// GetTimeRange parses the time range from a string
// Example time range format: 13/01/2020 - 17/01/2020
// Returns err if failed to parse
func GetTimeRange(timeString string) ([]*time.Time, error) {
	var startDate, endDate *time.Time

	// string input is not a time range
	if !isTimeRange(timeString) {
		startDate = tryParseTime(timeString)
		if startDate != nil {
			return []*time.Time{startDate}, nil
		}
		return nil, errors.New("cannot parse time")
	}

	// If in time range format
	ranges := strings.Split(timeString, "-")
	if len(ranges) < 2 {
		return nil, errors.New("cannot parse wrong time range format")
	}
	startDate = tryParseTime(ranges[0])
	endDate = tryParseTime(ranges[1])

	if startDate == nil || endDate == nil {
		return nil, errors.New("cannot parse time")
	}

	return []*time.Time{startDate, endDate}, nil
}

func isTimeRange(s string) bool {
	// This regexp will validate whether s string input is a time range
	// Example time range format: 13/01/2020 - 17/01/2020
	// Explain regexp:
	// (0[1-9] | [1-9] | [1-2][0-9] | 3[0-1]) / (0[1-9] | [1-9] | 1[0-2]) / (20[0-9]{2})
	// (01->09 OR 1->9 OR 10->29 OR 30->31) / 01->09 OR 1->9 OR 10->12 / 20(00->99)
	// (day) / (month) / (year)
	timeRangeRegexp := regexp.MustCompile(`(0[1-9]|[1-9]|[1-2][0-9]|3[0-1])/(0[1-9]|[1-9]|1[0-2])/(20[0-9]{2}) - (0[1-9]|[1-9]|[1-2][0-9]|3[0-1])/(0[1-9]|[1-9]|1[0-2])/(20[0-9]{2})`)
	return timeRangeRegexp.FindStringSubmatch(s) != nil
}

func tryParseTime(timeString string) *time.Time {
	timeString = strings.TrimSpace(timeString)
	time, err := time.Parse("2/1/2006", timeString)
	if err != nil {
		return nil
	}
	return &time
}

// FormatMonthYear formats a month string (YYYY-MM) to "Month Year" format
// Example: "2025-12" -> "December 2025"
func FormatMonthYear(month string) string {
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return month
	}
	year := parts[0]
	monthNum, err := strconv.Atoi(parts[1])
	if err != nil || monthNum < 1 || monthNum > 12 {
		return month
	}
	return fmt.Sprintf("%s %s", time.Month(monthNum).String(), year)
}

// GetWorkingDaysInMonth returns all weekdays (Mon-Fri) for a given month
func GetWorkingDaysInMonth(year, month int) []time.Time {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	var workingDays []time.Time
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		if IsWeekday(d) {
			workingDays = append(workingDays, d)
		}
	}
	return workingDays
}

// IsWeekday checks if a date is Monday-Friday
func IsWeekday(date time.Time) bool {
	return date.Weekday() != time.Saturday && date.Weekday() != time.Sunday
}

func ChunkDateRange(start, end time.Time) [][2]time.Time {
	var weeks [][2]time.Time

	currentDate := start
	for currentDate.Before(end) || currentDate.Equal(end) {
		if currentDate.Weekday() >= time.Monday && currentDate.Weekday() <= time.Friday {
			weekStart := currentDate
			for currentDate.Weekday() >= time.Monday && currentDate.Weekday() <= time.Friday && (currentDate.Before(end) || currentDate.Equal(end)) {
				currentDate = currentDate.AddDate(0, 0, 1)
			}
			weekEnd := currentDate.AddDate(0, 0, -1)

			weeks = append(weeks, [2]time.Time{weekStart, weekEnd})
		} else {
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return weeks
}
