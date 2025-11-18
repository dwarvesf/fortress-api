# Expense Flow Test Strategy

## Testing Levels
- **Unit:** Handler validation, integration selection, config parsing, NocoDB payload parsers, service metadata persistence.
- **End-to-End (manual/QA):** Staging run that submits expenses via NocoDB UI, ensures Fortress receives webhooks, and accounting reconciliation completes.
- **(No automated integration tests):** Deferred per product guidance on 2025-11-17; coverage relies on unit + staged manual verification.

## Coverage Rationale
- Parser + validation logic is high-risk because incorrect extraction leads to wrong payouts.
- Webhook signature and provider switching protect us from unauthorized events and misconfiguration.
- Attachment upload reliability affects finance review, so retry/fallback scenarios must be simulated (validated via unit tests and staging e2e rather than automated integration tests).

## Test Data Requirements
- Sample payloads for: valid submission, malformed title, unsupported currency, completion, uncompletion, deletion.
- Mock employee directory to map approvers/responsible parties.
- Fixture for attachment binary to exercise upload path.

## Assumptions & Constraints
- NocoDB webhook secret shared securely; tests can stub value.
- Accounting transaction creation already covered by existing integration tests; here we verify invocation, not internals.
- No dual-write requirement simplifies scenarios (no parallel provider assertions).

## Entry / Exit Criteria
- **Entry:** Planning specs approved, ExpenseIntegration interface defined.
- **Exit:** All unit/integration cases implemented & passing; staging validation checklist signed before production cutover.
