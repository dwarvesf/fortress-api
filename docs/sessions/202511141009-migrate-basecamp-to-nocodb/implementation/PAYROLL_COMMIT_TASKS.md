# Payroll Commit - Implementation Task Breakdown

**Session**: 202511141009-migrate-basecamp-to-nocodb
**Feature**: Expense & Accounting Todo Persistence during Payroll Commit
**Plan Reference**: `plan/payroll/PAYROLL_COMMIT_EXPENSE_PERSISTENCE.md`

## Overview

Implement persistence of expense submissions and accounting todos during payroll commit. This ensures individual expense records and accounting transactions are created, and NocoDB todos are marked as completed.

## Task List

### Phase 1: Add NocoDB Update Method
**File**: `pkg/service/nocodb/accounting_todo.go`

- [ ] **Task 1.1**: Add `MarkTodoAsCompleted` method signature
  - Method: `func (a *AccountingTodoService) MarkTodoAsCompleted(todoID int) error`
  - Location: After `GetLists` method (line 339+)

- [ ] **Task 1.2**: Implement NocoDB PATCH request
  - Validate client and config
  - Build PATCH request to `/tables/{tableID}/records`
  - Payload: `{"Id": todoID, "status": "completed"}`
  - Handle response (2xx success, non-2xx error)

- [ ] **Task 1.3**: Add comprehensive debug logging
  - Log: "Marking accounting todo {todoID} as completed in NocoDB"
  - Log: "Successfully marked accounting todo {todoID} as completed"
  - Log errors with response body

- [ ] **Task 1.4**: Add error handling
  - Return error if client is nil
  - Return error if tableID is empty
  - Return error if PATCH request fails
  - Return error if response status >= 300

### Phase 2: Extract Data from Payroll
**File**: `pkg/handler/payroll/commit.go`

- [ ] **Task 2.1**: Add `ExpenseSubmissionData` struct
  - Fields: `RecordID string`, `EmployeeID model.UUID`, `EmployeeName string`, `Amount float64`, `Currency string`, `Description string`
  - Location: After imports, before handler methods

- [ ] **Task 2.2**: Add `AccountingTodoData` struct
  - Fields: `TodoID int`, `EmployeeID model.UUID`, `EmployeeName string`, `Amount float64`, `Currency string`, `Description string`
  - Location: After `ExpenseSubmissionData` struct

- [ ] **Task 2.3**: Implement `extractExpenseSubmissionsFromPayroll` method
  - Loop through `payrolls []model.Payroll`
  - Unmarshal `ProjectBonusExplain` JSON
  - Identify expense submissions by source or pattern
  - Extract: RecordID, EmployeeID, EmployeeName, Amount, Currency, Description
  - Add debug logging for extraction count
  - Return `[]ExpenseSubmissionData`

- [ ] **Task 2.4**: Implement `extractAccountingTodosFromPayroll` method
  - Loop through `payrolls []model.Payroll`
  - Unmarshal `ProjectBonusExplain` JSON
  - Identify accounting todos by source or pattern (BasecampTodoID, title pattern)
  - Extract: TodoID, EmployeeID, EmployeeName, Amount, Currency, Description
  - Add debug logging for extraction count
  - Return `[]AccountingTodoData`

### Phase 3: Store Expense Submissions
**File**: `pkg/handler/payroll/commit.go`

- [ ] **Task 3.1**: Implement `storeExpenseSubmissions` method signature
  - Method: `func (h *handler) storeExpenseSubmissions(expenses []ExpenseSubmissionData, batchDate time.Time) error`

- [ ] **Task 3.2**: Add idempotency check
  - Query existing expenses by `task_provider="nocodb"` AND `task_ref=RecordID`
  - Skip if already exists
  - Log: "Expense submission {RecordID} already exists, skipping"

- [ ] **Task 3.3**: Create AccountingTransaction for each expense
  - Fetch currency by code
  - Calculate conversion amount and rate
  - Build transaction with:
    - `Amount`, `ConversionAmount`, `Name` (description + employee name)
    - `Category` = `AccountingExpense`
    - `Currency`, `Date` = batchDate
    - `Organization` = "Dwarves Foundation"
    - `Metadata` = JSON `{"source": "expense_submission", "expense_ref": RecordID}`
    - `ConversionRate`, `Type`
  - Insert to DB via `h.store.AccountingTransaction.Create()`

- [ ] **Task 3.4**: Create Expense record linked to transaction
  - Build Expense with:
    - `EmployeeID`, `CurrencyID`, `Amount`, `Reason` (description)
    - `IssuedDate` = batchDate
    - `TaskProvider` = "nocodb", `TaskRef` = RecordID, `TaskBoard` = "expense_submissions"
    - `AccountingTransactionID` = transaction ID from step 3.3
  - Insert to DB via `h.store.Expense.Create()`

- [ ] **Task 3.5**: Add comprehensive debug logging
  - Log: "Storing {count} expense submissions"
  - Log per expense: "Created AccountingTransaction {ID} for expense submission {RecordID}"
  - Log per expense: "Created Expense record {ID} linked to transaction {TransactionID}"
  - Log: "Successfully stored {count} expense submissions"

- [ ] **Task 3.6**: Add error handling
  - Return error if currency lookup fails
  - Return error if transaction creation fails
  - Return error if expense creation fails

### Phase 4: Store Accounting Todo Transactions
**File**: `pkg/handler/payroll/commit.go`

- [ ] **Task 4.1**: Implement `storeAccountingTodoTransactions` method signature
  - Method: `func (h *handler) storeAccountingTodoTransactions(todos []AccountingTodoData, batchDate time.Time) error`

- [ ] **Task 4.2**: Add idempotency check
  - Query existing transactions by metadata `{"source": "accounting_todo", "todo_id": TodoID}`
  - Skip if already exists
  - Log: "Accounting todo transaction {TodoID} already exists, skipping"

- [ ] **Task 4.3**: Create AccountingTransaction for each todo
  - Fetch currency by code
  - Calculate conversion amount and rate
  - Build transaction with:
    - `Amount`, `ConversionAmount`
    - `Name` = "{description} - {employeeName}"
    - `Category` = `AccountingExpense`
    - `Currency`, `Date` = batchDate
    - `Organization` = "Dwarves Foundation"
    - `Metadata` = JSON `{"source": "accounting_todo", "todo_id": TodoID}`
    - `ConversionRate`, `Type`
  - Insert to DB via `h.store.AccountingTransaction.Create()`

- [ ] **Task 4.4**: Add comprehensive debug logging
  - Log: "Storing {count} accounting todo transactions"
  - Log per todo: "Created AccountingTransaction {ID} for accounting todo {TodoID}"
  - Log: "Successfully stored {count} accounting todo transactions"

- [ ] **Task 4.5**: Add error handling
  - Return error if currency lookup fails
  - Return error if transaction creation fails

### Phase 5: Mark Todos as Completed
**File**: `pkg/handler/payroll/commit.go`

- [ ] **Task 5.1**: Implement `markAccountingTodosAsCompleted` method signature
  - Method: `func (h *handler) markAccountingTodosAsCompleted(todos []AccountingTodoData) error`

- [ ] **Task 5.2**: Check if provider is NocoDB
  - Check `h.service.PayrollAccountingTodoProvider != nil`
  - Type assert to `*nocodb.AccountingTodoService`
  - If not NocoDB, log and return nil (Basecamp uses different flow)

- [ ] **Task 5.3**: Call `MarkTodoAsCompleted` for each todo
  - Loop through todos
  - Call `service.MarkTodoAsCompleted(todo.TodoID)`
  - Log success: "Marked accounting todo {TodoID} as completed in NocoDB"
  - Log failure but continue: "Failed to mark todo {TodoID} as completed: {error}"

- [ ] **Task 5.4**: Add comprehensive debug logging
  - Log: "Marking {count} accounting todos as completed in NocoDB"
  - Log per todo success/failure
  - Log: "Completed marking accounting todos (success: {success}, failed: {failed})"

- [ ] **Task 5.5**: Add error handling
  - Continue on individual todo failures (don't abort)
  - Collect errors and return aggregated error if any failed
  - Allow non-fatal errors (log-only, don't abort commit)

### Phase 6: Update Commit Handler
**File**: `pkg/handler/payroll/commit.go`

- [ ] **Task 6.1**: Extract expense submissions after payroll insertion
  - Location: After `h.store.Payroll.InsertList(db, payrolls)` (line 194)
  - Call: `expenseSubmissions := h.extractExpenseSubmissionsFromPayroll(payrolls)`
  - Log: "Extracted {count} expense submissions from payroll"

- [ ] **Task 6.2**: Extract accounting todos after payroll insertion
  - Location: After task 6.1
  - Call: `accountingTodos := h.extractAccountingTodosFromPayroll(payrolls)`
  - Log: "Extracted {count} accounting todos from payroll"

- [ ] **Task 6.3**: Store expense submissions
  - Check: `if len(expenseSubmissions) > 0`
  - Call: `err = h.storeExpenseSubmissions(expenseSubmissions, batchDate)`
  - Handle error: `return fmt.Errorf("failed to store expense submissions: %w", err)`
  - Log: "Stored {count} expense submissions"

- [ ] **Task 6.4**: Store accounting todo transactions
  - Check: `if len(accountingTodos) > 0`
  - Call: `err = h.storeAccountingTodoTransactions(accountingTodos, batchDate)`
  - Handle error: `return fmt.Errorf("failed to store accounting todo transactions: %w", err)`
  - Log: "Stored {count} accounting todo transactions"

- [ ] **Task 6.5**: Mark todos as completed in NocoDB
  - Check: `if len(accountingTodos) > 0`
  - Call: `err = h.markAccountingTodosAsCompleted(accountingTodos)`
  - Handle error: Log only, don't abort (non-fatal)
  - Log error: "failed to mark accounting todos as completed in NocoDB"
  - Continue with email sending even if this fails

- [ ] **Task 6.6**: Verify existing flow unchanged
  - Ensure no changes to existing payroll commit logic
  - Ensure Basecamp compatibility maintained
  - Ensure aggregated transactions still created

## Testing Tasks

### Unit Tests
**File**: `pkg/handler/payroll/commit_test.go`

- [ ] **Task T.1**: Write `TestExtractExpenseSubmissionsFromPayroll`
  - Test empty payrolls
  - Test payrolls with expense submissions
  - Test payrolls with mixed bonuses (expenses + commissions)
  - Verify correct extraction of all fields

- [ ] **Task T.2**: Write `TestExtractAccountingTodosFromPayroll`
  - Test empty payrolls
  - Test payrolls with accounting todos
  - Test payrolls with mixed bonuses (todos + commissions)
  - Verify correct extraction of all fields

- [ ] **Task T.3**: Write `TestStoreExpenseSubmissions`
  - Test valid expense submissions
  - Test empty list
  - Test idempotency (duplicate submissions)
  - Test database errors
  - Verify Expense and AccountingTransaction created
  - Verify correct linking via AccountingTransactionID

- [ ] **Task T.4**: Write `TestStoreAccountingTodoTransactions`
  - Test valid accounting todos
  - Test empty list
  - Test idempotency (duplicate todos)
  - Test database errors
  - Verify AccountingTransaction created with correct metadata

- [ ] **Task T.5**: Write `TestMarkAccountingTodosAsCompleted`
  - Test NocoDB provider
  - Test Basecamp provider (should skip)
  - Test individual todo update failures
  - Verify non-fatal error handling

### Integration Tests
**File**: `pkg/handler/payroll/commit_integration_test.go`

- [ ] **Task T.6**: Write full payroll commit integration test
  - Setup: Create test payroll with expenses and todos
  - Execute: Commit payroll
  - Verify: Expense records created in DB
  - Verify: Individual AccountingTransactions created for expenses
  - Verify: Individual AccountingTransactions created for todos
  - Verify: Expense.AccountingTransactionID correctly linked
  - Verify: NocoDB status updated (mock NocoDB API)
  - Verify: Existing aggregated transactions still created

## Dependencies

### Required Store Methods
- `h.store.Expense.Create(db, expense)` - Create expense record
- `h.store.AccountingTransaction.Create(db, transaction)` - Create transaction
- `h.store.Currency.OneByCode(db, currencyCode)` - Fetch currency by code
- `h.store.Expense.OneByTaskRef(db, provider, ref)` - Check existing expense
- `h.store.AccountingTransaction.OneByMetadata(db, metadata)` - Check existing transaction

### Required Service Methods
- `h.service.PayrollAccountingTodoProvider.(*nocodb.AccountingTodoService)` - Type assertion
- `service.MarkTodoAsCompleted(todoID)` - Mark todo completed (new method from Phase 1)

## Success Criteria

- ✅ Each expense submission creates `Expense` record
- ✅ Each expense submission creates individual `AccountingTransaction`
- ✅ `Expense.AccountingTransactionID` correctly linked
- ✅ Each accounting todo creates individual `AccountingTransaction`
- ✅ NocoDB `accounting_todos` status updated to "completed"
- ✅ Existing payroll commit flow works unchanged
- ✅ Basecamp compatibility maintained
- ✅ Comprehensive debug logging throughout
- ✅ Error handling doesn't break payroll commit
- ✅ Idempotency checks prevent duplicates
- ✅ All unit tests pass
- ✅ Integration test validates full flow

## Task Progress Tracking

**Total Tasks**: 35 implementation tasks + 6 testing tasks = 41 tasks
**Completed**: 0
**In Progress**: 0
**Pending**: 41

Update this section as tasks are completed.
