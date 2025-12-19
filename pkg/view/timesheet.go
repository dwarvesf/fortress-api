package view

import (
	"time"

	notionSvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// TimesheetEntry represents a timesheet entry in API responses
type TimesheetEntry struct {
	PageID       string     `json:"page_id"`
	ProjectID    string     `json:"project_id"`
	ProjectName  string     `json:"project_name,omitempty"`
	Discord      string     `json:"discord"`
	Date         *time.Time `json:"date"`
	TaskType     string     `json:"task_type"`
	Status       string     `json:"status"`
	Hours        float64    `json:"hours"`
	ProofOfWorks string     `json:"proof_of_works"`
}

// TimesheetWeeklySummary represents aggregated weekly timesheet data
type TimesheetWeeklySummary struct {
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	TotalHours float64            `json:"total_hours"`
	ByTaskType map[string]float64 `json:"by_task_type"`
	ByProject  map[string]float64 `json:"by_project"`
	EntryCount int                `json:"entry_count"`
}

// ToTimesheetEntry converts a notion TimesheetEntry to view TimesheetEntry
func ToTimesheetEntry(e notionSvc.TimesheetEntry) TimesheetEntry {
	return TimesheetEntry{
		PageID:       e.PageID,
		ProjectID:    e.ProjectID,
		ProjectName:  e.ProjectName,
		Discord:      e.Discord,
		Date:         e.Date,
		TaskType:     e.TaskType,
		Status:       e.Status,
		Hours:        e.Hours,
		ProofOfWorks: e.ProofOfWorks,
	}
}

// ToTimesheetEntries converts a slice of notion TimesheetEntry to view TimesheetEntry
func ToTimesheetEntries(entries []notionSvc.TimesheetEntry) []TimesheetEntry {
	result := make([]TimesheetEntry, 0, len(entries))
	for _, e := range entries {
		result = append(result, ToTimesheetEntry(e))
	}
	return result
}
