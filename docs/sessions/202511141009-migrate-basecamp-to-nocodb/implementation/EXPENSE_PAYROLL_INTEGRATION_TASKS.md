# NocoDB Expense Payroll Integration - Task Breakdown

## Overview
Refactor NocoDB expense workflow to match Basecamp's payroll integration pattern where expenses are fetched from NocoDB API during payroll calculation (not persisted on approval), and only saved to DB when payroll is committed.

## Status Legend
- 🔴 Not Started
- 🟡 In Progress
- 🟢 Completed

---

## Phase 1: Payroll Expense Fetcher Interface & Basecamp Adapter (Foundation)

**Important Note**: This is DIFFERENT from the existing `taskprovider.ExpenseProvider` interface (which handles webhook operations). We're creating a NEW interface specifically for fetching expenses during payroll calculation.

### Task 1.1: Define PayrollExpenseFetcher Interface 🟢
**File**: `pkg/service/basecamp/basecamp.go` (update existing file)

**Description**: The ExpenseProvider interface already exists in basecamp.go (lines 50-55), fixed typo in return types

**Interface** (corrected):
```go
// ExpenseProvider defines methods for fetching expense todos for payroll calculation.
type ExpenseProvider interface {
	GetAllInList(todolistID, projectID int) ([]model.Todo, error)
	GetGroups(todosetID, projectID int) ([]model.TodoList, error)
	GetLists(projectID, todosetID int) ([]model.TodoList, error)
}
```

**Changes Made**:
- ✅ Fixed typo: `model.Todolist` → `model.TodoList`
- ✅ Interface now compiles correctly

**Acceptance Criteria**:
- ✅ Interface exists and is correct
- ✅ Return types match actual Basecamp model types (TodoList, not Todolist)
- ✅ Build succeeds

**Status**: COMPLETED

---

### Task 1.2: Implement Basecamp Todo Service as ExpenseProvider 🔴
**File**: `pkg/service/basecamp/todo/todo.go` (update existing)

**Description**: Basecamp Todo service already has the required methods (GetAllInList, GetGroups, GetLists). We just need to verify the method signatures match the ExpenseProvider interface.

**Existing Methods** (already implemented):
- Line 142: `GetAllInList(todoListID, projectID int, query ...string) ([]model.Todo, error)`
- Line 200: `GetGroups(todoListID, projectID int) ([]model.TodoGroup, error)`
- Line 225: `GetLists(projectID, todoSetsID int) ([]model.TodoList, error)`

**Issues to fix**:
1. `GetGroups` returns `[]model.TodoGroup` but interface expects `[]model.Todolist`
2. `GetLists` returns `[]model.TodoList` but interface expects `[]model.Todolist`
3. `GetAllInList` has optional `query ...string` parameter not in interface

**Solution**: The interface definition in basecamp.go may be incorrect. We need to:
- Check if `model.Todolist` is an alias or different type
- Possibly update interface to match actual return types
- OR create adapter methods that convert between types

**Acceptance Criteria**:
- ✅ Method signatures verified
- ✅ Type mismatches resolved
- ✅ Basecamp Todo service can satisfy ExpenseProvider interface

**Estimated Time**: 30 minutes

---

## Phase 2: NocoDB Expense Service (Core Implementation)

### Task 2.1: Create NocoDB Expense Service Structure 🔴
**File**: `pkg/service/nocodb/expense.go` (new file)

**Description**: Create service struct and constructor for fetching expenses from NocoDB API

**Requirements**:
- Service struct with dependencies:
  - `client *Service` - NocoDB API client
  - `cfg *config.Config` - Configuration
  - `store *store.Store` - Database store for employee lookups
  - `repo store.DBRepo` - Database repository
  - `logger logger.Logger` - Logger
- Constructor: `func NewExpenseService(client *Service, cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *ExpenseService`

**Code Template**:
```go
package nocodb

import (
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
    "github.com/dwarvesf/fortress-api/pkg/store"
)

type ExpenseService struct {
    client *Service
    cfg    *config.Config
    store  *store.Store
    repo   store.DBRepo
    logger logger.Logger
}

func NewExpenseService(client *Service, cfg *config.Config, store *store.Store, repo store.DBRepo, logger logger.Logger) *ExpenseService {
    return &ExpenseService{
        client: client,
        cfg:    cfg,
        store:  store,
        repo:   repo,
        logger: logger,
    }
}
```

**Acceptance Criteria**:
- ✅ File created with package declaration
- ✅ ExpenseService struct defined with all dependencies
- ✅ Constructor function implemented
- ✅ Imports added for required packages

**Estimated Time**: 20 minutes

---

### Task 2.2: Implement GetAllInList Method 🔴
**File**: `pkg/service/nocodb/expense.go`

**Description**: Fetch approved expenses from NocoDB and transform to Basecamp Todo format

**Requirements**:
- Map `todolistID` to NocoDB table/view ID (config-based mapping)
- Query NocoDB API: `GET /api/v2/tables/{tableId}/records`
- Filter for `status = 'approved'` or `status = 'completed'`
- For each expense record:
  - Fetch employee by `requester_email` to get `basecamp_id`
  - Parse `amount`, `currency`, `reason`, `task_board`
  - Transform to `bcModel.Todo` structure
- Handle pagination if needed
- Log all API calls for debugging

**Transformation Logic**:
```go
// NocoDB record → Basecamp Todo
Todo{
    ID: int(nocoRecord["Id"]),
    Title: fmt.Sprintf("%s %d %s", reason, amount, currency),
    Assignees: []Assignee{{ID: employeeBasecampID}},
    Bucket: Bucket{ID: parsedBucketID, Name: taskBoard},
    Completed: true,
}
```

**Error Handling**:
- Return error if NocoDB API call fails
- Return error if employee not found by email
- Log warning if record has invalid data (skip, don't fail)

**Acceptance Criteria**:
- ✅ Method signature matches interface: `GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error)`
- ✅ Queries NocoDB API with correct endpoint
- ✅ Filters for approved/completed expenses only
- ✅ Transforms all fields correctly to `bcModel.Todo`
- ✅ Handles employee lookup by email
- ✅ Logs debug information for each API call
- ✅ Returns errors appropriately

**Estimated Time**: 2-3 hours

---

### Task 2.3: Implement GetGroups Method 🔴
**File**: `pkg/service/nocodb/expense.go`

**Description**: Fetch expense groups from NocoDB (e.g., "Out" group in accounting)

**Requirements**:
- Map `todosetID` to NocoDB table ID with grouping configuration
- Query NocoDB API with grouping: `GET /api/v2/tables/{tableId}/records?groupBy=task_group`
- Transform grouped results to `bcModel.Todolist` structure
- Each group becomes a separate Todolist

**Transformation Logic**:
```go
// Group → Todolist
Todolist{
    ID: generatedID,  // Generate from group name or index
    Name: groupName,
    TodosCount: len(todosInGroup),
}
```

**Acceptance Criteria**:
- ✅ Method signature: `GetGroups(todosetID, projectID int) ([]bcModel.Todolist, error)`
- ✅ Queries NocoDB with groupBy parameter
- ✅ Transforms groups to Todolist format
- ✅ Handles empty groups gracefully
- ✅ Logs API calls

**Estimated Time**: 1-2 hours

---

### Task 2.4: Implement GetLists Method 🔴
**File**: `pkg/service/nocodb/expense.go`

**Description**: Fetch expense lists/views from NocoDB

**Requirements**:
- Map `todosetID` to NocoDB base/workspace ID
- Query NocoDB API: `GET /api/v2/tables/{tableId}/views`
- Transform views to `bcModel.Todolist` structure

**Transformation Logic**:
```go
// View → Todolist
Todolist{
    ID: int(view["Id"]),
    Name: view["title"],
    // Other fields as needed
}
```

**Acceptance Criteria**:
- ✅ Method signature: `GetLists(projectID, todosetID int) ([]bcModel.Todolist, error)`
- ✅ Queries NocoDB views API
- ✅ Transforms views to Todolist format
- ✅ Returns error on API failure

**Estimated Time**: 1 hour

---

## Phase 3: Provider Wiring & Injection

### Task 3.1: Update Service Layer Provider Initialization 🔴
**File**: `pkg/service/service.go`

**Description**: Wire up the new expense provider in service initialization

**Requirements**:
- Import new Basecamp expense adapter
- Import NocoDB expense service
- Create expense provider based on `TASK_PROVIDER` config
- Inject into Service struct

**Code Changes**:
```go
// After line 220 (after nocodbProvider initialization)

var expenseProvider taskprovider.ExpenseProvider
if selectedProvider == string(taskprovider.ProviderNocoDB) && nocodbProvider != nil {
    // Create NocoDB expense service
    nocoExpenseService := nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
    expenseProvider = nocoExpenseService
}
if expenseProvider == nil {
    // Fallback to Basecamp
    expenseProvider = tpbasecamp.NewExpenseProvider(basecampSvc)
}

// Add to Service struct return (line 249+)
return &Service{
    // ... existing fields
    ExpenseProvider: expenseProvider,  // Add this line
}
```

**Acceptance Criteria**:
- ✅ Expense provider created based on `TASK_PROVIDER` config
- ✅ NocoDB provider used when `TASK_PROVIDER=nocodb`
- ✅ Basecamp provider used as fallback
- ✅ Provider injected into Service struct
- ✅ Build succeeds with no errors

**Estimated Time**: 30 minutes

---

### Task 3.2: Update Payroll Handler Initialization 🔴
**File**: `pkg/handler/payroll/payroll.go`

**Description**: Inject ExpenseProvider into payroll handler

**Requirements**:
- Add `expenseProvider taskprovider.ExpenseProvider` field to handler struct
- Update constructor to accept and inject provider
- Provider should come from service layer

**Code Changes**:
```go
type handler struct {
    // ... existing fields
    expenseProvider taskprovider.ExpenseProvider
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, worker worker.Worker, logger logger.Logger) *handler {
    return &handler{
        // ... existing fields
        expenseProvider: service.ExpenseProvider,
    }
}
```

**Acceptance Criteria**:
- ✅ Handler struct has `expenseProvider` field
- ✅ Constructor injects provider from service layer
- ✅ Build succeeds

**Estimated Time**: 15 minutes

---

## Phase 4: Payroll Calculator Refactoring

### Task 4.1: Update calculatePayrolls to Use ExpenseProvider 🔴
**File**: `pkg/handler/payroll/payroll_calculator.go`

**Description**: Replace direct Basecamp service calls with expense provider interface

**Requirements**:
- Replace all `h.service.Basecamp.Todo.*` calls with `h.expenseProvider.*`
- Maintain exact same logic flow
- No changes to payroll calculation logic

**Code Changes**:
```go
// Line 54-110: Replace Basecamp calls
// OLD:
opsTodoLists, err := h.service.Basecamp.Todo.GetAllInList(opsExpenseID, opsID)
todolists, err := h.service.Basecamp.Todo.GetGroups(expenseID, woodlandID)

// NEW:
opsTodoLists, err := h.expenseProvider.GetAllInList(opsExpenseID, opsID)
todolists, err := h.expenseProvider.GetGroups(expenseID, woodlandID)
```

**Files to Update**:
- Lines 54-110: Main expense fetching logic
- Any other direct Basecamp Todo calls in calculatePayrolls method

**Acceptance Criteria**:
- ✅ All Basecamp.Todo calls replaced with expenseProvider calls
- ✅ No logic changes to payroll calculation
- ✅ Method signatures remain the same
- ✅ Error handling preserved
- ✅ Build succeeds
- ✅ Tests pass (if existing)

**Estimated Time**: 1 hour

---

### Task 4.2: Update getAccountingExpense to Use ExpenseProvider 🔴
**File**: `pkg/handler/payroll/payroll_calculator.go`

**Description**: Refactor accounting expense fetching to use provider interface

**Requirements**:
- Replace Basecamp service calls in `getAccountingExpense` method
- Lines 397-437

**Code Changes**:
```go
// Line 397-437
func (h *handler) getAccountingExpense(batch int) (res []bcModel.Todo, err error) {
    // OLD:
    lists, err := h.service.Basecamp.Todo.GetLists(accountingID, accountingTodoID)
    groups, err := h.service.Basecamp.Todo.GetGroups(lists[i].ID, accountingID)
    todos, err := h.service.Basecamp.Todo.GetAllInList(groups[j].ID, accountingID)

    // NEW:
    lists, err := h.expenseProvider.GetLists(accountingID, accountingTodoID)
    groups, err := h.expenseProvider.GetGroups(lists[i].ID, accountingID)
    todos, err := h.expenseProvider.GetAllInList(groups[j].ID, accountingID)
}
```

**Acceptance Criteria**:
- ✅ All Basecamp.Todo calls replaced in getAccountingExpense
- ✅ No logic changes
- ✅ Build succeeds

**Estimated Time**: 30 minutes

---

## Phase 5: Webhook Handler Cleanup

### Task 5.1: Remove DB Persistence from NocoDB Webhook Handler 🔴
**File**: `pkg/service/taskprovider/nocodb/provider.go`

**Description**: Update CreateExpense to only validate, not persist to DB

**Requirements**:
- Remove all DB insert logic for `expenses` table
- Remove all DB insert logic for `accounting_transactions` table
- Keep validation logic
- Keep feedback/logging
- Return success after validation

**Code Changes**:
```go
func (p *Provider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
    p.logger.Info("Processing expense approval webhook from NocoDB")

    // Validation only
    if payload.EmployeeID == nil {
        p.logger.Error(errors.New("employee not found"), "expense validation failed")
        return nil, errors.New("employee not found")
    }

    // Log approval event for debugging
    p.logger.Info("Expense approved in NocoDB",
        "employee_id", *payload.EmployeeID,
        "amount", payload.Amount,
        "currency", payload.Currency,
        "task_ref", payload.TaskRef,
    )

    // Return success - no DB persistence
    return &taskprovider.ExpenseTaskRef{
        TaskRef:   payload.TaskRef,
        ProjectID: payload.ProjectID,
    }, nil
}
```

**Acceptance Criteria**:
- ✅ All DB insert code removed
- ✅ All accounting transaction code removed
- ✅ Validation logic preserved
- ✅ Logging added for debugging
- ✅ Returns success after validation
- ✅ Build succeeds

**Estimated Time**: 1 hour

---

## Phase 6: Configuration & Environment

### Task 6.1: Add NocoDB Expense Configuration 🔴
**File**: `pkg/config/config.go`

**Description**: Add configuration for NocoDB expense integration

**Requirements**:
- Add to NocoConfig struct:
  - ExpenseBaseID
  - ExpenseTableID
  - ExpenseViewID
- Add environment variable mapping

**Code Changes**:
```go
type NocoConfig struct {
    // ... existing fields

    // Expense integration
    ExpenseBaseID  string `env:"NOCODB_EXPENSE_BASE_ID"`
    ExpenseTableID string `env:"NOCODB_EXPENSE_TABLE_ID"`
    ExpenseViewID  string `env:"NOCODB_EXPENSE_VIEW_ID"`
}
```

**Acceptance Criteria**:
- ✅ Config struct updated with new fields
- ✅ Environment variables mapped correctly
- ✅ Build succeeds

**Estimated Time**: 15 minutes

---

### Task 6.2: Update .env.example 🔴
**File**: `.env.example`

**Description**: Document new environment variables

**Code Changes**:
```bash
# NocoDB Expense Integration
NOCODB_EXPENSE_BASE_ID=<base_id>
NOCODB_EXPENSE_TABLE_ID=<table_id>
NOCODB_EXPENSE_VIEW_ID=<view_id>
```

**Acceptance Criteria**:
- ✅ All new env vars documented
- ✅ Example values provided
- ✅ Comments explain purpose

**Estimated Time**: 10 minutes

---

## Phase 7: Testing & Validation

### Task 7.1: Manual Testing - Basecamp Flow 🔴

**Description**: Verify existing Basecamp flow still works

**Test Steps**:
1. Set `TASK_PROVIDER=basecamp` in .env
2. Restart application
3. Trigger payroll calculation for test month
4. Verify expenses fetched from Basecamp
5. Verify payroll calculation includes expenses
6. Commit payroll
7. Verify expenses persisted to DB

**Acceptance Criteria**:
- ✅ Basecamp provider used when `TASK_PROVIDER=basecamp`
- ✅ Expenses fetched correctly
- ✅ Payroll calculation works
- ✅ Expenses saved to DB on commit
- ✅ No errors in logs

**Estimated Time**: 1 hour

---

### Task 7.2: Manual Testing - NocoDB Flow 🔴

**Description**: Verify new NocoDB flow works correctly

**Test Steps**:
1. Set `TASK_PROVIDER=nocodb` in .env
2. Set NocoDB expense config (BASE_ID, TABLE_ID, VIEW_ID)
3. Restart application
4. Approve expense in NocoDB (status = "approved")
5. Verify webhook returns 200 OK
6. Verify NO expense record created in DB
7. Trigger payroll calculation
8. Verify expenses fetched from NocoDB API
9. Verify payroll calculation includes expenses
10. Commit payroll
11. Verify expenses NOW persisted to DB

**Acceptance Criteria**:
- ✅ NocoDB provider used when `TASK_PROVIDER=nocodb`
- ✅ Webhook succeeds but doesn't create DB records
- ✅ Expenses fetched from NocoDB API during payroll
- ✅ Transformation to bcModel.Todo works correctly
- ✅ Employee matching by email works
- ✅ Payroll calculation includes expenses
- ✅ Expenses saved to DB on payroll commit
- ✅ No errors in logs

**Estimated Time**: 2 hours

---

### Task 7.3: Data Migration - Clean Up Existing Expenses 🔴

**Description**: Handle expenses created by old webhook implementation

**Options**:
1. **Recommended**: Delete expense records created before implementation date
2. **Alternative**: Keep for historical reference

**Implementation** (Option 1):
```sql
-- Delete expenses created by webhook (before payroll commit)
-- Check by looking for expenses without corresponding payroll records
DELETE FROM expenses
WHERE created_at >= '2025-11-14'  -- Date new webhook was deployed
AND id NOT IN (
    SELECT DISTINCT expense_id
    FROM payroll_expenses  -- Adjust based on actual schema
    WHERE expense_id IS NOT NULL
);
```

**Acceptance Criteria**:
- ✅ Migration plan documented
- ✅ SQL script tested on staging
- ✅ Backup taken before execution
- ✅ Verified no active payroll data affected

**Estimated Time**: 1 hour

---

## Phase 8: Documentation

### Task 8.1: Update Developer Documentation 🔴

**Files to Update**:
- `CLAUDE.md` - Add expense provider pattern to architecture section
- `README.md` - Document new env vars

**Content**:
- Explain expense provider abstraction
- Document NocoDB vs Basecamp flow difference
- Add configuration instructions
- Update architecture diagrams if needed

**Acceptance Criteria**:
- ✅ Documentation updated
- ✅ Configuration examples provided
- ✅ Architecture section reflects changes

**Estimated Time**: 1 hour

---

## Summary & Checklist

### Implementation Order
1. ✅ Phase 1: Provider Interface & Basecamp Adapter (1 hour)
2. ✅ Phase 2: NocoDB Expense Service (4-6 hours)
3. ✅ Phase 3: Provider Wiring (45 minutes)
4. ✅ Phase 4: Payroll Calculator Refactoring (1.5 hours)
5. ✅ Phase 5: Webhook Cleanup (1 hour)
6. ✅ Phase 6: Configuration (25 minutes)
7. ✅ Phase 7: Testing & Migration (4 hours)
8. ✅ Phase 8: Documentation (1 hour)

### Total Estimated Time: 13-15 hours

### Critical Path Items
1. Task 2.2: GetAllInList implementation (most complex transformation logic)
2. Task 7.2: NocoDB flow testing (validates entire implementation)
3. Task 7.3: Data migration (production impact)

### Dependencies
- Phase 2 depends on Phase 1 (interface must exist)
- Phase 3 depends on Phase 2 (NocoDB service must exist)
- Phase 4 depends on Phase 3 (provider must be injected)
- Phase 7 depends on all previous phases

### Risk Areas
1. **NocoDB API transformation**: Mapping NocoDB records to Basecamp Todo format may have edge cases
2. **Employee matching**: Email-based lookup might fail if emails don't match
3. **Data migration**: Existing expenses in DB need careful handling
4. **Configuration mapping**: todolistID/projectID to NocoDB IDs requires clear mapping logic

### Success Criteria
- ✅ Both Basecamp and NocoDB providers work correctly
- ✅ Payroll calculation unchanged (same logic, different data source)
- ✅ Expenses NOT created on approval for NocoDB
- ✅ Expenses created on payroll commit for both providers
- ✅ All existing tests pass
- ✅ No production errors after deployment
