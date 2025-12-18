package google

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// CalendarService handles Google Calendar operations
type CalendarService struct {
	config *config.Config
	logger logger.Logger
}

// NewCalendarService creates a new Google Calendar service
func NewCalendarService(cfg *config.Config, log logger.Logger) *CalendarService {
	return &CalendarService{
		config: cfg,
		logger: log,
	}
}

// CalendarEvent represents a calendar event to be created
type CalendarEvent struct {
	Summary     string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	AllDay      bool
	Email       string   // Email of the employee (for description/context)
	Attendees   []string // List of attendee emails
}

// CreateLeaveEvent creates a Google Calendar event for a leave request
func (s *CalendarService) CreateLeaveEvent(ctx context.Context, event CalendarEvent) (*calendar.Event, error) {
	l := s.logger.Fields(logger.Fields{
		"service": "google.calendar",
		"method":  "CreateLeaveEvent",
	})

	l.Debugf("creating leave event for: %s (start: %s, end: %s)", event.Email, event.StartDate, event.EndDate)

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     s.config.Google.ClientID,
		ClientSecret: s.config.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{calendar.CalendarScope},
	}

	// Create token from refresh token
	token := &oauth2.Token{
		RefreshToken: s.config.Google.TeamGoogleRefreshToken,
	}

	l.Debug("creating Google Calendar client with refresh token")

	// Create HTTP client with OAuth2 token
	httpClient := oauth2Config.Client(ctx, token)

	// Create Calendar service
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		l.Errorf(err, "failed to create calendar service")
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Build event
	calEvent := &calendar.Event{
		Summary:     event.Summary,
		Description: event.Description,
	}

	// Add attendees if provided
	if len(event.Attendees) > 0 {
		var attendees []*calendar.EventAttendee
		for _, email := range event.Attendees {
			attendees = append(attendees, &calendar.EventAttendee{
				Email:    email,
				Optional: true,
			})
		}
		calEvent.Attendees = attendees
		l.Debugf("added %d attendees to calendar event (all marked as optional)", len(attendees))
	}

	if event.AllDay {
		// All-day event - use date format
		calEvent.Start = &calendar.EventDateTime{
			Date: event.StartDate.Format("2006-01-02"),
		}
		calEvent.End = &calendar.EventDateTime{
			Date: event.EndDate.Add(24 * time.Hour).Format("2006-01-02"), // End date is exclusive for all-day events
		}
		l.Debugf("creating all-day event: start=%s end=%s", calEvent.Start.Date, calEvent.End.Date)
	} else {
		// Timed event - use datetime format
		calEvent.Start = &calendar.EventDateTime{
			DateTime: event.StartDate.Format(time.RFC3339),
		}
		calEvent.End = &calendar.EventDateTime{
			DateTime: event.EndDate.Format(time.RFC3339),
		}
		l.Debugf("creating timed event: start=%s end=%s", calEvent.Start.DateTime, calEvent.End.DateTime)
	}

	// Get the calendar ID - prioritize Dwarves Calendar, then accounting email, then primary
	calendarID := "primary"
	if s.config.Google.DwarvesCalendarID != "" {
		calendarID = s.config.Google.DwarvesCalendarID
		l.Debugf("using Dwarves Calendar ID: %s", calendarID)
	} else if s.config.Google.AccountingEmailID != "" {
		calendarID = s.config.Google.AccountingEmailID
		l.Debugf("using Accounting Email ID: %s", calendarID)
	}

	l.Debugf("inserting event to calendar: %s", calendarID)

	// Create the event
	createdEvent, err := calendarService.Events.Insert(calendarID, calEvent).Context(ctx).Do()
	if err != nil {
		l.Errorf(err, "failed to create calendar event")
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	l.Debugf("successfully created calendar event: id=%s link=%s", createdEvent.Id, createdEvent.HtmlLink)

	return createdEvent, nil
}
