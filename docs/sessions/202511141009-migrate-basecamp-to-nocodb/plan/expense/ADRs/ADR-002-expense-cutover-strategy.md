# ADR-002: Expense Cutover Strategy (No Dual-Write)

## Context
- Unlike invoice/accounting flows, the user explicitly requested a direct switch from Basecamp to NocoDB for expenses without a dual-write validation window.
- Expense events involve money movement and accounting reconciliation, so accuracy is critical even without dual-write.

## Decision
- Perform targeted end-to-end validation in a non-production/staging environment before enabling NocoDB in production.
- In production, gate the switch behind a runtime config (using `TASK_PROVIDER` + optional expense override). When the flag toggles to `nocodb`, Basecamp handlers stop processing new events immediately; only historical replay tooling remains for rollback.
- Maintain observability guardrails (metrics + structured logs) and a rollback checklist to revert `TASK_PROVIDER` to `basecamp` quickly if anomalies are detected.

## Consequences
- + Simplifies implementation work (no shadow writes or reconciliation logs).
- + Reduces Basecamp API usage as soon as NocoDB is ready.
- − Higher risk during cutover; depends heavily on pre-production validation quality.
- − Requires robust monitoring and alerting to catch issues immediately after the flag flip.
