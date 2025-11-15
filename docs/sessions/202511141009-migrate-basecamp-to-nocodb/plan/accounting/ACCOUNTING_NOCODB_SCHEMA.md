# Accounting Flow – NocoDB Schema Draft

## 1. `AccountingTodos`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | Primary key managed by NocoDB |
| `board_label` | string | "Accounting \| {Month Year}"; unique per cycle |
| `group` | enum(`in`,`out`) | Mirrors Basecamp In/Out groups |
| `template_id` | string | Optional reference to operational_service template ID |
| `title` | string | Todo content (e.g., `Office Rental | 1.5B | VND`) |
| `description` | text | Extra note (e.g., CBRE split) |
| `due_on` | date | ISO date; matches Basecamp due date logic |
| `assignee_ids` | array<string> | Provider-specific assignee identifiers |
| `status` | enum(`open`,`ready`,`paid`) | Workflow status |
| `provider_ref` | string | Back-reference used by webhook payloads |
| `metadata` | json | Stores salary flag, month/year, TM project ID |

## 2. `AccountingTransactions`
| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | Primary key |
| `todo_row_id` | string | FK to `AccountingTodos.id` |
| `group` | enum(`in`,`out`) | Copied from todo |
| `bucket` | enum | Derived bucket (office_space, office_services, etc.) |
| `amount` | decimal | Amount parsed from todo or form field |
| `currency` | string | ISO currency code |
| `actor` | string | User who changed status |
| `status` | enum | `completed`, `reopened`, etc. |
| `occurred_at` | timestamp | When change happened |
| `payload` | json | Raw webhook payload stored for debugging |

## 3. Webhook Contract
- Headers: `Authorization: Bearer <token>` (shared secret stored in `NOCO_WEBHOOK_SECRET`)
- Body keys: `tableId`, `rowId`, `event`, `old`, `new`, `triggeredBy`
- Accounting payload adds `board_label`, `group`, `amount`, `currency`, `metadata.month`, `metadata.year`

Captured sample payload: `docs/sessions/202511141009-migrate-basecamp-to-nocodb/research/nocodb/accounting_webhook_sample.json`.

## 4. `accounting_task_refs` (Fortress DB)
- Captures every todo created by the cronjob: `month`, `year`, `group_name`, `task_provider`, `task_ref`, `task_board`, optional `template_id` (operational service) or `project_id` (TM project), plus a JSON metadata blob.
- Primary key: UUID; uniqueness enforced on (`task_provider`, `task_ref`) for fast lookups during webhook ingestion.
- Serves as the bridge between NocoDB/Basecamp tasks and downstream accounting transactions (no more title parsing).
