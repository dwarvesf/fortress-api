package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Google provides public iCal feeds for holidays - no auth needed.
// Format: https://calendar.google.com/calendar/ical/{calendarID}/public/basic.ics
const vnHolidayICalURL = "https://calendar.google.com/calendar/ical/en.vietnamese%23holiday%40group.v.calendar.google.com/public/basic.ics"

type Holiday struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}

func main() {
	year := time.Now().Year()
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &year)
	}

	fmt.Printf("Fetching Vietnam public holidays via iCal feed...\n\n")

	holidays, err := fetchHolidaysFromICal(vnHolidayICalURL, year)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d holidays in %d:\n\n", len(holidays), year)
	fmt.Printf("%-12s %-12s %s\n", "Start", "End", "Name")
	fmt.Printf("%-12s %-12s %s\n", "----------", "----------", "----")

	for _, h := range holidays {
		end := h.EndDate.AddDate(0, 0, -1) // iCal DTEND is exclusive
		if end.Before(h.StartDate) {
			end = h.StartDate
		}
		fmt.Printf("%-12s %-12s %s\n", h.StartDate.Format("2006-01-02"), end.Format("2006-01-02"), h.Name)
	}

	// Print expanded dates for integration
	fmt.Printf("\n--- Individual holiday dates for %d ---\n", year)
	for _, h := range holidays {
		end := h.EndDate // iCal DTEND is exclusive, so iterate while Before(end)
		for d := h.StartDate; d.Before(end); d = d.AddDate(0, 0, 1) {
			fmt.Printf("  %s  %s\n", d.Format("2006-01-02"), h.Name)
		}
	}
}

func fetchHolidaysFromICal(icalURL string, year int) ([]Holiday, error) {
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

func parseICal(r io.Reader, year int) ([]Holiday, error) {
	scanner := bufio.NewScanner(r)

	var holidays []Holiday
	var current *Holiday
	inEvent := false

	for scanner.Scan() {
		line := scanner.Text()

		// Handle line unfolding (RFC 5545: continuation lines start with space/tab)
		for scanner.Scan() {
			next := scanner.Text()
			if strings.HasPrefix(next, " ") || strings.HasPrefix(next, "\t") {
				line += strings.TrimLeft(next, " \t")
			} else {
				// Process current line, then handle `next` in the next iteration
				processLine(line, &current, &inEvent, &holidays, year)
				line = next
				continue
			}
		}
		processLine(line, &current, &inEvent, &holidays, year)
	}

	return holidays, scanner.Err()
}

func processLine(line string, current **Holiday, inEvent *bool, holidays *[]Holiday, year int) {
	switch {
	case line == "BEGIN:VEVENT":
		*inEvent = true
		*current = &Holiday{}

	case line == "END:VEVENT":
		if *current != nil && *inEvent {
			// Filter by year
			if (*current).StartDate.Year() == year {
				*holidays = append(*holidays, **current)
			}
		}
		*inEvent = false
		*current = nil

	case *inEvent && *current != nil:
		if strings.HasPrefix(line, "SUMMARY:") {
			(*current).Name = strings.TrimPrefix(line, "SUMMARY:")
		} else if strings.HasPrefix(line, "DTSTART;VALUE=DATE:") {
			dateStr := strings.TrimPrefix(line, "DTSTART;VALUE=DATE:")
			if t, err := time.Parse("20060102", dateStr); err == nil {
				(*current).StartDate = t
			}
		} else if strings.HasPrefix(line, "DTSTART:") {
			dateStr := strings.TrimPrefix(line, "DTSTART:")
			if t, err := time.Parse("20060102", dateStr); err == nil {
				(*current).StartDate = t
			} else if t, err := time.Parse("20060102T150405Z", dateStr); err == nil {
				(*current).StartDate = t
			}
		} else if strings.HasPrefix(line, "DTEND;VALUE=DATE:") {
			dateStr := strings.TrimPrefix(line, "DTEND;VALUE=DATE:")
			if t, err := time.Parse("20060102", dateStr); err == nil {
				(*current).EndDate = t
			}
		} else if strings.HasPrefix(line, "DTEND:") {
			dateStr := strings.TrimPrefix(line, "DTEND:")
			if t, err := time.Parse("20060102", dateStr); err == nil {
				(*current).EndDate = t
			} else if t, err := time.Parse("20060102T150405Z", dateStr); err == nil {
				(*current).EndDate = t
			}
		}
	}
}
