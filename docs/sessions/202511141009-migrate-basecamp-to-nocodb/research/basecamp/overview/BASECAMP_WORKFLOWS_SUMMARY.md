# Basecamp Workflows - Quick Reference Summary

## 1. INVOICE FLOW

**Endpoint:** `PUT /webhooks/basecamp/operation/invoice`

**Trigger:** Todo marked complete in Basecamp with title format `MM/YYYY - #INVOICE_NUMBER`

**Key Files:**
- Handler: `pkg/handler/webhook/basecamp.go`, `basecamp_invoice.go`
- Controller: `pkg/controller/invoice/update_status.go`
- Store: `pkg/store/invoice/invoice.go`, `pkg/store/accounting/accounting.go`

**Database Models:**
- `invoices` (main) → links to `projects`, `bank_accounts`, `employees`
- `accounting_transactions` (created when paid)

**Basecamp Integration:**
- Parses todo title via regex for invoice number
- Validates "Paid" comment from approver
- Posts success/failure comment to todo
- Completes todo in Basecamp

**Key Validations:**
1. Title format: `MM/YYYY - #YYYY-XXX-###`
2. Invoice status must be `sent` or `overdue`
3. Must have "Paid" comment from HanBasecampID (prod) or NamNguyenBasecampID (dev)
4. Ignores AutoBotID (25727627)

**Hardcoded Constants:**
- Woodland (Prod): 9403032
- Playground (Dev): 12984857
- Woodland TodoList: 1346305133
- Playground TodoList: 1941398075

**Error Handling:** Returns 200 OK always; logs and posts comments on failure

---

## 2. ACCOUNTING FLOW

### 2.1 Monthly Todo Creation (Cronjob)

**Endpoint:** `POST /api/v1/cronjobs/sync-monthly-accounting-todo`

**Trigger:** Monthly cronjob (automated)

**Key Files:**
- Handler: `pkg/handler/accounting/accounting.go:42-118`
- Store: `pkg/store/operational_service/` (service templates)

**What It Creates in Basecamp:**

1. **Monthly Todo List**: `"Accounting | {Month} {Year}"` (e.g., "Accounting | January 2025")

2. **Two Groups:**
   - **"In" Group** - Revenue/Income todos
     - Source: Active Time & Material projects from database
     - Format: `"{ProjectName} {Month}/{Year}"` (e.g., "Voconic 1/2025")
     - Assigned to: Giang Than (26160802)
     - Due: Last day of month (or 23rd for Voconic)

   - **"Out" Group** - Expenses/Payments todos
     - Source: `operational_service` database table + hardcoded salary
     - Format: `"{ServiceName} | {Amount} | {Currency}"` or "salary 15th/1st"
     - Assigned to: Quang (22659105) + Han for salary (21562923)
     - Due: Last day of month (or specific dates for salary: 12th, 27th)

**Example Structure:**
```
Accounting | January 2025
├─ In Group
│  ├─ Voconic 1/2025 (due 23rd)
│  ├─ Dwarves Foundation 1/2025
│  └─ ConsultingWare 1/2025
└─ Out Group
   ├─ Office Rental | 1.500.000 | VND
   ├─ CBRE | 800.000 | VND
   ├─ Tiền điện | 300.000 | VND
   ├─ salary 15th (due 12th)
   └─ salary 1st (due 27th)
```

### 2.2 Transaction Creation (Webhook)

**Endpoint:** `POST /webhooks/basecamp/operation/accounting-transaction`

**Trigger:** Todo created in Accounting project with title format `Description | Amount | Currency`

**Key Files:**
- Handler: `pkg/handler/webhook/basecamp_accounting.go`
- Store: `pkg/store/accounting/accounting.go`

**Database Models:**
- `accounting_transactions` (created directly from todo)
- Links via metadata: `{source: "basecamp_accounting", id: todoID}`

**Basecamp Integration:**
- Fetches parent todo list to extract month/year
- Reads-only (no updates sent back)
- Parses amount with thousands separator (dots)

**Key Validations:**
1. Parent category must match: `Accounting | MONTH_NAME YEAR`
2. Todo title regex: `[S|s]alary\s*(1st|15th)|(.*)\|\s*([0-9\.]+)\s*\|\s*([a-zA-Z]{3})`
3. Amount parsed as integer (dots removed)
4. Currency must exist in database
5. Project validation: Accounting (15258324) in prod

**Hardcoded Constants:**
- Accounting Project (Prod): 15258324
- Accounting Todo Set (Prod): 2329633561
- Playground (Dev): 12984857
- Playground Todo Set (Dev): 1941398075

**Categorization:**
- Office rental/CBRE → `Office Space`
- Default → `Office Services`

**Error Handling:** Returns 400/500 on error; no comment posting

---

## 3. EXPENSE FLOW

**Endpoints:**
- Validation: `POST /webhooks/basecamp/expense/validate`
- Creation: `POST /webhooks/basecamp/expense`
- Deletion: `DELETE /webhooks/basecamp/expense`

**Triggers:**
- `todo_created` → validation + assignment
- `todo_completed` → expense creation
- `todo_uncompleted` → expense deletion

**Key Files:**
- Handler: `pkg/handler/webhook/basecamp_expense.go`
- Service: `pkg/service/basecamp/integration.go`
- Store: `pkg/store/expense/expense.go`, `pkg/store/accounting/accounting.go`

**Database Models:**
- `expenses` (main) → links to `employees`, `currencies`
- `accounting_transactions` (created when approved)

**Basecamp Integration:**
- Parses title: `Reason | Amount | Currency`
- Auto-assigns to HanBasecampID (prod) or NamNguyenBasecampID (dev)
- Posts validation comment (pass/fail)
- Posts creation/deletion comment (success/fail)

**Amount Parsing (k/tr/m support):**
- `500` → 500
- `500k` → 500,000
- `500k500` → 500,500
- `5m` → 5,000,000
- `5tr` → 5,000,000

**Key Validations:**
1. Title format: `reason | amount | currency` (3+ parts)
2. Amount must parse to non-zero
3. Currency: VND or USD only
4. Bucket: "Woodland" (prod) or "Fortress | Playground" (dev)
5. Employee and currency must exist in database

**Hardcoded Constants:**
- Woodland (Prod): 9403032
- Playground (Dev): 12984857
- Woodland ExpenseTodo: 2353511928
- Playground ExpenseTodo: 2436015405
- Prod Approver: HanBasecampID (21562923)
- Dev Approver: NamNguyenBasecampID (21581534)

**Error Handling:** Returns 200 OK always; posts comments for validation feedback

---

## 4. ON-LEAVE REQUEST FLOW

**Endpoints:**
- Validation: `POST /webhooks/basecamp/onleave/validate`
- Approval: `POST /webhooks/basecamp/onleave`

**Triggers:**
- `todo_created` → validation
- `todo_completed` → approval + schedule creation

**Key Files:**
- Handler: `pkg/handler/webhook/onleave.go`
- Store: `pkg/store/onleaverequest/onleave_request.go`

**Database Models:**
- `on_leave_requests` (created on approval)
- Links: creator_id, approver_id, assignee_ids (JSONB array)

**Basecamp Integration:**
- Parses title: `Name | Type | DateRange [| Shift]`
  - Example: `John Doe | Off | 15/01/2025 - 20/01/2025 | Morning`
- Creates monthly schedule entries in Basecamp calendar
- Sets subscribers to assignees + ops team
- Posts validation comment (pass/fail)

**Date Format:** `DD/MM/YYYY` (single or range with `-`)

**Key Validations:**
1. Title format: 3+ pipe-separated parts
2. Type: "off" or "remote" (case-insensitive)
3. Date range: start >= today, end >= start
4. Parent group: OnleaveID (6935836756 prod, 2243342506 dev)
5. Creator and approver must exist as employees
6. Environment check: Prod only processes "Woodland" bucket

**Hardcoded Constants:**
- Woodland Schedule (Prod): 1346305137
- Playground Schedule (Dev): 1941398077
- OnLeave (Prod): 6935836756
- OnLeave Playground (Dev): 2243342506
- Prod Ops: HuyNguyenBasecampID, GiangThanBasecampID
- Dev Ops: NamNguyenBasecampID

**Calendar Creation:** Chunks multi-month requests into monthly entries

**Error Handling:** Returns 200 OK; logs errors; posts validation comments

---

## 5. PAYROLL FLOW

### 5.1 Payroll Calculation

**Entrypoint:** `pkg/handler/payroll/payroll_calculator.go` (`calculatePayrolls`)

**Trigger:** Payroll batch execution (1st & 15th) pulls Basecamp data before generating payroll rows.

**Basecamp Reads:**
- **Ops Expenses:** `Basecamp.Todo.GetAllInList(opsExpenseID, opsID)` then `Basecamp.Comment.Gets()` to find approver comments containing “approve”.
- **Team Expenses:** `Basecamp.Todo.GetGroups(expenseTodoList, woodlandID)` → `GetAllInList()` per group → same approval check.
- **Accounting "Out" Todos:** `Basecamp.Todo.GetLists(accountingID, accountingTodoID)` → `GetGroups()` → `GetAllInList()` for the “Out” group; keeps todos assigned to a single person ≠ Han.
- Approved todos assigned to each employee add reimbursement entries via `getReimbursement()` (parses `Reason | Amount | Currency`, uses `Basecamp.ExtractBasecampExpenseAmount`).

**Outputs:** Reimbursements stored inside `ProjectBonusExplain` with `BasecampBucketID`/`BasecampTodoID`, plus bonus/commission explains for payroll records.

### 5.2 Payroll Commit

**Entrypoint:** `pkg/handler/payroll/commit.go` (`markBonusAsDone`).

**Basecamp Writes:**
- For reimbursement explains tied to Woodland bucket: `Basecamp.Todo.Complete(bucketID, todoID)` to mark todo done after payroll payout.
- For accounting explains: builds mention via `Basecamp.BasecampMention(employee.BasecampID)` and enqueues a worker job with `Basecamp.BuildCommentMessage()` to notify that reimbursement was deposited.

**Key Dependencies & Validations:**
1. Approver IDs: `HanBasecampID` (prod) vs `NamNguyenBasecampID` (dev) used during approval detection.
2. Bucket/list IDs for Ops, Expense, and Accounting vary per environment.
3. Reimbursement parsing relies on todo title format and presence of assignee matching employee Basecamp ID.

**Error Handling:** Logging on API failures; commit stops on Basecamp errors to avoid marking payroll done without closing todos.

---

## CROSS-FLOW PATTERNS

### Shared Services
- **Basecamp Service:** `pkg/service/basecamp/`
- **Currency Conversion:** Wise API (uses `service.Wise.Convert()`)
- **Employee Lookup:** By Basecamp ID (`store.Employee.OneByBasecampID()`)
- **Worker Queue:** Async comment posting via `worker.Enqueue()`

### Shared Constants (pkg/service/basecamp/consts/consts.go)
- Project IDs (Woodland, Playground, Accounting, OnLeave)
- Todo List IDs
- Person IDs (HanBasecampID, AutoBotID, etc.)
- Bucket names

### Metadata Pattern
All flows store metadata in related accounting transactions:
```json
{
  "source": "flow_type",  // "invoice", "expense", "basecamp_accounting"
  "id": "record_id"       // UUID or Basecamp ID
}
```

### Webhook Message Structure
```
BasecampWebhookMessage
├─ Kind: "todo_created", "todo_completed", "todo_uncompleted"
├─ Recording: {ID, Title, URL, Creator, Bucket, Parent, UpdatedAt}
└─ Creator: {ID, Name, Email}
```

---

## ENVIRONMENT-SPECIFIC CONFIGURATION

### Production (env == "prod")
| Flow | Bucket | TodoList/Set | Project | Assignee/Approver |
|------|--------|--------------|---------|-------------------|
| Invoice | Woodland (9403032) | 1346305133 | - | HanBasecampID |
| Accounting (Cronjob) | Accounting (15258324) | 2329633561 | 15258324 | Quang (Out), Giang (In) |
| Accounting (Webhook) | Accounting (15258324) | 2329633561 | 15258324 | - |
| Expense | Woodland (9403032) | 2353511928 | - | HanBasecampID |
| On-Leave | OnLeave (6935836756) | - | Woodland | HuyNguyenBasecampID, GiangThanBasecampID |

### Development (env != "prod")
| Flow | Bucket | TodoList/Set | Project | Assignee/Approver |
|------|--------|--------------|---------|-------------------|
| Invoice | Playground (12984857) | 1941398075 | - | NamNguyenBasecampID |
| Accounting (Cronjob) | Playground (12984857) | 1941398075 | 12984857 | Quang (Out), Giang (In) |
| Accounting (Webhook) | Playground (12984857) | - | 12984857 | - |
| Expense | Playground (12984857) | 2436015405 | - | NamNguyenBasecampID |
| On-Leave | OnLeave Playground (2243342506) | - | Playground | NamNguyenBasecampID |

---

## MIGRATION CHECKLIST

### Per-Flow Complexity
- **On-Leave:** ⭐⭐⭐⭐⭐ Most complex (calendar sync, date parsing, multiple approvals)
- **Invoice:** ⭐⭐⭐⭐ (comment validation, multiple async tasks)
- **Expense:** ⭐⭐⭐ (amount parsing, approver assignment)
- **Accounting:** ⭐⭐ (simplest, mostly read-only)

### Key Data to Preserve
- [ ] Invoice status history + paid_at timestamps
- [ ] Accounting transaction metadata (source, id)
- [ ] Expense basecamp_id mappings
- [ ] On-leave request creator/approver relationships
- [ ] Currency conversion rates

### Key Code to Update
- [ ] Basecamp API calls → NocoDB queries
- [ ] Title regex parsing → Form field reads
- [ ] Approval workflows → Status field transitions
- [ ] Comment posting → Activity feed or webhook notifications
- [ ] Calendar creation → NocoDB calendar or external sync

### Key Constants to Remap
- [ ] Project/Bucket IDs → NocoDB table/workspace IDs
- [ ] Todo List IDs → NocoDB view IDs
- [ ] Person IDs → NocoDB user IDs (if applicable)
- [ ] Hardcoded approver IDs → Role-based assignments

---

## FILE LOCATIONS REFERENCE

```
pkg/
├─ handler/webhook/
│  ├─ basecamp.go              (Invoice handler)
│  ├─ basecamp_invoice.go      (Invoice logic)
│  ├─ basecamp_accounting.go   (Accounting logic)
│  ├─ basecamp_expense.go      (Expense logic)
│  └─ onleave.go               (On-Leave logic)
│
├─ controller/invoice/
│  └─ update_status.go         (Invoice paid processing)
│
├─ store/
│  ├─ invoice/invoice.go
│  ├─ accounting/accounting.go
│  ├─ expense/expense.go
│  └─ onleaverequest/onleave_request.go
│
├─ model/
│  ├─ invoice.go
│  ├─ accounting.go
│  ├─ expense.go
│  └─ onleave_request.go
│
└─ service/basecamp/
   ├─ integration.go           (Expense/Accounting logic)
   ├─ consts/consts.go         (All hardcoded IDs)
   └─ [subservices]

pkg/routes/v1.go              (Webhook routes, lines 67-90)
migrations/schemas/
├─ 20221127165238-create_invoices_table.sql
├─ 20230106033811-accounting.sql
├─ 20230417095551-create-expenses-table.sql
└─ 20230531143430-add_onleave_request_table.sql
```
