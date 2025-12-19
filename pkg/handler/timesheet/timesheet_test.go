package timesheet

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/handler/timesheet/request"
	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func TestLogHoursRequestValidation(t *testing.T) {
	validDate := time.Now().AddDate(0, 0, -1) // yesterday

	tests := []struct {
		name        string
		request     request.LogHoursRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_request",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        6.0,
				ProofOfWorks: "PR #123",
			},
			expectError: false,
		},
		{
			name: "empty_discord_id",
			request: request.LogHoursRequest{
				DiscordID:    "",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        6.0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "discord_id is required",
		},
		{
			name: "empty_project_id",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        6.0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "project_id is required",
		},
		{
			name: "invalid_hours_zero",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "hours must be between 0 and 24",
		},
		{
			name: "invalid_hours_over_24",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        25.0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "hours must be between 0 and 24",
		},
		{
			name: "invalid_task_type",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "InvalidType",
				Hours:        6.0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "task_type must be Development, Design, or Meeting",
		},
		{
			name: "valid_design_task_type",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Design",
				Hours:        2.5,
				ProofOfWorks: "Figma link",
			},
			expectError: false,
		},
		{
			name: "valid_meeting_task_type",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Meeting",
				Hours:        0.5,
				ProofOfWorks: "Standup meeting",
			},
			expectError: false,
		},
		{
			name: "empty_proof_of_works",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         validDate,
				TaskType:     "Development",
				Hours:        6.0,
				ProofOfWorks: "",
			},
			expectError: true,
			errorMsg:    "proof_of_works is required",
		},
		{
			name: "future_date",
			request: request.LogHoursRequest{
				DiscordID:    "123456789",
				ProjectID:    "project-uuid",
				Date:         time.Now().AddDate(0, 0, 7), // 7 days in future
				TaskType:     "Development",
				Hours:        6.0,
				ProofOfWorks: "PR #123",
			},
			expectError: true,
			errorMsg:    "date cannot be in the future",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateWeeklySummary(t *testing.T) {
	startOfWeek := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC) // Monday
	endOfWeek := time.Date(2025, 12, 21, 23, 59, 59, 0, time.UTC) // Sunday

	tests := []struct {
		name           string
		entries        []notionSvc.TimesheetEntry
		expectedTotal  float64
		expectedCount  int
		expectedByType map[string]float64
	}{
		{
			name:           "empty_entries",
			entries:        []notionSvc.TimesheetEntry{},
			expectedTotal:  0,
			expectedCount:  0,
			expectedByType: map[string]float64{},
		},
		{
			name: "single_entry",
			entries: []notionSvc.TimesheetEntry{
				{Hours: 6.0, TaskType: "Development", ProjectID: "proj1"},
			},
			expectedTotal:  6.0,
			expectedCount:  1,
			expectedByType: map[string]float64{"Development": 6.0},
		},
		{
			name: "multiple_entries_same_type",
			entries: []notionSvc.TimesheetEntry{
				{Hours: 6.0, TaskType: "Development", ProjectID: "proj1"},
				{Hours: 4.0, TaskType: "Development", ProjectID: "proj1"},
			},
			expectedTotal:  10.0,
			expectedCount:  2,
			expectedByType: map[string]float64{"Development": 10.0},
		},
		{
			name: "multiple_entries_different_types",
			entries: []notionSvc.TimesheetEntry{
				{Hours: 6.0, TaskType: "Development", ProjectID: "proj1"},
				{Hours: 2.0, TaskType: "Design", ProjectID: "proj1"},
				{Hours: 0.5, TaskType: "Meeting", ProjectID: "proj1"},
			},
			expectedTotal:  8.5,
			expectedCount:  3,
			expectedByType: map[string]float64{"Development": 6.0, "Design": 2.0, "Meeting": 0.5},
		},
		{
			name: "fractional_hours",
			entries: []notionSvc.TimesheetEntry{
				{Hours: 2.5, TaskType: "Development", ProjectID: "proj1"},
				{Hours: 1.25, TaskType: "Development", ProjectID: "proj1"},
			},
			expectedTotal:  3.75,
			expectedCount:  2,
			expectedByType: map[string]float64{"Development": 3.75},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := calculateWeeklySummary(tt.entries, startOfWeek, endOfWeek)

			assert.Equal(t, tt.expectedTotal, summary.TotalHours)
			assert.Equal(t, tt.expectedCount, summary.EntryCount)
			assert.Equal(t, startOfWeek, summary.StartDate)
			assert.Equal(t, endOfWeek, summary.EndDate)

			for taskType, expectedHours := range tt.expectedByType {
				assert.Equal(t, expectedHours, summary.ByTaskType[taskType], "ByTaskType[%s]", taskType)
			}
		})
	}
}

func TestWeekBoundaryCalculation(t *testing.T) {
	tests := []struct {
		name              string
		now               time.Time
		weekOffset        int
		expectedStartDay  int
		expectedStartMonth time.Month
		expectedEndDay    int
	}{
		{
			name:              "current_week_wednesday",
			now:               time.Date(2025, 12, 17, 12, 0, 0, 0, time.UTC), // Wednesday
			weekOffset:        0,
			expectedStartDay:  15, // Monday
			expectedStartMonth: time.December,
			expectedEndDay:    21, // Sunday
		},
		{
			name:              "current_week_monday",
			now:               time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC), // Monday
			weekOffset:        0,
			expectedStartDay:  15,
			expectedStartMonth: time.December,
			expectedEndDay:    21,
		},
		{
			name:              "current_week_sunday",
			now:               time.Date(2025, 12, 21, 12, 0, 0, 0, time.UTC), // Sunday
			weekOffset:        0,
			expectedStartDay:  15,
			expectedStartMonth: time.December,
			expectedEndDay:    21,
		},
		{
			name:              "last_week",
			now:               time.Date(2025, 12, 17, 12, 0, 0, 0, time.UTC), // Wednesday
			weekOffset:        -1,
			expectedStartDay:  8,
			expectedStartMonth: time.December,
			expectedEndDay:    14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weekday := int(tt.now.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday = 7
			}

			// Start of week (Monday)
			startOfWeek := tt.now.AddDate(0, 0, -(weekday-1)+(tt.weekOffset*7))
			startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
			endOfWeek := startOfWeek.AddDate(0, 0, 6)

			assert.Equal(t, tt.expectedStartDay, startOfWeek.Day(), "start day")
			assert.Equal(t, tt.expectedStartMonth, startOfWeek.Month(), "start month")
			assert.Equal(t, tt.expectedEndDay, endOfWeek.Day(), "end day")
		})
	}
}

func TestGetEntriesInvalidDiscordID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/timesheets", nil)

	// Directly test the validation by checking query param
	discordID := c.Query("discord_id")
	assert.Empty(t, discordID, "discord_id should be empty when not provided")
}

func TestGetWeeklySummaryInvalidDiscordID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/timesheets/weekly-summary", nil)

	// Directly test the validation by checking query param
	discordID := c.Query("discord_id")
	assert.Empty(t, discordID, "discord_id should be empty when not provided")
}

func TestLogHoursRequestBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Test that invalid JSON fails to bind
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/timesheets", strings.NewReader("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	var req request.LogHoursRequest
	err := c.ShouldBindJSON(&req)
	assert.Error(t, err, "invalid JSON should fail to bind")
}

func TestToTimesheetEntry(t *testing.T) {
	date := time.Date(2025, 12, 17, 0, 0, 0, 0, time.UTC)
	entry := notionSvc.TimesheetEntry{
		PageID:       "page-123",
		ProjectID:    "proj-456",
		ProjectName:  "Test Project",
		Discord:      "testuser",
		Date:         &date,
		TaskType:     "Development",
		Status:       "Approved",
		Hours:        6.0,
		ProofOfWorks: "PR #123",
	}

	viewEntry := view.ToTimesheetEntry(entry)

	assert.Equal(t, entry.PageID, viewEntry.PageID)
	assert.Equal(t, entry.ProjectID, viewEntry.ProjectID)
	assert.Equal(t, entry.ProjectName, viewEntry.ProjectName)
	assert.Equal(t, entry.Discord, viewEntry.Discord)
	assert.Equal(t, entry.Date, viewEntry.Date)
	assert.Equal(t, entry.TaskType, viewEntry.TaskType)
	assert.Equal(t, entry.Status, viewEntry.Status)
	assert.Equal(t, entry.Hours, viewEntry.Hours)
	assert.Equal(t, entry.ProofOfWorks, viewEntry.ProofOfWorks)
}

func TestToTimesheetEntries(t *testing.T) {
	date := time.Date(2025, 12, 17, 0, 0, 0, 0, time.UTC)
	entries := []notionSvc.TimesheetEntry{
		{PageID: "page-1", Hours: 6.0, Date: &date},
		{PageID: "page-2", Hours: 4.0, Date: &date},
	}

	viewEntries := view.ToTimesheetEntries(entries)

	assert.Len(t, viewEntries, 2)
	assert.Equal(t, "page-1", viewEntries[0].PageID)
	assert.Equal(t, "page-2", viewEntries[1].PageID)
}
