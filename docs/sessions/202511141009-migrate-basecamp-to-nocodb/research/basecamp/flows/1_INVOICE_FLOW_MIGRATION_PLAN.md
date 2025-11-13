# Invoices Flow Migration Plan

## 1. Inventory Current Flow
- Trace `POST /api/v1/invoices/send` (handler + controller) to list all Basecamp interactions: todo creation, attachment upload, worker queue payloads, logs.
- Map async pieces in `pkg/worker/worker.go` and Basecamp helpers invoked for invoices (`todo`, `comment`, `attachment`).
- Capture webhook touchpoints (`pkg/handler/webhook/basecamp_invoice.go`, shared helpers) to understand status updates.

## 2. Define Target NocoDB Objects
- Decide NocoDB tables/boards for invoice tasks, comments, attachments, and status state.
- Document schema + metadata needed (project, month, invoice number, amount, assignees, files) so downstream parsing mirrors Basecamp.
- Determine NocoDB webhook payload format and map it to existing webhook expectations.

## 3. Build Minimal NocoDB Client Layer
- Implement `pkg/service/nocodb/client` satisfying the existing `client.Service` contract.
- Create NocoDB versions of the needed sub-services first (`todo`, `comment`, `attachment`) using identical interfaces.
- Externalize system IDs via config/env instead of hard-coded constants; add config struct to hold Basecamp + NocoDB identifiers.

## 4. Add Provider Selection & Factory
- Introduce provider-neutral `TaskIntegration` interface with invoice-centric methods (create todo, attach file, post comment).
- Extend service wiring (e.g., `cmd/server/main.go`, `pkg/service/service.go`) to instantiate Basecamp or NocoDB providers based on `CONFIG_TASK_PROVIDER` with Basecamp default.
- Inject the provider into controllers, handlers, and workers via existing dependency injection.

## 5. Adapt Invoice Send Controller
- Refactor `pkg/controller/invoice/send.go` to call the provider interface; move Basecamp-specific formatting/attachment prep into provider adapters.
- Ensure PDF attachments upload and link correctly in both systems; update worker payloads to carry provider context.

## 6. Update Background Worker
- In `pkg/worker/worker.go`, switch on provider before posting comments/attachments, reusing the abstraction instead of Basecamp-only code.
- Align failure handling/retry semantics across providers.

## 7. Replace / Extend Webhook Handling
- Create provider-neutral webhook handler that dispatches to Basecamp or NocoDB parser modules.
- Implement NocoDB webhook parser translating payloads to invoice status updates; register route or reuse existing endpoint with provider discriminator.
- Add fixtures/tests for NocoDB webhook payloads alongside Basecamp cases.

## 8. Testing & Cutover
- Add unit tests for new NocoDB services and provider selection logic.
- Create integration/mocked HTTP tests covering invoice send + worker comment flow for both providers.
- Prepare migration checklist: populate NocoDB tables, configure webhook endpoints, update `.env` with credentials, run dual-write verification, then flip `CONFIG_TASK_PROVIDER=nocodb`.
