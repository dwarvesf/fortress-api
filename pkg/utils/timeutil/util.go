package timeutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
)

const dataFormat = "2006-02-01"

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

//GetQuarterFromMonth get quarter from month
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

//BeginningOfYear return first day of the year
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
	return nil, errors.New(fmt.Sprintf(`parsing time %s does not match any remaining time format`, s))
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
