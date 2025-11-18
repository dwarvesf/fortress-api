# Expense Flow Specification

## 1. Overview
Replace Basecamp-driven expense submissions with a NocoDB-backed provider wired through the new `ExpenseIntegration` abstraction. The handler must parse webhook events, validate titles/metadata, persist expense/accounting records, and manage attachments/notifications with feature parity.

## 2. Goals
- Support create/complete/uncomplete expense lifecycle via NocoDB events.
- Preserve validation messaging (success/failure comments) with minimal behavior drift.
- Centralize configuration for workspace/list IDs, approver mapping, and webhook secrets.
- Provide safe rollback by toggling `TASK_PROVIDER` back to Basecamp.

## 3. Non-Goals
- No dual-write/shadow mode for expense submissions.
- No redesign of downstream accounting services; reuse existing APIs/stores.

## 4. Architecture
1. **ExpenseIntegration Interface**
   - Methods: `ValidateSubmission`, `CreateExpense`, `CompleteExpense`, `UncompleteExpense`, `PostFeedback`, `DeleteExpense`, `ParseWebhookPayload`.
   - Implementation: Basecamp + NocoDB (selected via provider flag). Both share DTOs defined in `pkg/service/taskintegration` to facilitate reuse.
2. **Handler Updates (`pkg/handler/webhook/expense.go`)**
   - Replace direct Basecamp references with integration interface.
   - Normalized flow: parse payload -> validate format -> call integration -> persist via `service.Expense` (existing) -> queue notifications.
3. **Service Layer**
   - Task provider factory instantiates ExpenseIntegration and injects into handler/service constructors in `pkg/service/service.go` / `cmd/server/main.go`.
   - Config struct `config.ExpenseIntegration` includes: `WorkspaceID`, `TableID`, `ViewID`, `ApproverMappings`, `WebhookSecret`, `AttachmentBucket`.
4. **Workflow**
   - Create Event: webhook `row.created` -> handler uses integration parser -> if format valid, create expense via `service.Expense` -> call `integration.PostFeedback(success)`.
   - Complete Event: `row.updated` with status `completed` -> mark expense paid + generate accounting transaction.
   - Uncomplete Event: `row.deleted` or `status` revert -> call `integration.UncompleteExpense` to reverse operations.

## 5. Data Modeling
- **NocoDB Table** `expense_submissions`
- Columns: `title` (required), `amount`, `currency` (SingleSelect limited to `VND` (default) and `USD`), `requester_team_email`, plus the operational tracking columns the webhook still needs (`status`, `attachment_files`, `metadata_json`).
  - Views for Prod vs Non-Prod buckets for targeted notifications; stored IDs in config.
  - Attachment handling: the table should use NocoDB's multi-file **Attachment** column so multiple receipts can live in one record. The backend currently consumes the first attachment entry (to keep behavior in parity with Basecamp) but we keep additional files in metadata for future expansion.
- **Metadata Persistence**
  - Extend existing expense DB table to store `task_provider`, `task_ref`, `attachment_url` if not already present.

## 6. API Contracts
- Webhook payload fields (NocoDB): `event` (create/update/delete), `tableId`, `viewId`, `rowId`, `rowData`, `oldRowData`, `triggerTime`, `signature`.
- Handler expects `signature` header for HMAC verification; rejects missing/mismatched signatures.
- Response semantics: always return 200 to avoid retries unless validation fails due to bad payload, in which case return 4xx.

## 7. Configuration
```
NOCO_EXPENSE_WORKSPACE_ID=
NOCO_EXPENSE_TABLE_ID=
NOCO_EXPENSE_WEBHOOK_SECRET=
NOCO_EXPENSE_APPROVER_MAPPING=han@example.com:123,nam@example.com:456
TASK_PROVIDER=basecamp|nocodb
```
- Mapping parsed at startup into struct for quick lookup.

## 8. Observability & Monitoring
- Log fields: `provider`, `row_id`, `event_type`, `status`, `amount`, `currency`.
- Metrics: `expense_webhook_events_total{provider,event}`, `expense_validation_failures_total`, `expense_create_duration_ms`.
- Alerts: high validation failure rate or webhook HMAC mismatches.

## 9. Risks & Mitigations
- **Risk:** Approver mapping drift. *Mitigation:* add startup validation ensuring every configured approver exists in employees table.
- **Risk:** Attachment upload failures. *Mitigation:* fallback to GCS link + comment instructing manual upload; include retries.
- **Risk:** Direct cutover without dual-write. *Mitigation:* stage/test, add monitoring, maintain instant rollback via flag.
