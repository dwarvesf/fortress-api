package accounting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildInvoiceTodo(t *testing.T) {
	// Test the helper function that builds invoice todos
	assigneeIDs := []int{123, 456}
	month := 10
	year := 2025
	projectName := "Test Project"

	todo := buildInvoiceTodo(projectName, month, year, assigneeIDs)

	// Verify todo structure
	require.Equal(t, "Test Project 10/2025", todo.Content)
	require.Equal(t, assigneeIDs, todo.AssigneeIDs)
	require.Equal(t, "31-10-2025", todo.DueOn) // Last day of October 2025
}

func TestGetProjectInvoiceDueOn(t *testing.T) {
	// Test special case for Voconic project
	dueOn := getProjectInvoiceDueOn("voconic", 10, 2025)
	require.Equal(t, "23-10-2025", dueOn)

	// Test regular project (last day of month)
	dueOn = getProjectInvoiceDueOn("Regular Project", 10, 2025)
	require.Equal(t, "31-10-2025", dueOn)

	// Test case insensitive
	dueOn = getProjectInvoiceDueOn("VoConIc", 10, 2025)
	require.Equal(t, "23-10-2025", dueOn)
}

func TestGetProjectInvoiceContent(t *testing.T) {
	content := getProjectInvoiceContent("Test Project", 10, 2025)
	require.Equal(t, "Test Project 10/2025", content)
}

func TestBuildInvoiceTodo_WithDifferentAssignees(t *testing.T) {
	// Test with different assignee IDs
	assigneeIDs := []int{999, 888}
	month := 12
	year := 2025
	projectName := "Another Project"

	todo := buildInvoiceTodo(projectName, month, year, assigneeIDs)

	require.Equal(t, "Another Project 12/2025", todo.Content)
	require.Equal(t, assigneeIDs, todo.AssigneeIDs)
	require.Equal(t, "31-12-2025", todo.DueOn)
}

func TestGetProjectInvoiceDueOn_February(t *testing.T) {
	// Test February in a non-leap year
	dueOn := getProjectInvoiceDueOn("Regular Project", 2, 2025)
	require.Equal(t, "28-2-2025", dueOn)

	// Test February in a leap year
	dueOn = getProjectInvoiceDueOn("Regular Project", 2, 2024)
	require.Equal(t, "29-2-2024", dueOn)
}

func TestGetProjectInvoiceDueOn_DifferentMonths(t *testing.T) {
	// Test different months
	testCases := []struct {
		month    int
		year     int
		expected string
	}{
		{1, 2025, "31-1-2025"},
		{4, 2025, "30-4-2025"},
		{6, 2025, "30-6-2025"},
		{9, 2025, "30-9-2025"},
		{11, 2025, "30-11-2025"},
	}

	for _, tc := range testCases {
		dueOn := getProjectInvoiceDueOn("Regular Project", tc.month, tc.year)
		require.Equal(t, tc.expected, dueOn, "Failed for month %d, year %d", tc.month, tc.year)
	}
}