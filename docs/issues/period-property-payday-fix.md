# Issue: Period Property Not Using Payday

**Status**: Identified
**Date**: 2026-01-16
**Severity**: Medium
**Area**: Invoice Generation, Contractor Payables

## Problem Description

The "Period" property in Contractor Payables is currently hardcoded to the 1st of the billing month, but it should use the Payday value from the contractor's Service Rate.

### Current Behavior

When generating a contractor invoice, the Period is set to:
```
2025-12-01  (for December 2025 invoice, regardless of Payday)
```

**Location**: `pkg/handler/invoice/invoice.go:451`
```go
Period: invoiceData.Month + "-01",  // Always hardcoded to 1st
```

### Expected Behavior

The Period should use the contractor's Payday:
```
2025-12-15  (for December 2025 invoice with Payday=15)
2025-12-01  (for December 2025 invoice with Payday=1)
```

## Root Cause

The Period calculation hardcodes "-01" suffix instead of using the `PayDay` field from the contractor's Service Rate.

### Payday Data Source

- **Database**: Contractor Service Rate
- **Property**: "00 Payday" (rollup from Service Rate)
- **Values**: Integer 1-31 (commonly 1 or 15)
- **Already Available**: `rateData.PayDay` is fetched at `pkg/controller/invoice/contractor_invoice.go:133`

## Solution

### Approach

Set Period to the billing month with Payday as the day:
- **Example**: December 2025 invoice (month=2025-12) with Payday=15 → Period=`2025-12-15`

### Implementation Steps

#### Step 1: Add PayDay to ContractorInvoiceData

**File**: `pkg/controller/invoice/contractor_invoice.go`

Add field to struct (around line 26-57):
```go
PayDay int // Pay day of month (1-31) from contractor's Service Rate
```

Populate in struct initialization (around line 596-626):
```go
PayDay: rateData.PayDay,
```

#### Step 2: Update Period Calculation

**File**: `pkg/handler/invoice/invoice.go`

Replace hardcoded calculation (line 451):
```go
// Before
Period: invoiceData.Month + "-01",

// After
Period: calculatePeriodDate(invoiceData.Month, invoiceData.PayDay),
```

#### Step 3: Add Helper Function

**File**: `pkg/handler/invoice/invoice.go`

Add helper function:
```go
// calculatePeriodDate calculates the Period date as YYYY-MM-DD
// using the billing month and Payday from contractor's Service Rate
func calculatePeriodDate(month string, payDay int) string {
    // Validate payDay (default to 1 if invalid)
    if payDay <= 0 || payDay > 31 {
        payDay = 1
    }

    // Parse month (format: YYYY-MM)
    monthTime, err := time.Parse("2006-01", month)
    if err != nil {
        // Fallback to month-01 if parse fails
        return month + "-01"
    }

    // Handle month boundaries (e.g., February with payDay=31)
    periodDate := time.Date(monthTime.Year(), monthTime.Month(), payDay, 0, 0, 0, 0, time.UTC)

    // If the day overflows (e.g., Feb 31 becomes Mar 3), use last day of billing month
    if periodDate.Month() != monthTime.Month() {
        // Get last day of billing month
        nextMonth := monthTime.AddDate(0, 1, 0)
        lastDay := nextMonth.AddDate(0, 0, -1)
        return lastDay.Format("2006-01-02")
    }

    return periodDate.Format("2006-01-02")
}
```

## Edge Cases Handled

1. **Invalid PayDay** (≤0 or >31): Default to 1
2. **Month boundary overflow** (e.g., February 31): Use last day of month (Feb 28/29)
3. **Parse errors**: Fallback to YYYY-MM-01 format

## Testing

### Test Cases

1. **Normal case - Payday=15**
   - Input: month="2025-12", payDay=15
   - Expected: Period="2025-12-15"

2. **Month boundary - February with Payday=31**
   - Input: month="2025-02", payDay=31
   - Expected: Period="2025-02-28"

3. **Invalid Payday - Payday=0**
   - Input: month="2025-12", payDay=0
   - Expected: Period="2025-12-01"

4. **First of month - Payday=1**
   - Input: month="2025-12", payDay=1
   - Expected: Period="2025-12-01"

### Verification Steps

1. Generate contractor invoice for December 2025 with Payday=15
   - Verify Period in Notion Contractor Payables shows `2025-12-15`

2. Query Contractor Payables by Period
   - Verify filtering by Period still works correctly

3. Test edge cases in different months
   - January (31 days), February (28/29 days), April (30 days)

## Files Modified

1. `pkg/controller/invoice/contractor_invoice.go`
   - Add `PayDay int` field to ContractorInvoiceData struct
   - Populate PayDay in struct initialization

2. `pkg/handler/invoice/invoice.go`
   - Replace hardcoded Period calculation
   - Add `calculatePeriodDate()` helper function

## Related Documentation

- Notion Schema: `docs/specs/notion/schema/contractor-payables.md`
- Contractor Payouts Schema: `docs/specs/notion/schema/payouts.md`
- Service Rate Structure: `pkg/service/notion/contractor_rates.go`

## Impact Assessment

### Low Risk
- PayDay data is already being fetched and validated
- Change is isolated to Period calculation
- Backward compatible (existing queries still work)
- No database migration required (Notion property already supports any date)

### Benefits
- Accurate Period dates matching contractor payment schedules
- Better alignment with actual invoice issue dates
- Improved data consistency

## Status Log

- **2026-01-16**: Issue identified and documented
- **2026-01-16**: Solution designed and approved
- **Pending**: Implementation
- **Pending**: Testing and verification
