# Payroll Commit - Expense & Accounting Todo Persistence

## Overview

When committing payroll, expenses and accounting todos from NocoDB need to be persisted:

### Expense Submissions (from `expense_submissions` table)
1. Create `Expense` records in database
2. Create individual `AccountingTransaction` records
3. Link via `expense.accounting_transaction_id`

### Accounting Todos (from `accounting_todos` table)
1. Create individual `AccountingTransaction` records in database
2. Mark as completed (`status=completed`) in NocoDB

## Current Situation

### What Happens Now

**Payroll Commit Flow** (`pkg/handler/payroll/commit.go`):
1. Fetch cached payroll snapshot
2. Mark bonuses as done (Basecamp only)
3. Insert payroll records to DB
4. Create accounting transactions (aggregated):
   - Payroll TW (base salary)
   - Payroll BHXH (contract)
   - Bonus (aggregated: commissions + bonuses - reimbursements)
5. Send emails

### What's Missing

**Expense submissions are NOT persisted**:
- Fetched from `expense_submissions` table during payroll calculation
- Counted in employee's bonus/reimbursement
- Aggregated into single "Bonus" transaction
- **NOT stored** in `expenses` table
- **NOT stored individually** in `accounting_transactions`

**Accounting todos are NOT persisted**:
- Fetched from `accounting_todos` table during payroll calculation
- Counted in employee's bonus/reimbursement
- Aggregated into single "Bonus" transaction
- **NOT stored individually** in `accounting_transactions`
- **NOT marked as completed** in NocoDB

## Requirements

### 1. Create Expense Records

For each expense submission used in payroll:
- Create `Expense` record in database
- Fields:
  - `EmployeeID`: Employee UUID
  - `CurrencyID`: Currency UUID
  - `Amount`: Expense amount
  - `Reason`: Expense title/description
  - `IssuedDate`: Current timestamp
  - `TaskProvider`: "nocodb"
  - `TaskRef`: NocoDB record ID
  - `TaskBoard`: "expense_submissions"
  - `AccountingTransactionID`: Link to transaction (created below)

### 2. Create Individual Accounting Transactions for Expenses

For each expense submission used in payroll:
- Create separate `AccountingTransaction` record
- Fields:
  - `Amount`: Expense amount
  - `ConversionAmount`: VND equivalent
  - `Name`: Expense description + employee name
  - `Category`: Accounting category (e.g., `AccountingExpense`)
  - `Currency`: Expense currency
  - `Date`: Current timestamp
  - `Organization`: "Dwarves Foundation"
  - `Metadata`: JSON `{"source": "expense_submission", "expense_ref": ID}`
  - `ConversionRate`: Exchange rate
  - `Type`: Transaction type

### 3. Create Individual Accounting Transactions for Accounting Todos

For each accounting todo used in payroll:
- Create separate `AccountingTransaction` record
- Fields:
  - `Amount`: Todo amount (from title parsing)
  - `ConversionAmount`: VND equivalent
  - `Name`: Todo description (e.g., "Tiền điện 12/2025 - Lê Minh Quang")
  - `Category`: Accounting category (e.g., `AccountingExpense`)
  - `Currency`: Todo currency (default VND)
  - `Date`: Current timestamp
  - `Organization`: "Dwarves Foundation"
  - `Metadata`: JSON `{"source": "accounting_todo", "todo_id": ID}`
  - `ConversionRate`: Exchange rate
  - `Type`: Transaction type

### 4. Mark Accounting Todos as Completed in NocoDB

- Update `status` field to `"completed"` in `accounting_todos` table
- For each todo ID used in payroll
- Use NocoDB PATCH API

## Implementation Plan

### Phase 1: Add NocoDB Update Method

**File**: `pkg/service/nocodb/accounting_todo.go`

Add method:
```go
// MarkTodoAsCompleted marks an accounting todo as completed in NocoDB
func (a *AccountingTodoService) MarkTodoAsCompleted(todoID int) error
```

**Implementation**:
- Validate client and config
- Build PATCH request to `/tables/{tableID}/records`
- Payload: `{"Id": todoID, "status": "completed"}`
- Handle response and errors
- Add debug logging

### Phase 2: Extract Expenses and Accounting Todos from Payroll

**File**: `pkg/handler/payroll/commit.go`

Add data structures:
```go
// ExpenseSubmissionData represents an expense submission extracted from payroll
type ExpenseSubmissionData struct {
    RecordID     string
    EmployeeID   model.UUID
    EmployeeName string
    Amount       float64
    Currency     string
    Description  string
}

// AccountingTodoData represents an accounting todo extracted from payroll
type AccountingTodoData struct {
    TodoID       int
    EmployeeID   model.UUID
    EmployeeName string
    Amount       float64
    Currency     string
    Description  string
}
```

Add helper methods:
```go
// extractExpenseSubmissionsFromPayroll extracts expense submissions from payroll bonus explains
func (h *handler) extractExpenseSubmissionsFromPayroll(payrolls []model.Payroll) []ExpenseSubmissionData

// extractAccountingTodosFromPayroll extracts accounting todos from payroll bonus explains
func (h *handler) extractAccountingTodosFromPayroll(payrolls []model.Payroll) []AccountingTodoData
```

**Implementation**:
- Loop through payrolls
- Unmarshal `ProjectBonusExplain` JSON
- Identify by source in metadata or title pattern
- Extract: ID, EmployeeID, Amount, Currency, Description
- Return structured lists

### Phase 3: Store Expense Submissions

**File**: `pkg/handler/payroll/commit.go`

Add method:
```go
// storeExpenseSubmissions creates Expense records and AccountingTransactions for expense submissions
func (h *handler) storeExpenseSubmissions(expenses []ExpenseSubmissionData, batchDate time.Time) error
```

**Implementation**:
- For each expense submission:
  - Create `AccountingTransaction` first
  - Create `Expense` record with `AccountingTransactionID` link
  - Insert to DB
- Add debug logging
- Return error if any insertion fails

### Phase 4: Store Accounting Todo Transactions

**File**: `pkg/handler/payroll/commit.go`

Add method:
```go
// storeAccountingTodoTransactions creates individual transactions for accounting todos
func (h *handler) storeAccountingTodoTransactions(todos []AccountingTodoData, batchDate time.Time) error
```

**Implementation**:
- For each accounting todo:
  - Create `AccountingTransaction` with:
    - Name: `"{description} - {employeeName}"`
    - Category: `AccountingExpense` or appropriate category
    - Amount/Currency from todo
    - Metadata: `{"source": "accounting_todo", "todo_id": ID}`
  - Insert to DB
- Add debug logging

### Phase 5: Mark Todos as Completed

**File**: `pkg/handler/payroll/commit.go`

Add method:
```go
// markAccountingTodosAsCompleted marks todos as completed in NocoDB
func (h *handler) markAccountingTodosAsCompleted(todos []AccountingTodoData) error
```

**Implementation**:
- Check if PayrollAccountingTodoProvider is NocoDB service
- Cast to `*nocodb.AccountingTodoService`
- For each todo:
  - Call `MarkTodoAsCompleted(todoID)`
  - Log success/failure
- Continue on individual failures (log but don't abort)

### Phase 6: Update Commit Handler

**File**: `pkg/handler/payroll/commit.go`

Update `commitPayrollHandler` method (line 112-220):

```go
// After line 194 (h.store.Payroll.InsertList)

// Extract expense submissions
expenseSubmissions := h.extractExpenseSubmissionsFromPayroll(payrolls)
h.logger.Debug(fmt.Sprintf("Extracted %d expense submissions from payroll", len(expenseSubmissions)))

// Extract accounting todos
accountingTodos := h.extractAccountingTodosFromPayroll(payrolls)
h.logger.Debug(fmt.Sprintf("Extracted %d accounting todos from payroll", len(accountingTodos)))

// Store expense submissions (Expense records + AccountingTransactions)
if len(expenseSubmissions) > 0 {
    err = h.storeExpenseSubmissions(expenseSubmissions, batchDate)
    if err != nil {
        return fmt.Errorf("failed to store expense submissions: %w", err)
    }
    h.logger.Debug(fmt.Sprintf("Stored %d expense submissions", len(expenseSubmissions)))
}

// Store accounting todo transactions
if len(accountingTodos) > 0 {
    err = h.storeAccountingTodoTransactions(accountingTodos, batchDate)
    if err != nil {
        return fmt.Errorf("failed to store accounting todo transactions: %w", err)
    }
    h.logger.Debug(fmt.Sprintf("Stored %d accounting todo transactions", len(accountingTodos)))

    // Mark todos as completed in NocoDB
    err = h.markAccountingTodosAsCompleted(accountingTodos)
    if err != nil {
        h.logger.Error(err, "failed to mark accounting todos as completed in NocoDB")
        // Don't return error - log only, continue with email sending
    }
}

// Continue with existing flow (line 197+)
```

## Files to Modify

1. `pkg/service/nocodb/accounting_todo.go`
   - Add `MarkTodoAsCompleted` method

2. `pkg/handler/payroll/commit.go`
   - Add `ExpenseSubmissionData` struct
   - Add `AccountingTodoData` struct
   - Add `extractExpenseSubmissionsFromPayroll` method
   - Add `extractAccountingTodosFromPayroll` method
   - Add `storeExpenseSubmissions` method
   - Add `storeAccountingTodoTransactions` method
   - Add `markAccountingTodosAsCompleted` method
   - Update `commitPayrollHandler` to call new methods

## Data Flow

```
Payroll Commit Triggered
    ↓
Fetch Cached Payroll (contains expenses & todos in ProjectBonusExplain)
    ↓
Extract Expense Submissions from Payroll.ProjectBonusExplain
    ↓
Extract Accounting Todos from Payroll.ProjectBonusExplain
    ↓
Create Expense Records + AccountingTransactions (for expense submissions)
    ↓
Create Individual AccountingTransactions (for accounting todos)
    ↓
Mark Todos as Completed in NocoDB (status=completed)
    ↓
Continue Normal Flow (emails, etc.)
```

## Key Considerations

### 1. Identifying Sources in ProjectBonusExplain

**Expense Submissions**:
- Check metadata or title pattern
- May have `BasecampTodoID` as NocoDB record ID
- Title format: "Description | Amount | Currency"

**Accounting Todos**:
- Check metadata or title pattern
- `BasecampTodoID`: Stores NocoDB record ID
- Title format: "Description | Amount | Currency"

### 2. Error Handling

- **Expense/Transaction creation failure**: Return error, abort commit
- **NocoDB update failure**: Log error, continue (don't abort commit)
- Rationale: Payroll already calculated, emails should be sent even if NocoDB update fails

### 3. Basecamp Compatibility

- Check if providers are NocoDB services
- Only call NocoDB-specific methods for NocoDB
- Basecamp flow uses existing `markBonusAsDone` method

### 4. Transaction Categories

Both expense submissions and accounting todos:
- Office expenses: `AccountingExpense`
- Could be enhanced to parse from metadata

### 5. Idempotency

- Check if records already exist before creating
- For expenses: Query by `task_provider + task_ref`
- For transactions: Query by metadata
- Skip if exists

## Testing Requirements

### Unit Tests

1. `TestExtractExpenseSubmissionsFromPayroll`
2. `TestExtractAccountingTodosFromPayroll`
3. `TestStoreExpenseSubmissions`
4. `TestStoreAccountingTodoTransactions`
5. `TestMarkAccountingTodosAsCompleted`

### Integration Tests

1. Full payroll commit with expenses and todos
2. Verify Expense records created
3. Verify AccountingTransactions created
4. Verify NocoDB status updated

## Success Criteria

- ✅ Each expense submission creates `Expense` record
- ✅ Each expense submission creates individual `AccountingTransaction`
- ✅ `Expense.AccountingTransactionID` correctly linked
- ✅ Each accounting todo creates individual `AccountingTransaction`
- ✅ NocoDB `accounting_todos` status updated to "completed"
- ✅ Existing payroll commit flow works unchanged
- ✅ Basecamp compatibility maintained
- ✅ Comprehensive debug logging
- ✅ Error handling doesn't break payroll commit
