# Specification 003: Hourly Rate Detection and Aggregation Logic

**Feature**: Hourly Rate-Based Service Fee Display
**Date**: 2026-01-07
**Status**: Draft

## Overview

This specification defines the algorithmic logic for detecting hourly-rate Service Fees and aggregating them into a single line item. Includes step-by-step algorithms, decision trees, and edge case handling.

## Algorithm 1: Hourly Rate Data Fetching

### Function: fetchHourlyRateData

**Purpose**: Fetch and validate hourly rate and hours data for a Service Fee payout.

**Location**: `pkg/controller/invoice/contractor_invoice.go` (private helper)

**Signature**:
```go
func fetchHourlyRateData(
    ctx context.Context,
    payout notion.PayoutEntry,
    ratesService *notion.ContractorRatesService,
    taskOrderService *notion.TaskOrderLogService,
    l logger.Logger,
) *hourlyRateData
```

**Returns**:
- `*hourlyRateData`: Successfully fetched and validated data
- `nil`: Any condition failed (use default display)

### Algorithm Steps

```
START fetchHourlyRateData

INPUT: payout, ratesService, taskOrderService, logger

STEP 1: Check ServiceRateID present
    IF payout.ServiceRateID == "" THEN
        LOG "[FALLBACK] payout {id}: no ServiceRateID"
        RETURN nil
    END IF

STEP 2: Fetch Contractor Rate
    LOG "[HOURLY_RATE] fetching rate: serviceRateID={id}"
    TRY:
        rateData = ratesService.FetchContractorRateByPageID(ctx, payout.ServiceRateID)
    CATCH error:
        LOG "[FALLBACK] payout {id}: failed to fetch rate - {error}"
        RETURN nil
    END TRY

    LOG "[HOURLY_RATE] fetched rate: billingType={type} hourlyRate={rate} currency={curr}"

STEP 3: Validate BillingType
    IF rateData.BillingType != "Hourly Rate" THEN
        LOG "[INFO] payout {id}: billingType={type} (not hourly)"
        RETURN nil
    END IF

STEP 4: Fetch Task Order Hours (with graceful degradation)
    hours = 0.0

    IF payout.TaskOrderID != "" THEN
        LOG "[HOURLY_RATE] fetching hours: taskOrderID={id}"
        TRY:
            hours = taskOrderService.FetchTaskOrderHoursByPageID(ctx, payout.TaskOrderID)
            LOG "[HOURLY_RATE] fetched hours: {hours}"
        CATCH error:
            LOG "[FALLBACK] payout {id}: failed to fetch hours, using 0"
            hours = 0.0  // Continue with 0 hours
        END TRY
    ELSE:
        LOG "[FALLBACK] payout {id}: no TaskOrderID, using 0 hours"
    END IF

STEP 5: Create hourlyRateData
    RETURN &hourlyRateData{
        HourlyRate:    rateData.HourlyRate,
        Hours:         hours,
        Currency:      rateData.Currency,
        BillingType:   rateData.BillingType,
        ServiceRateID: payout.ServiceRateID,
        TaskOrderID:   payout.TaskOrderID,
    }

END fetchHourlyRateData
```

### Decision Tree

```
                               START
                                 |
                                 v
                     ┌───────────────────────┐
                     │ ServiceRateID         │
                     │ present?              │
                     └───────┬───────────────┘
                             │
                     NO ─────┼───── YES
                     │       │
                     v       v
              ┌──────────┐  Fetch Contractor Rate
              │ RETURN   │       │
              │ nil      │       │
              └──────────┘  ┌────v──────┐
                             │ Fetch     │
                             │ Success?  │
                             └─┬─────────┘
                               │
                       NO ─────┼───── YES
                       │       │
                       v       v
                ┌──────────┐  ┌─────────────────┐
                │ RETURN   │  │ BillingType ==  │
                │ nil      │  │ "Hourly Rate"?  │
                └──────────┘  └─┬───────────────┘
                                │
                        NO ─────┼───── YES
                        │       │
                        v       v
                 ┌──────────┐  Fetch Task Order Hours
                 │ RETURN   │       │
                 │ nil      │       │
                 └──────────┘  ┌────v──────┐
                                │ Fetch     │
                                │ Success?  │
                                └─┬─────────┘
                                  │
                          NO ─────┼───── YES
                          │       │
                          v       v
                      ┌─────────┐ ┌─────────┐
                      │ hours=0 │ │ hours=N │
                      └────┬────┘ └────┬────┘
                           │           │
                           └─────┬─────┘
                                 v
                          ┌──────────────┐
                          │ RETURN       │
                          │ hourlyRateData│
                          └──────────────┘
```

### Implementation

```go
func fetchHourlyRateData(
    ctx context.Context,
    payout notion.PayoutEntry,
    ratesService *notion.ContractorRatesService,
    taskOrderService *notion.TaskOrderLogService,
    l logger.Logger,
) *hourlyRateData {
    // STEP 1: Check ServiceRateID present
    if payout.ServiceRateID == "" {
        l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no ServiceRateID", payout.PageID))
        return nil
    }

    // STEP 2: Fetch Contractor Rate
    l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: serviceRateID=%s", payout.ServiceRateID))
    rateData, err := ratesService.FetchContractorRateByPageID(ctx, payout.ServiceRateID)
    if err != nil {
        l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch rate", payout.PageID))
        return nil
    }

    l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
        rateData.BillingType, rateData.HourlyRate, rateData.Currency))

    // STEP 3: Validate BillingType
    if rateData.BillingType != "Hourly Rate" {
        l.Debug(fmt.Sprintf("[INFO] payout %s: billingType=%s (not hourly)", payout.PageID, rateData.BillingType))
        return nil
    }

    // STEP 4: Fetch Task Order hours (graceful degradation)
    var hours float64
    if payout.TaskOrderID != "" {
        l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching hours: taskOrderID=%s", payout.TaskOrderID))
        hours, err = taskOrderService.FetchTaskOrderHoursByPageID(ctx, payout.TaskOrderID)
        if err != nil {
            l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch hours, using 0", payout.PageID))
            hours = 0
        } else {
            l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: %.2f", hours))
        }
    } else {
        l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no TaskOrderID, using 0 hours", payout.PageID))
        hours = 0
    }

    // STEP 5: Create hourlyRateData
    return &hourlyRateData{
        HourlyRate:    rateData.HourlyRate,
        Hours:         hours,
        Currency:      rateData.Currency,
        BillingType:   rateData.BillingType,
        ServiceRateID: payout.ServiceRateID,
        TaskOrderID:   payout.TaskOrderID,
    }
}
```

---

## Algorithm 2: Hourly Service Fee Aggregation

### Function: aggregateHourlyServiceFees

**Purpose**: Consolidate all hourly-rate Service Fee items into a single line item.

**Location**: `pkg/controller/invoice/contractor_invoice.go` (private helper)

**Signature**:
```go
func aggregateHourlyServiceFees(
    lineItems []ContractorInvoiceLineItem,
    month string,
    l logger.Logger,
) []ContractorInvoiceLineItem
```

**Returns**: Modified lineItems slice with aggregated item replacing hourly items

### Algorithm Steps

```
START aggregateHourlyServiceFees

INPUT: lineItems[], month, logger

STEP 1: Partition line items
    hourlyItems = []
    otherItems = []

    FOR EACH item IN lineItems DO
        IF item.IsHourlyRate == true THEN
            hourlyItems.append(item)
        ELSE
            otherItems.append(item)
        END IF
    END FOR

    LOG "[AGGREGATE] found {count} hourly-rate Service Fee items"

    IF len(hourlyItems) == 0 THEN
        LOG "[AGGREGATE] no hourly items, returning unchanged"
        RETURN lineItems  // No aggregation needed
    END IF

STEP 2: Aggregate hourly items
    agg = hourlyRateAggregation{
        TotalHours:   0,
        HourlyRate:   0,
        TotalAmount:  0,
        Currency:     "",
        Descriptions: [],
        TaskOrderIDs: [],
    }

    FOR EACH item IN hourlyItems DO
        // Sum hours
        agg.TotalHours += item.Hours

        // Sum amounts
        agg.TotalAmount += item.Amount

        // Use first item's rate and currency
        IF agg.HourlyRate == 0 THEN
            agg.HourlyRate = item.Rate
            agg.Currency = item.OriginalCurrency
        ELSE
            // Validate same rate (log warning if different)
            IF item.Rate != agg.HourlyRate THEN
                LOG "[WARN] multiple hourly rates found: {agg.HourlyRate} vs {item.Rate}, using first"
            END IF
            IF item.OriginalCurrency != agg.Currency THEN
                LOG "[WARN] multiple currencies found: {agg.Currency} vs {item.OriginalCurrency}, using first"
            END IF
        END IF

        // Collect descriptions
        IF item.Description != "" THEN
            agg.Descriptions.append(strings.TrimSpace(item.Description))
        END IF

        // Collect task order IDs (for logging)
        IF item.TaskOrderID != "" THEN
            agg.TaskOrderIDs.append(item.TaskOrderID)
        END IF
    END FOR

    LOG "[AGGREGATE] totalHours={hours} rate={rate} totalAmount={amount} currency={curr}"

STEP 3: Generate title
    title = generateServiceFeeTitle(month)

STEP 4: Concatenate descriptions
    description = concatenateDescriptions(agg.Descriptions)

STEP 5: Create aggregated line item
    aggregatedItem = ContractorInvoiceLineItem{
        Title:            title,
        Description:      description,
        Hours:            agg.TotalHours,
        Rate:             agg.HourlyRate,
        Amount:           agg.TotalAmount,
        AmountUSD:        agg.TotalAmount,  // Assume already converted (TODO: verify)
        Type:             "Contractor Payroll",
        OriginalAmount:   agg.TotalAmount,
        OriginalCurrency: agg.Currency,
        IsHourlyRate:     false,  // Not marked (already aggregated)
        ServiceRateID:    "",     // N/A for aggregated item
        TaskOrderID:      "",     // Multiple orders, stored in agg.TaskOrderIDs
    }

    LOG "[AGGREGATE] created aggregated item with title: {title}"
    LOG "[AGGREGATE] aggregated {count} items from task orders: {taskOrderIDs}"

STEP 6: Rebuild line items
    RETURN append(otherItems, aggregatedItem)

END aggregateHourlyServiceFees
```

### Implementation

```go
func aggregateHourlyServiceFees(
    lineItems []ContractorInvoiceLineItem,
    month string,
    l logger.Logger,
) []ContractorInvoiceLineItem {
    // STEP 1: Partition line items
    var hourlyItems []ContractorInvoiceLineItem
    var otherItems []ContractorInvoiceLineItem

    for _, item := range lineItems {
        if item.IsHourlyRate {
            hourlyItems = append(hourlyItems, item)
        } else {
            otherItems = append(otherItems, item)
        }
    }

    l.Debug(fmt.Sprintf("[AGGREGATE] found %d hourly-rate Service Fee items to aggregate", len(hourlyItems)))

    if len(hourlyItems) == 0 {
        l.Debug("[AGGREGATE] no hourly items, returning unchanged")
        return lineItems
    }

    // STEP 2: Aggregate hourly items
    agg := &hourlyRateAggregation{}

    for _, item := range hourlyItems {
        // Sum hours and amounts
        agg.TotalHours += item.Hours
        agg.TotalAmount += item.Amount

        // Use first item's rate and currency
        if agg.HourlyRate == 0 {
            agg.HourlyRate = item.Rate
            agg.Currency = item.OriginalCurrency
        } else {
            // Validate consistency (log warnings)
            if item.Rate != agg.HourlyRate {
                l.Warn(fmt.Sprintf("[WARN] multiple hourly rates found: %.2f vs %.2f, using first",
                    agg.HourlyRate, item.Rate))
            }
            if item.OriginalCurrency != agg.Currency {
                l.Warn(fmt.Sprintf("[WARN] multiple currencies found: %s vs %s, using first",
                    agg.Currency, item.OriginalCurrency))
            }
        }

        // Collect descriptions
        if strings.TrimSpace(item.Description) != "" {
            agg.Descriptions = append(agg.Descriptions, strings.TrimSpace(item.Description))
        }

        // Collect task order IDs (for logging)
        if item.TaskOrderID != "" {
            agg.TaskOrderIDs = append(agg.TaskOrderIDs, item.TaskOrderID)
        }
    }

    l.Debug(fmt.Sprintf("[AGGREGATE] totalHours=%.2f rate=%.2f totalAmount=%.2f currency=%s",
        agg.TotalHours, agg.HourlyRate, agg.TotalAmount, agg.Currency))

    // STEP 3: Generate title
    title := generateServiceFeeTitle(month)

    // STEP 4: Concatenate descriptions
    description := concatenateDescriptions(agg.Descriptions)

    // STEP 5: Create aggregated line item
    aggregatedItem := ContractorInvoiceLineItem{
        Title:            title,
        Description:      description,
        Hours:            agg.TotalHours,
        Rate:             agg.HourlyRate,
        Amount:           agg.TotalAmount,
        AmountUSD:        agg.TotalAmount,  // Already converted
        Type:             "Contractor Payroll",
        OriginalAmount:   agg.TotalAmount,
        OriginalCurrency: agg.Currency,
        IsHourlyRate:     false,  // Already aggregated
        ServiceRateID:    "",
        TaskOrderID:      "",
    }

    l.Debug(fmt.Sprintf("[AGGREGATE] created aggregated item with title: %s", title))
    l.Debug(fmt.Sprintf("[AGGREGATE] aggregated %d items from task orders: %v",
        len(hourlyItems), agg.TaskOrderIDs))

    // STEP 6: Rebuild line items
    return append(otherItems, aggregatedItem)
}
```

---

## Helper Function 1: generateServiceFeeTitle

### Purpose
Generate title with invoice month date range.

### Algorithm

```
START generateServiceFeeTitle

INPUT: month (format: "2006-01")

STEP 1: Parse month
    TRY:
        monthTime = time.Parse("2006-01", month)
    CATCH error:
        RETURN "Service Fee"  // Fallback to generic title
    END TRY

STEP 2: Calculate date range
    startDate = first day of month (YYYY-MM-01)
    endDate = last day of month (YYYY-MM-DD)

STEP 3: Format title
    RETURN "Service Fee (Development work from {startDate} to {endDate})"

END generateServiceFeeTitle
```

### Implementation

```go
func generateServiceFeeTitle(month string) string {
    // STEP 1: Parse month
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return "Service Fee"  // Fallback
    }

    // STEP 2: Calculate date range
    startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)

    // STEP 3: Format title
    return fmt.Sprintf("Service Fee (Development work from %s to %s)",
        startDate.Format("2006-01-02"),
        endDate.Format("2006-01-02"))
}
```

### Examples

```go
// January 2026
generateServiceFeeTitle("2026-01")
// Returns: "Service Fee (Development work from 2026-01-01 to 2026-01-31)"

// February 2024 (leap year)
generateServiceFeeTitle("2024-02")
// Returns: "Service Fee (Development work from 2024-02-01 to 2024-02-29)"

// Invalid format
generateServiceFeeTitle("invalid")
// Returns: "Service Fee"
```

---

## Helper Function 2: concatenateDescriptions

### Purpose
Join descriptions with double line breaks, filtering empty strings.

### Algorithm

```
START concatenateDescriptions

INPUT: descriptions[] (array of strings)

STEP 1: Filter empty strings
    filtered = []
    FOR EACH desc IN descriptions DO
        trimmed = strings.TrimSpace(desc)
        IF trimmed != "" THEN
            filtered.append(trimmed)
        END IF
    END FOR

STEP 2: Join with double line breaks
    RETURN strings.Join(filtered, "\n\n")

END concatenateDescriptions
```

### Implementation

```go
func concatenateDescriptions(descriptions []string) string {
    // STEP 1: Filter empty strings
    var filtered []string
    for _, desc := range descriptions {
        trimmed := strings.TrimSpace(desc)
        if trimmed != "" {
            filtered = append(filtered, trimmed)
        }
    }

    // STEP 2: Join with double line breaks
    return strings.Join(filtered, "\n\n")
}
```

### Examples

```go
// Multiple descriptions
concatenateDescriptions([]string{
    "Work on Project X",
    "Implemented feature Y",
    "Fixed bug Z",
})
// Returns: "Work on Project X\n\nImplemented feature Y\n\nFixed bug Z"

// With empty strings
concatenateDescriptions([]string{
    "Work A",
    "",
    "Work B",
    "   ",  // Whitespace only
    "Work C",
})
// Returns: "Work A\n\nWork B\n\nWork C"

// All empty
concatenateDescriptions([]string{"", "  ", ""})
// Returns: ""

// Single description
concatenateDescriptions([]string{"Single work item"})
// Returns: "Single work item"
```

---

## Edge Cases

### Edge Case 1: No Hourly Items

**Input**:
```go
lineItems := []ContractorInvoiceLineItem{
    {Type: "Commission", Amount: 100, IsHourlyRate: false},
    {Type: "Refund", Amount: 50, IsHourlyRate: false},
}
```

**Processing**:
- `aggregateHourlyServiceFees` partitions: 0 hourly, 2 other
- Returns early with unchanged lineItems

**Output**: Same as input (no aggregation)

### Edge Case 2: Single Hourly Item

**Input**:
```go
lineItems := []ContractorInvoiceLineItem{
    {Type: "Contractor Payroll", Hours: 10, Rate: 50, Amount: 500, IsHourlyRate: true},
}
```

**Processing**:
- Aggregates into single item (for consistent title)
- Title: "Service Fee (Development work from 2026-01-01 to 2026-01-31)"

**Output**:
```go
[]ContractorInvoiceLineItem{
    {
        Title: "Service Fee (Development work from 2026-01-01 to 2026-01-31)",
        Hours: 10,
        Rate: 50,
        Amount: 500,
        IsHourlyRate: false,  // Aggregation complete
    },
}
```

### Edge Case 3: Multiple Hourly Rates (Shouldn't Happen)

**Input**:
```go
lineItems := []ContractorInvoiceLineItem{
    {Hours: 10, Rate: 50, Amount: 500, IsHourlyRate: true},   // $50/hr
    {Hours: 5, Rate: 60, Amount: 300, IsHourlyRate: true},    // $60/hr (unexpected)
}
```

**Processing**:
- Log warning: "multiple hourly rates found: 50.00 vs 60.00, using first"
- Use first rate (50)
- Sum hours: 10 + 5 = 15
- Sum amounts: 500 + 300 = 800

**Output**:
```go
ContractorInvoiceLineItem{
    Hours: 15,
    Rate: 50,      // First rate used
    Amount: 800,   // Sum of actual amounts (not recalculated)
}
```

### Edge Case 4: Zero Hours

**Input**:
```go
lineItems := []ContractorInvoiceLineItem{
    {Hours: 0, Rate: 50, Amount: 500, IsHourlyRate: true},  // Hours fetch failed
}
```

**Processing**:
- Aggregates normally
- TotalHours = 0

**Output**:
```go
ContractorInvoiceLineItem{
    Title: "Service Fee (Development work from ...)",
    Hours: 0,      // Displays "0" hours
    Rate: 50,
    Amount: 500,   // Still shows actual amount
}
```

**Invoice Display**: "0 hours @ $50/hr = $500" (amount from payout)

### Edge Case 5: Empty Descriptions

**Input**:
```go
lineItems := []ContractorInvoiceLineItem{
    {Description: "", IsHourlyRate: true},
    {Description: "  ", IsHourlyRate: true},  // Whitespace only
    {Description: "Actual work", IsHourlyRate: true},
}
```

**Processing**:
- `concatenateDescriptions` filters empty/whitespace
- Only "Actual work" included

**Output**:
```go
ContractorInvoiceLineItem{
    Description: "Actual work",  // Single description
}
```

---

## Success Criteria

1. All hourly Service Fees correctly identified
2. Aggregation produces single line item with correct totals
3. Title includes correct date range from invoice month
4. Descriptions concatenated with proper formatting
5. Edge cases handled gracefully
6. Warnings logged for unexpected conditions
7. No errors for valid inputs
8. Unit tests cover all scenarios

## Related Documents

- ADR-002: Aggregation Approach
- ADR-003: Error Handling Strategy
- Specification: spec-001-data-structures.md
- Specification: spec-002-service-methods.md
- Specification: spec-004-integration.md
