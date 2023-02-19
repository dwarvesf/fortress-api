package timeutil

import (
	"testing"
	"time"
)

func TestLastDayOfMonth(t *testing.T) {
	testcases := []struct {
		name    string
		month   int
		year    int
		wantDay int
	}{
		{
			name:    "case last day is 31",
			month:   8,
			year:    2019,
			wantDay: 31,
		},
		{
			name:    "case last day is 30",
			month:   9,
			year:    2019,
			wantDay: 30,
		},
		{
			name:    "case last day is 29",
			month:   2,
			year:    2020,
			wantDay: 29,
		},
		{
			name:    "case last day is 28",
			month:   2,
			year:    2019,
			wantDay: 28,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out := LastDayOfMonth(tc.month, tc.year)
			if out.Day() != tc.wantDay {
				t.Errorf("timeutil.LastDayOfMonth() want output: %v, got output: %v", tc.wantDay, out.Day())
			}
		})
	}
}

func TestLastMonthYear(t *testing.T) {
	testcases := []struct {
		name      string
		month     int
		year      int
		wantMonth int
		wantYear  int
	}{
		{
			name:      "case same year",
			month:     8,
			year:      2019,
			wantMonth: 7,
			wantYear:  2019,
		},
		{
			name:      "case different year",
			month:     1,
			year:      2020,
			wantMonth: 12,
			wantYear:  2019,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m, y := LastMonthYear(tc.month, tc.year)
			if m != tc.wantMonth || y != tc.wantYear {
				t.Errorf("timeutil.LastMonthYear() want output: %v/%v, got output: %v/%v", tc.wantYear, tc.wantMonth, y, m)
			}
		})
	}
}

func TestCountWeekendDays(t *testing.T) {
	testcases := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantDay int
	}{
		{
			name:    "case in one month",
			from:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Now().Location()),
			to:      time.Date(2020, time.January, 31, 0, 0, 0, 0, time.Now().Location()),
			wantDay: 8,
		},
		{
			name:    "case from date is weekend",
			from:    time.Date(2020, time.January, 12, 0, 0, 0, 0, time.Now().Location()),
			to:      time.Date(2020, time.January, 31, 0, 0, 0, 0, time.Now().Location()),
			wantDay: 5,
		},
		{
			name:    "case two different months",
			from:    time.Date(2020, time.January, 15, 0, 0, 0, 0, time.Now().Location()),
			to:      time.Date(2020, time.February, 24, 0, 0, 0, 0, time.Now().Location()),
			wantDay: 12,
		},
		{
			name:    "case two different year",
			from:    time.Date(2019, time.December, 23, 0, 0, 0, 0, time.Now().Location()),
			to:      time.Date(2020, time.January, 17, 0, 0, 0, 0, time.Now().Location()),
			wantDay: 6,
		},
		{
			name:    "case from > to",
			from:    time.Date(2020, time.January, 17, 0, 0, 0, 0, time.Now().Location()),
			to:      time.Date(2019, time.December, 23, 0, 0, 0, 0, time.Now().Location()),
			wantDay: 6,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			days := CountWeekendDays(tc.from, tc.to)
			if days != tc.wantDay {
				t.Errorf("timeutil.CountWeekendDays() want output: %v, got output: %v", tc.wantDay, days)
			}
		})
	}
}
