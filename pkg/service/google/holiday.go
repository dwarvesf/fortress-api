package google

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// Google provides public iCal feeds for holidays - no auth needed.
const vnHolidayICalURL = "https://calendar.google.com/calendar/ical/en.vietnamese%23holiday@group.v.calendar.google.com/public/basic.ics"

// HolidayEntry represents a single public holiday
type HolidayEntry struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time // Exclusive (iCal convention)
}

// HolidayService fetches and caches Vietnam public holidays from Google Calendar
type HolidayService struct {
	logger logger.Logger

	mu    sync.RWMutex
	cache map[int]map[string]string // year -> date(YYYY-MM-DD) -> holiday name
}

// NewHolidayService creates a new holiday service
func NewHolidayService(log logger.Logger) *HolidayService {
	return &HolidayService{
		logger: log,
		cache:  make(map[int]map[string]string),
	}
}

// GetHolidaysForMonth returns a set of holiday date strings (YYYY-MM-DD) for a given month.
// Also returns "working day" dates (makeup Saturdays that should count as working days).
func (s *HolidayService) GetHolidaysForMonth(year, month int) (holidays map[string]string, workingDays map[string]string, err error) {
	yearHolidays, err := s.getYearHolidays(year)
	if err != nil {
		return nil, nil, err
	}

	holidays = make(map[string]string)
	workingDays = make(map[string]string)

	prefix := fmt.Sprintf("%d-%02d-", year, month)
	for date, name := range yearHolidays {
		if !strings.HasPrefix(date, prefix) {
			continue
		}
		if strings.HasPrefix(name, "Working day for") {
			workingDays[date] = name
		} else if isOfficialHoliday(name) {
			holidays[date] = name
		}
	}

	return holidays, workingDays, nil
}

// getYearHolidays returns all holidays for a year, fetching from cache or iCal feed
func (s *HolidayService) getYearHolidays(year int) (map[string]string, error) {
	s.mu.RLock()
	if cached, ok := s.cache[year]; ok {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	s.logger.Debugf("fetching Vietnam public holidays for year=%d from Google Calendar", year)

	entries, err := fetchHolidaysFromICal(vnHolidayICalURL, year)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to fetch holidays for year=%d", year))
		return nil, err
	}

	// Expand multi-day entries into individual dates
	dateMap := make(map[string]string)
	for _, e := range entries {
		for d := e.StartDate; d.Before(e.EndDate); d = d.AddDate(0, 0, 1) {
			dateMap[d.Format("2006-01-02")] = e.Name
		}
	}

	s.logger.Debugf("cached %d holiday dates for year=%d", len(dateMap), year)

	s.mu.Lock()
	s.cache[year] = dateMap
	s.mu.Unlock()

	return dateMap, nil
}

// isOfficialHoliday filters out observational/non-official holidays.
// Only returns true for days that are actual days off in Vietnam.
func isOfficialHoliday(name string) bool {
	// These are actual government-mandated holidays / days off
	officialKeywords := []string{
		"New Year",           // International New Year's Day, New Year's Day Holiday
		"Vietnamese New Year", // Tet
		"Tet Holiday",        // Tet extended days
		"Liberation Day",     // Apr 30
		"Reunification Day",  // Apr 30 alternate name
		"Labor Day",          // May 1
		"Independence Day",   // Sep 2
		"Hung Kings",         // Hung Kings Festival
		"Day off for",        // Compensatory days off
	}

	for _, kw := range officialKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}

	return false
}

func fetchHolidaysFromICal(icalURL string, year int) ([]HolidayEntry, error) {
	resp, err := http.Get(icalURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return parseICal(resp.Body, year)
}

func parseICal(r io.Reader, year int) ([]HolidayEntry, error) {
	scanner := bufio.NewScanner(r)

	var entries []HolidayEntry
	var current *HolidayEntry
	inEvent := false

	// Collect all lines first, handling line unfolding (RFC 5545)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && len(lines) > 0 {
			lines[len(lines)-1] += strings.TrimLeft(line, " \t")
		} else {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for _, line := range lines {
		switch {
		case line == "BEGIN:VEVENT":
			inEvent = true
			current = &HolidayEntry{}

		case line == "END:VEVENT":
			if current != nil && inEvent && current.StartDate.Year() == year {
				entries = append(entries, *current)
			}
			inEvent = false
			current = nil

		case inEvent && current != nil:
			if after, ok := strings.CutPrefix(line, "SUMMARY:"); ok {
				current.Name = after
			} else if after, ok := strings.CutPrefix(line, "DTSTART;VALUE=DATE:"); ok {
				if t, err := time.Parse("20060102", after); err == nil {
					current.StartDate = t
				}
			} else if after, ok := strings.CutPrefix(line, "DTSTART:"); ok {
				current.StartDate = parseICalDate(after)
			} else if after, ok := strings.CutPrefix(line, "DTEND;VALUE=DATE:"); ok {
				if t, err := time.Parse("20060102", after); err == nil {
					current.EndDate = t
				}
			} else if after, ok := strings.CutPrefix(line, "DTEND:"); ok {
				current.EndDate = parseICalDate(after)
			}
		}
	}

	return entries, nil
}

func parseICalDate(s string) time.Time {
	if t, err := time.Parse("20060102", s); err == nil {
		return t
	}
	if t, err := time.Parse("20060102T150405Z", s); err == nil {
		return t
	}
	return time.Time{}
}
