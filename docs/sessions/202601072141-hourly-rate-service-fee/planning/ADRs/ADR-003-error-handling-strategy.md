# ADR-003: Error Handling and Fallback Strategy

**Status:** Proposed
**Date:** 2026-01-07
**Context:** Hourly Rate-Based Service Fee Display Implementation

## Context

The hourly rate feature introduces multiple points where data fetching can fail:
- ServiceRateID missing from Contractor Payouts
- Contractor Rate fetch failure (network, permissions, missing data)
- BillingType is not "Hourly Rate" (valid scenario, not error)
- Task Order Log fetch failure (network, permissions, missing data)
- Missing or invalid hours data

Per requirements (NFR-2): "No invoice generation failures due to missing hourly rate data"

The challenge is designing an error handling strategy that:
- Ensures invoice generation always succeeds
- Provides graceful degradation to fallback behavior
- Offers clear debugging information via logs
- Maintains backward compatibility

## Decision

**We will implement defensive programming with graceful degradation at every decision point.**

### Core Principle

> Every step that could fail must have:
> 1. Error detection
> 2. Logged reason for failure
> 3. Fallback to safe default behavior
> 4. Continuation of invoice generation

### Error Handling Decision Tree

```
┌──────────────────────────────────────────────────────────────┐
│ Processing Service Fee Payout                                 │
└─────────────────────┬────────────────────────────────────────┘
                      │
                      v
                ┌─────────────┐
                │ ServiceRateID│
                │ present?     │
                └──┬───────┬───┘
                   NO      YES
                   │        │
                   v        v
            ┌─────────────────────────┐
            │ DEFAULT DISPLAY         │    Fetch Contractor Rate
            │ Qty = 1                 │    by ServiceRateID
            │ Unit Cost = Amount      │           │
            │                         │           │
            │ Log: [FALLBACK]         │    ┌──────v──────┐
            │  No ServiceRateID       │    │ Fetch       │
            └─────────────────────────┘    │ Success?    │
                                           └──┬───────┬──┘
                                              NO      YES
                                              │        │
                                              v        v
                                       ┌─────────────────────────┐
                                       │ DEFAULT DISPLAY         │    Check BillingType
                                       │ Qty = 1                 │           │
                                       │ Unit Cost = Amount      │           │
                                       │                         │    ┌──────v──────────┐
                                       │ Log: [FALLBACK]         │    │ BillingType =   │
                                       │  Rate fetch failed      │    │ "Hourly Rate"?  │
                                       └─────────────────────────┘    └──┬───────────┬──┘
                                                                         NO          YES
                                                                         │            │
                                                                         v            v
                                                                  ┌─────────────────────────┐
                                                                  │ DEFAULT DISPLAY         │    Fetch Task Order hours
                                                                  │ Qty = 1                 │           │
                                                                  │ Unit Cost = Amount      │           │
                                                                  │                         │    ┌──────v──────┐
                                                                  │ Log: [INFO]             │    │ Fetch       │
                                                                  │  Not hourly rate        │    │ Success?    │
                                                                  └─────────────────────────┘    └──┬───────┬──┘
                                                                                                    NO      YES
                                                                                                    │        │
                                                                                                    v        v
                                                                                             ┌─────────────────────────┐
                                                                                             │ HOURLY DISPLAY (0 hrs)  │    HOURLY DISPLAY
                                                                                             │ Qty = 0                 │    Qty = hours
                                                                                             │ Unit Cost = HourlyRate  │    Unit Cost = HourlyRate
                                                                                             │ Amount = PayoutAmount   │    Amount = PayoutAmount
                                                                                             │                         │
                                                                                             │ Log: [FALLBACK]         │    Log: [SUCCESS]
                                                                                             │  Hours fetch failed     │     Hourly rate applied
                                                                                             └─────────────────────────┘
```

## Rationale

### Why Graceful Degradation?

1. **User Experience**: Invoice always generates, even with incomplete data
2. **Operational Safety**: No production outages due to Notion API issues
3. **Backward Compatibility**: Fallback behavior matches current implementation
4. **Debuggability**: Logs show exactly what failed and why

### Why Log Every Decision?

1. **Troubleshooting**: Support can diagnose issues from logs
2. **Monitoring**: Can track error rates and patterns
3. **Audit Trail**: Know which invoices used fallback behavior
4. **Performance**: Can identify slow API calls

### Why Continue on Errors?

1. **Requirements**: NFR-2 mandates no invoice generation failures
2. **Business Logic**: Better to show invoice with Qty=1 than no invoice
3. **Manual Review**: Team can review and correct if needed
4. **Partial Success**: One failed item doesn't break entire invoice

## Error Handling Specifications

### Error Category 1: Missing Data (Not Errors)

These are valid scenarios where fallback is expected behavior:

```go
// ServiceRateID missing
if payout.ServiceRateID == "" {
    l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no ServiceRateID, using default display (Qty=1)", payout.PageID))
    // Use default: Qty=1, Unit Cost=Amount
    return createDefaultLineItem(payout)
}
```

```go
// BillingType is not "Hourly Rate"
if rateData.BillingType != "Hourly Rate" {
    l.Debug(fmt.Sprintf("[INFO] payout %s: billingType=%s (not hourly), using default display (Qty=1)",
        payout.PageID, rateData.BillingType))
    // Use default: Qty=1, Unit Cost=Amount
    return createDefaultLineItem(payout)
}
```

### Error Category 2: Fetch Failures (Actual Errors)

These are error conditions requiring fallback:

```go
// Contractor Rate fetch failed
rateData, err := ratesService.FetchContractorRateByPageID(ctx, payout.ServiceRateID)
if err != nil {
    l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch contractor rate (serviceRateID=%s), using default display (Qty=1)",
        payout.PageID, payout.ServiceRateID))
    // Use default: Qty=1, Unit Cost=Amount
    return createDefaultLineItem(payout)
}
```

```go
// Task Order hours fetch failed
hours, err := taskOrderService.FetchTaskOrderHoursByPageID(ctx, payout.TaskOrderID)
if err != nil {
    l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch task order hours (taskOrderID=%s), using 0 hours",
        payout.PageID, payout.TaskOrderID))
    // Use 0 hours, continue with hourly rate display
    hours = 0
}
```

### Error Category 3: Data Validation (Warnings)

These are unexpected but non-fatal conditions:

```go
// Multiple hourly items with different rates
if aggregation.hourlyRates.Count() > 1 {
    l.Warn(fmt.Sprintf("[WARN] multiple hourly rates found for invoice: rates=%v, using first rate=%.2f",
        aggregation.hourlyRates.Values(), aggregation.hourlyRates.First()))
}
```

```go
// Multiple currencies in hourly items (shouldn't happen)
if aggregation.currencies.Count() > 1 {
    l.Warn(fmt.Sprintf("[WARN] multiple currencies found in hourly items: currencies=%v, using first=%s",
        aggregation.currencies.Values(), aggregation.currencies.First()))
}
```

## Logging Strategy

### Log Levels

- **DEBUG**: Decision points, data fetching, normal flow
- **ERROR**: API failures, missing required data (with fallback)
- **WARN**: Unexpected conditions that don't break flow

### Log Prefixes

- `[HOURLY_RATE]`: Hourly rate detection and processing
- `[FALLBACK]`: Fallback to default behavior (with reason)
- `[SUCCESS]`: Successful hourly rate application
- `[INFO]`: Normal non-hourly flow
- `[AGGREGATE]`: Aggregation processing
- `[WARN]`: Unexpected but handled conditions

### Log Format

```go
// Every decision point
l.Debug(fmt.Sprintf("[HOURLY_RATE] payout %s: serviceRateID=%s taskOrderID=%s",
    payout.PageID, payout.ServiceRateID, payout.TaskOrderID))

// Successful fetch
l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: pageID=%s billingType=%s hourlyRate=%.2f currency=%s",
    rateData.PageID, rateData.BillingType, rateData.HourlyRate, rateData.Currency))

// Successful hours fetch
l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: taskOrderID=%s hours=%.2f",
    payout.TaskOrderID, hours))

// Fallback with reason
l.Debug(fmt.Sprintf("[FALLBACK] payout %s: %s, using default display (Qty=1, Unit Cost=%.2f)",
    payout.PageID, reason, payout.Amount))

// Success
l.Debug(fmt.Sprintf("[SUCCESS] payout %s: applied hourly rate display (Hours=%.2f, Rate=%.2f, Amount=%.2f)",
    payout.PageID, hours, rate, amount))
```

## Alternatives Considered

### Alternative 1: Fail Fast

**Approach**: Return error immediately on any fetch failure, halt invoice generation.

**Rejected Because**:
- Violates NFR-2 requirement
- Poor user experience (no invoice generated)
- Operational risk (production outages)
- Over-engineering for low-risk feature

### Alternative 2: Retry Logic

**Approach**: Retry failed API calls with exponential backoff.

**Rejected Because**:
- Adds complexity (retry logic, timeout management)
- Increases invoice generation time
- Notion API failures are typically not transient (missing data, permissions)
- Fallback to default display is acceptable per requirements

### Alternative 3: Error Accumulation

**Approach**: Collect all errors, return at end with partial invoice data.

**Rejected Because**:
- Still breaks invoice generation (return error)
- Complicates control flow
- No benefit over graceful degradation
- Harder to debug

### Alternative 4: Silent Fallback

**Approach**: Fallback without logging errors.

**Rejected Because**:
- No visibility into issues
- Can't diagnose problems
- Can't track error rates
- Support team can't help users

## Consequences

### Positive

1. **Resilient**: Invoice always generates successfully
2. **Safe**: No production outages from Notion API issues
3. **Debuggable**: Clear log trail for every decision
4. **Transparent**: Support can see what happened
5. **Backward Compatible**: Fallback matches current behavior
6. **Testable**: Each error path can be unit tested

### Negative

1. **Silent Degradation**: Invoices may have Qty=1 when should be hourly
   - Mitigation: Clear logs for manual review
2. **Log Volume**: More DEBUG logs per invoice
   - Mitigation: Use DEBUG level, production typically logs INFO+

### Neutral

1. **Manual Review**: Some invoices may need manual correction
   - Acceptable: Feature is enhancement, not critical path

## Implementation Examples

### Helper Function: Create Default Line Item

```go
// createDefaultLineItem creates line item with default display (Qty=1, Unit Cost=Amount)
// Used as fallback when hourly rate data unavailable
func createDefaultLineItem(payout notion.PayoutEntry, amountUSD float64, description string) ContractorInvoiceLineItem {
    return ContractorInvoiceLineItem{
        Title:            "",
        Description:      description,
        Hours:            1,    // Default quantity
        Rate:             amountUSD,
        Amount:           amountUSD,
        AmountUSD:        amountUSD,
        Type:             string(payout.SourceType),
        OriginalAmount:   payout.Amount,
        OriginalCurrency: payout.Currency,
        IsHourlyRate:     false, // Explicitly mark as non-hourly
    }
}
```

### Helper Function: Fetch Hourly Rate Data with Fallback

```go
// fetchHourlyRateData fetches contractor rate and task order hours
// Returns nil if any step fails (caller uses default display)
func fetchHourlyRateData(
    ctx context.Context,
    payout notion.PayoutEntry,
    services *invoiceServices,
    l logger.Logger,
) *hourlyRateData {
    // Check ServiceRateID present
    if payout.ServiceRateID == "" {
        l.Debug(fmt.Sprintf("[FALLBACK] payout %s: no ServiceRateID", payout.PageID))
        return nil
    }

    // Fetch Contractor Rate
    l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching contractor rate: serviceRateID=%s", payout.ServiceRateID))
    rateData, err := services.ContractorRates.FetchContractorRateByPageID(ctx, payout.ServiceRateID)
    if err != nil {
        l.Error(err, fmt.Sprintf("[FALLBACK] payout %s: failed to fetch rate", payout.PageID))
        return nil
    }

    l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s",
        rateData.BillingType, rateData.HourlyRate, rateData.Currency))

    // Check BillingType
    if rateData.BillingType != "Hourly Rate" {
        l.Debug(fmt.Sprintf("[INFO] payout %s: billingType=%s (not hourly)", payout.PageID, rateData.BillingType))
        return nil
    }

    // Fetch Task Order hours
    var hours float64
    if payout.TaskOrderID != "" {
        l.Debug(fmt.Sprintf("[HOURLY_RATE] fetching hours: taskOrderID=%s", payout.TaskOrderID))
        hours, err = services.TaskOrderLog.FetchTaskOrderHoursByPageID(ctx, payout.TaskOrderID)
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

### Integration Example

```go
// In GenerateContractorInvoice, line item building:
for i, payout := range payouts {
    amountUSD := amountsUSD[i]

    // ... existing description logic ...

    // NEW: Attempt to fetch hourly rate data
    var lineItem ContractorInvoiceLineItem
    if payout.SourceType == notion.PayoutSourceTypeServiceFee {
        hourlyData := fetchHourlyRateData(ctx, payout, services, l)
        if hourlyData != nil {
            // SUCCESS: Use hourly rate display
            l.Debug(fmt.Sprintf("[SUCCESS] payout %s: applying hourly rate display", payout.PageID))
            lineItem = ContractorInvoiceLineItem{
                Title:            "", // Set later during aggregation
                Description:      description,
                Hours:            hourlyData.Hours,
                Rate:             hourlyData.HourlyRate,
                Amount:           payout.Amount, // Use original amount from payout
                AmountUSD:        amountUSD,
                Type:             string(payout.SourceType),
                OriginalAmount:   payout.Amount,
                OriginalCurrency: payout.Currency,
                IsHourlyRate:     true, // Mark for aggregation
                ServiceRateID:    payout.ServiceRateID,
                TaskOrderID:      payout.TaskOrderID,
            }
        } else {
            // FALLBACK: Use default display
            lineItem = createDefaultLineItem(payout, amountUSD, description)
        }
    } else {
        // Non-Service Fee: use default display
        lineItem = createDefaultLineItem(payout, amountUSD, description)
    }

    lineItems = append(lineItems, lineItem)
}
```

## Testing Strategy

### Unit Tests for Error Paths

```go
func TestFetchHourlyRateData_ErrorScenarios(t *testing.T) {
    tests := []struct {
        name           string
        payout         notion.PayoutEntry
        mockRateErr    error
        mockHoursErr   error
        expectedResult *hourlyRateData
        expectedLog    string
    }{
        {
            name:           "missing ServiceRateID returns nil",
            payout:         notion.PayoutEntry{ServiceRateID: ""},
            expectedResult: nil,
            expectedLog:    "[FALLBACK] no ServiceRateID",
        },
        {
            name:           "rate fetch error returns nil",
            payout:         notion.PayoutEntry{ServiceRateID: "rate123"},
            mockRateErr:    errors.New("notion API error"),
            expectedResult: nil,
            expectedLog:    "[FALLBACK] failed to fetch rate",
        },
        {
            name: "non-hourly billing type returns nil",
            payout: notion.PayoutEntry{ServiceRateID: "rate123"},
            mockRate: &notion.ContractorRateData{BillingType: "Monthly Fixed"},
            expectedResult: nil,
            expectedLog: "[INFO] billingType=Monthly Fixed (not hourly)",
        },
        {
            name: "hours fetch error uses 0 hours",
            payout: notion.PayoutEntry{ServiceRateID: "rate123", TaskOrderID: "task456"},
            mockRate: &notion.ContractorRateData{BillingType: "Hourly Rate", HourlyRate: 50},
            mockHoursErr: errors.New("task order not found"),
            expectedResult: &hourlyRateData{HourlyRate: 50, Hours: 0},
            expectedLog: "[FALLBACK] failed to fetch hours, using 0",
        },
    }
}
```

### Integration Tests

- Test invoice generation with various Notion API failures
- Verify all fallback paths produce valid invoices
- Check log output contains expected error messages

## Monitoring and Alerting

### Metrics to Track

1. **Fallback Rate**: % of Service Fees using fallback display
2. **Error Types**: Count by error category (rate fetch, hours fetch, etc.)
3. **API Failures**: Track Notion API error rates
4. **Zero Hours Count**: Track Service Fees with 0 hours (may indicate issue)

### Alert Thresholds

- Fallback rate > 20%: Investigate Notion API or data quality issues
- Zero hours > 5 per day: Check Task Order Log data completeness

## Success Criteria

1. Invoice generation never fails due to missing hourly rate data
2. Every error logged with clear reason and context
3. All fallback paths covered by unit tests
4. Integration tests verify fallback behavior
5. Support team can diagnose issues from logs
6. Monitoring shows < 10% fallback rate in production

## Related Documents

- Requirements: NFR-2 (Error Handling requirement)
- ADR-001: Data Fetching Strategy
- ADR-002: Aggregation Approach
- ADR-004: Code Organization
