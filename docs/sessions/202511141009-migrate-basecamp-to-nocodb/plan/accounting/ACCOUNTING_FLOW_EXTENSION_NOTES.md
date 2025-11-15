# Accounting Flow – NocoDB Extension Notes (2025-11-15)

## Context
The original migration scope covered writing accounting todos/webhooks via the shared `TaskIntegration` abstraction and storing transactions in Postgres. The invoice-specific tables in Noco (`invoice_tasks`) are now complete, and we want tighter coupling between Accounting In-group todos in Noco and the invoice workflow that previously lived entirely in Basecamp.

## Newly Agreed Scope
1. **Attach Invoice PDFs to both invoice_tasks and accounting_todos (Group In)**
   - Hybrid approach: add an `attachment_url` column to `invoice_tasks` for the canonical link, and surface the invoice number string (no attachment) on the Accounting todo description/metadata once the invoice is sent so downstream lookups can match on that marker.
   - Sequence:
     1. Cron still creates In-group todos with empty descriptions; every insert records an `accounting_task_refs` row (project_id, month/year, group, provider ref).
     2. When `controller.invoice.Send` runs, upload the PDF once via `TaskProvider.UploadAttachment`, persist the returned URL/markup in the new invoice column, and use `accounting_task_refs` to find the matching Accounting todo (project/month/year).
     3. Patch that todo (via Noco API) to append `<invoice_number>` to its description (used later for lookups) and store `invoice_id`, `invoice_number`, `invoice_task_id`, `attachment_url` in metadata for webhook lookups.

2. **Propagate payment state back to invoices (todo-only flow)**
   - `/webhooks/nocodb/accounting` no longer needs `amount/currency`; its sole job is to react when an `accounting_todos` row in the In group transitions to `completed`.
   - On completion, update the matching `invoice_tasks` row in Noco to `status=paid` and call `MarkInvoiceAsPaidWithTaskRef` so the Fortress invoice flips to `paid`. No separate `accounting_transactions` insert is required.
   - This relies entirely on the metadata we stamped when sending the invoice (invoice_id, invoice_task_id, invoice_number).

   - Accounting todo webhook updates invoice status immediately; `/webhooks/nocodb/invoice` remains as a safety net when the invoice table itself changes (manual edits, retries).

3. **Noco schema simplification**
   - `accounting_todos.status` only needs `open`/`completed` states. Remove unused columns such as `template_id` (now stored in metadata) to reduce sync noise.

## Outstanding Questions / Next Steps
- Define the exact metadata contract shared between invoice tasks and accounting todos (fields + types).
- Decide where to host the new Noco `PATCH` helpers (likely `pkg/service/nocodb/service.go`).
- Ensure the webhook’s security model covers the new cross-table updates (probably reuse the existing bearer secret/signature verification).

## OUT OF SCOPE / FUTURE WORK
- Handling invoices whose accounting todo is created on-demand (not via the monthly cron) — acknowledged as an edge case, out of current scope, to be implemented later.


## Notes

### Attachment of invoice PDF to accounting todo in Basecamp
- **Q:** When an invoice is sent, is the PDF attached to the accounting todo in Basecamp?
- **A:** Yes. During `controller.Send`, the code calls `dispatchInvoiceTask`, which uploads the generated PDF through the current task provider (`UploadAttachment`) and stores the returned Basecamp markup in `iv.TodoAttachment`  
  *(see `pkg/controller/invoice/send.go:184–205`)*.

  When the Basecamp-specific provider creates or updates the invoice todo, it appends that markup into the todo description  
  *(see `pkg/service/basecamp/todo/todo.go:399–417`)*.

  Therefore, whenever an invoice is sent with `TASK_PROVIDER=basecamp`, the PDF is uploaded and embedded in the accounting todo, ensuring the attachment stays with the Basecamp workflow.

### 

#### Problem
- `attachment_url` currently points directly to a PDF stored in GCS (e.g., `https://storage.googleapis.com/...`).
- NocoDB does **not** treat external URLs as real attachments, so:
  - No attachment chip in UI  
  - No preview  
  - No CDN/permissions handling

#### Correct Approach with NocoDB
1. **Upload the PDF to NocoDB** using their file-upload API:  
   `POST /api/v1/storage/upload` or `POST /api/v2/storage/upload` with multipart form-data + `xc-token`.
2. **Use the returned file descriptor** (URL or signedUrl + metadata: name, mimeType, size, etc.) and write it into the attachment column.  
   → This enables proper preview, CDN storage, and versioning.
3. If you still want the original **GCS URL**, keep it in metadata — but the **primary** `attachment_url` must come from NocoDB’s upload response.

#### Follow-up Steps
- Add a helper in `pkg/service/nocodb/service.go`:
  - `UploadFile(ctx, fileName, contentType, bytes)` → performs upload and returns the NocoDB descriptor.
- Update the invoice-sending flow:
  - After generating the PDF → upload to NocoDB first.  
  - Only fallback to GCS if upload fails.
- Store the NocoDB-provided URL/descriptor in `invoice_tasks.attachment_url`.  