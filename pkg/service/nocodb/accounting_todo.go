package nocodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// AccountingTodoService fetches accounting todos from NocoDB API for payroll calculation.
// Focuses on "Out" group todos (expenses to be paid out).
type AccountingTodoService struct {
	client *Service
	cfg    *config.Config
	store  *store.Store
	repo   store.DBRepo
	logger logger.Logger
}

// NewAccountingTodoService creates a new NocoDB accounting todo service for payroll.
func NewAccountingTodoService(client *Service, cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *AccountingTodoService {
	return &AccountingTodoService{
		client: client,
		cfg:    cfg,
		store:  store,
		repo:   repo,
		logger: logger,
	}
}

// GetAllInList fetches accounting todos from "Out" group for payroll calculation.
// todolistID is ignored for NocoDB (uses configured accounting_todos table)
// projectID is ignored for NocoDB (uses configured workspace)
func (a *AccountingTodoService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
	if a.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	ctx := context.Background()
	tableID := a.cfg.AccountingIntegration.Noco.TodosTableID
	if tableID == "" {
		a.logger.Error(errors.New("accounting todos table id not configured"), "nocodb accounting todos table id is empty")
		return nil, errors.New("nocodb accounting todos table id not configured")
	}

	a.logger.Debug(fmt.Sprintf("Fetching accounting todos from NocoDB table %s (todolistID=%d, projectID=%d)", tableID, todolistID, projectID))

	// Query NocoDB API for "Out" group todos (not completed)
	path := fmt.Sprintf("/tables/%s/records", tableID)
	query := url.Values{}
	query.Set("where", "(task_group,eq,out)~and(status,neq,completed)")
	query.Set("limit", "100")

	resp, err := a.client.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		a.logger.Error(err, "failed to fetch accounting todos from nocodb")
		return nil, fmt.Errorf("nocodb get accounting todos failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		a.logger.Error(fmt.Errorf("nocodb returned status %s", resp.Status), fmt.Sprintf("nocodb get accounting todos failed: %s", string(body)))
		return nil, fmt.Errorf("nocodb get accounting todos failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		a.logger.Error(err, "failed to decode nocodb accounting todos response")
		return nil, fmt.Errorf("failed to decode nocodb response: %w", err)
	}

	a.logger.Debug(fmt.Sprintf("Fetched %d accounting todo records from NocoDB", len(result.List)))

	// Debug: log raw records
	for i, record := range result.List {
		recordJSON, _ := json.Marshal(record)
		a.logger.Debug(fmt.Sprintf("Raw NocoDB record %d: %s", i+1, string(recordJSON)))
	}

	// Transform each NocoDB record to bcModel.Todo
	todos := make([]bcModel.Todo, 0, len(result.List))
	for _, record := range result.List {
		todo, err := a.transformRecordToTodo(record)
		if err != nil {
			// Log warning but continue processing other records
			a.logger.Error(err, fmt.Sprintf("failed to transform accounting todo record %s", extractRecordID(record)))
			continue
		}
		todos = append(todos, *todo)
	}

	a.logger.Debug(fmt.Sprintf("Transformed %d accounting todos to Todo format", len(todos)))

	return todos, nil
}

// transformRecordToTodo transforms a NocoDB accounting todo record to Basecamp Todo format.
func (a *AccountingTodoService) transformRecordToTodo(record map[string]interface{}) (*bcModel.Todo, error) {
	// Extract fields from NocoDB record
	recordID := extractRecordID(record)
	if recordID == "" {
		return nil, errors.New("missing record ID")
	}

	title := extractString(record, "title")
	if title == "" {
		return nil, errors.New("missing title")
	}

	taskGroup := extractString(record, "task_group")
	assigneeIDsRaw := record["assignee_ids"]

	// Parse title to extract description, amount, and currency
	// Expected format: "Description | Amount | Currency"
	description, amount, currency, err := a.parseTodoTitle(title)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todo title: %w", err)
	}

	// Debug: log extracted fields
	a.logger.Debug(fmt.Sprintf("NocoDB accounting todo record %s - title: '%s', taskGroup: '%s', assigneeIds: %v, amount: %.0f, currency: %s",
		recordID, title, taskGroup, assigneeIDsRaw, amount, currency))

	// Parse assignee_ids array
	assigneeIDs, err := a.parseAssigneeIDs(assigneeIDsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse assignee_ids: %w", err)
	}

	// Resolve assignees to employees and filter out Han
	assignees, err := a.resolveAssignees(assigneeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve assignees: %w", err)
	}

	// Require exactly 1 assignee (excluding Han)
	if len(assignees) != 1 {
		a.logger.Debug(fmt.Sprintf("Skipping accounting todo %s - expected 1 assignee, got %d", recordID, len(assignees)))
		return nil, fmt.Errorf("expected 1 assignee, got %d", len(assignees))
	}

	// Convert record ID to int
	id, err := strconv.Atoi(recordID)
	if err != nil {
		a.logger.Error(err, fmt.Sprintf("failed to convert record ID to int: %s", recordID))
		return nil, fmt.Errorf("invalid record ID %s: %w", recordID, err)
	}

	// Build Todo title in Basecamp format: "description | amount | currency"
	// This matches the format expected by payroll calculator's getReimbursement function
	todoTitle := fmt.Sprintf("%s | %.0f | %s", description, amount, currency)

	// Create Todo object
	todo := &bcModel.Todo{
		ID:        id,
		Title:     todoTitle,
		Assignees: assignees,
		Bucket: bcModel.Bucket{
			ID:   id,
			Name: taskGroup,
		},
		Completed: false, // Accounting todos are open until paid
	}

	a.logger.Debug(fmt.Sprintf("Transformed accounting todo record %s to Todo: %s (original: %s)", recordID, todoTitle, title))

	return todo, nil
}

// parseTodoTitle parses the todo title in format "Description | Amount | Currency"
// Returns description, amount (as float64), and currency
func (a *AccountingTodoService) parseTodoTitle(title string) (string, float64, string, error) {
	parts := strings.Split(title, "|")
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid title format, expected 'Description | Amount | Currency', got: %s", title)
	}

	description := strings.TrimSpace(parts[0])
	amountStr := strings.TrimSpace(parts[1])
	currency := strings.TrimSpace(parts[2])

	// Default to VND if currency is empty
	if currency == "" {
		currency = "VND"
		a.logger.Debug(fmt.Sprintf("Currency empty in title '%s', defaulting to VND", title))
	}

	// Parse amount (remove any thousand separators)
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.ReplaceAll(amountStr, ".", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("failed to parse amount '%s': %w", amountStr, err)
	}

	return description, amount, currency, nil
}

// parseAssigneeIDs parses assignee_ids field which can be:
// - JSON array: ["123", "456"]
// - Comma-separated string: "123,456"
// - Array of interface{}: []interface{}{123, 456}
func (a *AccountingTodoService) parseAssigneeIDs(assigneeIDsRaw interface{}) ([]int, error) {
	if assigneeIDsRaw == nil {
		return []int{}, nil
	}

	var assigneeIDs []int

	switch v := assigneeIDsRaw.(type) {
	case []interface{}:
		// Array of interface{} (from JSON unmarshaling)
		for _, item := range v {
			switch id := item.(type) {
			case float64:
				assigneeIDs = append(assigneeIDs, int(id))
			case string:
				if intID, err := strconv.Atoi(id); err == nil {
					assigneeIDs = append(assigneeIDs, intID)
				}
			}
		}
	case string:
		// JSON string or comma-separated
		str := strings.TrimSpace(v)
		if str == "" {
			return []int{}, nil
		}

		// Try parse as JSON array
		var jsonIDs []interface{}
		if err := json.Unmarshal([]byte(str), &jsonIDs); err == nil {
			for _, item := range jsonIDs {
				switch id := item.(type) {
				case float64:
					assigneeIDs = append(assigneeIDs, int(id))
				case string:
					if intID, err := strconv.Atoi(id); err == nil {
						assigneeIDs = append(assigneeIDs, intID)
					}
				}
			}
		} else {
			// Try comma-separated
			parts := strings.Split(str, ",")
			for _, part := range parts {
				if intID, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
					assigneeIDs = append(assigneeIDs, intID)
				}
			}
		}
	}

	a.logger.Debug(fmt.Sprintf("Parsed assignee_ids: %v → %v", assigneeIDsRaw, assigneeIDs))
	return assigneeIDs, nil
}

// resolveAssignees fetches employees by basecamp_id and filters out Han.
func (a *AccountingTodoService) resolveAssignees(assigneeIDs []int) ([]bcModel.Assignee, error) {
	const HanBasecampID = 23147886 // Production Han basecamp_id

	assignees := []bcModel.Assignee{}
	for _, basecampID := range assigneeIDs {
		// Skip Han (approver)
		if basecampID == HanBasecampID {
			a.logger.Debug(fmt.Sprintf("Skipping Han's basecamp_id: %d", basecampID))
			continue
		}

		// Fetch employee by basecamp_id
		employee, err := a.store.Employee.OneByBasecampID(a.repo.DB(), basecampID)
		if err != nil {
			a.logger.Error(err, fmt.Sprintf("failed to find employee by basecamp_id: %d", basecampID))
			// Skip this assignee but continue processing
			continue
		}

		assignees = append(assignees, bcModel.Assignee{
			ID:   basecampID,
			Name: employee.FullName,
		})
	}

	a.logger.Debug(fmt.Sprintf("Resolved %d assignees (filtered Han)", len(assignees)))
	return assignees, nil
}

// GetGroups returns the "Out" group for accounting todos.
// For NocoDB, this is simplified since we filter task_group="out" in GetAllInList.
func (a *AccountingTodoService) GetGroups(todosetID, projectID int) ([]bcModel.TodoGroup, error) {
	if a.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	a.logger.Debug(fmt.Sprintf("Returning 'Out' group for accounting todos (todosetID=%d, projectID=%d)", todosetID, projectID))

	// Return single "Out" group
	groups := []bcModel.TodoGroup{
		{
			ID:    todosetID,
			Title: "out", // Lowercase to match filter in getAccountingExpense
		},
	}

	return groups, nil
}

// GetLists returns a default list for accounting todos.
// For NocoDB, we use a simplified single-list approach.
func (a *AccountingTodoService) GetLists(projectID, todosetID int) ([]bcModel.TodoList, error) {
	if a.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	a.logger.Debug(fmt.Sprintf("Returning default list for accounting todos (projectID=%d, todosetID=%d)", projectID, todosetID))

	// Return default list
	lists := []bcModel.TodoList{
		{
			ID:   todosetID,
			Name: "Accounting Todos",
		},
	}

	return lists, nil
}

// MarkTodoAsCompleted marks an accounting todo as completed in NocoDB
func (a *AccountingTodoService) MarkTodoAsCompleted(todoID int) error {
	if a.client == nil {
		return errors.New("nocodb client is nil")
	}

	tableID := a.cfg.AccountingIntegration.Noco.TodosTableID
	if tableID == "" {
		a.logger.Error(errors.New("accounting todos table id not configured"), "nocodb accounting todos table id is empty")
		return errors.New("nocodb accounting todos table id not configured")
	}

	a.logger.Debug(fmt.Sprintf("Marking accounting todo %d as completed in NocoDB table %s", todoID, tableID))

	ctx := context.Background()

	// Build PATCH payload
	payload := map[string]interface{}{
		"status": "completed",
	}

	a.logger.Debug(fmt.Sprintf("Sending PATCH request to NocoDB for todo %d with status: completed", todoID))

	// Use service's UpdateAccountingTodo method
	err := a.client.UpdateAccountingTodo(ctx, strconv.Itoa(todoID), payload)
	if err != nil {
		a.logger.Error(err, fmt.Sprintf("failed to mark accounting todo %d as completed in nocodb", todoID))
		return fmt.Errorf("nocodb mark todo as completed failed: %w", err)
	}

	a.logger.Debug(fmt.Sprintf("Successfully marked accounting todo %d as completed in NocoDB", todoID))
	return nil
}
