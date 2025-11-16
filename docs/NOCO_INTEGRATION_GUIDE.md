# NocoDB Integration Guide

_Source: Session `202511141009-migrate-basecamp-to-nocodb` (research ▸ plan ▸ implementation folders)._

## 1. Goals & Scope

- Sunset Basecamp as the task + webhook provider for invoice and accounting flows without disrupting Ops/Finance stakeholders.
- Reuse the shared `TaskIntegration` abstraction so Fortress can switch between Basecamp and NocoDB via config (`TASK_PROVIDER=basecamp|nocodb`).
- Provide schemas, API contracts, webhook payloads, and rollout steps so client implementers can build/extend the NocoDB provider confidently.

## 2. Architecture Overview

1. **Provider Interface** – `TaskIntegration` exposes invoice methods (`CreateInvoiceTask`, `AttachInvoiceFile`, `PostInvoiceComment`, `CompleteTask`, `ParseInvoiceWebhook`) and accounting methods (`CreateMonthlyPlan`, `CreateAccountingTodo`, `ParseAccountingWebhook`). All existing controllers/cronjobs now call this interface instead of Basecamp-specific code.
2. **Provider Factory** – Initialized in `cmd/server/main.go` / `pkg/service/service.go`, reading env-backed config structs (`InvoiceIntegration`, `AccountingIntegration`). Provider IDs, list/table IDs, webhook secrets, and auth tokens all come from `.env` (see implementation tasks A1–A4 in both Invoice/Accounting lists).
3. **NocoDB Client Layer** – Wraps the REST APIs documented in `research/nocodb/openapi.json` (notably `/api/v2/storage/upload` and `/api/v2/tables/{tableId}/records[/{recordId}]`). The adapters keep HTTP concerns (xc-token header, base URL) isolated from business logic.
4. **Metadata Bridge** – New Postgres table `accounting_task_refs` stores `(task_provider, task_ref, task_board, project_id, template_id, metadata)` for every todo the cronjob generates. Invoice send flow and accounting webhook both use this to correlate Noco rows back to Fortress invoices (plan/accounting/ACCOUNTING_NOCODB_SCHEMA.md).

## 3. Data Modeling

### 3.1 Invoice Assets (`plan/invoice/INVOICE_NOCODB_SCHEMA.md`)

| Table | Highlights |
| --- | --- |
| `invoice_tasks` | Columns: `id` (UUID PK), `invoice_number` (unique), `month`, `year`, `status` (`draft|sent|paid|overdue|error`), `amount`, `currency`, `attachment_url` (Noco upload descriptor), `fortress_invoice_id` (UUID string), timestamps. Index on `(status,due_date)` for views. Stores provider metadata for cross-linking to accounting todos. |
| `invoice_comments` | Optional activity log referencing `invoice_tasks.id` with `author`, `message`, `type`. Used when worker mirrors comment streams. |
| `invoice_webhook_events` | Optional audit store of received webhook payloads for debugging/replay. |

### 3.2 Accounting Assets (`plan/accounting/ACCOUNTING_NOCODB_SCHEMA.md`)

| Table | Highlights |
| --- | --- |
| `accounting_todos` | Columns: `id`, `board_label` (`Accounting | {Month Year}`), `task_group` (`in`/`out`), `title`, `description`, `due_on`, `assignee_ids` array, `status` (`open`/`ready`/`paid` → simplified to `open`/`completed` per extension notes), `metadata` (JSON storing month/year/template/project IDs + invoice linkage). |
| `accounting_transactions` | Stores webhook-derived state changes (todo_row_id, group, bucket, amount, currency, actor, status, occurred_at, payload JSON). For the simplified flow this is optional because invoice payment now relies on todo completion metadata. |

### 3.3 Cross-link Metadata (plan/accounting/ACCOUNTING_FLOW_EXTENSION_NOTES.md)

- When an invoice is sent, look up the matching `accounting_todos` row via `accounting_task_refs` (match project/month/year) and patch its metadata/description with `invoice_id`, `invoice_number`, `invoice_task_id`, and the attachment descriptor.
- Accounting todo completion webhook updates both `invoice_tasks.status` and Fortress invoice status via `MarkInvoiceAsPaidWithTaskRef`.

## 4. NocoDB API Surface (research/nocodb/openapi.json)

| Purpose | Endpoint | Notes |
| --- | --- | --- |
| Upload invoice PDF | `POST /api/v2/storage/upload` | Multipart form-data; returns attachment descriptor (`url`, `signedUrl`, `mimetype`, `size`). Store descriptor in `invoice_tasks.attachment_url`. |
| Create rows | `POST /api/v2/tables/{tableId}/records` | Accepts object or array payload; set fields by API name (e.g., `invoice_number`). Use `xc-token` header for auth. |
| Update rows | `PATCH /api/v2/tables/{tableId}/records` | Send partial payload including `Id` field for metadata/status updates. NocoDB accepts record ID in request body, not URL path. |
| Count rows / query views | `GET /api/v2/tables/{tableId}/records` / `count` | Support filters for reconciliation scripts. |
| Manage links | `/api/v2/tables/{tableId}/links/{linkFieldId}/records/{recordId}` | Optional if we later add linked tables (e.g., clients). |

Implementation note: keep `NOCO_BASE_URL` without `/api/v2`; adapters append paths. All requests require `xc-token: <NOCO_TOKEN>`.

## 5. Invoice Workflow (plan/invoice/INVOICE_FLOW_EXECUTION_PLAN.md & implementation checklist)

1. **Send Invoice**
   - Controller generates PDF, calls provider `UploadInvoiceAttachment` (implementation task G1–G3) to push file to Noco storage, receives descriptor.
   - Provider `CreateInvoiceTask` inserts row into `invoice_tasks` with `invoice_number`, `project_id`, amount totals, `status=sent`, and `attachment_url` set to descriptor (fallback: keep GCS URL in metadata).
   - Provider `AttachInvoiceFile`/`PostInvoiceComment` optionally append context or notify assignees.
2. **Sync with Accounting**
   - Use `accounting_task_refs` to locate the In-group todo for the same project/month/year.
   - Patch the accounting todo with invoice metadata + mention (extension notes #1).
3. **Webhook Handling**
   - `/webhooks/nocodb/invoice` receives row updates from `invoice_tasks` (status transitions). Parser verifies `event`, table name, HMAC signature (`x-nocodb-signature`), loads invoice, and updates DB status.
   - `/webhooks/nocodb/accounting` listens for `accounting_todos` completions (In group). Handler reads metadata, updates `invoice_tasks.status=paid`, calls `MarkInvoiceAsPaidWithTaskRef`, and logs Discord event.
4. **Dual-Run Strategy**
   - Execution plan originally called for dual-write (Basecamp + Noco) but was de-scoped; provider flag ensures Basecamp path remains available for rollback.

## 6. Accounting Workflow (plan/accounting/ACCOUNTING_FLOW_EXECUTION_PLAN.md & implementation checklist)

1. **Cronjob Writer**
   - `pkg/handler/accounting/accounting.go` now calls provider APIs to create monthly boards and todos for In/Out, salary, and TM projects. Each insert records a row in `accounting_task_refs` with provider metadata (implementation tasks C8–C11).
   - Metadata includes month/year, template/service IDs, and (post-extension) invoice linkage placeholders.
2. **Webhook Receiver**
   - `/webhooks/accounting` is provider-neutral. For NocoDB it accepts payloads like `research/nocodb/accounting_webhook_sample.json` and transforms them into `AccountingTransactionCreateInput` (tasks D12–D14). Signature or bearer token stored in `NOCO_WEBHOOK_SECRET` (schema doc §3).
   - After extension scope (F4), the handler no longer creates accounting transactions for invoices; it focuses on invoice synchronization.
3. **Sample Payload**
   ```json
   {
     "tableId": "tbl_accounting_transactions",
     "rowId": "row_123",
     "event": "row.updated",
     "triggeredBy": {"name": "Giang Than", "email": "giang@example.com"},
     "new": {
       "todo_row_id": "row_todo_456",
       "board_label": "Accounting | January 2025",
       "group": "out",
       "title": "Office Rental | 1.500.000 | VND",
       "amount": 1500000,
       "currency": "VND",
       "status": "completed",
       "metadata": {"month": 1, "year": 2025, "template_id": "svc_office_rental"}
     },
     "old": {"status": "open"}
   }
   ```
4. **Status Loop**
   - When Ops marks a todo completed in NocoDB, webhook → Fortress updates accounting tables + invoice statuses (if metadata present) <5 minutes (success metric).

## 7. Security & Validation

- **Authentication** – All API calls use `xc-token` (personal/workspace token). Store per environment.
- **Webhooks** – Invoices use HMAC SHA-256 signature header; Accounting uses Bearer secret header (`Authorization: Bearer <NOCO_WEBHOOK_SECRET>` per schema doc §3). Both handlers validate table name + event type before mutating state.
- **Email Restrictions** – Invoice send path still enforces `@dwarves...` addresses outside prod (see `pkg/handler/invoice/request/request.go`) even though provider changed.
- **Attachments** – Always upload to Noco storage (extension notes). Keep external GCS link in metadata only for fallback.

## 8. Environment & Setup (plan/accounting/NOCO_SETUP_NOTES.md)

1. **Provision Tables** – Use `scripts/local/create_nocodb_accounting_tables.sh` with `NOCO_BASE_URL`, `NOCO_TOKEN`, `NOCO_BASE_ID`. Script prints table IDs for `.env` (`NOCO_ACCOUNTING_TODOS_TABLE_ID`, etc.). Rename legacy camelCase tables to snake_case to match automation names.
2. **Meta API Tips** – Primary keys must be declared explicitly; choose supported `uidt` values (`AutoNumber`, `SingleLineText`, `JSON`, …). Keep base URL without `/api/v2`.
3. **Configuration Keys** – Add to `.env` / secrets:
   - `TASK_PROVIDER`
   - `NOCO_BASE_URL`, `NOCO_TOKEN`, `NOCO_BASE_ID`
   - Table IDs for invoice/accounting assets
   - Webhook secrets (`NOCO_INVOICE_WEBHOOK_SECRET`, `NOCO_ACCOUNTING_WEBHOOK_SECRET`)
   - Optional: upload bucket overrides, board/list IDs

## 9. Cutover Strategy

### Invoice Flow
- Plan called for dual-write shadowing, but implementation cut scope after full provider switch once unit/integration tests passed. Keep Basecamp path behind the flag for rollback.
- Monitor `/webhooks/nocodb/invoice` success rate, ensure Ops confirm attachments + notifications.

### Accounting Flow (plan/accounting/CUTOVER_CHECKLIST.md)
1. Stage: set `TASK_PROVIDER=nocodb`, run migrations, trigger cronjob, validate data + webhook in staging.
2. Prod: deploy config, trigger cronjob manually for upcoming month, disable Basecamp automation, enable Noco webhook.
3. Monitor: watch logs for `StoreNocoAccountingTransaction`, compare counts to Basecamp baseline, revert via flag if needed. After one healthy cycle, strip Basecamp IDs.

## 10. Implementation Status (implementation/*.md)

| Area | Status | Notes |
| --- | --- | --- |
| Provider infrastructure | ✅ | Interfaces, factories, config wiring completed for both invoice + accounting (Invoice tasks A1–A4, Accounting tasks A1–A4). |
| Noco schema + webhook setup | ✅ | Tables defined + payload docs captured (Invoice tasks B5–B7, Accounting tasks B5–B7). |
| Invoice send path migration | ✅ | Controller, worker, webhook refactors done (tasks C8–C10, D12–D13). Dual-write skipped but documented. |
| Accounting cron + webhook migration | ✅ | Cron + webhook call provider, metadata persisted (tasks C8–C11, D12–D14). |
| Attachment & metadata sync | ✅ | Upload helper + metadata patching implemented (Accounting tasks F1–F4, G1–G3). |
| Outstanding items | ⚠️ | - End-to-end validation of invoice↔accounting handshake still open (Accounting task F5).<br>- Broader regression tests + rollout documentation for attachment upload (Accounting task G4). |

## 11. Next Steps & Recommendations

1. **Complete E2E Validation** – Run manual flow (send invoice → patch accounting todo → complete todo → verify invoice status) before next production cycle.
2. **Finalize Testing Docs** – Capture attachment upload regression plan per implementation task G4 and add to QA checklist.
3. **Monitor Metadata Drift** – Add periodic job or dashboard comparing `invoice_tasks`, `accounting_todos`, and `accounting_task_refs` to catch orphaned rows after cutover.
4. **Document Config in README** – Mirror this guide in the public developer docs so future flows (Expense, Payroll) can reuse the pattern.

With the above architecture, schemas, API surface, and rollout steps, integrating NocoDB as the unified task provider should be straightforward while keeping an escape hatch back to Basecamp via the provider flag.
