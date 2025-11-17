# Expense Flow Research

## Existing Basecamp Flow Findings
- Handlers live in `pkg/handler/webhook/basecamp_expense.go`; they listen to Basecamp `todo_created`, `todo_completed`, `todo_uncompleted` events and call Basecamp APIs for todo assignment, validation comments, and deletion.
- Expense title parsing expects the format `<Reason> | <Amount> | <Currency>` with hardcoded bucket/project IDs and assignee IDs defined in `pkg/service/basecamp/consts` (e.g., Han Ngo vs Nam Nguyen between prod/non-prod). Validation comments are posted back to the todo when parsing fails or succeeds.
- Amount parsing invokes `service.Basecamp.ExtractBasecampExpenseAmount`, currency limited to VND/USD. Attachments are fetched via `Recording.TryToGetInvoiceImageURL` and stored as metadata for downstream accounting/expense stores.
- Upon success the handler calls `service.Basecamp.CreateBasecampExpense` which orchestrates store writes (`store/expense`, `store/accounting`, `store/employee`), Wise currency conversions, and accounting transaction creation.

## Target State Considerations
- We must replicate validation + response UX: auto-assigning approvers, leaving success/failure comments, and ensuring expense metadata reaches accounting services.
- Unlike invoices/accounting, the expense cutover will flip directly to NocoDB once validated (no dual-write). Therefore, feature-flag behavior must allow targeted rollout (likely via existing `TASK_PROVIDER` flag) but we need explicit gating per flow (expense vs invoice) to avoid unintended switches.
- ExpenseIntegration should encapsulate: workspace/table IDs, approver mapping, webhook secret, and file upload endpoints. Even though `TASK_PROVIDER` selects Basecamp vs NocoDB globally, expense-specific config (table/list IDs, view IDs, automation IDs) must live under a dedicated config struct to avoid mixing with invoice/accounting settings.

## NocoDB Modeling & API Research
- Table fields should capture: reason text, amount (decimal), currency enum, requester (relation to employees table or raw email), approver relation, supporting docs (file attachments with descriptor URLs), status (draft/validated/approved/rejected), metadata (Basecamp legacy ID, Fortress expense ID, timestamps), and bucket classification for accounting.
- We must ensure attachments can be uploaded once (similar to invoice flow). NocoDB supports multipart file columns returning descriptor IDs; reuse the `pkg/service/nocodb` upload helper built for invoices, but confirm size limits and retention settings.
- Automations: configure row-created and row-updated webhooks to mimic Basecamp events. Payload will include `view_id`, `row_id`, `oldRow`, `newRow`. We'll need to document mapping from event type to internal handler actions (validate/create/completion/uncompletion) and include signature verification using NocoDB's HMAC secret.

## Security & Compliance Notes
- Webhook endpoint `/webhooks/nocodb/expense` (or unified `/webhooks/expense`) must validate HMAC signature, timestamp, and optional IP allow list. Replay protection is recommended (store last seen event IDs or timestamps per row).
- Attachments may contain receipts with PII; ensure uploaded files are stored in NocoDB with access restricted to Finance workspace or mirrored into Fortress-managed storage with signed URLs.
- Logging should redact currency amounts only when necessary; ensure error logs do not leak invoice images URLs if they include tokens.

## Testing & Observability
- Unit tests need to cover validation permutations (invalid format, unsupported currency, zero amount) and provider selection logic.
- Integration/fake-server tests should assert REST payloads for create/update/delete operations, ensuring metadata (row IDs, status transitions) matches expectation.
- Observability: add metrics for expense webhook events by provider, validation failures, and create/uncheck outcomes. Plan to include structured logging fields (row_id, provider, status) for easier cutover monitoring.

## Open Questions / Risks
- Approver mapping: do we have NocoDB user IDs for all finance approvers? Need a sync/mapping strategy (maybe store in config or DB table) before implementation.
- Attachments: if NocoDB file storage is insufficient, may need hybrid approach (store in GCS but link from row). Determine acceptable retention/performance.
- Accounting transaction coupling: confirm whether `CreateBasecampExpense` responsibilities are reusable or need new provider-neutral service boundaries.
