# Accounting Flow – Implementation Tasks

## A. Provider & Config (Week 1)
~~1. Extend `TaskIntegration` with accounting-specific contract (`CreateMonthlyPlan`, `CreateAccountingTodo`, `ParseAccountingWebhook`).~~
~~2. Introduce `AccountingIntegration` config struct; add env variables for project/list/group IDs, webhook secret, service template overrides.~~
~~3. Wire provider factory + service registry to instantiate Basecamp/NocoDB accounting adapters under `TASK_PROVIDER` flag.~~
~~4. Document new config knobs in `.env.example` and developer README.~~

## B. NocoDB Data Modeling (Week 1)
~~5. Create `AccountingTodos` table/view in NocoDB (columns: board/month label, group, due date, assignees, template_id, provider metadata).~~
~~6. Create `AccountingTransactions` table/view for webhook outputs (fields: todo_row_id, group, bucket, amount, currency, actor, status, provider).~~
~~7. Capture schema diagrams + JSON payload samples in `docs/sessions/.../research/nocodb/` for reuse by other flows.~~

## C. Cronjob Writer Migration (Week 2)
~~8. Refactor `pkg/handler/accounting/accounting.go` to use provider abstraction instead of Basecamp client.~~
~~9. Implement NocoDB adapter logic for monthly board creation, In/Out group setup, service template + salary todo generation, TM project todo creation.~~
~~10. Persist provider metadata for every created todo: add a DB table/columns storing `task_provider`, `task_ref` (NocoDB row ID/Basecamp todo ID), `task_board`, and template/project identifiers so webhooks can look up structured context instead of relying on title parsing.~~
~~11. Roll out via the unified `TASK_PROVIDER` flag; no dual-write required (staging validation → flip flag in prod, keep Basecamp code for rollback).~~

## D. Webhook Receiver Migration (Week 3)
~~12. Define NocoDB accounting webhook payload + register endpoint (reuse `/webhooks/accounting` with provider discriminator or new route).~~
~~13. Implement parser translating payload into `AccountingTransactionCreateInput`, covering month/year inference and bucket detection.~~
~~14. Update webhook handler to branch on provider, reuse validation/transaction creation logic, and add signature verification + retry handling (Basecamp path considered deprecated).~~

## E. Testing & Cutover (Week 4)
~~15. Unit tests for cronjob handler: template expansion, salary todos, provider selection (mock adapters).~~
~~16. Unit tests for webhook parser using captured Basecamp + NocoDB fixtures.~~
~~17. Cutover execution: staging validation → flip `TASK_PROVIDER` to `nocodb`, monitor logs/transactions, keep Basecamp code for rollback until stable (documented in `plan/accounting/CUTOVER_CHECKLIST.md`).~~

## F. Invoice Attachment & Status Sync (Week 5)
~~1. Add Noco `invoice_tasks.attachment_url` column + migration; update config/docs.~~
~~2. Build repo helper to fetch In-group accounting_task_refs by project/month/year.~~
~~3. Update `controller.invoice.Send` to patch matching accounting todo descriptions with `<invoice_number>` and metadata via new Noco API helper.~~
~~4. **Todo-completion webhook flow** (no accounting_transactions insert):~~
   - ~~4.1 Remove the Wise/transaction logic from `/webhooks/nocodb/accounting`; handler only performs invoice linkage.~~
   - ~~4.2 Loose parsing: `ParseAccountingWebhook` must accept todo payloads without amount/currency and just track group/status/metadata.~~
   - ~~4.3 On completion: update `invoice_tasks` status + call `MarkInvoiceAsPaidWithTaskRef` using metadata.~~
~~5. Validate e2e: send invoice → todo metadata synced → mark todo completed → invoice tasks + Fortress invoice flip to `paid`.~~
6. ~~Fix webhook parser to read `task_group` so In-bucket todos trigger invoice sync (2025-11-15).~~
7. ~~Accounting webhook now only updates Noco invoice task status (DB status handled by invoice webhook) – 2025-11-15.~~

## G. Invoice Attachment Issue (NocoDB)
1. ~~Create `UploadFile` helper in `pkg/service/nocodb/service.go` (multipart upload → return descriptor URL) plus a thin `UploadInvoiceAttachment` wrapper.~~
2. ~~Wire `pkg/controller/invoice/send.go` to call the helper after PDF generation, store the returned Noco URL in `invoice_tasks.attachment_url`, and keep GCS URL only as metadata fallback.~~
3. ~~Ensure task provider payloads + accounting metadata can carry attachment info without relying on external URLs (update DTOs if needed).~~
~~4. Tests & verification: unit cover the upload helper (mock HTTP), regression test invoice send path, document validation steps once live.~~
