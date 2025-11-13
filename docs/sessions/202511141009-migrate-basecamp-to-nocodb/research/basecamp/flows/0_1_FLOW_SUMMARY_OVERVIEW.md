# Basecamp Flow Business Logic Summaries

## Invoice Flow
1. `POST /api/v1/invoices/send` creates a Basecamp todo titled `{Project} {Month}/{Year} - #{InvoiceNumber}` in the Accounting project, attaches the invoice PDF, and enqueues a worker job to post a confirmation comment with the PDF asynchronously.
2. Webhook `PUT /webhooks/basecamp/operation/invoice` listens for the todo being completed, parses the title to extract the invoice, checks the “Paid” comment author (must be approver), updates invoice status in DB, posts success/failure comment, and closes the todo.

## Accounting Flow
1. Monthly cron `POST /api/v1/cronjobs/sync-monthly-accounting-todo` builds a todo list “Accounting | {Month} {Year}”, adds "In" (per active T&M project) and "Out" (operational services + salaries) groups with due dates/assignees derived from DB templates and hardcoded salary rules.
2. Webhook `POST /webhooks/basecamp/operation/accounting-transaction` reacts to each todo: parses titles like `Service | Amount | Currency`, derives month/year from parent list, validates currency/bucket, then inserts accounting transactions linked via todo metadata.

## Expense Flow
1. `todo_created` webhook validates todo title `Reason | Amount | Currency`, parses amount (supports k/tr/m), ensures bucket + currency, auto-assigns to approver, and posts validation feedback.
2. `todo_completed` webhook creates an expense record, converts currency via Wise if needed, generates an accounting transaction, and posts confirmation.
3. `todo_uncompleted` webhook removes the expense/transaction and logs via comment.

## On-Leave Flow
1. `todo_created` webhook validates leave todo titles `Name | Type | DateRange [| Shift]`, confirming type (`off`/`remote`), date order, bucket, and participants; posts pass/fail feedback.
2. `todo_completed` webhook creates `on_leave_requests` entries, splits date ranges into monthly calendar events, writes them to Basecamp schedule with subscribers (assignees + Ops), and records creator/approver metadata.

## Payroll Flow
1. Payroll calculation pulls approved Ops/Team/Accounting reimbursements by listing Basecamp todos (Ops + Woodland expense lists, Accounting "Out" group) and filtering for approver comments containing “approve” from Han/Nam; reimbursements tied to employees update bonus/reimbursement explains.
2. Commit phase reads `ProjectBonusExplain` entries with Basecamp IDs, completes reimbursement todos (Woodland) via `Todo.Complete`, and posts accounting notifications via worker-enqueued comments mentioning the employee.
