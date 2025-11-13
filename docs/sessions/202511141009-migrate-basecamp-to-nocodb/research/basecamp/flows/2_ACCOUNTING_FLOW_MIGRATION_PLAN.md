# Accounting Flow Migration Plan

## 1. Inventory & Document Current Behavior
- Capture both components: (a) monthly cronjob `POST /api/v1/cronjobs/sync-monthly-accounting-todo`; (b) webhook `POST /webhooks/basecamp/operation/accounting-transaction`.
- List every Basecamp dependency: todo list creation, group creation, assignments, due date logic, and identifiers from `pkg/service/basecamp/consts`.
- Map data sources: `operational_service` store, Time & Material project queries, hardcoded salary rows, and transaction stores.
- Record validation logic in `pkg/handler/accounting/accounting.go` and `pkg/handler/webhook/basecamp_accounting.go` (regex parsing, bucket validation, amount parsing).

## 2. Define NocoDB Data Structures
- Decide how a monthly "Accounting" board/list is represented (e.g., `AccountingTodos` table with month/year columns + type column for In/Out).
- Model dependency for dynamic rows: service templates, salary tasks, TM projects; ensure schema supports tags like `group=in/out`, `due_date`, `assignee_ids`.
- Design an `AccountingTransactions` table/view (or reused existing) where webhook-created rows will land with metadata fields equivalent to Basecamp todo ID.
- Enumerate environment-specific configuration (prod vs dev) and wire them as `config.TaskProvider.Accounting` entries instead of constants.

## 3. Implement NocoDB Writer for Monthly Cronjob
- Build helper in new provider (`pkg/service/nocodb/accounting`) that can: create/find monthly board, create group buckets, insert todos with assignees & due dates.
- Update handler `pkg/handler/accounting/accounting.go` to call provider-agnostic interface rather than Basecamp-specific todo service, preserving logging + metrics.
- Move special-case logic (Voconic due date, salary multi-todo creation) into reusable functions shared by both providers.
- Ensure transaction metadata (IDs for later webhook matching) is persisted in DB (e.g., store NocoDB row ID next to todo IDs).

## 4. Implement NocoDB Webhook Receiver
- Define webhook payload contract for accounting transactions; document expected fields and register endpoint (reuse existing path with provider discriminator or create `/webhooks/nocodb/accounting`).
- Write parser that translates payload into `AccountingTransactionCreateInput`, mirroring Basecamp logic (month/year inference, reason | amount | currency parsing, salary detection).
- Replace Basecamp-specific fetches (e.g., GET parent todo list) with equivalent NocoDB API calls to retrieve board metadata if webhook lacks month context.
- Keep validation + categorization (Office Space vs Office Services) identical, but ensure IDs/foreign keys reference new tables.

## 5. Shared Provider Selection & Config
- Reuse global `TaskIntegration` toggle created for invoices; extend it with `AccountingIntegration` interface exposing `CreateMonthlyPlan` + `ParseTransactionWebhook`.
- Configure `CONFIG_TASK_PROVIDER_ACCOUNTING` (or reuse single provider variable) to choose Basecamp vs NocoDB per environment.
- Centralize ID mappings (project/list/group) into config so both UI + cronjob read the same data.

## 6. Testing & Cutover
- Add table-driven tests for cronjob handler verifying payload built for both providers (mock NocoDB client, assert correct groups/todos). Simulate service templates + TM projects fixtures.
- Unit test webhook parser with fixtures derived from Basecamp + NocoDB sample payloads to guarantee identical accounting transaction outputs.
- Dry-run plan: run cronjob in sandbox to populate both Basecamp + NocoDB (dual write) until verified; once validated flip provider flag for cronjob, then reroute webhook endpoint.
