# Payroll Handler – Current Basecamp Integration

## 1. Entry Points
- `pkg/handler/payroll/payroll_calculator.go`
  - `calculatePayrolls()` gathers Basecamp data before computing payroll records.
  - `getBonus()` and helpers derive bonuses, commissions, reimbursements using Basecamp todos/comments.
- `pkg/handler/payroll/commit.go`
  - `markBonusAsDone()` closes Basecamp todos or posts notifications once payroll is paid.

## 2. Basecamp Reads
1. **Ops Expense Approvals**
   - IDs: `opsExpenseID`/`opsID` (playground vs prod) from `consts`.
   - Calls `Basecamp.Todo.GetAllInList(opsExpenseID, opsID)`.
   - Comments via `Basecamp.Comment.Gets()` to find approver (Han or Nam) mentioning “approve”. Approved todos are added to the reimbursement pool.

2. **Team Expense Reimbursements**
   - Gets expense groups with `Basecamp.Todo.GetGroups(expenseID, woodlandID)` then todos via `GetAllInList()`.
   - Uses the same comment-approval filter to keep only approved expenses.

3. **Accounting Expenses**
   - `getAccountingExpense()` fetches lists using `Basecamp.Todo.GetLists(accountingID, accountingTodoID)`.
   - For each list, collects groups, finds the “Out” group, and gets all todos.
   - Keeps todos with exactly one assignee that isn’t Han—these become reimbursements tied to accounting tasks.

4. **Reimbursement Extraction**
   - For each approved todo assigned to an employee, `getReimbursement()` parses titles formatted `Reason | Amount | Currency`.
   - Uses `Basecamp.ExtractBasecampExpenseAmount()` for `k/tr/m` suffix parsing and Wise conversion when currency isn’t VND.

## 3. Basecamp Writes
1. **Todo Completion (Expense Reimbursements)**
   - After payroll is committed, `markBonusAsDone()` iterates `ProjectBonusExplain` entries.
   - For reimbursements whose `BasecampBucketID` equals Woodland (Playground in dev), calls `Basecamp.Todo.Complete(bucketID, todoID)` to mark the todo done.

2. **Accounting Notifications**
   - When `BasecampBucketID` equals Accounting, creates a mention via `Basecamp.BasecampMention(employee.BasecampID)` and builds a comment message `Basecamp.BuildCommentMessage()` stating the amount was deposited in payroll.
   - Enqueues the comment through `h.worker.Enqueue(bcModel.BasecampCommentMsg, payload)` for async posting.

## 4. Supporting Details
- Approval logic relies on constants `HanBasecampID` / `NamNguyenBasecampID` for prod/dev.
- Environment toggles switch bucket/list IDs: `WoodlandID`, `ExpenseTodoID`, `AccountingID`, `AccountingTodoID`, `OperationID`, etc.
- `ProjectBonusExplain` JSON (stored on payrolls) includes `BasecampBucketID`/`BasecampTodoID` so later stages know which todo/comment to target.
- Worker queue uses `pkg/worker/worker.go`, which handles `BasecampCommentMsg` by calling `service.Basecamp.Comment.Create()`.

## 5. Migration Considerations
- Need NocoDB equivalents for:
  - Pulling approved Ops/Team/Accounting expense rows with approver identity and comment text.
  - Mention/comment notifications or status updates when reimbursements land in payroll.
  - Completing tasks or updating statuses for reimbursements to avoid duplicate payouts.
- Approval detection must shift from parsing “approve” comments to NocoDB status fields or structured approval records while preserving employee mapping.
- Worker payloads should be abstracted to target Basecamp or NocoDB comment endpoints interchangeably.
