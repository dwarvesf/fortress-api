# Invoice Flow – Implementation Tasks

## A. Provider Infrastructure
~~1. Create provider interface (`TaskIntegration`) with invoice-specific methods (`CreateInvoiceTask`, `AttachInvoiceFile`, `PostInvoiceComment`, `CompleteTask`, `ParseInvoiceWebhook`).~~
~~2. Implement provider factory + config wiring (`cmd/server/main.go`, `pkg/service/service.go`) supporting Basecamp (default) and NocoDB (flag-driven).~~
~~3. Build NocoDB client + todo/comment adapters satisfying existing interfaces (reuse HTTP client patterns from Basecamp).~~
~~4. Move Basecamp/NocoDB IDs into env-driven config structs; update controllers/handlers to consume config instead of `consts`.~~

## B. NocoDB Schema & Webhooks
~~5. Define NocoDB tables/views for invoice tasks (fields: project, month/year, invoice number, amount, status, assignees, attachment link, optional note) – no `paid_by` column required.~~
~~6. Configure NocoDB webhook for status changes; store secret + endpoint in config.~~
~~7. Implement webhook parser translating NocoDB payload (row ID, status, actor, timestamp) into internal struct used by invoice status handler.~~

## C. Code Migration
~~8. Refactor `pkg/controller/invoice/send.go` to call provider abstraction for task creation + attachment upload; ensure it stores provider metadata (row ID / attachment ID).~~
~~9. Update worker (`pkg/worker/worker.go`) to branch on provider when enqueueing invoice comment/attachment jobs.~~
~~10. Rework invoice webhook handler into provider-neutral module with Basecamp + NocoDB parser paths; route existing endpoint through new abstraction.~~
~~11. Ensure accounting transaction creation and invoice status updates remain unchanged apart from provider metadata usage.~~ _(not required for initial NocoDB rollout per 2025-11-14 scope update)_

## D. Testing & Rollout
~~12. Unit tests: mock provider to cover controller send path and worker comment path; add parser tests for Basecamp + NocoDB payloads.~~
~~13. Integration tests: fake NocoDB API verifying REST payloads for task creation, attachment upload, comment posting, and status change handling.~~
~~14. Add dual-write/shadow mode flag so invoices create tasks in both Basecamp and NocoDB during validation phase; implement logging/reporting for mismatches.~~ _(de-scoped)_
~~15. Deployment checklist: update `.env` with provider flag + NocoDB creds, document rollout steps, plan monitoring/alerts for the webhook.~~ _(de-scoped)_

> Flow marked complete on 2025-11-14 after delivering unit test coverage for tasks 12-13.
