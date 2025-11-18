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

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// ExpenseService fetches approved expenses from NocoDB API and transforms them to Basecamp Todo format
// for payroll calculation. This service mimics the Basecamp Todo service interface.
type ExpenseService struct {
	client *Service
	cfg    *config.Config
	store  *store.Store
	repo   store.DBRepo
	logger logger.Logger
}

// NewExpenseService creates a new NocoDB expense service for payroll expense fetching.
func NewExpenseService(client *Service, cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *ExpenseService {
	return &ExpenseService{
		client: client,
		cfg:    cfg,
		store:  store,
		repo:   repo,
		logger: logger,
	}
}

// GetAllInList fetches all approved expenses from NocoDB and transforms them to Basecamp Todo format.
// todolistID is mapped to NocoDB table/view ID
// projectID is mapped to NocoDB workspace/base ID (currently ignored, using config)
func (e *ExpenseService) GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error) {
	if e.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	ctx := context.Background()
	tableID := e.cfg.ExpenseIntegration.Noco.TableID
	if tableID == "" {
		e.logger.Error(errors.New("expense table id not configured"), "nocodb expense table id is empty")
		return nil, errors.New("nocodb expense table id not configured")
	}

	e.logger.Debug(fmt.Sprintf("Fetching expenses from NocoDB table %s (todolistID=%d, projectID=%d)", tableID, todolistID, projectID))

	// Query NocoDB API for approved expenses
	path := fmt.Sprintf("/tables/%s/records", tableID)
	query := url.Values{}
	// Filter for approved expenses only
	query.Set("where", "(status,eq,approved)")
	query.Set("limit", "100") // Adjust limit as needed

	resp, err := e.client.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		e.logger.Error(err, "failed to fetch expenses from nocodb")
		return nil, fmt.Errorf("nocodb get expenses failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		e.logger.Error(fmt.Errorf("nocodb returned status %s", resp.Status), fmt.Sprintf("nocodb get expenses failed: %s", string(body)))
		return nil, fmt.Errorf("nocodb get expenses failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		e.logger.Error(err, "failed to decode nocodb expense response")
		return nil, fmt.Errorf("failed to decode nocodb response: %w", err)
	}

	e.logger.Debug(fmt.Sprintf("Fetched %d expense records from NocoDB", len(result.List)))

	// Transform each NocoDB record to bcModel.Todo
	todos := make([]bcModel.Todo, 0, len(result.List))
	for _, record := range result.List {
		todo, err := e.transformRecordToTodo(record)
		if err != nil {
			// Log warning but continue processing other records
			e.logger.Error(err, fmt.Sprintf("failed to transform expense record %s", extractRecordID(record)))
			continue
		}
		todos = append(todos, *todo)
	}

	e.logger.Debug(fmt.Sprintf("Transformed %d expenses to Todo format", len(todos)))

	return todos, nil
}

// GetGroups fetches expense groups from NocoDB (e.g., "Out" group in accounting).
// todosetID is mapped to NocoDB table ID with grouping configuration
// projectID is mapped to NocoDB workspace/base ID (currently ignored)
func (e *ExpenseService) GetGroups(todosetID, projectID int) ([]bcModel.TodoGroup, error) {
	if e.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	ctx := context.Background()
	tableID := e.cfg.ExpenseIntegration.Noco.TableID
	if tableID == "" {
		e.logger.Error(errors.New("expense table id not configured"), "nocodb expense table id is empty")
		return nil, errors.New("nocodb expense table id not configured")
	}

	e.logger.Debug(fmt.Sprintf("Fetching expense groups from NocoDB table %s (todosetID=%d, projectID=%d)", tableID, todosetID, projectID))

	// Query NocoDB API with grouping
	// Note: NocoDB grouping API may vary - this is a placeholder implementation
	// Adjust based on actual NocoDB API capabilities
	path := fmt.Sprintf("/tables/%s/records", tableID)
	query := url.Values{}
	query.Set("where", "(status,eq,approved)~or(status,eq,completed)")
	query.Set("groupBy", "task_group") // Adjust field name based on actual schema
	query.Set("limit", "100")

	resp, err := e.client.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		e.logger.Error(err, "failed to fetch expense groups from nocodb")
		return nil, fmt.Errorf("nocodb get expense groups failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		e.logger.Error(fmt.Errorf("nocodb returned status %s", resp.Status), fmt.Sprintf("nocodb get expense groups failed: %s", string(body)))
		return nil, fmt.Errorf("nocodb get expense groups failed: %s - %s", resp.Status, string(body))
	}

	// Parse grouped response
	// Note: Response format may vary based on NocoDB version/configuration
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		e.logger.Error(err, "failed to decode nocodb expense groups response")
		return nil, fmt.Errorf("failed to decode nocodb groups response: %w", err)
	}

	// Transform to TodoGroup
	// This is a simplified implementation - adjust based on actual NocoDB response structure
	groups := []bcModel.TodoGroup{
		{
			ID:   todosetID, // Use todosetID as default group ID
			Name: "Expenses",
		},
	}

	e.logger.Debug(fmt.Sprintf("Returning %d expense groups", len(groups)))

	return groups, nil
}

// GetLists fetches expense lists/views from NocoDB.
// projectID is mapped to NocoDB base/workspace ID
// todosetID is currently ignored
func (e *ExpenseService) GetLists(projectID, todosetID int) ([]bcModel.TodoList, error) {
	if e.client == nil {
		return nil, errors.New("nocodb client is nil")
	}

	ctx := context.Background()
	tableID := e.cfg.ExpenseIntegration.Noco.TableID
	if tableID == "" {
		e.logger.Error(errors.New("expense table id not configured"), "nocodb expense table id is empty")
		return nil, errors.New("nocodb expense table id not configured")
	}

	e.logger.Debug(fmt.Sprintf("Fetching expense lists from NocoDB table %s (projectID=%d, todosetID=%d)", tableID, projectID, todosetID))

	// Query NocoDB API for views
	// Note: This endpoint may vary based on NocoDB API version
	path := fmt.Sprintf("/tables/%s/views", tableID)

	resp, err := e.client.makeRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		e.logger.Error(err, "failed to fetch expense lists from nocodb")
		// Fallback: return default list if views endpoint not available
		return e.getDefaultExpenseList(todosetID), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		e.logger.Error(fmt.Errorf("nocodb returned status %s", resp.Status), fmt.Sprintf("nocodb get expense lists failed: %s", string(body)))
		// Fallback: return default list
		return e.getDefaultExpenseList(todosetID), nil
	}

	var result struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		e.logger.Error(err, "failed to decode nocodb expense lists response")
		// Fallback: return default list
		return e.getDefaultExpenseList(todosetID), nil
	}

	// Transform views to TodoList
	lists := make([]bcModel.TodoList, 0, len(result.List))
	for _, view := range result.List {
		list := bcModel.TodoList{
			ID:   extractIntID(view),
			Name: extractString(view, "title"),
		}
		lists = append(lists, list)
	}

	if len(lists) == 0 {
		// Fallback if no views returned
		return e.getDefaultExpenseList(todosetID), nil
	}

	e.logger.Debug(fmt.Sprintf("Returning %d expense lists", len(lists)))

	return lists, nil
}

// transformRecordToTodo transforms a NocoDB expense record to Basecamp Todo format.
func (e *ExpenseService) transformRecordToTodo(record map[string]interface{}) (*bcModel.Todo, error) {
	// Extract fields from NocoDB record
	recordID := extractRecordID(record)
	if recordID == "" {
		return nil, errors.New("missing record ID")
	}

	requesterEmail := extractString(record, "requester_team_email")
	if requesterEmail == "" {
		return nil, errors.New("missing requester_team_email")
	}

	amount := extractFloat(record, "amount")
	currency := extractString(record, "currency")
	title := extractString(record, "title")       // NocoDB uses "title" field
	taskBoard := extractString(record, "task_board")

	// Debug: log extracted fields
	e.logger.Debug(fmt.Sprintf("NocoDB expense record %s - title: '%s', amount: %.0f, currency: '%s', taskBoard: '%s'",
		recordID, title, amount, currency, taskBoard))

	// Fetch employee by email to get basecamp_id
	employee, err := e.store.Employee.OneByEmail(e.repo.DB(), requesterEmail)
	if err != nil {
		e.logger.Error(err, fmt.Sprintf("failed to find employee by email: %s", requesterEmail))
		return nil, fmt.Errorf("employee not found for email %s: %w", requesterEmail, err)
	}

	if employee.BasecampID == 0 {
		e.logger.Error(errors.New("employee has no basecamp_id"), fmt.Sprintf("employee missing basecamp_id: %s", requesterEmail))
		return nil, fmt.Errorf("employee %s has no basecamp_id", requesterEmail)
	}

	// Convert record ID to int
	id, err := strconv.Atoi(recordID)
	if err != nil {
		e.logger.Error(err, fmt.Sprintf("failed to convert record ID to int: %s", recordID))
		return nil, fmt.Errorf("invalid record ID %s: %w", recordID, err)
	}

	// Build Todo title in Basecamp format: "title | amount | currency"
	// This matches the format expected by payroll calculator's getReimbursement function
	todoTitle := fmt.Sprintf("%s | %.0f | %s", title, amount, currency)

	// Create Todo object
	todo := &bcModel.Todo{
		ID:    id,
		Title: todoTitle,
		Assignees: []bcModel.Assignee{
			{
				ID:   employee.BasecampID,
				Name: employee.FullName,
			},
		},
		Bucket: bcModel.Bucket{
			ID:   id, // Use record ID as bucket ID
			Name: taskBoard,
		},
		Completed: true, // Expense is approved/completed
	}

	e.logger.Debug(fmt.Sprintf("Transformed expense record %s to Todo: %s", recordID, todoTitle))

	return todo, nil
}

// getDefaultExpenseList returns a default expense list when views cannot be fetched.
func (e *ExpenseService) getDefaultExpenseList(todosetID int) []bcModel.TodoList {
	return []bcModel.TodoList{
		{
			ID:   todosetID,
			Name: "Expenses",
		},
	}
}

// Helper functions

func extractString(record map[string]interface{}, key string) string {
	if val, ok := record[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func extractFloat(record map[string]interface{}, key string) float64 {
	if val, ok := record[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func extractIntID(record map[string]interface{}) int {
	idStr := extractRecordID(record)
	if idStr == "" {
		return 0
	}
	if id, err := strconv.Atoi(idStr); err == nil {
		return id
	}
	return 0
}

// MarkExpenseAsCompleted marks an expense submission as completed in NocoDB
func (e *ExpenseService) MarkExpenseAsCompleted(expenseID int) error {
	if e.client == nil {
		return errors.New("nocodb client is nil")
	}

	tableID := e.cfg.ExpenseIntegration.Noco.TableID
	if tableID == "" {
		e.logger.Error(errors.New("expense table id not configured"), "nocodb expense table id is empty")
		return errors.New("nocodb expense table id not configured")
	}

	e.logger.Debug(fmt.Sprintf("Marking expense submission %d as completed in NocoDB table %s", expenseID, tableID))

	ctx := context.Background()
	path := fmt.Sprintf("/tables/%s/records", tableID)

	// Build PATCH payload
	payload := map[string]interface{}{
		"Id":     expenseID,
		"status": "completed",
	}

	payloadJSON, _ := json.Marshal(payload)
	e.logger.Debug(fmt.Sprintf("Sending PATCH request to NocoDB: %s with payload: %s", path, string(payloadJSON)))

	resp, err := e.client.makeRequest(ctx, http.MethodPatch, path, nil, payload)
	if err != nil {
		e.logger.Error(err, fmt.Sprintf("failed to mark expense %d as completed in nocodb", expenseID))
		return fmt.Errorf("nocodb mark expense as completed failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	e.logger.Debug(fmt.Sprintf("NocoDB PATCH response status: %s, body: %s", resp.Status, string(body)))

	if resp.StatusCode >= 300 {
		e.logger.Error(fmt.Errorf("nocodb returned status %s", resp.Status), fmt.Sprintf("nocodb mark expense %d as completed failed: %s", expenseID, string(body)))
		return fmt.Errorf("nocodb mark expense %d as completed failed: %s - %s", expenseID, resp.Status, string(body))
	}

	e.logger.Debug(fmt.Sprintf("Successfully marked expense submission %d as completed in NocoDB", expenseID))
	return nil
}
