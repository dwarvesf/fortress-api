# Planning Phase Status

## Status: COMPLETED

## Date: 2026-01-04

## Summary

Planning phase for refactoring payout functions to align with updated Notion schema is complete.

## Documents Created

### ADRs
- [001-schema-alignment.md](./ADRs/001-schema-alignment.md) - Decision to align code with new schema

### Specifications
- [spec-payout-types.md](./specifications/spec-payout-types.md) - Changes to payout_types.go
- [spec-contractor-payouts-service.md](./specifications/spec-contractor-payouts-service.md) - Changes to contractor_payouts.go
- [spec-handler.md](./specifications/spec-handler.md) - Changes to handler

## Key Changes

| Category | Change |
|----------|--------|
| Constant | `PayoutSourceTypeContractorPayroll` → `PayoutSourceTypeServiceFee` |
| Struct | Remove `Direction` from `PayoutEntry` |
| Struct | Rename `ContractorFeesID` → `TaskOrderID` |
| Relation | `"Billing"` → `"00 Task Order"` |
| Relation | `"Refund"` → `"01 Refund"` |
| Relation | `"Invoice Split"` → `"02 Invoice Split"` |
| Remove | Property writes for `Type`, `Month`, `Direction` |
| Add | `Description` field to create inputs |

## Files to Modify
1. `pkg/service/notion/payout_types.go`
2. `pkg/service/notion/contractor_payouts.go`
3. `pkg/handler/notion/contractor_payouts.go`

## Next Phase
Proceed to implementation task breakdown (Phase 3)
