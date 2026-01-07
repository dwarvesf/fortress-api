# ADR-002: Aggregation Approach for Hourly Rate Service Fees

**Status:** Proposed
**Date:** 2026-01-07
**Context:** Hourly Rate-Based Service Fee Display Implementation

## Context

Per requirements (FR-4), all hourly-rate Service Fee items for the same invoice must be aggregated into a single line item with:
- Title: "Service Fee (Development work from YYYY-MM-01 to YYYY-MM-DD)"
- Quantity: Sum of all Final Hours Worked
- Unit Cost: Hourly Rate (assumes all use same rate)
- Amount: Sum of all payout amounts
- Description: Concatenated proof of works

The challenge is deciding where and how to perform this aggregation while maintaining:
- Code clarity and maintainability
- Testability
- Backward compatibility with existing invoice generation flow

## Decision

**We will use post-processing aggregation after line item creation.**

### Approach

1. **Phase 1: Build All Line Items** (existing flow maintained):
   - Process each payout sequentially
   - Create ContractorInvoiceLineItem for each payout
   - Mark hourly-rate items with metadata (IsHourlyRate flag)
   - Store all items in lineItems slice

2. **Phase 2: Identify Hourly Service Fees** (new):
   - Scan lineItems for items marked as hourly-rate Service Fees
   - Collect hourly items into separate slice
   - Keep non-hourly items unchanged

3. **Phase 3: Aggregate Hourly Items** (new):
   - Sum all hours from hourly items
   - Sum all amounts from hourly items
   - Concatenate descriptions with line breaks
   - Create single aggregated line item with consolidated data
   - Replace multiple hourly items with single aggregated item

4. **Phase 4: Continue Existing Flow**:
   - Proceed with commission grouping
   - Apply sorting (Service Fee last, amount ASC)
   - Multi-currency calculations
   - Invoice generation

### Data Flow

```
┌──────────────────────────────────────────────────────────────┐
│ Existing: Build Line Items (per payout)                      │
│                                                               │
│ For each payout:                                              │
│   - Fetch data (rate, hours if hourly)                       │
│   - Create ContractorInvoiceLineItem                          │
│   - Set Hours, Rate, Amount                                   │
│   - Mark IsHourlyRate = true if applicable                    │
│   - Append to lineItems[]                                     │
│                                                               │
│ Result: lineItems[] with mixed types                          │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         v
┌──────────────────────────────────────────────────────────────┐
│ NEW: Aggregate Hourly Service Fees                           │
│                                                               │
│ Step 1: Partition line items                                 │
│   hourlyItems := []LineItem with IsHourlyRate=true           │
│   otherItems  := []LineItem with IsHourlyRate=false          │
│                                                               │
│ Step 2: Aggregate hourly items                               │
│   totalHours   := sum(hourlyItems[].Hours)                   │
│   totalAmount  := sum(hourlyItems[].Amount)                  │
│   hourlyRate   := hourlyItems[0].Rate (same for all)         │
│   description  := join(hourlyItems[].Description, "\n\n")    │
│   title        := "Service Fee (Development work from ...)"  │
│                                                               │
│ Step 3: Create aggregated item                               │
│   aggregatedItem := ContractorInvoiceLineItem{               │
│       Title:       title,                                     │
│       Description: description,                               │
│       Hours:       totalHours,                                │
│       Rate:        hourlyRate,                                │
│       Amount:      totalAmount,                               │
│       Type:        "Contractor Payroll",                      │
│   }                                                           │
│                                                               │
│ Step 4: Rebuild line items                                   │
│   lineItems = append(otherItems, aggregatedItem)             │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         v
┌──────────────────────────────────────────────────────────────┐
│ Existing: Commission Grouping, Sorting, Calculations         │
└──────────────────────────────────────────────────────────────┘
```

## Rationale

### Why Post-Processing?

1. **Minimal Disruption to Existing Flow**:
   - Line item building logic unchanged
   - Each payout still creates its own line item initially
   - Aggregation happens as separate, isolated step
   - Easy to disable/remove without affecting other code

2. **Clear Separation of Concerns**:
   - Build phase: Create individual line items with all data
   - Aggregation phase: Consolidate hourly items
   - Display phase: Format and render
   - Each phase testable independently

3. **Easier Error Handling**:
   - If aggregation fails, can fallback to showing individual items
   - Errors in one item don't affect others during building
   - Aggregation errors are isolated and logged

4. **Better Debuggability**:
   - Can log all individual items before aggregation
   - Can see what gets aggregated and why
   - Clear audit trail in logs

### Why Not Pre-Aggregation?

Pre-aggregation (during payout processing) would require:
- Complex state management to track partial aggregation
- Decision logic mixed with data fetching
- Harder to handle errors (what if 3rd item fails?)
- Difficult to maintain individual item context

### Data Structure Design

```go
// Add to ContractorInvoiceLineItem struct
type ContractorInvoiceLineItem struct {
    // ... existing fields ...

    // NEW: Metadata for aggregation
    IsHourlyRate      bool    // Mark as hourly-rate Service Fee
    ServiceRateID     string  // For logging/debugging
    TaskOrderID       string  // For logging/debugging
}

// Helper struct for aggregation process
type hourlyRateAggregation struct {
    TotalHours       float64
    HourlyRate       float64
    TotalAmount      float64
    Currency         string
    Descriptions     []string
    TaskOrderIDs     []string // For logging
}
```

## Alternatives Considered

### Alternative 1: Real-Time Aggregation

**Approach**: Maintain running aggregation state during payout processing.

```go
var hourlyAggregator *hourlyRateAggregation
for payout := range payouts {
    if isHourlyRate {
        hourlyAggregator.add(hours, amount, description)
    } else {
        lineItems = append(lineItems, createLineItem(payout))
    }
}
if hourlyAggregator != nil {
    lineItems = append(lineItems, hourlyAggregator.toLineItem())
}
```

**Rejected Because**:
- Complex state management
- Harder to handle errors in middle of aggregation
- Mixing concerns: fetching + aggregation
- Difficult to test: requires mocking entire payout processing
- Less clear what happens if aggregation fails mid-way

### Alternative 2: Database-Level Aggregation

**Approach**: Query Notion with filters to pre-aggregate at database level.

**Rejected Because**:
- Notion API doesn't support aggregation queries
- Would require multiple queries: get IDs, then aggregate separately
- More complex than application-level aggregation
- Harder to debug and monitor

### Alternative 3: Pre-Group Before Processing

**Approach**: Group payouts by billing type before processing, handle each group differently.

**Rejected Because**:
- Requires knowing billing type before fetching (chicken-and-egg)
- Would need to fetch all rates upfront (violates lazy loading from ADR-001)
- Breaks existing sequential payout processing pattern

## Consequences

### Positive

1. **Clean Separation**: Build → Identify → Aggregate → Display
2. **Testable**: Each function testable independently with mock data
3. **Maintainable**: Clear boundaries between existing and new code
4. **Reversible**: Remove aggregation call, everything works as before
5. **Debuggable**: Log individual items + aggregated result
6. **Error Resilient**: Aggregation failure doesn't break invoice generation

### Negative

1. **Temporary Data Duplication**: Individual items exist briefly before aggregation
   - Mitigation: Happens in memory, negligible overhead
2. **Two-Pass Processing**: Build all items, then aggregate
   - Mitigation: Low volume (3-5 items max), performance impact negligible

### Neutral

1. **Memory Usage**: Slightly higher (stores individual + aggregated items temporarily)
   - Impact: Negligible for invoice size (< 10 items typically)

## Implementation Details

### Aggregation Function Signature

```go
// pkg/controller/invoice/contractor_invoice.go

// aggregateHourlyServiceFees consolidates all hourly-rate Service Fee items
// into a single line item with summed hours and amounts.
// Returns modified lineItems slice with aggregated item replacing individual items.
func aggregateHourlyServiceFees(
    lineItems []ContractorInvoiceLineItem,
    month string,
    logger logger.Logger,
) []ContractorInvoiceLineItem {
    // Implementation details in spec-003
}
```

### Integration Point

```go
// In GenerateContractorInvoice, after line item building (line ~286):

// 4. Build line items from payouts
var lineItems []ContractorInvoiceLineItem
for i, payout := range payouts {
    // ... existing line item creation logic ...
    lineItems = append(lineItems, lineItem)
}

// NEW: 4.5 Aggregate hourly-rate Service Fees
l.Debug("aggregating hourly-rate Service Fee items")
lineItems = aggregateHourlyServiceFees(lineItems, month, l)

// 4.6 Group Commission items by Project (existing code continues)
l.Debug("grouping commission items by project")
// ... rest of existing logic ...
```

### Title Generation

```go
// Generate title with invoice month date range
func generateServiceFeeTitle(month string) string {
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return "Service Fee"
    }

    startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
    endDate := startDate.AddDate(0, 1, -1)

    return fmt.Sprintf("Service Fee (Development work from %s to %s)",
        startDate.Format("2006-01-02"),
        endDate.Format("2006-01-02"))
}
```

### Description Concatenation

```go
// Concatenate proof of works with double line breaks
func concatenateDescriptions(descriptions []string) string {
    var filtered []string
    for _, desc := range descriptions {
        if strings.TrimSpace(desc) != "" {
            filtered = append(filtered, strings.TrimSpace(desc))
        }
    }
    return strings.Join(filtered, "\n\n")
}
```

## Edge Cases Handled

1. **No Hourly Items**: Return lineItems unchanged
2. **Single Hourly Item**: Still aggregate (for consistent title format)
3. **Empty Descriptions**: Skip empty strings in concatenation
4. **Different Hourly Rates**: Log warning, use first rate (shouldn't happen per requirements)
5. **Different Currencies**: Log warning, use first currency (shouldn't happen per requirements)
6. **Zero Hours**: Allow (may happen if Task Order fetch failed)

## Testing Strategy

### Unit Tests

```go
func TestAggregateHourlyServiceFees(t *testing.T) {
    tests := []struct {
        name     string
        input    []ContractorInvoiceLineItem
        month    string
        expected []ContractorInvoiceLineItem
    }{
        {
            name: "no hourly items - return unchanged",
            input: []ContractorInvoiceLineItem{
                {Type: "Commission", Amount: 100},
            },
            expected: // same as input
        },
        {
            name: "single hourly item - aggregate with title",
            input: []ContractorInvoiceLineItem{
                {Type: "Contractor Payroll", IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500},
            },
            expected: []ContractorInvoiceLineItem{
                {Title: "Service Fee (Development work from 2026-01-01 to 2026-01-31)", Hours: 10, Rate: 50, Amount: 500},
            },
        },
        {
            name: "multiple hourly items - aggregate into one",
            input: []ContractorInvoiceLineItem{
                {IsHourlyRate: true, Hours: 10, Rate: 50, Amount: 500, Description: "Work A"},
                {IsHourlyRate: true, Hours: 5, Rate: 50, Amount: 250, Description: "Work B"},
                {Type: "Commission", Amount: 100},
            },
            expected: []ContractorInvoiceLineItem{
                {Type: "Commission", Amount: 100},
                {Title: "Service Fee (...)", Hours: 15, Rate: 50, Amount: 750, Description: "Work A\n\nWork B"},
            },
        },
    }
}
```

### Integration Tests

- Test with real Notion test data
- Verify aggregated invoice output matches golden file
- Test multi-currency scenarios (VND and USD hourly fees)

## Monitoring and Logging

```go
l.Debug(fmt.Sprintf("[AGGREGATE] found %d hourly-rate Service Fee items to aggregate", len(hourlyItems)))
l.Debug(fmt.Sprintf("[AGGREGATE] totalHours=%.2f hourlyRate=%.2f totalAmount=%.2f currency=%s",
    totalHours, hourlyRate, totalAmount, currency))
l.Debug(fmt.Sprintf("[AGGREGATE] created aggregated item with title: %s", title))
```

## Success Criteria

1. All hourly Service Fees consolidated into single line item
2. Title includes correct date range from invoice month
3. Hours and amounts sum correctly
4. Descriptions concatenated with proper formatting
5. Non-hourly items unchanged
6. Unit test coverage > 90% for aggregation function
7. No regression in existing invoice generation

## Related Documents

- Requirements: FR-4 (Aggregation requirement)
- ADR-001: Data Fetching Strategy
- ADR-003: Error Handling Strategy
- Specification: `spec-003-detection-logic.md`
- Specification: `spec-004-integration.md`
