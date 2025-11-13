# NocoDB Schema & Webhook Design – Invoice Flow

## 1. Workspace Structure
- **Base (Workspace):** `Fortress Ops`
- **Tables:**
  1. `invoice_tasks` – task board replacement for Basecamp todos.
  2. `invoice_comments` – optional activity log (captures system + human notes).
  3. `invoice_webhook_events` – audit log of incoming webhook payloads (optional, helps debugging).

## 2. `invoice_tasks` Table
| Column | Type | Required | Notes |
|---|---|---|---|
| `id` | UUID (PK) | ✅ | Auto-generated.
| `invoice_number` | Text | ✅ | Unique index; matches Fortress invoice number (`2025-ABC-001`).
| `month` | Integer | ✅ | 1-12.
| `year` | Integer | ✅ | 4-digit year.
| `status` | Enum | ✅ | `draft`, `sent`, `paid`, `overdue`, `error` (maps 1:1 with Fortress enum).
| `amount` | Numeric | ✅ | Invoice total in project currency.
| `currency` | Text | ✅ | Currency code (ISO).
| `attachment_url` | Text | ✅ | Public GCS link to invoice PDF.
| `attachment_file` | File | optional | Mirror PDF inside NocoDB if desired.
| `fortress_invoice_id` | Text | ✅ | UUID string from Fortress DB; indexed for fast joins.
| `created_at` | Timestamp | auto | System-managed.
| `updated_at` | Timestamp | auto | System-managed.

### Indexes
- Unique index on `invoice_number`.
- B-Tree on `(status, due_date)` for views.
- Optional partial index on `status = 'sent'` to fetch pending approvals.

## 3. `invoice_comments` Table (optional)
| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `invoice_task_id` | UUID FK → `invoice_tasks.id` | cascade delete |
| `author` | Text | `system`, `user email`, etc. |
| `message` | Long Text | Markdown/HTML allowed |
| `type` | Enum | `info`, `success`, `warning`, `error` |
| `created_at` | Timestamp | auto |

Used by worker jobs to mirror the Basecamp comment stream when TaskProvider=NocoDB.

## 4. Webhook Design
- **Endpoint:** `/webhooks/nocodb/invoice`
- **Trigger:** NocoDB automation on row update (`status` changes) or manual “Mark Paid” button.
- **Payload (JSON):**
```json
{
  "event": "row.updated",
  "table": "invoice_tasks",
  "payload": {
    "id": "uuid",
    "invoice_number": "2025-DF-001",
    "status": "paid",
    "updated_by": "han@d.foundation",
    "attachment_url": "https://...",
    "fortress_invoice_id": "uuid"
  },
  "signature": "HMAC-SHA256"
}
```
- **Security:** HMAC header `x-nocodb-signature` using `NOCO_INVOICE_WEBHOOK_SECRET`.
- **Fortress Handler Logic:**
  1. Verify signature + table name.
  2. Verify `status` transitioned from `sent|overdue` → `paid`.
  3. Load invoice via `fortress_invoice_id` or `invoice_number`.
  4. Call `MarkInvoiceAsPaidWithTaskRef(invoice, ref, true)` where `ref.Provider = nocodb`, `ref.ExternalID = payload.id`.

## 5. Automation Recipes (inside NocoDB)
1. **On Row Insert (Status=sent):** auto-assign `assignee_email`, set `due_date`, send Slack/email notification (optional).
2. **On Row Update (Status=paid):**
   - Trigger webhook to Fortress.
   - Optionally lock row (read-only) or move to “Done” Kanban group.
3. **Validation:** ensure `invoice_number` unique, `status` constrained, required fields enforced via form UI.

## 6. Dual-Write Support
- Add `provider_metadata` JSON storing Basecamp IDs during transition: `{ "basecamp": { "bucket_id": 15258324, "todo_id": 123456 } }`.
- TaskProvider keeps writing to both Basecamp + NocoDB until flag flips; after cutover, Basecamp entries become optional.

## 7. Next Steps
- Create tables via NocoDB UI or API using this schema.
- Configure automation & webhook secret, store values in `.env` (`NOCO_INVOICE_TABLE_ID`, `NOCO_INVOICE_WEBHOOK_SECRET`).
- Update TaskProvider NocoDB implementation to call the REST endpoints (list rows by invoice number, create row with metadata, post comments to `invoice_comments`, etc.).
