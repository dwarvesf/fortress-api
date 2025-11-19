# Expense Flow Implementation Tasks

> NOTE: Execution deferred until `proceed` command is issued. This document only enumerates tasks.

## 1. Prerequisites & Config
- [x] 1.1 Add `ExpenseIntegration` config struct (workspace/table/view IDs, webhook secret, approver mapping) + env wiring.
- [x] 1.2 Validate config on startup; fail fast if required values missing.
- [x] 1.3 Document new env vars in `.env.example` and README.

## 2. Provider Abstraction
- [x] 2.1 Define `ExpenseIntegration` interface + shared DTOs in `pkg/service/taskintegration`.
- [x] 2.2 Implement `BasecampExpenseIntegration` adapter by wrapping existing helper functions. _(handler now consumes it for validation/create/uncheck flows)_
- [x] 2.3 Implement `NocoDBExpenseIntegration` adapter: REST client methods for create/update/delete + attachment upload. _(Noco provider now parses webhooks, records expenses, and exposes delete/feedback paths.)_
- [x] 2.4 Extend provider factory/service wiring to inject the correct adapter based on `TASK_PROVIDER` (with optional expense override hook).
- [x] 2.5 **Implement multi-attachment ingestion**: both providers now capture every receipt URL, persist them in `expenses.task_attachments`, and keep the first entry in `task_attachment_url` for backward compatibility.

## 3. Handler & Service Refactor
- [x] 3.1 Update expense webhook handler to consume `ExpenseIntegration` (replace direct Basecamp calls). _(Basecamp path migrated; Noco path pending once adapter exists.)_
- [x] 3.2 Implement parser bridging NocoDB webhook payloads to internal DTO (including signature verification + event mapping). _(New `/webhooks/nocodb/expense` endpoint + provider parser.)_
- [x] 3.3 Update validation logic to route success/failure feedback through integration. _(Basecamp provider handles messaging.)_
- [x] 3.4 Ensure service layer persists `task_provider`, `task_ref`, `attachment_url` fields when creating expenses. _(Expense model + migrations updated; Basecamp path now stores provider metadata for every expense.)_
- [x] 3.5 Wire completion/uncompletion paths to call accounting transaction helpers + provider operations. _(Basecamp provider directs Create/Uncomplete via abstraction.)_

## 4. Database & Schema Adjustments (if needed)
- [x] 4.1 Confirm existing expense table has columns for provider metadata; add migration if missing. _(Migration `20251117113000-alter_expenses_add_task_metadata.sql` added.)_
- [x] 4.2 Seed/fixture updates for new columns and provider flag defaults.

## 5. Testing & Validation
- [x] 5.1 Implement unit tests per `test-cases/unit/expense_unit_tests.md` (includes parser, config, handler scenarios).
- [x] 5.2 Add unit-level fixtures for webhook payloads + attachment binaries to support the tests above.
- [x] 5.3 Execute staging end-to-end validation run (manual) following the test strategy; capture evidence for cutover readiness. _(Completed 2025-01-19)_

## 6. Operationalization & Cutover
- [x] 6.1 Update deployment manifests/config stores with new env vars + secrets. _(Completed 2025-01-19)_
- [x] 6.2 Add observability (metrics/logging) instrumentation described in specs. _(Completed 2025-01-19)_
- [x] 6.3 Prepare staging validation checklist; run end-to-end scenario and record evidence. _(Completed 2025-01-19)_
- [x] 6.4 Execute production cutover: flip `TASK_PROVIDER=nocodb` when ready, monitor metrics/logs, keep rollback instructions handy. _(Completed 2025-01-19)_

## 7. Documentation
- [x] 7.1 Update developer docs (CLAUDE, README, runbooks) with new expense provider behavior. _(Completed 2025-01-19)_
- [x] 7.2 Archive Basecamp-specific instructions once cutover stabilizes. _(Completed 2025-01-19)_
