# ADR-001: Data Fetching Strategy for Hourly Rate Service Fees

**Status:** Proposed
**Date:** 2026-01-07
**Context:** Hourly Rate-Based Service Fee Display Implementation

## Context

When generating contractor invoices, we need to fetch additional data from Notion to display hourly-rate Service Fees correctly:
- Contractor Rate data (Billing Type, Hourly Rate, Currency)
- Task Order Log data (Final Hours Worked)

The challenge is deciding when and how to fetch this data while maintaining:
- Performance (acceptable API call volume)
- Code maintainability
- Error resilience
- Backward compatibility

## Decision

**We will use sequential, lazy-loading data fetching with early detection.**

### Approach

1. **Early Detection Phase** (during payout processing):
   - Extract `ServiceRateID` from PayoutEntry's `00 Service Rate` relation
   - Store ServiceRateID in PayoutEntry struct for later use
   - No immediate fetching - just capture the reference

2. **Lazy Loading Phase** (during line item enrichment):
   - For each Service Fee payout with ServiceRateID present:
     - Fetch Contractor Rate by ServiceRateID
     - Check if BillingType = "Hourly Rate"
     - If yes, fetch Task Order hours by TaskOrderID
     - If no, skip additional fetching

3. **Sequential Processing**:
   - Process payouts one at a time
   - Fetch rate data only when needed
   - Fetch hours data only for confirmed hourly-rate items

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Query Pending Payouts                               │
│ - Extract ServiceRateID from "00 Service Rate" relation     │
│ - Extract TaskOrderID from "00 Task Order" relation         │
│ - Store in PayoutEntry struct                               │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      v
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Build Line Items (per payout)                       │
│ - For Service Fee with ServiceRateID:                       │
│   ├─> Fetch Contractor Rate (ServiceRateID)                 │
│   ├─> Check BillingType = "Hourly Rate"?                    │
│   │   ├─ YES: Fetch Task Order hours (TaskOrderID)          │
│   │   │       Mark as hourly-rate item                      │
│   │   └─ NO:  Use default display (Qty=1)                   │
│   └─> On any error: Fallback to default display             │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      v
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Aggregate Hourly Service Fees                       │
│ - Identify all hourly-rate items                            │
│ - Sum hours, amount, concatenate descriptions               │
│ - Create single aggregated line item                        │
└─────────────────────────────────────────────────────────────┘
```

## Rationale

### Why Sequential Over Parallel?

1. **Low Volume**: Requirements specify 1-3 Service Fees per invoice (NFR-3)
   - 2-6 additional API calls per invoice is acceptable
   - Parallelization overhead outweighs benefits for this volume

2. **Simpler Error Handling**:
   - Each fetch can handle errors independently
   - Clear logging at each step
   - Easy to debug failures

3. **Matches Existing Patterns**:
   - Current code processes payouts sequentially
   - Minimal architectural disruption
   - Easier code review and maintenance

### Why Lazy Loading?

1. **Avoid Unnecessary Fetches**:
   - Only fetch rate data for Service Fees with ServiceRateID
   - Only fetch hours data for confirmed hourly-rate items
   - Non-hourly items (Monthly Fixed) skip additional API calls

2. **Clear Separation of Concerns**:
   - Detection phase: Identify what needs fetching
   - Fetch phase: Get required data
   - Aggregation phase: Build final display

3. **Graceful Degradation**:
   - Missing ServiceRateID → Skip fetching, use default display
   - Rate fetch fails → Skip hours fetch, use default display
   - Hours fetch fails → Continue with 0 hours, log error

## Alternatives Considered

### Alternative 1: Parallel Upfront Fetching

**Approach**: Fetch all Contractor Rates for the contractor at the start in parallel with payouts.

**Rejected Because**:
- Would fetch rate data for all months/periods, not just the relevant one
- More complex to match rates to payouts
- Unnecessary data fetching for non-Service Fee payouts (Commission, Refund)
- Over-engineering for low volume scenario

### Alternative 2: Bulk Task Order Query

**Approach**: Collect all TaskOrderIDs, then fetch hours data in single bulk query.

**Rejected Because**:
- Notion API doesn't support efficient bulk queries by page ID
- Would require OR filter with multiple page IDs (query complexity)
- Marginal performance benefit (1 query vs 2-3 queries)
- Added complexity for minimal gain

### Alternative 3: Embedded Data in Payout

**Approach**: Rely on Notion formulas/rollups to pre-calculate and embed hours data in Contractor Payouts.

**Rejected Because**:
- Requires Notion schema changes (out of scope per constraints)
- Tight coupling between databases
- Reduces flexibility for future changes
- Formula maintenance burden

## Consequences

### Positive

1. **Simple Implementation**: Straightforward sequential logic, easy to understand
2. **Testable**: Each fetch operation can be unit tested independently
3. **Debuggable**: Clear log trail showing each decision point
4. **Maintainable**: Follows existing codebase patterns
5. **Resilient**: Each step can fail gracefully without breaking invoice generation

### Negative

1. **Slightly Slower**: 2-6 sequential API calls per invoice (vs parallel approach)
   - Mitigation: Acceptable per NFR-3 for low volume
2. **N+1 Pattern**: One fetch per Service Fee item
   - Mitigation: Maximum 3 Service Fees per invoice per requirements

### Neutral

1. **API Call Volume**: 2-6 additional calls per invoice
   - Within acceptable limits per NFR-3
   - Logged for monitoring

## Implementation Notes

### Service Layer Changes

```go
// pkg/service/notion/contractor_payouts.go
type PayoutEntry struct {
    // ... existing fields ...
    ServiceRateID   string // NEW: From "00 Service Rate" relation
}

// Add extraction in QueryPendingPayoutsByContractor:
entry.ServiceRateID = s.extractFirstRelationID(props, "00 Service Rate")
```

### New Service Methods

```go
// pkg/service/notion/contractor_rates.go
func (s *ContractorRatesService) FetchContractorRateByPageID(ctx context.Context, pageID string) (*ContractorRateData, error)

// pkg/service/notion/task_order_log.go
func (s *TaskOrderLogService) FetchTaskOrderHoursByPageID(ctx context.Context, pageID string) (float64, error)
```

### Controller Helper

```go
// pkg/controller/invoice/contractor_invoice.go
func fetchHourlyRateData(ctx context.Context, serviceRateID, taskOrderID string, services) (*hourlyRateData, error) {
    // Fetch rate
    // Check billing type
    // Fetch hours if hourly rate
    // Return aggregated data or nil
}
```

## Monitoring and Logging

Every decision point will log:
```go
l.Debug(fmt.Sprintf("[HOURLY_RATE] payout %s: serviceRateID=%s taskOrderID=%s", payoutID, serviceRateID, taskOrderID))
l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched rate: billingType=%s hourlyRate=%.2f currency=%s", billingType, rate, currency))
l.Debug(fmt.Sprintf("[HOURLY_RATE] fetched hours: taskOrderID=%s hours=%.2f", taskOrderID, hours))
l.Debug(fmt.Sprintf("[HOURLY_RATE] fallback to default display: reason=%s", reason))
```

## Success Criteria

1. Invoice generation time increases by < 500ms per invoice
2. All fetches have error handling and fallback behavior
3. DEBUG logs show clear decision trail
4. Unit tests cover all fetch scenarios (success, failure, missing data)
5. No regression in existing invoice generation

## Related Documents

- Requirements: `docs/sessions/202601072141-hourly-rate-service-fee/requirements/requirements.md`
- ADR-002: Aggregation Approach
- ADR-003: Error Handling Strategy
- Specification: `spec-002-service-methods.md`
