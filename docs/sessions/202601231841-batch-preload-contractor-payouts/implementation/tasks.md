# Implementation Tasks: Batch Pre-Loading for Contractor Payouts

## Overview
Optimize `/cronjobs/create-contractor-payouts` by pre-fetching data in batches before worker pool processing, eliminating N+1 API calls.

---

## Tasks

### Task 1: Add Batch Contractor Rates Fetcher to Service Layer
- **Status**: COMPLETED
- **File(s)**: `pkg/service/notion/contractor_rates.go`
- **Description**: Add a new method `FindActiveRatesByContractors(ctx, contractorIDs []string, date time.Time) (map[string]*ContractorRateData, error)` that fetches rates for multiple contractors in a single Notion API query
- **Implementation Notes**:
  - Method returns a map keyed by contractor ID
  - Queries all active rates and filters by contractor IDs in memory
  - Fetches contractor details in parallel with maxConcurrent=5

### Task 2: Add Batch Payout Existence Checker to Service Layer
- **Status**: COMPLETED
- **File(s)**: `pkg/service/notion/contractor_payouts.go`
- **Description**: Add `CheckPayoutsExistByContractorFees(ctx, taskOrderPageIDs []string) (map[string]string, error)` 
- **Implementation Notes**:
  - Returns map of taskOrderID -> existingPayoutPageID
  - Queries payouts with "00 Task Order" relation set, filters in memory
  - Early exit when all target IDs found

### Task 3: Add Contractor Positions Cache Structure
- **Status**: COMPLETED
- **File(s)**: `pkg/service/notion/contractor_payouts.go`
- **Description**: Add `GetContractorPositionsBatch(ctx, contractorPageIDs []string) map[string][]string`
- **Implementation Notes**:
  - Runs position fetches in parallel with maxConcurrent=5
  - Deduplicates contractor IDs before fetching
  - Returns map of contractorID -> []positions

### Task 4: Refactor processContractorPayrollPayouts - Pre-Loading Phase
- **Status**: COMPLETED
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**: Added batch pre-loading phase before worker pool
- **Implementation Notes**:
  - Collects unique contractor IDs and task order IDs
  - Runs 3 batch fetches in parallel (rates, payouts, positions)
  - Logs pre-loading duration for performance monitoring
  - Data passed to workers via closure (read-only maps)

### Task 5: Refactor Worker Loop to Use Pre-Loaded Data
- **Status**: COMPLETED
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**: Modified worker loop to use pre-loaded data instead of individual API calls
- **Implementation Notes**:
  - Lookups from preloadedRates map (no API call)
  - Lookups from preloadedPayouts map (no API call)
  - Lookups from preloadedPositions map (no API call)
  - Only CreatePayout API call remains per order (unavoidable)

### Task 6: Apply Same Pattern to processRefundPayouts
- **Status**: COMPLETED
- **File(s)**: `pkg/service/notion/contractor_payouts.go`, `pkg/handler/notion/contractor_payouts.go`
- **Description**: Applied batch pre-loading pattern to refund payouts
- **Implementation Notes**:
  - Added `CheckPayoutsExistByRefundRequests(ctx, refundRequestPageIDs []string)` batch method
  - Pre-loads payout existence before worker pool
  - Workers use preloaded cache instead of individual API calls

### Task 7: Apply Same Pattern to processInvoiceSplitPayouts
- **Status**: COMPLETED
- **File(s)**: `pkg/service/notion/contractor_payouts.go`, `pkg/handler/notion/contractor_payouts.go`
- **Description**: Applied batch pre-loading pattern to invoice split payouts
- **Implementation Notes**:
  - Added `CheckPayoutsExistByInvoiceSplits(ctx, invoiceSplitPageIDs []string)` batch method
  - Pre-loads payout existence before worker pool
  - Workers use preloaded cache instead of individual API calls

### Task 8: Add Debug Logging for Performance Monitoring
- **Status**: COMPLETED
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`, `pkg/service/notion/contractor_rates.go`, `pkg/service/notion/contractor_payouts.go`
- **Description**: Added DEBUG-level logging throughout
- **Implementation Notes**:
  - [PRELOAD] prefix for pre-loading phase logs
  - [BATCH_RATES], [BATCH_PAYOUT_CHECK], [BATCH_POSITIONS] prefixes in service layer
  - [BATCH_REFUND_CHECK], [BATCH_SPLIT_CHECK] prefixes for other payout types
  - All logs include counts, durations, and relevant IDs (no PII)

---

## Dependency Order
```
Task 1 ─┬─→ Task 4 ─→ Task 5 ─┬─→ Task 6 ─→ Task 7 ─→ Task 8
Task 2 ─┘                     │
Task 3 ───────────────────────┘
```

Tasks 1, 2, 3 can be done in parallel. Task 4 requires 1 & 2. Task 5 requires 3 & 4. Tasks 6, 7, 8 are sequential after Task 5.

---

## Summary of API Call Reduction

| Handler | Before | After |
|---------|--------|-------|
| processContractorPayrollPayouts | O(3N) per order | O(3) total + O(N) creates |
| processRefundPayouts | O(N) per refund | O(1) total + O(N) creates |
| processInvoiceSplitPayouts | O(N) per split | O(1) total + O(N) creates |

For 50 contractors/orders:
- **Before**: ~150 API calls (3 per order: rate + payout check + positions)
- **After**: ~53 API calls (3 batch + 50 creates)

---

## Files Modified

1. `pkg/service/notion/contractor_rates.go`
   - Added: `FindActiveRatesByContractors()` batch method

2. `pkg/service/notion/contractor_payouts.go`
   - Added: `CheckPayoutsExistByContractorFees()` batch method
   - Added: `CheckPayoutsExistByRefundRequests()` batch method
   - Added: `CheckPayoutsExistByInvoiceSplits()` batch method
   - Added: `GetContractorPositionsBatch()` parallel fetch method

3. `pkg/handler/notion/contractor_payouts.go`
   - Refactored: `processContractorPayrollPayouts()` with batch pre-loading
   - Refactored: `processRefundPayouts()` with batch pre-loading
   - Refactored: `processInvoiceSplitPayouts()` with batch pre-loading
