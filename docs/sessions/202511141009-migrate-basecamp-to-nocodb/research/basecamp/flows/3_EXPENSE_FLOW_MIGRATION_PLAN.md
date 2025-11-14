# Expense Flow Migration Plan

## 1. Inventory Existing Expense Workflow
- Review handlers `pkg/handler/webhook/basecamp_expense.go` to catalog triggers (`todo_created`, `todo_completed`, `todo_uncompleted`) and Basecamp APIs called (todo assignment, comments, deletion).
- Capture dependency on `pkg/service/basecamp/integration.go` for amount parsing, Wise currency conversion, and accounting transaction creation.
- Document store usage: `pkg/store/expense`, `pkg/store/accounting`, `pkg/store/employee`, plus metadata stored in expenses table.
- Note all validations (title parsing, currency restrictions, bucket validation, approver requirements) and hardcoded IDs.

## 2. Design NocoDB Expense Objects & Views
- Define a table for expense submissions with fields mapping 1:1 to parsed data (reason, amount, currency, requester, approver, supporting docs, status).
- Decide how approval assignments are represented (user fields vs relation to employees) to replace Basecamp auto-assign.
- Determine where validation feedback lives (status column + system comments vs activity log) and how attachments/PDFs are stored.
- Map environment-specific list IDs + approvers into config, enabling role-based lookup instead of constant person IDs.

## 3. Implement NocoDB Expense Provider
- Create provider methods for: creating submission rows, updating status (validated/approved/rejected), posting comments/attachments, deleting rows on uncomplete.
- Integrate provider behind a new `ExpenseIntegration` interface that mirrors current Basecamp helper functions (validate, create, delete) so handler logic stays mostly the same.
- Support attachment upload path (Basecamp currently attaches invoice PDF) by wiring to NocoDB file columns or external storage referenced in row.
- Reuse existing amount parsing utilities; expose them under provider-neutral helpers.

## 4. Webhook & Event Translation
- Configure NocoDB automations/webhooks to emit events analogous to Basecamp `todo_*` events (e.g., row created/updated/deleted) and document payload.
- Build parser translating these payloads into handler inputs (kind, creator, title-equivalent fields). Maintain idempotency by storing NocoDB row IDs on the expense record.
- Replace Basecamp-specific comment posting with provider-neutral status writes + optional Slack/Discord notification if NocoDB lacks inline comments.

## 5. Validation & Auto-Assignment Logic
- Implement role lookup using employee data tied to NocoDB user IDs; maintain mapping table so approvals route correctly without hardcoded Basecamp IDs.
- Enforce same validation sequence (format, amount > 0, currency in allowed list, bucket restrictions) inside the handler before calling provider.
- Mirror assignment/responsibility: ensure new provider sets `approver_id`/`assignee_ids` fields and triggers notification (email/webhook) as Basecamp previously did via todo assignment.

## 6. Testing & Cutover
- Unit test expense handler with provider mock to exercise validation paths, creation, completion, and deletion scenarios.
- Add integration tests hitting a fake NocoDB API (httptest server) verifying correct REST payloads for create/update/delete.
- Execute staged rollout: dual-write expenses to both Basecamp + NocoDB, verify accounting transactions stay consistent, then switch incoming webhooks + disable Basecamp bucket.
