# ADR-001: Use Background Worker Queue for Invoice Splits Generation

## Status
Accepted

## Context
When marking a client invoice as paid via `?inv paid` command, we need to generate invoice split records for commission tracking. This involves:
- Querying line items from Notion
- Creating multiple split records in Notion
- Updating the invoice's "Splits Generated" flag

These operations are I/O intensive and could slow down the mark-paid response.

## Decision
Use the existing Worker queue infrastructure (`pkg/worker/worker.go`) to process invoice splits generation in the background.

**Flow:**
1. `processNotionInvoicePaid()` completes status update
2. Enqueue `GenerateInvoiceSplitsMsg` with invoice page ID
3. Return response immediately
4. Worker processes splits generation asynchronously

## Consequences

### Positive
- Fast response time for mark-paid command
- Leverages existing infrastructure
- Decouples split generation from status update
- Failures don't block the main flow

### Negative
- Slight delay between paid and splits created
- Need error handling/retry for failed jobs
- Debugging requires checking worker logs

## Alternatives Considered

1. **Synchronous processing**: Rejected - too slow for user experience
2. **Separate cronjob**: Rejected - unnecessary polling, delayed processing
