package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"
	"strings"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// ExpenseService fetches approved expenses from Notion API and transforms them to Basecamp Todo format
// for payroll calculation. This service implements the basecamp.ExpenseProvider interface.
type ExpenseService struct {
	*baseService
	store *store.Store
	repo  store.DBRepo

	// idMapping stores the mapping from hashed integer ID to original Notion page UUID.
	// This is populated during GetAllInList and used during MarkExpenseAsCompleted.
	idMapping map[int]string
}

// NewExpenseService creates a new Notion expense service for payroll expense fetching.
func NewExpenseService(cfg *config.Config, store *store.Store, repo store.DBRepo, l logger.Logger) *ExpenseService {
	base := newBaseService(cfg, l)
	if base == nil {
		return nil
	}

	return &ExpenseService{
		baseService: base,
		store:       store,
		repo:        repo,
		idMapping:   make(map[int]string),
	}
}

// GetAllInList fetches all approved expenses from Notion and transforms them to Basecamp Todo format.
// todolistID and projectID are ignored as Notion uses database IDs from configuration.
func (e *ExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
	if e.client == nil {
		return nil, errors.New("notion client is nil")
	}

	expenseDBID := e.cfg.ExpenseIntegration.Notion.ExpenseDBID
	if expenseDBID == "" {
		e.logger.Error(errors.New("expense database id not configured"), "notion expense database id is empty")
		return nil, errors.New("notion expense database id not configured")
	}

	e.logger.Debug(fmt.Sprintf("Fetching expenses from Notion database %s (todolistID=%d, projectID=%d)", expenseDBID, todolistID, projectID))

	// Fetch approved expenses from Notion
	pages, err := e.fetchApprovedExpenses(expenseDBID)
	if err != nil {
		return nil, err
	}

	e.logger.Debug(fmt.Sprintf("Fetched %d expense records from Notion", len(pages)))

	// Transform each Notion page to bcModel.Todo
	todos := make([]bcModel.Todo, 0, len(pages))
	for _, page := range pages {
		todo, err := e.transformPageToTodo(page)
		if err != nil {
			// Log warning but continue processing other records
			e.logger.Error(err, fmt.Sprintf("failed to transform expense page %s", page.ID))
			continue
		}
		todos = append(todos, *todo)
	}

	e.logger.Debug(fmt.Sprintf("Transformed %d expenses to Todo format", len(todos)))

	return todos, nil
}

// GetGroups returns expense groups. For Notion, we return a single default group.
// todosetID and projectID are ignored as Notion uses database IDs from configuration.
func (e *ExpenseService) GetGroups(todosetID, projectID int) ([]bcModel.TodoGroup, error) {
	if e.client == nil {
		return nil, errors.New("notion client is nil")
	}

	e.logger.Debug(fmt.Sprintf("Returning default expense group (todosetID=%d, projectID=%d)", todosetID, projectID))

	// Return a single default group for expenses
	groups := []bcModel.TodoGroup{
		{
			ID:   todosetID,
			Name: "Expenses",
		},
	}

	return groups, nil
}

// GetLists returns expense lists. For Notion, we return a single default list.
// projectID and todosetID are ignored as Notion uses database IDs from configuration.
func (e *ExpenseService) GetLists(projectID, todosetID int) ([]bcModel.TodoList, error) {
	if e.client == nil {
		return nil, errors.New("notion client is nil")
	}

	e.logger.Debug(fmt.Sprintf("Returning default expense list (projectID=%d, todosetID=%d)", projectID, todosetID))

	// Return a single default list for expenses
	lists := []bcModel.TodoList{
		{
			ID:   todosetID,
			Name: "Expenses",
		},
	}

	return lists, nil
}

// MarkExpenseAsCompleted marks an expense as "Paid" in Notion by updating the Status property.
// expenseID is the Notion page ID (UUID format stored as integer via hash).
func (e *ExpenseService) MarkExpenseAsCompleted(expensePageID string) error {
	if e.client == nil {
		return errors.New("notion client is nil")
	}

	e.logger.Debug(fmt.Sprintf("Marking expense %s as Paid in Notion", expensePageID))

	ctx := context.Background()

	// Update page status to "Paid" (complete status in Notion)
	_, err := e.client.UpdatePage(ctx, expensePageID, nt.UpdatePageParams{
		DatabasePageProperties: nt.DatabasePageProperties{
			"Status": nt.DatabasePageProperty{
				Status: &nt.SelectOptions{
					Name: "Paid",
				},
			},
		},
	})
	if err != nil {
		e.logger.Error(err, fmt.Sprintf("failed to mark expense %s as Paid in Notion", expensePageID))
		return fmt.Errorf("failed to update expense status: %w", err)
	}

	e.logger.Debug(fmt.Sprintf("Successfully marked expense %s as Paid in Notion", expensePageID))
	return nil
}

// fetchApprovedExpenses queries Notion for expenses with Status = "Approved"
// It uses either the data source ID (for multi-source databases) or falls back to database ID
func (e *ExpenseService) fetchApprovedExpenses(databaseID string) ([]nt.Page, error) {
	// Check if we have a data source ID configured (for multi-source databases)
	dataSourceID := e.cfg.ExpenseIntegration.Notion.DataSourceID
	if dataSourceID != "" {
		e.logger.Debug(fmt.Sprintf("Using data source query for ID: %s", dataSourceID))
		return e.queryDataSource(dataSourceID)
	}

	// Fallback to standard database query for single-source databases
	e.logger.Debug(fmt.Sprintf("Using standard database query for ID: %s", databaseID))
	return e.queryDatabase(databaseID)
}

// queryDatabase queries a standard Notion database (single data source)
func (e *ExpenseService) queryDatabase(databaseID string) ([]nt.Page, error) {
	ctx := context.Background()

	// Query for expenses with status "Approved" (in_progress status type in Notion)
	filter := &nt.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Status: &nt.StatusDatabaseQueryFilter{
				Equals: "Approved",
			},
		},
	}

	var allPages []nt.Page
	var cursor string

	for {
		query := &nt.DatabaseQuery{
			Filter:   filter,
			PageSize: 100,
		}
		if cursor != "" {
			query.StartCursor = cursor
		}

		resp, err := e.client.QueryDatabase(ctx, databaseID, query)
		if err != nil {
			e.logger.Error(err, "failed to query Notion expense database")
			return nil, fmt.Errorf("notion query failed: %w", err)
		}

		allPages = append(allPages, resp.Results...)

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		cursor = *resp.NextCursor
	}

	return allPages, nil
}

// DataSourceQueryRequest represents the request body for data source query
type DataSourceQueryRequest struct {
	Filter      *nt.DatabaseQueryFilter `json:"filter,omitempty"`
	PageSize    int                     `json:"page_size,omitempty"`
	StartCursor string                  `json:"start_cursor,omitempty"`
}

// DataSourcePage represents a Notion page from data source query response
// This is needed because go-notion doesn't support the new "data_source_id" parent type
type DataSourcePage struct {
	ID             string                     `json:"id"`
	CreatedTime    string                     `json:"created_time"`
	LastEditedTime string                     `json:"last_edited_time"`
	Parent         DataSourcePageParent       `json:"parent"`
	Archived       bool                       `json:"archived"`
	Properties     nt.DatabasePageProperties  `json:"properties"`
	URL            string                     `json:"url"`
}

// DataSourcePageParent represents the parent of a page in data source query response
type DataSourcePageParent struct {
	Type         string `json:"type"`
	DataSourceID string `json:"data_source_id,omitempty"`
	DatabaseID   string `json:"database_id,omitempty"`
}

// DataSourceQueryResponse represents the response from data source query
type DataSourceQueryResponse struct {
	Results    []DataSourcePage `json:"results"`
	HasMore    bool             `json:"has_more"`
	NextCursor *string          `json:"next_cursor"`
}

// queryDataSource queries a Notion data source directly using HTTP
// This is required for multi-source databases that aren't supported by go-notion
func (e *ExpenseService) queryDataSource(dataSourceID string) ([]nt.Page, error) {
	// Query for expenses with status "Approved"
	filter := &nt.DatabaseQueryFilter{
		Property: "Status",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Status: &nt.StatusDatabaseQueryFilter{
				Equals: "Approved",
			},
		},
	}

	var allPages []nt.Page
	var cursor string

	for {
		reqBody := DataSourceQueryRequest{
			Filter:   filter,
			PageSize: 100,
		}
		if cursor != "" {
			reqBody.StartCursor = cursor
		}

		resp, err := e.executeDataSourceQuery(dataSourceID, reqBody)
		if err != nil {
			return nil, err
		}

		// Convert DataSourcePage to nt.Page
		for _, dsPage := range resp.Results {
			page := e.convertDataSourcePageToPage(dsPage)
			allPages = append(allPages, page)
		}

		if !resp.HasMore || resp.NextCursor == nil {
			break
		}
		cursor = *resp.NextCursor
	}

	return allPages, nil
}

// convertDataSourcePageToPage converts a DataSourcePage to nt.Page
func (e *ExpenseService) convertDataSourcePageToPage(dsPage DataSourcePage) nt.Page {
	return nt.Page{
		ID:         dsPage.ID,
		URL:        dsPage.URL,
		Archived:   dsPage.Archived,
		Properties: dsPage.Properties,
	}
}

// executeDataSourceQuery makes the HTTP request to query a data source
// Uses POST method and API version 2025-09-03 as per Notion's API reference
func (e *ExpenseService) executeDataSourceQuery(dataSourceID string, reqBody DataSourceQueryRequest) (*DataSourceQueryResponse, error) {
	// Normalize data source ID - remove hyphens if present
	normalizedID := strings.ReplaceAll(dataSourceID, "-", "")
	url := fmt.Sprintf("https://api.notion.com/v1/data_sources/%s/query", normalizedID)
	e.logger.Debug(fmt.Sprintf("Data source query URL: %s", url))

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	e.logger.Debug(fmt.Sprintf("Querying data source %s with body: %s", dataSourceID, string(jsonBody)))

	// Use POST method for data source query as per Notion API reference
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+e.cfg.Notion.Secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2025-09-03") // Required for data source queries

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		e.logger.Error(err, "failed to execute data source query")
		return nil, fmt.Errorf("data source query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		e.logger.Error(errors.New(string(body)), fmt.Sprintf("data source query returned status %d", resp.StatusCode))
		return nil, fmt.Errorf("data source query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result DataSourceQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// transformPageToTodo transforms a Notion expense page to Basecamp Todo format
func (e *ExpenseService) transformPageToTodo(page nt.Page) (*bcModel.Todo, error) {
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast page properties")
	}

	// Extract expense fields
	// Build title from Notes (rich_text) or Description (rich_text) or Refund ID (title)
	title := e.extractTitle(props)
	amount := ExtractNumber(props,"Amount")
	currency := ExtractSelect(props,"Currency")
	// Try "Reason" first (Refund Requests DB), fallback to "Expense Category" (legacy)
	category := ExtractSelect(props,"Reason")
	if category == "" {
		category = ExtractSelect(props,"Expense Category")
		e.logger.Debug(fmt.Sprintf("Using legacy 'Expense Category' property: %s", category))
	} else {
		e.logger.Debug(fmt.Sprintf("Using 'Reason' property for category: %s", category))
	}

	// Get requestor email via rollup or fallback to relation query
	email, err := e.extractRequestorEmail(page.ID, props)
	if err != nil {
		return nil, fmt.Errorf("failed to extract requestor email: %w", err)
	}

	if email == "" {
		return nil, errors.New("missing requestor email")
	}

	// Lookup employee by team email
	employee, err := e.store.Employee.OneByEmail(e.repo.DB(), email)
	if err != nil {
		e.logger.Error(err, fmt.Sprintf("failed to find employee by email: %s", email))
		return nil, fmt.Errorf("employee not found for email %s: %w", email, err)
	}

	if employee.BasecampID == 0 {
		e.logger.Error(errors.New("employee has no basecamp_id"), fmt.Sprintf("employee missing basecamp_id: %s", email))
		return nil, fmt.Errorf("employee %s has no basecamp_id", email)
	}

	// Convert UUID to integer using hash and store mapping for later lookup
	todoID := e.uuidToInt(page.ID)
	e.idMapping[todoID] = page.ID

	// Build Todo title in format: "description | amount | currency"
	// This matches the format expected by payroll calculator's getReimbursement function
	todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)

	// Create Todo object
	// AppURL stores the original Notion page UUID for later use in MarkExpenseAsCompleted
	todo := &bcModel.Todo{
		ID:    todoID,
		Title: todoTitle,
		AppURL: page.ID, // Store original Notion page UUID
		Assignees: []bcModel.Assignee{
			{
				ID:   employee.BasecampID,
				Name: employee.FullName,
			},
		},
		Bucket: bcModel.Bucket{
			ID:   todoID,
			Name: category,
		},
		Completed: true, // Expense is approved
	}

	e.logger.Debug(fmt.Sprintf("Transformed expense page %s to Todo: %s (employee: %s)", page.ID, todoTitle, email))

	return todo, nil
}

// extractTitle extracts the title/description from Notion page properties
// Priority: Notes (rich_text) -> Description (rich_text) -> Refund ID (title) -> Title (title) -> Reason (title, legacy)
func (e *ExpenseService) extractTitle(props nt.DatabasePageProperties) string {
	// Try "Notes" first (Refund Requests DB uses this for description)
	if notesProp, ok := props["Notes"]; ok && len(notesProp.RichText) > 0 {
		var parts []string
		for _, rt := range notesProp.RichText {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		if result != "" {
			e.logger.Debug(fmt.Sprintf("Extracted title from 'Notes' property: %s", result))
			return result
		}
	}

	// Try "Description" (rich_text)
	if descProp, ok := props["Description"]; ok && len(descProp.RichText) > 0 {
		var parts []string
		for _, rt := range descProp.RichText {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		if result != "" {
			e.logger.Debug(fmt.Sprintf("Extracted title from 'Description' property: %s", result))
			return result
		}
	}

	// Try "Refund ID" (title property in Refund Requests DB)
	if refundIDProp, ok := props["Refund ID"]; ok && len(refundIDProp.Title) > 0 {
		var parts []string
		for _, rt := range refundIDProp.Title {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		if result != "" {
			e.logger.Debug(fmt.Sprintf("Extracted title from 'Refund ID' property: %s", result))
			return result
		}
	}

	// Fallback to "Title" (legacy)
	if titleProp, ok := props["Title"]; ok && len(titleProp.Title) > 0 {
		var parts []string
		for _, rt := range titleProp.Title {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		if result != "" {
			e.logger.Debug(fmt.Sprintf("Extracted title from 'Title' property: %s", result))
			return result
		}
	}

	// Fallback to "Reason" as title type (legacy Expense Request DB)
	if reasonProp, ok := props["Reason"]; ok && len(reasonProp.Title) > 0 {
		var parts []string
		for _, rt := range reasonProp.Title {
			parts = append(parts, rt.PlainText)
		}
		result := strings.TrimSpace(strings.Join(parts, ""))
		if result != "" {
			e.logger.Debug(fmt.Sprintf("Extracted title from 'Reason' (title) property: %s", result))
			return result
		}
	}

	e.logger.Debug("No title property found in Notion page")
	return ""
}


// extractRequestorEmail extracts the requestor's email from the Notion page.
// Uses the Team Email property (email type).
func (e *ExpenseService) extractRequestorEmail(pageID string, props nt.DatabasePageProperties) (string, error) {
	// Get email from Team Email property
	if emailProp, ok := props["Team Email"]; ok && emailProp.Email != nil {
		e.logger.Debug(fmt.Sprintf("extractRequestorEmail: found Team Email: %s", *emailProp.Email))
		if *emailProp.Email != "" {
			return *emailProp.Email, nil
		}
	}

	e.logger.Debug("extractRequestorEmail: Team Email property not found or empty")
	return "", errors.New("no Team Email found")
}

// uuidToInt converts a UUID string to an integer using FNV-1a hash of the last 8 hex characters.
// This provides deterministic, unique-enough integer IDs for the existing payroll system.
func (e *ExpenseService) uuidToInt(uuid string) int {
	// Remove hyphens and get the last 8 characters
	clean := strings.ReplaceAll(uuid, "-", "")
	if len(clean) < 8 {
		// Fallback for invalid UUID
		h := fnv.New32a()
		h.Write([]byte(uuid))
		return int(h.Sum32())
	}

	// Take last 8 hex characters
	suffix := clean[len(clean)-8:]

	// Parse as hex and convert to int
	val, err := strconv.ParseInt(suffix, 16, 64)
	if err != nil {
		// Fallback to FNV hash if parsing fails
		h := fnv.New32a()
		h.Write([]byte(uuid))
		return int(h.Sum32())
	}

	return int(val)
}
