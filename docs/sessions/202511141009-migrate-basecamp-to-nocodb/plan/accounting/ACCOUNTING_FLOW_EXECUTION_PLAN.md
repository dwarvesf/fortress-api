# Accounting Flow Execution Plan

## Objective
Replace Basecamp as the transport for monthly accounting todos and webhook-driven accounting transactions by reusing the TaskIntegration scaffold, wiring a NocoDB-backed implementation, and cutting over with zero disruption to Ops/Accounting reconciliation.

## Preconditions
- ✅ Research artifacts ready (`overview/ACCOUNTING_IN_OUT_DETAILS.md`, `flows/2_ACCOUNTING_FLOW_MIGRATION_PLAN.md`).
- ✅ Shared TaskIntegration + provider factory merged from invoice work.
- ☐ NocoDB Accounting workspace + tables/views provisioned (`AccountingTodos`, `AccountingTransactions`).
- ☐ Environment config entries for accounting provider (IDs, webhook secret, service templates) defined and injected.
- ☐ Sample NocoDB webhook payloads captured for parser fixtures.

## Workstream A – Provider & Config Extension (Week 1)
1. Extend `TaskIntegration` (or sibling interface) with accounting-specific methods: `CreateMonthlyPlan`, `CreateAccountingTodo`, `ParseAccountingWebhook`.
2. Add `AccountingIntegration` struct in config; load project/list/group identifiers and service template overrides from env/secret store.
3. Update provider factory to instantiate Basecamp and NocoDB accounting adapters under the same flag used for invoices (`TASK_PROVIDER=basecamp|nocodb`).
4. Document config contract in `README`/`.env.example`; ensure dev/staging values exist.

## Workstream B – NocoDB Data Modeling (Week 1)
1. Design `AccountingTodos` schema mirroring "Accounting | {Month Year}" todo list, `group` (In/Out), `due_on`, assignees array, `template_id`, and metadata for salary/service types.
2. Define `AccountingTransactions` (or reuse existing table) with references back to todo rows (row_id, provider) plus bucket/category fields.
3. Configure NocoDB views/filters to emulate Basecamp group separation and expose webhook triggers on insert/update.
4. Export schema docs + sample payload JSON into `research/nocodb/` for future flows.

## Workstream C – Cronjob Writer Migration (Week 2)
1. Refactor `pkg/handler/accounting/accounting.go` to call provider abstraction; keep existing validation/logging but eliminate direct Basecamp client usage.
2. Implement NocoDB adapter logic for: monthly board creation, In/Out grouping, service template iteration, salary todo generation, and TM project todo creation (reuse `ACCOUNTING_IN_OUT_DETAILS.md` mapping).
3. Persist provider metadata (NocoDB row IDs, board IDs) alongside existing Basecamp IDs in the database for reconciliation/rollback.
4. Roll out behind the `TASK_PROVIDER` flag, switching directly to NocoDB once verified in staging (no dual-write).

## Workstream D – Webhook Receiver Migration (Week 3)
1. Define `/webhooks/accounting` payload contract for NocoDB (mirror Basecamp fields: todo content, group, status, actor, timestamp, amount metadata).
2. Implement parser translating payload into `AccountingTransactionCreateInput`, including month/year inference and bucket detection logic.
3. Route `/webhooks/nocodb/accounting` through the provider abstraction. Basecamp endpoint is deprecated (kept only for historical replay), so no further Basecamp changes required.
4. Add retry/backoff + signature validation for the new webhook endpoint.

## Workstream E – Testing, Dual Run & Cutover (Week 4)
1. Add table-driven unit tests for cronjob handler covering template expansion, salary todo creation, and provider selection (mock Basecamp/NocoDB adapters).
2. Create parser tests using captured Basecamp + NocoDB payload fixtures to ensure identical transaction outputs.
3. Cutover plan: promote `TASK_PROVIDER=nocodb` after staging validation, monitor logs/transactions, keep Basecamp code as a rollback option only.

## Cutover Checklist
1. Verify NocoDB tables populated for upcoming month via cronjob dry-run.
2. Confirm webhook secrets/URLs deployed and receiving events (staging → prod).
3. Toggle provider flag to `nocodb` for cronjob + webhook simultaneously; monitor logs, accounting transaction counts, and Ops confirmation.
4. After stable cycle, remove Basecamp-only config and archive unused constants.

## Owners & Dependencies
- **Primary Dev:** TBD (needs familiarity with accounting cron + TaskIntegration).
- **Reviewer:** Finance Platform engineer (validates transaction accuracy + config rollout).
- **Dependencies:** Operational service templates, project data, Wise currency service remain unchanged; coordinate with Ops for NocoDB board access.

## Success Metrics
- 100% of monthly todos (In/Out, salaries, TM projects) created in NocoDB with correct metadata.
- Accounting webhook generates identical transaction records (amount, bucket, currency) compared to Basecamp baseline.
- Basecamp accounting endpoints disabled after cutover; no new Basecamp API calls for cron/webhook flow.
- Ops/Accounting confirms visibility, notifications, and reconciliation timelines are unaffected or improved.
