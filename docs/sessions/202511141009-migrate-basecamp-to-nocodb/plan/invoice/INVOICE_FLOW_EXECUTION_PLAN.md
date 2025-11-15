# Invoice Flow Execution Plan

## Objective
Cut Basecamp out of the invoice lifecycle (send, comment, webhook) by standing up a NocoDB-backed provider, dual-running for validation, then switching permanently with zero downtime for accounting and ops teams.

## Preconditions
- ✅ Migration documentation complete (`overview/` + `flows/INVOICE_FLOW_MIGRATION_PLAN.md`).
- ☐ NocoDB workspace + tables for invoice tasks/comments exist (fields: project, month, invoice number, status, assignees, attachment links).
- ☐ Feature flag / config entry (e.g., `TASK_PROVIDER`) wired through environment configs.
- ☐ NocoDB credentials + webhook endpoint registered.

## Workstream A – Shared Infrastructure (Week 1)
1. Finalize provider abstraction (`TaskIntegration`) with methods: `CreateInvoiceTask`, `AttachInvoiceFile`, `PostInvoiceComment`, `CompleteTask`, `ParseInvoiceWebhook`.
2. Implement NocoDB client + todo/comment sub-services using existing interface contracts.
3. Externalize Basecamp/NocoDB IDs via config struct (no more `consts` usage in controllers/handlers).
4. Add provider selection logic in `cmd/server/main.go` and `pkg/service/service.go` (default Basecamp, flag for NocoDB).

## Workstream B – Data Modeling & Webhooks (Week 2)
1. Define NocoDB schema for invoice board (tables/views, status enums, attachment storage policy). Single approver handles all invoices, so we can skip a `paid_by` column and instead rely on the webhook actor/automation metadata for verification.
2. Configure NocoDB webhooks for task status changes; document payload format.
3. Build webhook parser translating NocoDB payloads to internal structs (invoice number, status, approver, files) while keeping Basecamp handler intact.
4. Seed dev data and run smoke tests by hitting webhook endpoint with sample payloads.

## Workstream C – Code Migration (Week 3)
1. Update `pkg/controller/invoice/send.go` to call provider interface for todo creation + attachment upload; remove Basecamp-specific code paths.
2. Update worker (`pkg/worker/worker.go`) to branch on provider when posting comments/attachments.
3. Refactor webhook handler (`pkg/handler/webhook/basecamp_invoice.go`) into provider-neutral module; add NocoDB parser and route alias if needed.
4. Ensure invoice controller still records metadata (todo IDs, attachment sgids or NocoDB row IDs) for later reconciliation.

## Workstream D – Testing & Dual Run (Week 4)
1. Unit tests: provider selection, controller send path (mock provider), webhook parser for both Basecamp + NocoDB payloads.
2. Integration tests: hit fake NocoDB API to verify REST payloads and responses.
3. Dual-write strategy: when Basecamp remains primary, optionally create NocoDB tasks in shadow mode to compare data; log discrepancies.
4. Release checklist: 
   - Update `.env` + deployment manifests with provider flag and NocoDB creds.
   - Grant ops access to NocoDB board + webhook monitoring dashboard.

## Cutover Steps
1. Enable dual-write in staging → validate 3 invoice cycles.
2. Enable dual-write in production with NocoDB shadow logging for 1-2 weeks.
3. Flip provider flag to NocoDB; keep Basecamp code path behind flag for rollback.
4. Monitor logs/webhooks; once stable, clean up unused Basecamp IDs and close dual-write.

## Owners & Dependencies
- **Primary Dev:** TBD (needs Go + NocoDB familiarity)
- **Reviewer:** Platform/Infra engineer (for config + secret rollout)
- **Dependencies:** Google Drive/Mail services unchanged; ensure attachments accessible via new task provider.

## Success Metrics
- 100% of new invoices create NocoDB tasks with correct metadata and attachments.
- Invoice paid webhook updates status in <5 minutes with correct approver info.
- No Basecamp API calls from invoice handlers once flag is on.
- Ops/Accounting confirm visibility and notifications match prior Basecamp behavior.
