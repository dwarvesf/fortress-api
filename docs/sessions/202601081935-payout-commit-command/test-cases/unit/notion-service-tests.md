# Notion Service Unit Test Cases - Payout Commit

## Overview

This document defines unit test cases for Notion service methods required for payout commit functionality. Tests focus on Notion API interactions, property extraction, pagination handling, and error scenarios with mocked Notion client.

## Test Files Location

- `pkg/service/notion/contractor_payables_test.go`
- `pkg/service/notion/contractor_payouts_test.go`
- `pkg/service/notion/invoice_split_test.go`
- `pkg/service/notion/refund_requests_test.go`

## Test Setup Pattern

```go
// Mock Notion client
type mockNotionClient struct {
    queryDatabaseFunc   func(ctx context.Context, dbID string, query *nt.DatabaseQuery) (*nt.DatabaseQueryResponse, error)
    findPageByIDFunc    func(ctx context.Context, pageID string) (*nt.Page, error)
    updatePageFunc      func(ctx context.Context, pageID string, params nt.UpdatePageParams) (*nt.Page, error)
}

// Test setup helper
func setupServiceTest() (*ContractorPayablesService, *mockNotionClient) {
    mockClient := &mockNotionClient{}
    logger := logger.NewLogrusLogger("error")
    config := &config.Config{
        Notion: config.NotionConfig{
            Databases: config.Databases{
                ContractorPayables: "db-payables-id",
                ContractorPayouts:  "db-payouts-id",
                InvoiceSplit:       "db-split-id",
                RefundRequest:      "db-refund-id",
                ServiceRate:        "db-rate-id",
            },
        },
    }

    service := &ContractorPayablesService{
        client: mockClient,
        logger: logger,
        config: config,
    }

    return service, mockClient
}
```

---

## Test Suite 1: ContractorPayablesService - QueryPendingPayablesByPeriod

### Test 1.1: Query Success - Multiple Results

**Test Name**: `TestQueryPendingPayablesByPeriod_Success_MultipleResults`

**Description**: Verify service correctly queries and parses multiple pending payables.

**Input/Setup**:
- Period: "2025-01-01"
- Mock QueryDatabase returns 3 pages with properties:
  ```go
  {
      PageID: "page-1",
      Properties: {
          "Payment Status": {Status: {Name: "Pending"}},
          "Period": {Date: {Start: "2025-01-01"}},
          "Contractor": {Relation: [{ID: "contractor-1"}]},
          "Total": {Number: 5000.00},
          "Currency": {Select: {Name: "USD"}},
          "Payout Items": {Relation: [{ID: "payout-1"}, {ID: "payout-2"}]},
      }
  }
  ```

**Expected Output**:
- No error returned
- Result contains 3 PendingPayable items
- All properties correctly extracted
- Payout Items array contains multiple IDs

**Edge Cases Covered**:
- Multiple results parsing
- Multiple relations in Payout Items

---

### Test 1.2: Query Success - Single Result

**Test Name**: `TestQueryPendingPayablesByPeriod_Success_SingleResult`

**Description**: Verify service handles single result correctly.

**Input/Setup**:
- Period: "2025-01-01"
- Mock QueryDatabase returns 1 page

**Expected Output**:
- Result slice contains 1 item
- All properties correctly extracted

**Edge Cases Covered**:
- Single result handling

---

### Test 1.3: Query Success - Empty Results

**Test Name**: `TestQueryPendingPayablesByPeriod_Success_EmptyResults`

**Description**: Verify service returns empty slice when no results found.

**Input/Setup**:
- Period: "2025-01-01"
- Mock QueryDatabase returns empty Results array

**Expected Output**:
- No error returned
- Result is empty slice (not nil)

**Edge Cases Covered**:
- Empty result set

---

### Test 1.4: Query Success - Pagination Handling

**Test Name**: `TestQueryPendingPayablesByPeriod_Pagination`

**Description**: Verify service correctly handles paginated results.

**Input/Setup**:
- Period: "2025-01-01"
- Mock QueryDatabase returns:
  - First call: 100 results, HasMore=true, NextCursor="cursor-1"
  - Second call: 50 results, HasMore=false

**Expected Output**:
- QueryDatabase called twice
- Second call includes StartCursor="cursor-1"
- Result contains all 150 items

**Edge Cases Covered**:
- Pagination with multiple pages
- Cursor handling

---

### Test 1.5: Query Filter Construction

**Test Name**: `TestQueryPendingPayablesByPeriod_FilterConstruction`

**Description**: Verify correct filter is constructed for query.

**Input/Setup**:
- Period: "2025-01-01"
- Capture QueryDatabase call arguments

**Expected Output**:
- Filter contains AND condition with:
  - Payment Status Status equals "Pending"
  - Period Date equals "2025-01-01"
- Page size set to 100

**Edge Cases Covered**:
- Filter construction logic

---

### Test 1.6: Property Extraction - Missing Contractor Relation

**Test Name**: `TestQueryPendingPayablesByPeriod_MissingContractorRelation`

**Description**: Verify service handles missing Contractor relation gracefully.

**Input/Setup**:
- Period: "2025-01-01"
- Mock returns page with empty Contractor relation array

**Expected Output**:
- ContractorPageID is empty string
- No error returned (or logged warning)
- Other properties still extracted

**Edge Cases Covered**:
- Missing required relation

---

### Test 1.7: Property Extraction - Empty Payout Items

**Test Name**: `TestQueryPendingPayablesByPeriod_EmptyPayoutItems`

**Description**: Verify service handles payable with no payout items.

**Input/Setup**:
- Period: "2025-01-01"
- Mock returns page with empty Payout Items relation

**Expected Output**:
- PayoutItemPageIDs is empty array
- No error returned

**Edge Cases Covered**:
- Empty relation array

---

### Test 1.8: Notion API Error

**Test Name**: `TestQueryPendingPayablesByPeriod_NotionAPIError`

**Description**: Verify service propagates Notion API errors.

**Input/Setup**:
- Period: "2025-01-01"
- Mock QueryDatabase returns error: "notion API rate limit"

**Expected Output**:
- Error returned wrapping original error
- Error message contains context

**Edge Cases Covered**:
- API error handling

---

### Test 1.9: Invalid Period Format

**Test Name**: `TestQueryPendingPayablesByPeriod_InvalidPeriodFormat`

**Description**: Verify service handles invalid period format (if validation implemented).

**Input/Setup**:
- Period: "invalid-date"

**Expected Output**:
- Error returned or invalid filter constructed (depends on implementation)

**Edge Cases Covered**:
- Input validation

---

### Test 1.10: Property Type Casting Errors

**Test Name**: `TestQueryPendingPayablesByPeriod_PropertyTypeCastError`

**Description**: Verify service handles unexpected property types gracefully.

**Input/Setup**:
- Period: "2025-01-01"
- Mock returns page with unexpected property type (e.g., Total as string instead of number)

**Expected Output**:
- Error logged or default value used
- Processing continues (best-effort)

**Edge Cases Covered**:
- Property type mismatches

---

## Test Suite 2: ContractorPayablesService - UpdatePayableStatus

### Test 2.1: Update Success

**Test Name**: `TestUpdatePayableStatus_Success`

**Description**: Verify service correctly updates payable status and payment date.

**Input/Setup**:
- PageID: "page-1"
- Status: "Paid"
- PaymentDate: "2025-01-15"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- UpdatePage called with correct parameters:
  - Payment Status: Status type with "Paid"
  - Payment Date: Date type with "2025-01-15"

**Edge Cases Covered**:
- Standard update operation

---

### Test 2.2: Update Parameters Construction

**Test Name**: `TestUpdatePayableStatus_ParametersConstruction`

**Description**: Verify correct update parameters are constructed.

**Input/Setup**:
- PageID: "page-1"
- Status: "Paid"
- PaymentDate: "2025-01-15"
- Capture UpdatePage call arguments

**Expected Output**:
- Parameters contain:
  - Payment Status using Status property type
  - Payment Date using Date property type
  - Date is nt.NewDateTime with correct format

**Edge Cases Covered**:
- Parameter construction logic

---

### Test 2.3: Empty Page ID

**Test Name**: `TestUpdatePayableStatus_EmptyPageID`

**Description**: Verify service handles empty page ID.

**Input/Setup**:
- PageID: ""
- Status: "Paid"
- PaymentDate: "2025-01-15"

**Expected Output**:
- Error returned before API call
- UpdatePage NOT called

**Edge Cases Covered**:
- Input validation

---

### Test 2.4: Notion API Error

**Test Name**: `TestUpdatePayableStatus_NotionAPIError`

**Description**: Verify service propagates Notion API errors.

**Input/Setup**:
- PageID: "page-1"
- Status: "Paid"
- PaymentDate: "2025-01-15"
- Mock UpdatePage returns error: "page not found"

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

### Test 2.5: Idempotency

**Test Name**: `TestUpdatePayableStatus_Idempotency`

**Description**: Verify update can be safely repeated.

**Input/Setup**:
- PageID: "page-1" (already has Status="Paid")
- Status: "Paid"
- PaymentDate: "2025-01-15"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- Update succeeds even if status already "Paid"

**Edge Cases Covered**:
- Idempotent operations

---

## Test Suite 3: ContractorPayablesService - GetContractorPayDay

### Test 3.1: Query Success - PayDay Found

**Test Name**: `TestGetContractorPayDay_Success`

**Description**: Verify service correctly retrieves PayDay from Service Rate.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock QueryDatabase returns 1 result with:
  ```go
  Properties: {
      "PayDay": {Select: {Name: "15"}},
  }
  ```

**Expected Output**:
- No error returned
- PayDay returned as integer: 15

**Edge Cases Covered**:
- Standard query and extraction

---

### Test 3.2: Query Success - PayDay=1

**Test Name**: `TestGetContractorPayDay_PayDayOne`

**Description**: Verify service handles PayDay=1 correctly.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock returns PayDay: "1"

**Expected Output**:
- No error returned
- PayDay returned as integer: 1

**Edge Cases Covered**:
- Alternative PayDay value

---

### Test 3.3: No Service Rate Found

**Test Name**: `TestGetContractorPayDay_NoServiceRate`

**Description**: Verify service returns error when no Service Rate found.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock QueryDatabase returns empty results

**Expected Output**:
- Error returned: "no service rate found for contractor contractor-1"

**Edge Cases Covered**:
- Missing Service Rate record

---

### Test 3.4: PayDay Property Missing

**Test Name**: `TestGetContractorPayDay_PayDayPropertyMissing`

**Description**: Verify service handles missing PayDay property.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock QueryDatabase returns result without PayDay property

**Expected Output**:
- Error returned: "PayDay property not found"

**Edge Cases Covered**:
- Missing property

---

### Test 3.5: PayDay Invalid Value

**Test Name**: `TestGetContractorPayDay_InvalidPayDayValue`

**Description**: Verify service handles non-numeric PayDay value.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock returns PayDay: "invalid"

**Expected Output**:
- Error returned: "invalid PayDay value: invalid"

**Edge Cases Covered**:
- Invalid property value

---

### Test 3.6: Query Filter Construction

**Test Name**: `TestGetContractorPayDay_FilterConstruction`

**Description**: Verify correct filter is constructed for Service Rate query.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Capture QueryDatabase call arguments

**Expected Output**:
- Filter contains:
  - Property: "Contractor"
  - Relation Contains: "contractor-1"
- Page size: 1

**Edge Cases Covered**:
- Query construction

---

### Test 3.7: Multiple Service Rates

**Test Name**: `TestGetContractorPayDay_MultipleServiceRates`

**Description**: Verify service uses first result when multiple found.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock QueryDatabase returns 2 results with different PayDay values

**Expected Output**:
- No error returned
- PayDay from first result returned

**Edge Cases Covered**:
- Multiple results handling

---

### Test 3.8: Notion API Error

**Test Name**: `TestGetContractorPayDay_NotionAPIError`

**Description**: Verify service propagates Notion API errors.

**Input/Setup**:
- ContractorPageID: "contractor-1"
- Mock QueryDatabase returns error

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

### Test 3.9: Empty Contractor ID

**Test Name**: `TestGetContractorPayDay_EmptyContractorID`

**Description**: Verify service handles empty contractor ID.

**Input/Setup**:
- ContractorPageID: ""

**Expected Output**:
- Error returned before API call

**Edge Cases Covered**:
- Input validation

---

## Test Suite 4: ContractorPayoutsService - GetPayoutWithRelations

### Test 4.1: Payout With Invoice Split Only

**Test Name**: `TestGetPayoutWithRelations_InvoiceSplitOnly`

**Description**: Verify service extracts Invoice Split relation correctly.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns:
  ```go
  Properties: {
      "Status": {Status: {Name: "Pending"}},
      "02 Invoice Split": {Relation: [{ID: "split-1"}]},
      "01 Refund": {Relation: []},
  }
  ```

**Expected Output**:
- No error returned
- Result contains:
  - PageID: "payout-1"
  - Status: "Pending"
  - InvoiceSplitID: "split-1"
  - RefundRequestID: ""

**Edge Cases Covered**:
- Invoice Split relation only

---

### Test 4.2: Payout With Refund Only

**Test Name**: `TestGetPayoutWithRelations_RefundOnly`

**Description**: Verify service extracts Refund relation correctly.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns:
  ```go
  Properties: {
      "Status": {Status: {Name: "Pending"}},
      "02 Invoice Split": {Relation: []},
      "01 Refund": {Relation: [{ID: "refund-1"}]},
  }
  ```

**Expected Output**:
- Result contains:
  - InvoiceSplitID: ""
  - RefundRequestID: "refund-1"

**Edge Cases Covered**:
- Refund relation only

---

### Test 4.3: Payout With Both Relations

**Test Name**: `TestGetPayoutWithRelations_BothRelations`

**Description**: Verify service extracts both relations when present.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns both relations populated

**Expected Output**:
- Result contains both InvoiceSplitID and RefundRequestID

**Edge Cases Covered**:
- Both relations populated

---

### Test 4.4: Payout With No Relations

**Test Name**: `TestGetPayoutWithRelations_NoRelations`

**Description**: Verify service handles payout with no relations.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns empty relation arrays

**Expected Output**:
- No error returned
- Result contains:
  - InvoiceSplitID: ""
  - RefundRequestID: ""

**Edge Cases Covered**:
- No relations present

---

### Test 4.5: Empty Page ID

**Test Name**: `TestGetPayoutWithRelations_EmptyPageID`

**Description**: Verify service handles empty page ID.

**Input/Setup**:
- PayoutPageID: ""

**Expected Output**:
- Error returned before API call

**Edge Cases Covered**:
- Input validation

---

### Test 4.6: Notion API Error

**Test Name**: `TestGetPayoutWithRelations_NotionAPIError`

**Description**: Verify service propagates API errors.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns error: "page not found"

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

### Test 4.7: Property Type Cast Error

**Test Name**: `TestGetPayoutWithRelations_PropertyTypeCastError`

**Description**: Verify service handles property type cast failures.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns page with properties not castable to DatabasePageProperties

**Expected Output**:
- Error returned: "failed to cast payout page properties"

**Edge Cases Covered**:
- Property type casting

---

### Test 4.8: Multiple Relation IDs

**Test Name**: `TestGetPayoutWithRelations_MultipleRelationIDs`

**Description**: Verify service uses first relation ID when multiple present.

**Input/Setup**:
- PayoutPageID: "payout-1"
- Mock FindPageByID returns Invoice Split with 2 relation IDs

**Expected Output**:
- InvoiceSplitID is first relation ID

**Edge Cases Covered**:
- Multiple relation IDs handling

---

## Test Suite 5: ContractorPayoutsService - UpdatePayoutStatus

### Test 5.1: Update Success

**Test Name**: `TestUpdatePayoutStatus_Success`

**Description**: Verify service correctly updates payout status.

**Input/Setup**:
- PageID: "payout-1"
- Status: "Paid"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- UpdatePage called with Status property type

**Edge Cases Covered**:
- Standard update operation

---

### Test 5.2: Property Type - Status Not Select

**Test Name**: `TestUpdatePayoutStatus_PropertyType`

**Description**: Verify service uses Status property type (not Select).

**Input/Setup**:
- PageID: "payout-1"
- Status: "Paid"
- Capture UpdatePage parameters

**Expected Output**:
- Parameters use `.Status` field (not `.Select`)

**Edge Cases Covered**:
- Correct property type usage

---

### Test 5.3: Empty Page ID

**Test Name**: `TestUpdatePayoutStatus_EmptyPageID`

**Description**: Verify service handles empty page ID.

**Input/Setup**:
- PageID: ""
- Status: "Paid"

**Expected Output**:
- Error returned before API call

**Edge Cases Covered**:
- Input validation

---

### Test 5.4: Notion API Error

**Test Name**: `TestUpdatePayoutStatus_NotionAPIError`

**Description**: Verify service propagates API errors.

**Input/Setup**:
- PageID: "payout-1"
- Status: "Paid"
- Mock UpdatePage returns error

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

## Test Suite 6: InvoiceSplitService - UpdateInvoiceSplitStatus

### Test 6.1: Update Success

**Test Name**: `TestUpdateInvoiceSplitStatus_Success`

**Description**: Verify service correctly updates invoice split status.

**Input/Setup**:
- PageID: "split-1"
- Status: "Paid"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- UpdatePage called with Select property type

**Edge Cases Covered**:
- Standard update operation

---

### Test 6.2: Property Type - Select Not Status

**Test Name**: `TestUpdateInvoiceSplitStatus_PropertyType`

**Description**: Verify service uses Select property type (CRITICAL difference).

**Input/Setup**:
- PageID: "split-1"
- Status: "Paid"
- Capture UpdatePage parameters

**Expected Output**:
- Parameters use `.Select` field (NOT `.Status`)

**Edge Cases Covered**:
- Correct property type for Invoice Split

---

### Test 6.3: Empty Page ID

**Test Name**: `TestUpdateInvoiceSplitStatus_EmptyPageID`

**Description**: Verify service handles empty page ID.

**Input/Setup**:
- PageID: ""
- Status: "Paid"

**Expected Output**:
- Error returned before API call

**Edge Cases Covered**:
- Input validation

---

### Test 6.4: Notion API Error

**Test Name**: `TestUpdateInvoiceSplitStatus_NotionAPIError`

**Description**: Verify service propagates API errors.

**Input/Setup**:
- PageID: "split-1"
- Status: "Paid"
- Mock UpdatePage returns error

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

## Test Suite 7: RefundRequestsService - UpdateRefundRequestStatus

### Test 7.1: Update Success

**Test Name**: `TestUpdateRefundRequestStatus_Success`

**Description**: Verify service correctly updates refund request status.

**Input/Setup**:
- PageID: "refund-1"
- Status: "Paid"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- UpdatePage called with Status property type

**Edge Cases Covered**:
- Standard update operation

---

### Test 7.2: Property Type - Status Not Select

**Test Name**: `TestUpdateRefundRequestStatus_PropertyType`

**Description**: Verify service uses Status property type.

**Input/Setup**:
- PageID: "refund-1"
- Status: "Paid"
- Capture UpdatePage parameters

**Expected Output**:
- Parameters use `.Status` field (not `.Select`)

**Edge Cases Covered**:
- Correct property type usage

---

### Test 7.3: Status Transition - Approved to Paid

**Test Name**: `TestUpdateRefundRequestStatus_ApprovedToPaid`

**Description**: Verify service handles status transition from Approved to Paid.

**Input/Setup**:
- PageID: "refund-1"
- Status: "Paid"
- Mock UpdatePage succeeds

**Expected Output**:
- No error returned
- Update succeeds regardless of previous status

**Edge Cases Covered**:
- Status transition logic

---

### Test 7.4: Empty Page ID

**Test Name**: `TestUpdateRefundRequestStatus_EmptyPageID`

**Description**: Verify service handles empty page ID.

**Input/Setup**:
- PageID: ""
- Status: "Paid"

**Expected Output**:
- Error returned before API call

**Edge Cases Covered**:
- Input validation

---

### Test 7.5: Notion API Error

**Test Name**: `TestUpdateRefundRequestStatus_NotionAPIError`

**Description**: Verify service propagates API errors.

**Input/Setup**:
- PageID: "refund-1"
- Status: "Paid"
- Mock UpdatePage returns error

**Expected Output**:
- Error returned wrapping original error

**Edge Cases Covered**:
- API error handling

---

## Test Suite 8: Cross-Service Property Type Verification

### Test 8.1: Property Type Consistency Check

**Test Name**: `TestPropertyTypes_ConsistencyCheck`

**Description**: Verify all services use correct property types.

**Input/Setup**:
- Create update calls for all services
- Capture parameters for each

**Expected Output**:
- Contractor Payables: `.Status` for Payment Status
- Contractor Payouts: `.Status` for Status
- Invoice Split: `.Select` for Status (DIFFERENT)
- Refund Requests: `.Status` for Status

**Edge Cases Covered**:
- Property type consistency across services

---

## Test Execution Notes

### Testing Dependencies
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Mock Notion client interface
- Use real Notion SDK types for responses

### Mock Strategy
- Mock `QueryDatabase`, `FindPageByID`, `UpdatePage` methods
- Return properly structured Notion API response objects
- Verify correct database IDs and parameters passed

### Coverage Goals
- 100% line coverage for each service method
- All property extraction paths tested
- All error paths tested
- Pagination logic thoroughly tested

### Key Testing Focus Areas
1. **Property Type Differences**: Critical to test Select vs Status usage
2. **Pagination Logic**: Ensure cursor handling works correctly
3. **Relation Extraction**: Test single, multiple, and empty relations
4. **Error Propagation**: Verify errors wrapped with context
5. **Date Formatting**: Verify date properties formatted correctly

### Test Execution
```bash
# Run all Notion service tests
go test ./pkg/service/notion -v

# Run specific service tests
go test ./pkg/service/notion -run TestContractorPayables -v
go test ./pkg/service/notion -run TestContractorPayouts -v
go test ./pkg/service/notion -run TestInvoiceSplit -v
go test ./pkg/service/notion -run TestRefundRequests -v

# Run with coverage
go test ./pkg/service/notion -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Critical Test Cases
The following test cases are CRITICAL and must not fail:
1. Property type tests (Select vs Status) - prevents Notion API rejections
2. Pagination handling - prevents data loss with large result sets
3. Empty relation handling - prevents nil pointer errors
4. API error propagation - ensures proper error handling up the stack
