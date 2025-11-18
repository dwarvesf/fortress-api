# NocoDB Expense Payroll Integration - Implementation Plan

## Current State vs Target State

### Current Implementation (Incorrect)
- ✅ NocoDB webhook creates expense records in DB immediately on approval
- ❌ Payroll calculator expects to fetch from Basecamp API, not DB
- ❌ No NocoDB expense service to mimic Basecamp Todo service

### Target Implementation (Following Basecamp Flow)
- ❌ NocoDB webhook does NOT create DB records
- ✅ NocoDB expense service fetches from NocoDB API during payroll calculation
- ✅ Expenses persisted to DB only when payroll is committed

---

## Implementation Changes

### 1. Update NocoDB Webhook Handler
**File**: `pkg/service/taskprovider/nocodb/provider.go`

**Current behavior**:
```go
func (p *Provider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
    // Creates expense record in DB
    // Creates accounting transaction
    // Links them together
}
```

**New behavior**:
```go
func (p *Provider) CreateExpense(ctx context.Context, payload *taskprovider.ExpenseWebhookPayload) (*taskprovider.ExpenseTaskRef, error) {
    // Validation only
    // No DB persistence
    // Return success
}
```

**Changes**:
- Remove all DB insert logic for `expenses` table
- Remove all DB insert logic for `accounting_transactions` table
- Keep only validation/feedback logic
- Log approval event for debugging

---

### 2. Create NocoDB Expense Service
**New File**: `pkg/service/nocodb/expense.go`

**Purpose**: Fetch approved expenses from NocoDB API and transform to Basecamp Todo format

**Interface**:
```go
type ExpenseService interface {
    // Fetch all expenses in a specific list (e.g., team expenses)
    GetAllInList(todolistID, projectID int) ([]bcModel.Todo, error)

    // Fetch expense groups (e.g., "Out" group in accounting)
    GetGroups(todosetID, projectID int) ([]bcModel.Todolist, error)

    // Fetch accounting todo lists
    GetLists(projectID, todosetID int) ([]bcModel.Todolist, error)
}
```

**Implementation details**:

#### GetAllInList(todolistID, projectID int)
- Map `todolistID` to NocoDB view/table ID
- Map `projectID` to NocoDB workspace/base ID
- Query NocoDB API: `GET /api/v2/tables/{tableId}/records`
- Filter: `status = 'approved'` or `status = 'completed'`
- Transform each NocoDB record to `bcModel.Todo`:
  ```go
  Todo{
      ID: int(nocoRecord["Id"]),
      Title: fmt.Sprintf("%s %d %s", reason, amount, currency),
      Assignees: []Assignee{{ID: employeeBasecampID}},
      Bucket: Bucket{ID: parsedBucketID, Name: taskBoard},
      Completed: true,
  }
  ```

#### GetGroups(todosetID, projectID int)
- Map to NocoDB grouped view
- Query NocoDB API with grouping: `GET /api/v2/tables/{tableId}/records?groupBy=task_group`
- Return groups as `bcModel.Todolist` with group names

#### GetLists(projectID, todosetID int)
- Map to NocoDB views/boards
- Query NocoDB API: `GET /api/v2/tables/{tableId}/views`
- Transform views to `bcModel.Todolist`

**Mapping Logic**:
```go
// NocoDB record → Basecamp Todo
nocoRecord["requester_email"] → fetch employee.basecamp_id → Todo.Assignees
nocoRecord["amount"] + nocoRecord["currency"] → parse to Todo.Title format
nocoRecord["reason"] → prepend to Todo.Title
nocoRecord["task_board"] → Todo.Bucket.Name
nocoRecord["Id"] → Todo.ID
```

---

### 3. Update Payroll Calculator
**File**: `pkg/handler/payroll/payroll_calculator.go`

**Current code**:
```go
// Line 54-110: Fetch from Basecamp
opsTodoLists, err := h.service.Basecamp.Todo.GetAllInList(opsExpenseID, opsID)
todolists, err := h.service.Basecamp.Todo.GetGroups(expenseID, woodlandID)
accountingExpenses, err := h.getAccountingExpense(batch)
```

**New code**:
```go
// Use injected expense provider
opsTodoLists, err := h.expenseProvider.GetAllInList(opsExpenseID, opsID)
todolists, err := h.expenseProvider.GetGroups(expenseID, woodlandID)
accountingExpenses, err := h.getAccountingExpense(batch)
```

**Changes**:
- Add `expenseProvider` field to payroll handler struct
- Inject provider during handler initialization
- Provider = Basecamp or NocoDB based on `TASK_PROVIDER` config
- **No changes to payroll calculation logic**

---

### 4. Update getAccountingExpense
**File**: `pkg/handler/payroll/payroll_calculator.go`

**Current code** (line 397-437):
```go
func (h *handler) getAccountingExpense(batch int) (res []bcModel.Todo, err error) {
    lists, err := h.service.Basecamp.Todo.GetLists(accountingID, accountingTodoID)
    groups, err := h.service.Basecamp.Todo.GetGroups(lists[i].ID, accountingID)
    todos, err := h.service.Basecamp.Todo.GetAllInList(groups[j].ID, accountingID)
}
```

**New code**:
```go
func (h *handler) getAccountingExpense(batch int) (res []bcModel.Todo, err error) {
    lists, err := h.expenseProvider.GetLists(accountingID, accountingTodoID)
    groups, err := h.expenseProvider.GetGroups(lists[i].ID, accountingID)
    todos, err := h.expenseProvider.GetAllInList(groups[j].ID, accountingID)
}
```

**Changes**:
- Replace `h.service.Basecamp.Todo.*` with `h.expenseProvider.*`
- **No logic changes**

---

### 5. Payroll Commit Flow (No Changes Needed)
**File**: `pkg/handler/payroll/commit.go`

**Current flow**:
1. User commits payroll for a batch
2. For each payroll record:
   - Parse `bonusExplain` JSON (contains expense info)
   - For each expense in bonusExplain:
     - Check if expense already exists in DB by `BasecampTodoID`
     - If not exists:
       - Create `expenses` table record
       - Create `accounting_transactions` record
       - Link them via `accounting_transaction_id`
3. Mark payroll as paid

**Why no changes needed**:
- Expense persistence happens here, not in webhook
- Already uses `BasecampTodoID` from bonusExplain to track expenses
- For NocoDB, `BasecampTodoID` will be NocoDB record ID
- Works for both Basecamp and NocoDB

---

## Configuration

### Environment Variables
```bash
# Existing
TASK_PROVIDER=nocodb  # or basecamp

# New - NocoDB API credentials
NOCODB_EXPENSE_BASE_ID=<base_id>
NOCODB_EXPENSE_TABLE_ID=<table_id>
NOCODB_EXPENSE_VIEW_ID=<view_id>
NOCODB_API_TOKEN=<api_token>
```

### Provider Injection
**File**: `pkg/handler/payroll/payroll.go`

```go
type handler struct {
    // ... existing fields
    expenseProvider ExpenseProvider
}

func New(/* ... */, cfg *config.Config) *handler {
    var expenseProvider ExpenseProvider

    if strings.ToLower(cfg.TaskProvider) == "nocodb" {
        expenseProvider = nocodb.NewExpenseService(cfg, logger)
    } else {
        expenseProvider = basecamp.NewExpenseService(basecampSvc)
    }

    return &handler{
        // ... existing fields
        expenseProvider: expenseProvider,
    }
}
```

---

## Data Flow Diagram

### Expense Approval → Payroll Calculation

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. User approves expense in NocoDB                              │
│    - Changes status to "approved"                               │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. NocoDB webhook fires → POST /webhooks/nocodb/expense         │
│    - Webhook handler validates                                  │
│    - NO DB insert (changed from current implementation)         │
│    - Returns 200 OK                                             │
└─────────────────────────────────────────────────────────────────┘
             │
             │ (Expense stays in NocoDB only)
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Payroll calculation triggered (monthly)                      │
│    - GET /api/v1/payrolls/detail?month=11&year=2025            │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Payroll calculator calls expense provider                    │
│    - expenseProvider.GetAllInList() → NocoDB API                │
│    - expenseProvider.GetGroups() → NocoDB API                   │
│    - expenseProvider.GetLists() → NocoDB API                    │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. NocoDB Expense Service fetches from NocoDB API               │
│    - Queries approved expenses                                  │
│    - Transforms to bcModel.Todo format                          │
│    - Returns to calculator                                      │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Payroll calculator processes expenses                        │
│    - Matches expenses to employees                              │
│    - Adds to bonus amount                                       │
│    - Stores in bonusExplain JSON                                │
│    - Calculates total payroll                                   │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Returns payroll preview to user                              │
│    - Includes expenses in bonus breakdown                       │
│    - NOT YET COMMITTED TO DB                                    │
└─────────────────────────────────────────────────────────────────┘
```

### Payroll Commit → Expense Persistence

```
┌─────────────────────────────────────────────────────────────────┐
│ 8. User commits payroll                                         │
│    - POST /api/v1/payrolls/commit                              │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 9. Payroll commit handler processes                             │
│    - For each payroll record:                                   │
│      - Parse bonusExplain JSON                                  │
│      - For each expense in bonusExplain:                        │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 10. Create expense in DB (FIRST TIME)                           │
│     - INSERT INTO expenses (employee_id, amount, reason, ...)   │
│     - INSERT INTO accounting_transactions (...)                 │
│     - UPDATE expenses SET accounting_transaction_id = ...       │
└────────────┬────────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────────┐
│ 11. Mark payroll as paid                                        │
│     - UPDATE payrolls SET is_paid = true                        │
│     - Expense now persisted in DB                               │
└─────────────────────────────────────────────────────────────────┘
```

---

## Migration Notes

### Existing Expenses in DB
- Expenses created via current webhook implementation exist in DB
- These will NOT be counted in payroll (payroll reads from NocoDB API)
- Options:
  1. Delete existing expense records (recommended)
  2. Keep for historical reference but ignore in calculations

### Testing Strategy
1. Test with `TASK_PROVIDER=basecamp` (existing flow)
2. Test with `TASK_PROVIDER=nocodb` (new flow)
3. Verify expenses NOT created on approval
4. Verify expenses created on payroll commit
5. Verify payroll calculation matches between providers

---

## File Changes Summary

### New Files
- `pkg/service/nocodb/expense.go` - NocoDB expense service implementation

### Modified Files
- `pkg/service/taskprovider/nocodb/provider.go` - Remove DB persistence from CreateExpense
- `pkg/handler/payroll/payroll_calculator.go` - Use expense provider interface
- `pkg/handler/payroll/payroll.go` - Inject expense provider
- `pkg/service/basecamp/basecamp.go` - Add ExpenseProvider interface
- `pkg/service/nocodb/nocodb.go` - Add expense service to NocoDB service struct

### No Changes Needed
- `pkg/handler/payroll/commit.go` - Already handles expense persistence correctly
- `pkg/model/expense.go` - No model changes
- Database migrations - No schema changes
