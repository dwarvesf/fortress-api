# Specification 004: Integration with GenerateContractorInvoice

**Feature**: Hourly Rate-Based Service Fee Display
**Date**: 2026-01-07
**Status**: Draft

## Overview

This specification defines how hourly rate functionality integrates into the existing `GenerateContractorInvoice` method in `pkg/controller/invoice/contractor_invoice.go`. Includes exact integration points, code modifications, and sequencing.

## Current Flow (Before Changes)

```
GenerateContractorInvoice
├─ 1. Query Contractor Rates (by discord + month)
├─ 2. Query Pending Payouts + Bank Account (parallel)
├─ 3. Convert payout amounts to USD (parallel)
├─ 4. Build line items from payouts (sequential)
│   ├─ For each payout:
│   │   ├─ Determine description (from WorkDetails/Description)
│   │   └─ Create ContractorInvoiceLineItem
│   │       └─ Hours=1, Rate=AmountUSD, Amount=AmountUSD
│   └─ Result: lineItems[] with all payouts as individual items
├─ 4.5 Group Commission items by Project
├─ 5. Sort line items (Service Fee last, Amount ASC)
├─ 5.5-5.10 Calculate multi-currency subtotals
└─ 6. Build and return ContractorInvoiceData
```

## New Flow (With Hourly Rate Support)

```
GenerateContractorInvoice
├─ 1. Query Contractor Rates (by discord + month)
├─ 2. Query Pending Payouts + Bank Account (parallel)
├─ 3. Convert payout amounts to USD (parallel)
├─ 4. Build line items from payouts (sequential) [MODIFIED]
│   ├─ Create service instances [NEW]
│   │   ├─ ratesService = NewContractorRatesService()
│   │   └─ taskOrderService = NewTaskOrderLogService()
│   │
│   ├─ For each payout:
│   │   ├─ Determine description (from WorkDetails/Description)
│   │   │
│   │   ├─ IF SourceType == ServiceFee THEN [NEW]
│   │   │   ├─ hourlyData = fetchHourlyRateData(payout, services)
│   │   │   │
│   │   │   ├─ IF hourlyData != nil THEN
│   │   │   │   └─ Create hourly line item:
│   │   │   │       ├─ Hours = hourlyData.Hours
│   │   │   │       ├─ Rate = hourlyData.HourlyRate
│   │   │   │       ├─ Amount = payout.Amount (original)
│   │   │   │       └─ IsHourlyRate = true [NEW FIELD]
│   │   │   │
│   │   │   └─ ELSE (fallback)
│   │   │       └─ Create default line item (Hours=1)
│   │   │
│   │   └─ ELSE (non-Service Fee)
│   │       └─ Create default line item (Hours=1)
│   │
│   └─ Result: lineItems[] with mixed hourly/non-hourly items
│
├─ 4.5 Aggregate hourly Service Fees [NEW]
│   ├─ lineItems = aggregateHourlyServiceFees(lineItems, month)
│   └─ Result: Hourly items replaced with single aggregated item
│
├─ 4.6 Group Commission items by Project (renamed from 4.5)
├─ 5. Sort line items (Service Fee last, Amount ASC)
├─ 5.5-5.10 Calculate multi-currency subtotals
└─ 6. Build and return ContractorInvoiceData
```

## Integration Point 1: Service Instantiation

### Location
After line ~200 (after parallel conversions complete)

### Code Addition

```go
// After "l.Debug('[DEBUG] contractor_invoice: parallel currency conversions completed')"

// NEW: Create service instances for hourly rate processing
l.Debug("[DEBUG] contractor_invoice: creating services for hourly rate processing")
ratesService := notion.NewContractorRatesService(c.config, c.logger)
taskOrderService := notion.NewTaskOrderLogService(c.config, c.logger)

if ratesService == nil || taskOrderService == nil {
    l.Error(nil, "failed to create services for hourly rate processing")
    // Continue without hourly rate support (graceful degradation)
    ratesService = nil
    taskOrderService = nil
}
```

### Rationale
- Services created once, reused for all payouts
- Nil check allows graceful degradation if service creation fails
- Logged for debugging

## Integration Point 2: Line Item Building

### Location
Lines ~205-284 (payout processing loop)

### Current Code (Simplified)

```go
// 4. Build line items from payouts
var lineItems []ContractorInvoiceLineItem
var total float64

for i, payout := range payouts {
    amountUSD := amountsUSD[i]

    l.Debug(fmt.Sprintf("processing payout: pageID=%s name=%s sourceType=%s...",
        payout.PageID, payout.Name, payout.SourceType))

    // Determine description based on source type
    var description string
    switch payout.SourceType {
    case notion.PayoutSourceTypeServiceFee:
        // ... fetch description logic ...
    case notion.PayoutSourceTypeRefund:
        // ... refund logic ...
    default:
        description = payout.Description
    }

    // Initialize line item with default values
    title := ""
    lineItem := ContractorInvoiceLineItem{
        Title:             title,
        Description:       description,
        Hours:             1,         // Default quantity
        Rate:              amountUSD, // Unit cost = converted amount
        Amount:            amountUSD,
        AmountUSD:         amountUSD,
        Type:              string(payout.SourceType),
        CommissionRole:    payout.CommissionRole,
        CommissionProject: payout.CommissionProject,
        OriginalAmount:   payout.Amount,
        OriginalCurrency: payout.Currency,
    }

    lineItems = append(lineItems, lineItem)
    total += amountUSD
}
```

### Modified Code

```go
// 4. Build line items from payouts
var lineItems []ContractorInvoiceLineItem
var total float64

for i, payout := range payouts {
    amountUSD := amountsUSD[i]

    l.Debug(fmt.Sprintf("processing payout: pageID=%s name=%s sourceType=%s amount=%.2f currency=%s",
        payout.PageID, payout.Name, payout.SourceType, payout.Amount, payout.Currency))

    // Determine description based on source type (EXISTING LOGIC UNCHANGED)
    var description string
    switch payout.SourceType {
    case notion.PayoutSourceTypeServiceFee:
        // ... existing description fetching logic ...
    case notion.PayoutSourceTypeRefund:
        // ... existing refund logic ...
    default:
        description = payout.Description
    }

    // NEW: For Service Fees, attempt hourly rate display
    var lineItem ContractorInvoiceLineItem
    if payout.SourceType == notion.PayoutSourceTypeServiceFee {
        // Attempt to fetch hourly rate data
        var hourlyData *hourlyRateData
        if ratesService != nil && taskOrderService != nil {
            hourlyData = fetchHourlyRateData(ctx, payout, ratesService, taskOrderService, l)
        }

        if hourlyData != nil {
            // SUCCESS: Use hourly rate display
            l.Debug(fmt.Sprintf("[SUCCESS] payout %s: applying hourly rate display (hours=%.2f rate=%.2f)",
                payout.PageID, hourlyData.Hours, hourlyData.HourlyRate))

            lineItem = ContractorInvoiceLineItem{
                Title:            "",  // Set during aggregation
                Description:      description,
                Hours:            hourlyData.Hours,
                Rate:             hourlyData.HourlyRate,
                Amount:           payout.Amount,  // Use original payout amount
                AmountUSD:        amountUSD,
                Type:             string(payout.SourceType),
                CommissionRole:    payout.CommissionRole,
                CommissionProject: payout.CommissionProject,
                OriginalAmount:   payout.Amount,
                OriginalCurrency: payout.Currency,
                // NEW FIELDS
                IsHourlyRate:     true,
                ServiceRateID:    payout.ServiceRateID,
                TaskOrderID:      payout.TaskOrderID,
            }
        } else {
            // FALLBACK: Use default display
            l.Debug(fmt.Sprintf("[FALLBACK] payout %s: using default display (Qty=1, Unit Cost=%.2f)",
                payout.PageID, amountUSD))

            lineItem = createDefaultLineItem(payout, amountUSD, description)
        }
    } else {
        // Non-Service Fee: use default display
        lineItem = createDefaultLineItem(payout, amountUSD, description)
    }

    lineItems = append(lineItems, lineItem)
    total += amountUSD
}

l.Debug(fmt.Sprintf("built %d line items with total=%.2f USD", len(lineItems), total))
```

### New Helper Function: createDefaultLineItem

```go
// createDefaultLineItem creates line item with default display (Qty=1, Unit Cost=Amount).
// Used as fallback when hourly rate data unavailable or for non-Service Fee payouts.
func createDefaultLineItem(
    payout notion.PayoutEntry,
    amountUSD float64,
    description string,
) ContractorInvoiceLineItem {
    return ContractorInvoiceLineItem{
        Title:             "",
        Description:       description,
        Hours:             1,         // Default quantity
        Rate:              amountUSD, // Unit cost = converted amount
        Amount:            amountUSD,
        AmountUSD:         amountUSD,
        Type:              string(payout.SourceType),
        CommissionRole:    payout.CommissionRole,
        CommissionProject: payout.CommissionProject,
        OriginalAmount:    payout.Amount,
        OriginalCurrency:  payout.Currency,
        // NEW FIELDS (explicitly set to default)
        IsHourlyRate:      false,
        ServiceRateID:     payout.ServiceRateID,  // Preserved for logging
        TaskOrderID:       payout.TaskOrderID,     // Preserved for logging
    }
}
```

## Integration Point 3: Aggregation

### Location
After line ~286 (after line items built, before commission grouping)

### Code Addition

```go
l.Debug(fmt.Sprintf("built %d line items with total=%.2f USD", len(lineItems), total))

// NEW: 4.5 Aggregate hourly-rate Service Fees
l.Debug("aggregating hourly-rate Service Fee items")
lineItems = aggregateHourlyServiceFees(lineItems, month, l)
l.Debug(fmt.Sprintf("after aggregation: %d line items", len(lineItems)))

// EXISTING: 4.6 Group Commission items by Project (renamed from 4.5)
l.Debug("grouping commission items by project")
// ... rest of existing commission grouping logic ...
```

### Sequencing

```
Line Items Flow:

BEFORE Aggregation:
  [0] Commission - $100 (IsHourlyRate=false)
  [1] Service Fee - 10 hrs @ $50 = $500 (IsHourlyRate=true)
  [2] Service Fee - 5 hrs @ $50 = $250 (IsHourlyRate=true)
  [3] Refund - $50 (IsHourlyRate=false)

AFTER Aggregation (aggregateHourlyServiceFees):
  [0] Commission - $100
  [1] Refund - $50
  [2] Service Fee (Development work from 2026-01-01 to 2026-01-31) - 15 hrs @ $50 = $750

AFTER Commission Grouping:
  (No change - Commission not grouped by project in this example)

AFTER Sorting:
  [0] Commission - $50
  [1] Refund - $100
  [2] Service Fee (Development work from ...) - $750 (Service Fee last)
```

## Modified Function Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ GenerateContractorInvoice                                   │
└───────────────────┬─────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Query Contractor Rates (discord + month)                     │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Query Payouts + Bank Account (parallel)                      │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Convert amounts to USD (parallel)                            │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ [NEW] Create service instances                               │
│ - ContractorRatesService                                     │
│ - TaskOrderLogService                                        │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Build Line Items (for each payout)                           │
│                                                               │
│   ┌───────────────────────────────────────────────────┐     │
│   │ Determine description                              │     │
│   └───────────────┬───────────────────────────────────┘     │
│                   │                                           │
│                   v                                           │
│   ┌───────────────────────────────────────────────────┐     │
│   │ IF SourceType == ServiceFee?                      │     │
│   └───────┬───────────────────────────────────────────┘     │
│           │                                                   │
│     NO ───┼─── YES                                           │
│     │     │                                                   │
│     │     v                                                   │
│     │   ┌───────────────────────────────────────┐           │
│     │   │ [NEW] fetchHourlyRateData()           │           │
│     │   └───────┬───────────────────────────────┘           │
│     │           │                                             │
│     │     nil ──┼─── hourlyData                              │
│     │     │     │                                             │
│     │     │     v                                             │
│     │     │   ┌─────────────────────────────────┐           │
│     │     │   │ Create hourly line item         │           │
│     │     │   │ IsHourlyRate = true             │           │
│     │     │   └───────┬─────────────────────────┘           │
│     │     │           │                                       │
│     │     v           │                                       │
│     │   ┌─────────────┴─────────────┐                       │
│     │   │ Create default line item   │                       │
│     │   │ IsHourlyRate = false       │                       │
│     │   └───────┬───────────────────┘                       │
│     │           │                                             │
│     v           v                                             │
│   ┌─────────────┴───────────────────────────────┐           │
│   │ Append to lineItems[]                        │           │
│   └──────────────────────────────────────────────┘           │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ [NEW] Aggregate Hourly Service Fees                          │
│ - Partition: hourly vs non-hourly items                      │
│ - Sum hours, amounts, concatenate descriptions               │
│ - Create single aggregated item with title                   │
│ - Replace hourly items with aggregated item                  │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Group Commission items by Project                            │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Sort line items (Service Fee last, Amount ASC)               │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Calculate multi-currency subtotals                           │
└───────────────────┬───────────────────────────────────────────┘
                    │
                    v
┌───────────────────────────────────────────────────────────────┐
│ Build and return ContractorInvoiceData                       │
└───────────────────────────────────────────────────────────────┘
```

## Backward Compatibility

### Scenario 1: Service Created Before ServiceRateID Field Added

**Situation**: Existing payouts don't have ServiceRateID populated.

**Behavior**:
- `payout.ServiceRateID == ""` (empty string)
- `fetchHourlyRateData` returns nil immediately
- Falls back to default display (Qty=1)
- Invoice generates successfully (no breaking change)

**Log Output**:
```
[FALLBACK] payout payout-123: no ServiceRateID
```

### Scenario 2: No Hourly Rate Contractors

**Situation**: All contractors use "Monthly Fixed" billing.

**Behavior**:
- `fetchHourlyRateData` fetches rate successfully
- BillingType = "Monthly Fixed" (not "Hourly Rate")
- Returns nil
- Falls back to default display
- Invoice generates successfully

**Log Output**:
```
[INFO] payout payout-123: billingType=Monthly Fixed (not hourly)
```

### Scenario 3: Notion API Failures

**Situation**: Notion API is down or rate/hours fetch fails.

**Behavior**:
- Error logged with [FALLBACK] prefix
- Returns nil
- Falls back to default display
- Invoice generates successfully (no breaking change)

**Log Output**:
```
[FALLBACK] payout payout-123: failed to fetch rate
```

## Testing Strategy

### Unit Tests

**Test 1: Integration with Existing Flow**
```go
func TestGenerateContractorInvoice_HourlyRate_Integration(t *testing.T) {
    // Mock services return hourly rate data
    // Verify:
    // - Line items created with hourly display
    // - Aggregation produces single item
    // - Existing flow (commission, sorting) unchanged
}
```

**Test 2: Fallback to Default Display**
```go
func TestGenerateContractorInvoice_HourlyRate_Fallback(t *testing.T) {
    // Mock rate fetch failure
    // Verify:
    // - Invoice generates successfully
    // - Line items use default display (Qty=1)
    // - Logs show [FALLBACK] messages
}
```

**Test 3: Mixed Invoice**
```go
func TestGenerateContractorInvoice_Mixed(t *testing.T) {
    // Payouts: 1 hourly Service Fee, 1 Commission, 1 Refund
    // Verify:
    // - Hourly Service Fee aggregated
    // - Commission and Refund unchanged
    // - Correct sorting order
}
```

### Integration Tests

**Test 1: End-to-End with Real Notion Data**
- Use test Notion databases with known data
- Verify complete invoice generation with hourly rates
- Compare output to golden file

**Test 2: Multi-Currency Hourly Rates**
- Test with USD and VND hourly Service Fees
- Verify correct currency handling and display

## Performance Impact

### Additional Overhead

**Per Invoice**:
- Service instantiation: +2 objects (negligible memory)
- Per Service Fee payout:
  - +1-2 Notion API calls (rate + hours)
  - +~50ms latency per API call
- Aggregation: O(n) where n = number of line items (typically < 10)
  - Negligible CPU time

**Total Impact**:
- 1 Service Fee: +2 API calls, +~100-200ms
- 3 Service Fees: +6 API calls, +~300-600ms
- Within acceptable range per NFR-3

### Optimization Opportunities (Future)

1. **Batch rate fetching**: If same ServiceRateID used multiple times, cache result
2. **Parallel API calls**: Fetch rate and hours in parallel (minimal benefit for 2 calls)
3. **Rate caching**: Cache contractor rates for invoice month (requires cache invalidation)

Note: Current sequential approach is acceptable for low volume (1-3 Service Fees per invoice).

## Rollback Plan

### Quick Rollback (Without Code Changes)

Not possible - feature requires code deployment.

### Code Rollback

**To disable feature**:
1. Comment out aggregation call:
   ```go
   // lineItems = aggregateHourlyServiceFees(lineItems, month, l)
   ```
2. All invoices revert to default display (Qty=1)
3. No breaking changes (graceful degradation)

**To fully remove**:
1. Remove service instantiation code
2. Remove hourly rate detection in line item building
3. Remove aggregation call
4. Remove helper functions
5. Remove new fields from ContractorInvoiceLineItem (breaking change)

## Success Criteria

1. Invoice generation succeeds for all payout types
2. Hourly Service Fees display with correct hours and rate
3. Fallback to default display on any error
4. No regression in existing invoice generation
5. Performance impact < 500ms per invoice
6. All logs use consistent prefixes
7. Unit tests cover all integration points
8. Integration tests verify end-to-end flow

## Related Documents

- ADR-001: Data Fetching Strategy
- ADR-002: Aggregation Approach
- ADR-003: Error Handling Strategy
- ADR-004: Code Organization
- Specification: spec-001-data-structures.md
- Specification: spec-002-service-methods.md
- Specification: spec-003-detection-logic.md
