# ADR-001: ExpenseIntegration Abstraction & Flagging

## Context
- Expense flow currently depends on Basecamp-specific handlers/services with hardcoded IDs and comment behaviors.
- Existing TaskIntegration infrastructure introduced for invoices/accounting already supports Basecamp vs NocoDB selection via the `TASK_PROVIDER` env flag, but expense logic is not yet wired into that abstraction.
- Requirement: perform a full Basecamp → NocoDB migration for expenses without dual-write, while retaining ability to fall back quickly if needed.

## Decision
1. Introduce a dedicated `ExpenseIntegration` interface that mirrors the current Basecamp helper contract (validate/create/delete/update/comment) but is provider-agnostic.
2. Provide two implementations—`BasecampExpenseIntegration` (existing logic) and `NocoDBExpenseIntegration` (new). The service registry will instantiate the appropriate implementation based on `TASK_PROVIDER`.
3. Keep `TASK_PROVIDER` as the single flag controlling provider selection across invoice/accounting/expense flows, but allow per-flow overrides via config to facilitate targeted rollouts (e.g., `provider.expense.override` defaults to global flag).
4. Consolidate expense-specific configuration (list IDs, workspace/table identifiers, approver mapping, webhook secrets) into `config.ExpenseIntegration`, injected into both implementations.

## Consequences
- + Enables a clean swap between providers without touching handler logic; fosters reuse of validation + metadata persistence code.
- + Reduces risk by centralizing provider config and making fallbacks symmetrical across flows.
- − Requires reworking service wiring and dependency injection for expense handlers, including updates to tests/mocks.
- − Single `TASK_PROVIDER` flag means the first expense cutover must align with invoice/accounting readiness unless we implement per-flow override (which we plan to support.)
