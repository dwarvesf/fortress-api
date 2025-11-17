# Accounting Todo Payroll Integration Tasks

## Overview

Add NocoDB `accounting_todos` table support for payroll calculation. Currently, accounting expenses are only fetched from Basecamp. This implementation creates a separate `AccountingTodoService` to query NocoDB accounting todos with "Out" group filtering.

**Scope**: Payroll calculation accounting expenses only (not webhook handling)

---

## Phase 1: Create AccountingTodoService

### Task 1.1: Create Service Structure

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Create AccountingTodoService struct and constructor

**Requirements**:
- [ ] Create `AccountingTodoService` struct with fields:
  - `client *Service` (NocoDB client)
  - `cfg *config.Config`
  - `store *store.Store`
  - `repo store.DBRepo`
  - `logger logger.Logger`
- [ ] Create `NewAccountingTodoService` constructor
- [ ] Use `cfg.AccountingIntegration.Noco.TodosTableID` for table ID

**Code Changes**:
```go
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
```

**Acceptance Criteria**:
- [ ] AccountingTodoService struct defined
- [ ] Constructor creates instance with all dependencies
- [ ] Uses AccountingIntegration.Noco.TodosTableID config

---

### Task 1.2: Implement GetAllInList Method

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Fetch accounting todos filtered by "Out" group

**Requirements**:
- [ ] Query NocoDB API: `GET /tables/{accountingTodosTableID}/records`
- [ ] Filter: `where=(task_group,eq,out)~and(status,neq,completed)`
- [ ] Parse response: `{"list": [...]}`
- [ ] Transform each record to `bcModel.Todo`
- [ ] Extract fields: `title`, `assignee_ids`, `task_group`, `status`
- [ ] Parse assignee_ids array → resolve to employees
- [ ] Filter out Han's basecamp_id
- [ ] Handle single assignee requirement
- [ ] Add debug logging for transformation

**Code Changes**:
```go
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
```

**Acceptance Criteria**:
- [ ] Queries accounting_todos table with correct filter
- [ ] Returns only "Out" group todos
- [ ] Excludes completed todos
- [ ] Transforms to bcModel.Todo format
- [ ] Logs debug information

---

### Task 1.3: Implement transformRecordToTodo Method

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Transform NocoDB accounting todo record to Basecamp Todo format

**Requirements**:
- [ ] Extract record ID
- [ ] Extract title field
- [ ] Parse assignee_ids array (JSON string array)
- [ ] Resolve assignees to employees by basecamp_id
- [ ] Filter out Han's basecamp_id (consts.HanBasecampID)
- [ ] Require exactly 1 assignee after filtering
- [ ] Parse title format: "Description | Amount | Currency"
- [ ] Map task_group to Bucket name
- [ ] Set Completed=false (accounting todos are open)

**Code Changes**:
```go
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

	// Debug: log extracted fields
	a.logger.Debug(fmt.Sprintf("NocoDB accounting todo record %s - title: '%s', taskGroup: '%s', assigneeIds: %v",
		recordID, title, taskGroup, assigneeIDsRaw))

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

	// Create Todo object
	// Title is already in format: "Description | Amount | Currency"
	todo := &bcModel.Todo{
		ID:        id,
		Title:     title,
		Assignees: assignees,
		Bucket: bcModel.Bucket{
			ID:   id,
			Name: taskGroup,
		},
		Completed: false, // Accounting todos are open until paid
	}

	a.logger.Debug(fmt.Sprintf("Transformed accounting todo record %s to Todo: %s", recordID, title))

	return todo, nil
}
```

**Acceptance Criteria**:
- [ ] Extracts all required fields
- [ ] Parses assignee_ids correctly
- [ ] Filters Han's basecamp_id
- [ ] Returns error if not exactly 1 assignee
- [ ] Maps fields to bcModel.Todo

---

### Task 1.4: Implement Assignee Resolution Helpers

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Helper methods to parse and resolve assignee_ids

**Requirements**:
- [ ] Create `parseAssigneeIDs` - parse assignee_ids from various formats
- [ ] Create `resolveAssignees` - fetch employees and filter Han
- [ ] Handle assignee_ids as JSON array or comma-separated string
- [ ] Fetch employees by basecamp_id
- [ ] Skip employees not found (log warning)
- [ ] Filter out Han's basecamp_id (23147886 in prod, check consts)

**Code Changes**:
```go
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
	const HanBasecampID = 23147886 // TODO: Use consts.HanBasecampID

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
```

**Acceptance Criteria**:
- [ ] Parses assignee_ids from multiple formats
- [ ] Fetches employees by basecamp_id
- [ ] Filters out Han's basecamp_id
- [ ] Handles missing employees gracefully

---

### Task 1.5: Implement GetGroups Method

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Return "Out" group for accounting todos

**Requirements**:
- [ ] Return hardcoded TodoGroup with Title="Out"
- [ ] Use todosetID as group ID
- [ ] No actual groupBy query needed (we already filter in GetAllInList)

**Code Changes**:
```go
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
```

**Acceptance Criteria**:
- [ ] Returns TodoGroup with Title="out"
- [ ] Uses todosetID as group ID

---

### Task 1.6: Implement GetLists Method

**File**: `pkg/service/nocodb/accounting_todo.go`

**Description**: Return default list for accounting todos

**Requirements**:
- [ ] Return default TodoList with configured todosetID
- [ ] No views query needed (simplified implementation)

**Code Changes**:
```go
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
```

**Acceptance Criteria**:
- [ ] Returns TodoList with configured ID
- [ ] Name is descriptive

---

## Phase 2: Wire AccountingTodoService into Service Layer

### Task 2.1: Update Service Initialization

**File**: `pkg/service/service.go`

**Description**: Use AccountingTodoService for PayrollExpenseProvider when using NocoDB

**Requirements**:
- [ ] Import `nocodb` package
- [ ] Update PayrollExpenseProvider initialization (lines 250-259)
- [ ] Use AccountingTodoService instead of ExpenseService
- [ ] Keep ExpenseService for webhook handling (separate concern)

**Code Changes**:
```go
// Line 250-259 (current)
var payrollExpenseProvider basecamp.ExpenseProvider
if selectedProvider == string(taskprovider.ProviderNocoDB) && nocoSvc != nil {
	// Use NocoDB expense service for payroll
	payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
}
if payrollExpenseProvider == nil {
	// Fallback to Basecamp adapter
	payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}
```

**Updated code**:
```go
// Payroll expense fetcher provider (for fetching expenses during payroll calculation)
var payrollExpenseProvider basecamp.ExpenseProvider
if selectedProvider == string(taskprovider.ProviderNocoDB) && nocoSvc != nil {
	// Use NocoDB accounting todo service for payroll (accounting_todos table)
	payrollExpenseProvider = nocodb.NewAccountingTodoService(nocoSvc, cfg, store, repo, logger.L)
}
if payrollExpenseProvider == nil {
	// Fallback to Basecamp adapter
	payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}
```

**Acceptance Criteria**:
- [ ] PayrollExpenseProvider uses AccountingTodoService for NocoDB
- [ ] Basecamp fallback still works
- [ ] No changes to webhook ExpenseProvider

---

## Phase 3: Update Store Layer (if needed)

### Task 3.1: Check Employee.OneByBasecampID Exists

**File**: `pkg/store/employee/employee.go`

**Description**: Verify employee store has OneByBasecampID method

**Requirements**:
- [ ] Check if `OneByBasecampID(db *gorm.DB, basecampID int) (*model.Employee, error)` exists
- [ ] If not, create the method
- [ ] Query: `SELECT * FROM employees WHERE basecamp_id = ? AND deleted_at IS NULL`

**Code Changes** (if method doesn't exist):
```go
// OneByBasecampID gets an employee by basecamp_id
func (s *store) OneByBasecampID(db *gorm.DB, basecampID int) (*model.Employee, error) {
	var employee model.Employee
	return &employee, db.Where("basecamp_id = ?", basecampID).First(&employee).Error
}
```

**Acceptance Criteria**:
- [ ] Method exists in employee store
- [ ] Returns employee or error
- [ ] Respects soft deletes

---

## Phase 4: Testing

### Task 4.1: Create Unit Tests

**File**: `pkg/service/nocodb/accounting_todo_test.go`

**Description**: Unit tests for AccountingTodoService

**Requirements**:
- [ ] Test GetAllInList with mock NocoDB responses
- [ ] Test transformRecordToTodo with various inputs
- [ ] Test parseAssigneeIDs with different formats
- [ ] Test resolveAssignees with Han filtering
- [ ] Test GetGroups returns "Out" group
- [ ] Test GetLists returns default list
- [ ] Test error cases (missing fields, invalid data)

**Test Cases**:
1. GetAllInList success - returns filtered todos
2. GetAllInList filters out "in" group
3. GetAllInList filters out completed todos
4. transformRecordToTodo - valid record
5. transformRecordToTodo - missing title
6. transformRecordToTodo - multiple assignees (should error)
7. parseAssigneeIDs - JSON array
8. parseAssigneeIDs - comma-separated string
9. resolveAssignees - filters Han
10. resolveAssignees - handles missing employee

**Acceptance Criteria**:
- [ ] All test cases pass
- [ ] Code coverage > 80%
- [ ] Edge cases covered

---

## Phase 5: Documentation

### Task 5.1: Update NOCO_INTEGRATION_GUIDE.md

**File**: `docs/NOCO_INTEGRATION_GUIDE.md`

**Description**: Document AccountingTodoService usage

**Requirements**:
- [ ] Add section: "Accounting Todos for Payroll"
- [ ] Document task_group filtering ("out" vs "in")
- [ ] Document assignee_ids format and resolution
- [ ] Document Han filtering logic
- [ ] Add example accounting_todos records
- [ ] Document configuration (TodosTableID)

**Acceptance Criteria**:
- [ ] Guide explains accounting todo integration
- [ ] Configuration clearly documented
- [ ] Examples provided

---

### Task 5.2: Update CLAUDE.md

**File**: `/Users/quang/workspace/dwarvesf/fortress-api-feat-nocodb-migrate-expense/CLAUDE.md`

**Description**: Document architecture decision

**Requirements**:
- [ ] Add AccountingTodoService to architecture overview
- [ ] Explain separation between ExpenseService and AccountingTodoService
- [ ] Document when each service is used

**Acceptance Criteria**:
- [ ] Architecture section updated
- [ ] Clear distinction between services

---

### Task 5.3: Create ADR

**File**: `docs/adr/NNNN-accounting-todo-service.md`

**Description**: Architecture Decision Record for separate service

**Requirements**:
- [ ] Document decision to create separate AccountingTodoService
- [ ] List alternatives considered
- [ ] Explain rationale (different tables, schemas, domains)
- [ ] Document consequences

**Acceptance Criteria**:
- [ ] ADR created and numbered
- [ ] Decision clearly explained

---

## Summary Checklist

### Phase 1: AccountingTodoService
- [ ] Task 1.1: Service structure created
- [ ] Task 1.2: GetAllInList implemented
- [ ] Task 1.3: transformRecordToTodo implemented
- [ ] Task 1.4: Assignee resolution helpers implemented
- [ ] Task 1.5: GetGroups implemented
- [ ] Task 1.6: GetLists implemented

### Phase 2: Service Wiring
- [ ] Task 2.1: Service initialization updated

### Phase 3: Store Layer
- [ ] Task 3.1: OneByBasecampID method verified/created

### Phase 4: Testing
- [ ] Task 4.1: Unit tests created

### Phase 5: Documentation
- [ ] Task 5.1: NOCO_INTEGRATION_GUIDE updated
- [ ] Task 5.2: CLAUDE.md updated
- [ ] Task 5.3: ADR created

---

## Estimated Time

- Phase 1: 3-4 hours
- Phase 2: 30 minutes
- Phase 3: 30 minutes
- Phase 4: 1-2 hours
- Phase 5: 1 hour

**Total: 6-8 hours**

---

## Success Criteria

1. ✅ AccountingTodoService queries accounting_todos table
2. ✅ Only "Out" group todos fetched
3. ✅ Assignees correctly resolved by basecamp_id
4. ✅ Han's basecamp_id filtered out
5. ✅ Single assignee requirement enforced
6. ✅ Payroll calculation includes accounting expenses from NocoDB
7. ✅ No regression in Basecamp flow
8. ✅ All tests pass
9. ✅ Documentation complete
