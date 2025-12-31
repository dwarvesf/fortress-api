# ADR-001: Follow Existing Payout Processing Pattern

## Status
Accepted

## Context
We need to implement commission payout processing for the `create-contractor-payouts` cronjob. The endpoint already handles `contractor_payroll` and `refund` payout types with a consistent pattern.

## Decision
Follow the existing pattern established by `processContractorPayrollPayouts()` and `processRefundPayouts()`:

1. **Handler Pattern**: Add `processCommissionPayouts()` method following same structure
2. **Service Methods**: Add query and create methods to existing services
3. **Idempotency**: Use relation-based duplicate detection (Invoice Split relation)
4. **Response Format**: Match existing JSON response structure

## Consequences

### Positive
- Consistent codebase, easier maintenance
- Reuse existing service infrastructure
- Familiar pattern for future developers

### Negative
- None significant

## Implementation Notes
- Query method in `InvoiceSplitService`
- Check/Create methods in `ContractorPayoutsService`
- Handler in `contractor_payouts.go`
