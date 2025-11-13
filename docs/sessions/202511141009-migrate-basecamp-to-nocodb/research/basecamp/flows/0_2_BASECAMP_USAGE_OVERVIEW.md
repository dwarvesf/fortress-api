# Basecamp Usage Overview & Payroll Deep Dive

## 1. Repository-Wide Usage Map

### 1.1 Service Wiring
- `pkg/service/service.go` builds `service.Basecamp` and injects it into handlers, controllers, and workers.
- `pkg/service/basecamp/` contains the full integration: OAuth client (`client/client.go`), models, consts, and sub-services (`todo`, `comment`, `people`, `schedule`, `attachment`, `project`, `campfire`, `messageboard`, `recording`, `subscription`, `webhook`, `integration`).

### 1.2 Runtime Consumers
- **Invoice flow**: `pkg/controller/invoice/send.go`, `pkg/controller/invoice/update_status.go`, `pkg/controller/invoice/commission.go`, `pkg/handler/webhook/basecamp_invoice.go`, `pkg/worker/worker.go` (comment/todo jobs).
- **Accounting & cronjobs**: `pkg/handler/accounting/accounting.go`, `pkg/handler/webhook/basecamp_accounting.go`.
- **Expense & payroll**: `pkg/handler/webhook/basecamp_expense.go`, `pkg/handler/payroll/payroll_calculator.go`, `pkg/handler/payroll/commit.go`.
- **On-leave & general ops**: `pkg/handler/webhook/onleave.go`, `pkg/handler/webhook/basecamp.go`, `pkg/handler/profile/profile.go`, `pkg/controller/employee/update_employee_status.go`, `pkg/handler/discord/discord.go`.
- **Routing/tests**: `pkg/routes/v1.go` and `pkg/routes/v1_test.go` define all `/webhooks/basecamp/**` endpoints.

### 1.3 Persistence & Metadata
- Stores referencing `basecamp_id` columns: `pkg/store/expense/expense.go`, `pkg/store/employee/employee.go`, `pkg/store/recruitment/recruitment.go`.
- Models carrying Basecamp IDs: `pkg/model/basecamp.go`, `pkg/model/expense.go`, `pkg/model/project_info.go`, `pkg/model/project.go` (view), `pkg/model/employees.go`, `pkg/model/candidate.go`, `pkg/model/webhook.go`.
- Migrations/seeds defining those columns: `migrations/schemas/20221102153827-init_employees_schema.sql`, `20230417095551-create-expenses-table.sql`, payroll/payroll_commission migrations, plus `migrations/seed/*.sql` (employee Basecamp IDs, metadata).

### 1.4 Documentation & Specs
- Swagger definitions (`docs/swagger.yaml`, `docs/swagger.json`, `docs/docs.go`) list Basecamp webhook routes.
- Analysis docs under `docs/sessions/202511141009-migrate-basecamp-to-nocodb/overview/` capture the detailed workflow understanding.

## 2. Payroll Handler Deep Dive (Current State)

### 2.1 Entry Points
- `pkg/handler/payroll/payroll_calculator.go` → `calculatePayrolls()` pulls Basecamp data to compute project bonuses, reimbursements, and approvals.
- `pkg/handler/payroll/commit.go` → `markBonusAsDone()` completes todos or posts comments once payroll is paid.

### 2.2 Data Pulled from Basecamp
1. **Ops Expense Approvals**
   - Reads Ops expense list IDs: `opsExpenseID`/`opsID` (Playground vs prod) from `consts`.
   - Calls `Basecamp.Todo.GetAllInList()` for every Ops expense todo.
   - Fetches comments via `Basecamp.Comment.Gets()` and filters todos where the approver (Han or Nam) commented "approve".

2. **Team Expense Reimbursements**
   - Lists expense groups using `Basecamp.Todo.GetGroups()` then `GetAllInList()` per group.
   - Reuses comment-check logic to keep only approved expenses.

3. **Accounting Expenses**
   - `getAccountingExpense()` fetches lists via `Basecamp.Todo.GetLists(accountingID, accountingTodoID)`.
   - Iterates groups, grabs todos from the "Out" group, and keeps ones assigned to a single person (not Han) to treat as reimbursements.

4. **Reimbursement Extraction**
   - For each approved todo assigned to the employee, `getReimbursement()` parses `Title` in the format `Reason | Amount | Currency` using `Basecamp.ExtractBasecampExpenseAmount()` and optional Wise conversion.

### 2.3 Actions Performed Back to Basecamp
1. **Todo Completion**
   - In `markBonusAsDone()`, when a payroll record references `BasecampBucketID`/`BasecampTodoID`, the handler calls `Basecamp.Todo.Complete()` for reimbursement todos (Woodland bucket) after depositing the amount.

2. **Comment Notifications**
   - For accounting-related todos (Accounting bucket), builds a Basecamp mention for the employee via `Basecamp.BasecampMention()` and enqueues a worker job (`Basecamp.BuildCommentMessage`) to notify that the reimbursement is in payroll.

3. **Approver Lookup**
   - Throughout calculations the handler compares comment creator IDs to `approver` constant (Han or Nam) to decide if an expense is approved; these constants live in `pkg/service/basecamp/consts/consts.go`.

### 2.4 Dependencies & Touchpoints Summary
- **Services Used**: `Basecamp.Todo`, `Basecamp.Comment`, `Basecamp.BasecampMention`, `Basecamp.BuildCommentMessage`, `Basecamp.ExtractBasecampExpenseAmount`.
- **Config/Consts**: Project/todo list IDs (`WoodlandID`, `ExpenseTodoID`, `AccountingID`, etc.), approver IDs (`HanBasecampID`, `NamNguyenBasecampID`).
- **Worker Queue**: `h.worker.Enqueue()` enqueues `BasecampCommentMsg` payloads for asynchronous notifications.
- **Stores/Models**: Payroll records embed `BasecampBucketID`/`BasecampTodoID` in `ProjectBonusExplain` JSON so later commits know which todo to close/comment on.

### 2.5 Migration Implications
- Need NocoDB equivalents for: expense approval source data (Ops + Woodland expense boards), accounting reimbursement feed, mention/comment notifications, and todo completion semantics.
- Must preserve approval detection (who approved, comment text) with NocoDB status fields or activity logs.
- Worker payloads should become provider-neutral so payroll commits can notify regardless of backend.
