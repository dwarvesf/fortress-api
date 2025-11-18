# Expense Flow Migration Requirements

## Source
- User conversation on 2025-11-17.

## Scope
- Full replacement of Basecamp expense workflow with NocoDB-backed implementation (provider, webhooks, validations, cutover) using existing TaskIntegration flagging patterns.

## Clarifications
1. No dual-write phase is needed; once NocoDB path is validated we can switch directly.
2. Use the existing `TASK_PROVIDER` environment flag to select NocoDB (no new flag required).
3. Create a dedicated `ExpenseIntegration` abstraction that plugs into the shared TaskIntegration scaffolding but can encapsulate expense-specific config (list IDs, approver mappings, webhook secret, etc.).

## Assumptions
- Operational parity with Basecamp behavior is required (validations, notifications, accounting hooks).
- Existing research docs (e.g., `research/basecamp/flows/3_EXPENSE_FLOW_MIGRATION_PLAN.md`) remain the canonical source for historical context.
