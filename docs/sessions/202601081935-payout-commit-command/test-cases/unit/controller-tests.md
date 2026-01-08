# Controller Unit Test Cases - Contractor Payables

## Overview

This document defines unit test cases for the Contractor Payables controller layer (`pkg/controller/contractorpayables/`). These tests focus on business logic, cascade update orchestration, and error handling with mocked Notion services.

## Test File Location

`pkg/controller/contractorpayables/contractorpayables_test.go`

## Test Setup Pattern

```go
// Mock Notion services
type mockContractorPayablesService struct {
    queryFunc      func(ctx context.Context, period string) ([]notion.PendingPayable, error)
    updateFunc     func(ctx context.Context, pageID, status, paymentDate string) error
    getPayDayFunc  func(ctx context.Context, contractorPageID string) (int, error)
}

type mockContractorPayoutsService struct {
    getWithRelationsFunc func(ctx context.Context, payoutPageID string) (*notion.PayoutWithRelations, error)
    updateStatusFunc     func(ctx context.Context, pageID, status string) error
}

type mockInvoiceSplitService struct {
    updateStatusFunc func(ctx context.Context, pageID, status string) error
}

type mockRefundRequestService struct {
    updateStatusFunc func(ctx context.Context, pageID, status string) error
}

// Test setup helper
func setupControllerTest() (*controller, *mockNotionServices) {
    logger := logger.NewLogrusLogger("error")
    config := &config.Config{}

    mockServices := &mockNotionServices{
        contractorPayables: &mockContractorPayablesService{},
        contractorPayouts:  &mockContractorPayoutsService{},
        invoiceSplit:       &mockInvoiceSplitService{},
        refundRequests:     &mockRefundRequestService{},
    }

    service := &service.Service{
        Notion: mockServices,
    }

    ctrl := New(config, logger, service)
    return ctrl.(*controller), mockServices
}
```

---

## Test Suite 1: PreviewCommit Method

### Test 1.1: Valid Request - Multiple Payables

**Test Name**: `TestPreviewCommit_ValidRequest_MultiplePayables`

**Description**: Verify preview correctly queries and filters payables by PayDay.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns:
  ```go
  []notion.PendingPayable{
      {
          PageID: "payable-1",
          ContractorPageID: "contractor-1",
          ContractorName: "John Doe",
          Total: 5000.00,
          Currency: "USD",
          PayoutItemPageIDs: []string{"payout-1", "payout-2"},
      },
      {
          PageID: "payable-2",
          ContractorPageID: "contractor-2",
          ContractorName: "Jane Smith",
          Total: 7500.00,
          Currency: "USD",
          PayoutItemPageIDs: []string{"payout-3"},
      },
      {
          PageID: "payable-3",
          ContractorPageID: "contractor-3",
          ContractorName: "Bob Wilson",
          Total: 2500.00,
          Currency: "VND",
          PayoutItemPageIDs: []string{"payout-4"},
      },
  }, nil
  ```
- Mock GetContractorPayDay returns:
  - contractor-1: 15
  - contractor-2: 15
  - contractor-3: 1 (should be filtered out)

**Expected Output**:
- No error returned
- Response contains:
  - Month: "2025-01"
  - Batch: 15
  - Count: 2
  - TotalAmount: 12500.00
  - Contractors: 2 items (John Doe and Jane Smith only)
- Bob Wilson filtered out due to PayDay=1

**Edge Cases Covered**:
- Multiple payables with different PayDays
- PayDay filtering logic
- Total amount calculation
- Mixed currencies

---

### Test 1.2: Valid Request - All Filtered Out

**Test Name**: `TestPreviewCommit_AllPayablesFilteredOut`

**Description**: Verify preview returns empty result when all payables are filtered by PayDay.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 3 payables
- Mock GetContractorPayDay returns 1 for all contractors (all filtered out)

**Expected Output**:
- No error returned
- Response contains:
  - Count: 0
  - TotalAmount: 0.00
  - Contractors: empty array

**Edge Cases Covered**:
- All payables filtered out by PayDay
- Empty result handling

---

### Test 1.3: No Pending Payables Found

**Test Name**: `TestPreviewCommit_NoPendingPayables`

**Description**: Verify preview handles case when query returns no results.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns empty slice

**Expected Output**:
- No error returned
- Response contains:
  - Count: 0
  - TotalAmount: 0.00
  - Contractors: empty array

**Edge Cases Covered**:
- No pending payables in period

---

### Test 1.4: Query Error

**Test Name**: `TestPreviewCommit_QueryError`

**Description**: Verify preview propagates errors from Notion service.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns error: "notion API connection failed"

**Expected Output**:
- Error returned containing "failed to query pending payables"
- Error wraps original error

**Edge Cases Covered**:
- Service layer error propagation

---

### Test 1.5: GetPayDay Error - Skip Contractor

**Test Name**: `TestPreviewCommit_GetPayDayError_SkipContractor`

**Description**: Verify preview continues when GetPayDay fails for some contractors.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 3 payables
- Mock GetContractorPayDay:
  - contractor-1: returns 15
  - contractor-2: returns error "no service rate found"
  - contractor-3: returns 15

**Expected Output**:
- No error returned (best-effort approach)
- Response contains 2 contractors (contractor-2 skipped)
- Log should contain debug message about skipped contractor

**Edge Cases Covered**:
- Partial PayDay fetch failures
- Best-effort filtering

---

### Test 1.6: Single Payable Result

**Test Name**: `TestPreviewCommit_SinglePayable`

**Description**: Verify preview handles single payable correctly.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 1 payable
- Mock GetContractorPayDay returns 15

**Expected Output**:
- Response contains:
  - Count: 1
  - TotalAmount: payable amount
  - Contractors: 1 item

**Edge Cases Covered**:
- Single item result

---

### Test 1.7: Batch=1 Filtering

**Test Name**: `TestPreviewCommit_Batch1Filtering`

**Description**: Verify preview correctly filters for batch=1.

**Input/Setup**:
- Month: "2025-01"
- Batch: 1 (not 15)
- Mock QueryPendingPayablesByPeriod returns multiple payables
- Mock GetContractorPayDay returns mix of 1 and 15

**Expected Output**:
- Only contractors with PayDay=1 included
- Contractors with PayDay=15 filtered out

**Edge Cases Covered**:
- Batch=1 filtering (not just 15)

---

### Test 1.8: Period Date Conversion

**Test Name**: `TestPreviewCommit_PeriodDateConversion`

**Description**: Verify month parameter is correctly converted to period date format.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Verify QueryPendingPayablesByPeriod is called with "2025-01-01"

**Expected Output**:
- Service method called with period "2025-01-01" (YYYY-MM-01)

**Edge Cases Covered**:
- Date format conversion

---

### Test 1.9: Large Number of Payables

**Test Name**: `TestPreviewCommit_LargeNumberOfPayables`

**Description**: Verify preview handles large result sets efficiently.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 100 payables
- All have matching PayDay

**Expected Output**:
- Response contains all 100 payables
- Total amount correctly calculated
- No performance issues (test execution time)

**Edge Cases Covered**:
- Large result sets

---

### Test 1.10: Mixed Currencies

**Test Name**: `TestPreviewCommit_MixedCurrencies`

**Description**: Verify preview preserves currency information for each contractor.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns payables with USD, VND, GBP currencies

**Expected Output**:
- Each contractor preview includes correct currency
- Total amount is sum (without currency conversion)

**Edge Cases Covered**:
- Multiple currencies in result

---

## Test Suite 2: CommitPayables Method

### Test 2.1: Valid Request - Full Success

**Test Name**: `TestCommitPayables_ValidRequest_FullSuccess`

**Description**: Verify commit successfully updates all payables and related records.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 2 payables
- Mock GetContractorPayDay returns 15 for both
- Mock GetPayoutWithRelations returns payout with Invoice Split
- All update methods succeed

**Expected Output**:
- Response contains:
  - Updated: 2
  - Failed: 0
  - Errors: nil
- UpdatePayableStatus called twice with "Paid" status
- UpdatePayoutStatus called for each payout item
- UpdateInvoiceSplitStatus called for related invoice splits
- Payment date set to current date

**Edge Cases Covered**:
- Standard happy path with full cascade updates

---

### Test 2.2: Valid Request - With Refund Relations

**Test Name**: `TestCommitPayables_WithRefundRelations`

**Description**: Verify commit updates Refund Request status when present.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 1 payable with 1 payout item
- Mock GetPayoutWithRelations returns:
  ```go
  &PayoutWithRelations{
      PageID: "payout-1",
      Status: "Pending",
      InvoiceSplitID: "",
      RefundRequestID: "refund-1",
  }
  ```
- All update methods succeed

**Expected Output**:
- UpdateRefundRequestStatus called with "refund-1" and "Paid"
- UpdateInvoiceSplitStatus NOT called (empty ID)
- UpdatePayoutStatus called
- UpdatePayableStatus called

**Edge Cases Covered**:
- Payout with Refund relation only
- Conditional update logic

---

### Test 2.3: Valid Request - With Both Invoice and Refund

**Test Name**: `TestCommitPayables_WithBothRelations`

**Description**: Verify commit updates both Invoice Split and Refund when present.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock GetPayoutWithRelations returns:
  ```go
  &PayoutWithRelations{
      PageID: "payout-1",
      Status: "Pending",
      InvoiceSplitID: "split-1",
      RefundRequestID: "refund-1",
  }
  ```

**Expected Output**:
- UpdateInvoiceSplitStatus called with "split-1" and "Paid"
- UpdateRefundRequestStatus called with "refund-1" and "Paid"
- Both relations updated

**Edge Cases Covered**:
- Payout with both relations populated

---

### Test 2.4: Valid Request - No Relations

**Test Name**: `TestCommitPayables_NoRelations`

**Description**: Verify commit works when payout has no related records.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock GetPayoutWithRelations returns:
  ```go
  &PayoutWithRelations{
      PageID: "payout-1",
      Status: "Pending",
      InvoiceSplitID: "",
      RefundRequestID: "",
  }
  ```

**Expected Output**:
- UpdateInvoiceSplitStatus NOT called
- UpdateRefundRequestStatus NOT called
- UpdatePayoutStatus called
- UpdatePayableStatus called
- No errors returned

**Edge Cases Covered**:
- Payout with no relations

---

### Test 2.5: Multiple Payout Items Per Payable

**Test Name**: `TestCommitPayables_MultiplePayoutItems`

**Description**: Verify commit handles payables with multiple payout items.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns 1 payable with 3 payout items
- All update methods succeed

**Expected Output**:
- GetPayoutWithRelations called 3 times
- UpdatePayoutStatus called 3 times
- UpdatePayableStatus called once
- All updates successful

**Edge Cases Covered**:
- Multiple payout items per payable

---

### Test 2.6: Partial Failure - Payout Update Fails

**Test Name**: `TestCommitPayables_PartialFailure_PayoutUpdate`

**Description**: Verify commit continues on payout update failure (best-effort).

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns 2 payables
- GetPayoutWithRelations succeeds for all
- UpdatePayoutStatus fails for one payout item
- UpdatePayableStatus succeeds for both payables

**Expected Output**:
- Response shows partial success
- Error logged but not propagated
- Other payout items still processed
- Payable status still updated (best-effort)

**Edge Cases Covered**:
- Best-effort payout updates
- Error handling in cascade

---

### Test 2.7: Partial Failure - Invoice Split Update Fails

**Test Name**: `TestCommitPayables_PartialFailure_InvoiceSplit`

**Description**: Verify commit continues when Invoice Split update fails.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- UpdateInvoiceSplitStatus returns error
- Other updates succeed

**Expected Output**:
- Error logged
- UpdatePayoutStatus still called
- UpdatePayableStatus still called
- Commit continues (best-effort)

**Edge Cases Covered**:
- Related record update failures

---

### Test 2.8: Partial Failure - Refund Update Fails

**Test Name**: `TestCommitPayables_PartialFailure_RefundUpdate`

**Description**: Verify commit continues when Refund Request update fails.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- UpdateRefundRequestStatus returns error
- Other updates succeed

**Expected Output**:
- Error logged
- Payout and Payable updates still proceed
- Best-effort approach

**Edge Cases Covered**:
- Refund update failures

---

### Test 2.9: Partial Failure - Payable Update Fails

**Test Name**: `TestCommitPayables_PartialFailure_PayableUpdate`

**Description**: Verify commit returns error when payable update fails.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns 2 payables
- UpdatePayableStatus succeeds for first, fails for second

**Expected Output**:
- Response contains:
  - Updated: 1
  - Failed: 1
  - Errors: array with 1 error
- First payable marked as updated
- Second payable marked as failed with error details

**Edge Cases Covered**:
- Payable update failures
- Failure tracking

---

### Test 2.10: No Pending Payables - Returns Error

**Test Name**: `TestCommitPayables_NoPendingPayables`

**Description**: Verify commit returns error when no pending payables found.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns empty slice

**Expected Output**:
- Error returned: "no pending payables found for month 2025-01"
- No updates attempted

**Edge Cases Covered**:
- Empty result set handling

---

### Test 2.11: Query Error

**Test Name**: `TestCommitPayables_QueryError`

**Description**: Verify commit propagates query errors.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns error

**Expected Output**:
- Error returned wrapping original error
- No updates attempted

**Edge Cases Covered**:
- Service layer errors

---

### Test 2.12: All Payables Filtered by PayDay

**Test Name**: `TestCommitPayables_AllFilteredByPayDay`

**Description**: Verify commit returns error when all payables filtered out.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock QueryPendingPayablesByPeriod returns payables
- Mock GetContractorPayDay returns 1 for all (filtered out)

**Expected Output**:
- Error returned: "no pending payables found for month 2025-01 batch 15"
- No updates attempted

**Edge Cases Covered**:
- All payables filtered by PayDay

---

### Test 2.13: GetPayoutWithRelations Error

**Test Name**: `TestCommitPayables_GetPayoutError`

**Description**: Verify commit handles errors fetching payout relations.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock GetPayoutWithRelations returns error

**Expected Output**:
- Error logged
- Payout item processing continues (best-effort)
- Or payable marked as failed depending on implementation

**Edge Cases Covered**:
- Payout fetch errors

---

### Test 2.14: Payment Date Format

**Test Name**: `TestCommitPayables_PaymentDateFormat`

**Description**: Verify payment date is set in correct format.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns 1 payable
- All updates succeed
- Capture UpdatePayableStatus call arguments

**Expected Output**:
- UpdatePayableStatus called with payment date in "YYYY-MM-DD" format
- Date is current date

**Edge Cases Covered**:
- Date formatting

---

### Test 2.15: Commit Sequence Order

**Test Name**: `TestCommitPayables_UpdateSequenceOrder`

**Description**: Verify updates happen in correct cascade order.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns 1 payable with 1 payout item with Invoice Split
- Track call order

**Expected Output**:
- Call sequence:
  1. GetPayoutWithRelations
  2. UpdateInvoiceSplitStatus
  3. UpdatePayoutStatus
  4. UpdatePayableStatus
- Correct cascade order maintained

**Edge Cases Covered**:
- Update sequence ordering

---

### Test 2.16: Idempotency Check

**Test Name**: `TestCommitPayables_Idempotency`

**Description**: Verify commit can be safely re-run on same data.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns payables already with "Paid" status (simulating re-run)
- All update methods succeed (idempotent)

**Expected Output**:
- No errors returned
- Updates applied successfully
- No duplicate side effects

**Edge Cases Covered**:
- Idempotency of operations

---

### Test 2.17: Empty Payout Items Array

**Test Name**: `TestCommitPayables_EmptyPayoutItems`

**Description**: Verify commit handles payable with no payout items.

**Input/Setup**:
- Month: "2025-01"
- Batch: 15
- Mock returns payable with empty PayoutItemPageIDs array

**Expected Output**:
- No payout item updates attempted
- Payable status still updated
- No errors

**Edge Cases Covered**:
- Payable with no payout items

---

## Test Suite 3: Helper Methods

### Test 3.1: commitSinglePayable - Full Cascade

**Test Name**: `TestCommitSinglePayable_FullCascade`

**Description**: Verify single payable commit with full cascade updates.

**Input/Setup**:
- Payable with 2 payout items
- Each payout has Invoice Split and Refund
- All updates succeed

**Expected Output**:
- All related records updated
- Payable status updated
- No errors

**Edge Cases Covered**:
- Complete cascade update

---

### Test 3.2: commitSinglePayable - Payable Update Fails

**Test Name**: `TestCommitSinglePayable_PayableUpdateFails`

**Description**: Verify error returned when final payable update fails.

**Input/Setup**:
- All payout updates succeed
- UpdatePayableStatus returns error

**Expected Output**:
- Error returned
- Payout items already updated (not rolled back)

**Edge Cases Covered**:
- Final step failure

---

### Test 3.3: commitPayoutItem - All Relations

**Test Name**: `TestCommitPayoutItem_AllRelations`

**Description**: Verify payout item commit with all relation types.

**Input/Setup**:
- Payout with both Invoice Split and Refund
- All updates succeed

**Expected Output**:
- Both relations updated
- Payout status updated
- No errors

**Edge Cases Covered**:
- Complete payout item update

---

### Test 3.4: commitPayoutItem - GetPayoutWithRelations Fails

**Test Name**: `TestCommitPayoutItem_GetPayoutFails`

**Description**: Verify error handling when fetching payout fails.

**Input/Setup**:
- GetPayoutWithRelations returns error

**Expected Output**:
- Error returned immediately
- No update attempts made

**Edge Cases Covered**:
- Fetch failures

---

### Test 3.5: commitPayoutItem - Payout Update Fails

**Test Name**: `TestCommitPayoutItem_PayoutUpdateFails`

**Description**: Verify error returned when payout status update fails.

**Input/Setup**:
- GetPayoutWithRelations succeeds
- Related record updates succeed
- UpdatePayoutStatus fails

**Expected Output**:
- Error returned
- Related records already updated

**Edge Cases Covered**:
- Payout update failure

---

## Test Execution Notes

### Testing Dependencies
- Use `github.com/stretchr/testify/require` for assertions
- Mock all Notion service methods
- Use `context.Background()` for test contexts

### Mock Strategy
- Mock Notion service methods at interface level
- Track method call counts and arguments
- Return different results per test scenario
- Verify correct parameters passed to mocked methods

### Coverage Goals
- 100% line coverage for controller code
- All business logic paths tested
- All error paths tested
- Cascade update logic thoroughly tested

### Test Execution
```bash
# Run controller tests only
go test ./pkg/controller/contractorpayables -v

# Run with coverage
go test ./pkg/controller/contractorpayables -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Key Testing Focus Areas
1. **PayDay Filtering Logic**: Ensure correct contractors included/excluded
2. **Cascade Update Sequence**: Verify correct order of operations
3. **Error Handling**: Test best-effort vs. fail-fast scenarios
4. **Aggregation Logic**: Verify counts and totals calculated correctly
5. **Edge Cases**: Empty results, partial failures, large datasets
