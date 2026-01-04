# ADR-001: Use Task Order Log as Data Source for Service Fee Payouts

## Status
Accepted

## Context
The current `processContractorPayrollPayouts` handler queries Contractor Fees table to create Service Fee payouts. The Contractor Fees table acts as an intermediary that aggregates data from Task Order Log.

With the updated Notion schema, the Contractor Payouts table now has a direct "00 Task Order" relation to Task Order Log. This allows bypassing the Contractor Fees intermediary.

## Decision
Refactor `processContractorPayrollPayouts` to:
1. Query Task Order Log directly (Type=Order, Status=Approved)
2. Fetch contractor rates from Contractor Rates table
3. Calculate payout amount based on billing type
4. Create payout with direct Task Order Log reference

## Consequences

### Positive
- Simpler data flow (Task Order Log → Payout)
- Removes dependency on Contractor Fees table for payout creation
- Direct relation between Task Order Log and Payout enables better traceability

### Negative
- Requires new method to query Contractor Rates by contractor page ID
- Requires new method to update Task Order Log status
- Amount calculation logic moves from Notion formulas to Go code

### Risks
- Contractor Rates query must handle cases where no rate is found
- Must ensure month extraction from Date is consistent (YYYY-MM format)
