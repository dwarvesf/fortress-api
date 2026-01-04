# Specification: payout_types.go Changes

## File
`pkg/service/notion/payout_types.go`

## Changes Required

### 1. Update PayoutSourceType Constant

**Current:**
```go
PayoutSourceTypeContractorPayroll PayoutSourceType = "Contractor Payroll"
```

**New:**
```go
PayoutSourceTypeServiceFee PayoutSourceType = "Service Fee"
```

### 2. Deprecate PayoutDirection (Optional)

The `PayoutDirection` type and constants are no longer used since `Direction` was removed from the schema. Options:
- Keep for internal logic if needed
- Remove if not used elsewhere

**Recommendation:** Keep for now as it's used in `PayoutLineItem` struct for internal calculations.

### 3. Update PayoutLineItem struct

No changes needed - the struct is for internal use and can keep Direction for signed amount calculations.
