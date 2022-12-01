package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
)

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

func LastYearMonth(year int, month time.Month) (int, time.Month) {
	if month == time.January {
		return year - 1, time.December
	}
	return year, month - 1
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

func IsSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
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

func LastDayOfMonth(year int, month time.Month) time.Time {
	return now.New(time.Date(year, month, 1, 0, 0, 0, 0, time.Now().Location())).EndOfMonth()
}

func GetMonthFromString(s string) (int, error) {
	for i := range months {
		if strings.ToLower(s) == months[i] {
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf(`%s is not a month`, s)
}

func GetNextMonthInfo() (int, int) {
	now := time.Now()
	nxtMonth := now.AddDate(0, 1, 0)
	return int(nxtMonth.Month()), nxtMonth.Year()
}
