# Handler Unit Test Cases - Contractor Payables

## Overview

This document defines unit test cases for the Contractor Payables handler layer (`pkg/handler/contractorpayables/`). These tests focus on request validation, response formatting, and error handling without testing business logic.

## Test File Location

`pkg/handler/contractorpayables/contractorpayables_test.go`

## Test Setup Pattern

```go
// Mock controller interface
type mockController struct {
    previewCommitFunc func(ctx context.Context, month string, batch int) (*PreviewCommitResponse, error)
    commitPayablesFunc func(ctx context.Context, month string, batch int) (*CommitResponse, error)
}

// Test setup helper
func setupHandlerTest() (*handler, *mockController, *httptest.ResponseRecorder, *gin.Context) {
    gin.SetMode(gin.TestMode)
    mockCtrl := &mockController{}
    logger := logger.NewLogrusLogger("error")
    config := &config.Config{}

    h := New(mockCtrl, logger, config)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    return h.(*handler), mockCtrl, w, c
}
```

---

## Test Suite 1: PreviewCommit Endpoint

### Test 1.1: Valid Request - Success Response

**Test Name**: `TestPreviewCommit_ValidRequest_Success`

**Description**: Verify handler correctly processes valid query parameters and returns 200 OK with preview data.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=15`
- Mock controller returns:
  ```go
  &PreviewCommitResponse{
      Month: "2025-01",
      Batch: 15,
      Count: 3,
      TotalAmount: 15000.00,
      Contractors: []ContractorPreview{
          {Name: "John Doe", Amount: 5000.00, Currency: "USD", PayableID: "page-1"},
          {Name: "Jane Smith", Amount: 7500.00, Currency: "USD", PayableID: "page-2"},
          {Name: "Bob Wilson", Amount: 2500.00, Currency: "USD", PayableID: "page-3"},
      },
  }, nil
  ```

**Expected Output**:
- HTTP Status: 200 OK
- Response body contains `data` field with preview response
- Response structure matches `view.Response` format
- All contractor details present in response

**Edge Cases Covered**:
- Standard happy path with multiple contractors

---

### Test 1.2: Valid Request - Empty Results

**Test Name**: `TestPreviewCommit_ValidRequest_EmptyResults`

**Description**: Verify handler returns 200 OK with empty results when no pending payables found.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=15`
- Mock controller returns: `nil, errors.New("no pending payables found")`

**Expected Output**:
- HTTP Status: 200 OK
- Response body contains:
  ```json
  {
    "data": {
      "month": "2025-01",
      "batch": 15,
      "count": 0,
      "total_amount": 0,
      "contractors": []
    }
  }
  ```

**Edge Cases Covered**:
- No pending payables scenario (not treated as error)

---

### Test 1.3: Invalid Month Format - Missing Hyphen

**Test Name**: `TestPreviewCommit_InvalidMonthFormat_MissingHyphen`

**Description**: Verify handler rejects month format without hyphen.

**Input/Setup**:
- Query parameters: `month=202501&batch=15`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Response message: "Invalid month format. Use YYYY-MM (e.g., 2025-01)"
- Controller should NOT be called

**Edge Cases Covered**:
- Month format validation (missing separator)

---

### Test 1.4: Invalid Month Format - Wrong Length

**Test Name**: `TestPreviewCommit_InvalidMonthFormat_WrongLength`

**Description**: Verify handler rejects month format with incorrect length.

**Input/Setup**:
- Query parameters: `month=25-01&batch=15`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Response message: "Invalid month format. Use YYYY-MM (e.g., 2025-01)"
- Controller should NOT be called

**Edge Cases Covered**:
- Month format validation (incorrect year length)

---

### Test 1.5: Invalid Month Format - Invalid Characters

**Test Name**: `TestPreviewCommit_InvalidMonthFormat_InvalidCharacters`

**Description**: Verify handler rejects month format with non-numeric characters.

**Input/Setup**:
- Query parameters: `month=20XX-01&batch=15`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Response message: "Invalid month format. Use YYYY-MM (e.g., 2025-01)"
- Controller should NOT be called

**Edge Cases Covered**:
- Month format validation (invalid characters)

---

### Test 1.6: Invalid Batch Value - Zero

**Test Name**: `TestPreviewCommit_InvalidBatch_Zero`

**Description**: Verify handler rejects batch value of 0.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=0`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch must be 1 or 15

**Edge Cases Covered**:
- Batch validation (zero value)

---

### Test 1.7: Invalid Batch Value - Negative

**Test Name**: `TestPreviewCommit_InvalidBatch_Negative`

**Description**: Verify handler rejects negative batch values.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=-1`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch must be 1 or 15

**Edge Cases Covered**:
- Batch validation (negative number)

---

### Test 1.8: Invalid Batch Value - Non-PayDay

**Test Name**: `TestPreviewCommit_InvalidBatch_NonPayDay`

**Description**: Verify handler rejects batch values other than 1 or 15.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=10`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch must be 1 or 15

**Edge Cases Covered**:
- Batch validation (valid integer but not a PayDay)

---

### Test 1.9: Missing Required Parameter - Month

**Test Name**: `TestPreviewCommit_MissingParameter_Month`

**Description**: Verify handler rejects request with missing month parameter.

**Input/Setup**:
- Query parameters: `batch=15` (month omitted)

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating month is required

**Edge Cases Covered**:
- Required parameter validation

---

### Test 1.10: Missing Required Parameter - Batch

**Test Name**: `TestPreviewCommit_MissingParameter_Batch`

**Description**: Verify handler rejects request with missing batch parameter.

**Input/Setup**:
- Query parameters: `month=2025-01` (batch omitted)

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch is required

**Edge Cases Covered**:
- Required parameter validation

---

### Test 1.11: Controller Returns Error

**Test Name**: `TestPreviewCommit_ControllerError`

**Description**: Verify handler correctly handles controller errors (not "no pending payables").

**Input/Setup**:
- Query parameters: `month=2025-01&batch=15`
- Mock controller returns: `nil, errors.New("notion API error")`

**Expected Output**:
- HTTP Status: 500 Internal Server Error
- Response contains error details

**Edge Cases Covered**:
- Error propagation from controller layer

---

### Test 1.12: Single Contractor Result

**Test Name**: `TestPreviewCommit_SingleContractor`

**Description**: Verify handler correctly formats response with single contractor.

**Input/Setup**:
- Query parameters: `month=2025-01&batch=15`
- Mock controller returns single contractor in preview

**Expected Output**:
- HTTP Status: 200 OK
- Response contains array with one contractor
- Total amount equals single contractor amount

**Edge Cases Covered**:
- Single item in result list

---

### Test 1.13: Large Contractor List

**Test Name**: `TestPreviewCommit_LargeContractorList`

**Description**: Verify handler correctly handles large number of contractors (50+).

**Input/Setup**:
- Query parameters: `month=2025-01&batch=15`
- Mock controller returns 50+ contractors

**Expected Output**:
- HTTP Status: 200 OK
- Response contains all contractors
- Count matches number of contractors

**Edge Cases Covered**:
- Large result sets

---

## Test Suite 2: Commit Endpoint

### Test 2.1: Valid Request - Full Success

**Test Name**: `TestCommit_ValidRequest_FullSuccess`

**Description**: Verify handler correctly processes valid commit request with all payables updated successfully.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15
  }
  ```
- Mock controller returns:
  ```go
  &CommitResponse{
      Month: "2025-01",
      Batch: 15,
      Updated: 3,
      Failed: 0,
      Errors: nil,
  }, nil
  ```

**Expected Output**:
- HTTP Status: 200 OK
- Response body contains commit result with success counts
- Errors field is null or empty

**Edge Cases Covered**:
- Standard happy path with full success

---

### Test 2.2: Valid Request - Partial Success

**Test Name**: `TestCommit_ValidRequest_PartialSuccess`

**Description**: Verify handler returns 207 Multi-Status when some updates fail.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15
  }
  ```
- Mock controller returns:
  ```go
  &CommitResponse{
      Month: "2025-01",
      Batch: 15,
      Updated: 2,
      Failed: 1,
      Errors: []CommitError{
          {PayableID: "page-3", Error: "notion API timeout"},
      },
  }, nil
  ```

**Expected Output**:
- HTTP Status: 207 Multi-Status
- Response contains updated count, failed count, and error details
- Errors array contains specific failure information

**Edge Cases Covered**:
- Partial failure scenario

---

### Test 2.3: Invalid Request Body - Malformed JSON

**Test Name**: `TestCommit_InvalidRequestBody_MalformedJSON`

**Description**: Verify handler rejects malformed JSON.

**Input/Setup**:
- Request body: `{invalid json`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Response indicates JSON parsing error

**Edge Cases Covered**:
- Malformed request body

---

### Test 2.4: Invalid Request Body - Wrong Types

**Test Name**: `TestCommit_InvalidRequestBody_WrongTypes`

**Description**: Verify handler rejects request with wrong field types.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": 202501,
    "batch": "fifteen"
  }
  ```

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating type mismatch

**Edge Cases Covered**:
- Type validation

---

### Test 2.5: Invalid Month Format

**Test Name**: `TestCommit_InvalidMonthFormat`

**Description**: Verify handler validates month format in request body.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "01-2025",
    "batch": 15
  }
  ```

**Expected Output**:
- HTTP Status: 400 Bad Request
- Response message: "Invalid month format. Use YYYY-MM (e.g., 2025-01)"

**Edge Cases Covered**:
- Month format validation in POST body

---

### Test 2.6: Invalid Batch Value

**Test Name**: `TestCommit_InvalidBatch`

**Description**: Verify handler validates batch value in request body.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 20
  }
  ```

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch must be 1 or 15

**Edge Cases Covered**:
- Batch validation in POST body

---

### Test 2.7: No Pending Payables Found

**Test Name**: `TestCommit_NoPendingPayables`

**Description**: Verify handler returns 404 when no pending payables exist.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15
  }
  ```
- Mock controller returns: `nil, errors.New("no pending payables found for month 2025-01 batch 15")`

**Expected Output**:
- HTTP Status: 404 Not Found
- Response contains error message

**Edge Cases Covered**:
- Empty result set treated as 404 for POST

---

### Test 2.8: Controller Returns Generic Error

**Test Name**: `TestCommit_ControllerError`

**Description**: Verify handler correctly handles controller errors.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15
  }
  ```
- Mock controller returns: `nil, errors.New("notion API connection failed")`

**Expected Output**:
- HTTP Status: 500 Internal Server Error
- Response contains error details

**Edge Cases Covered**:
- Error propagation from controller

---

### Test 2.9: All Updates Failed

**Test Name**: `TestCommit_AllUpdatesFailed`

**Description**: Verify handler handles scenario where all updates failed (but controller didn't error out completely).

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15
  }
  ```
- Mock controller returns:
  ```go
  &CommitResponse{
      Month: "2025-01",
      Batch: 15,
      Updated: 0,
      Failed: 3,
      Errors: []CommitError{
          {PayableID: "page-1", Error: "API error"},
          {PayableID: "page-2", Error: "API error"},
          {PayableID: "page-3", Error: "API error"},
      },
  }, nil
  ```

**Expected Output**:
- HTTP Status: 207 Multi-Status (since Failed > 0)
- Response shows 0 updated, 3 failed with error details

**Edge Cases Covered**:
- Complete failure (but not controller error)

---

### Test 2.10: Missing Required Fields

**Test Name**: `TestCommit_MissingRequiredFields`

**Description**: Verify handler rejects request with missing required fields.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01"
  }
  ```

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation error indicating batch is required

**Edge Cases Covered**:
- Required field validation

---

### Test 2.11: Empty Request Body

**Test Name**: `TestCommit_EmptyRequestBody`

**Description**: Verify handler rejects completely empty request body.

**Input/Setup**:
- Request body: `{}`

**Expected Output**:
- HTTP Status: 400 Bad Request
- Validation errors for missing month and batch

**Edge Cases Covered**:
- Empty request body

---

### Test 2.12: Extra Fields in Request

**Test Name**: `TestCommit_ExtraFields`

**Description**: Verify handler ignores extra fields in request body.

**Input/Setup**:
- Request body:
  ```json
  {
    "month": "2025-01",
    "batch": 15,
    "extra_field": "ignored"
  }
  ```
- Mock controller returns success response

**Expected Output**:
- HTTP Status: 200 OK
- Extra fields are ignored, request processed successfully

**Edge Cases Covered**:
- Forward compatibility with extra fields

---

## Test Suite 3: Helper Functions

### Test 3.1: Month Format Validation - Valid Formats

**Test Name**: `TestIsValidMonthFormat_ValidFormats`

**Description**: Verify month format validator accepts valid formats.

**Input/Setup**:
- Test inputs:
  - "2025-01"
  - "2024-12"
  - "2023-06"

**Expected Output**:
- All inputs return true

**Edge Cases Covered**:
- Various valid month formats

---

### Test 3.2: Month Format Validation - Invalid Formats

**Test Name**: `TestIsValidMonthFormat_InvalidFormats`

**Description**: Verify month format validator rejects invalid formats.

**Input/Setup**:
- Test inputs:
  - "202501" (missing hyphen)
  - "25-01" (wrong year length)
  - "2025-1" (wrong month length)
  - "2025-13" (invalid month number - optional, depends on implementation)
  - "2025/01" (wrong separator)
  - "" (empty string)

**Expected Output**:
- All inputs return false

**Edge Cases Covered**:
- Various invalid formats

---

## Test Execution Notes

### Testing Dependencies
- Use `github.com/stretchr/testify/require` for assertions
- Use `net/http/httptest` for HTTP testing
- Use `github.com/gin-gonic/gin` test mode

### Mock Strategy
- Mock controller interface only (no Notion service mocking at handler level)
- Focus on handler's request/response transformation
- Verify controller is called with correct parameters

### Coverage Goals
- 100% line coverage for handler code
- All validation paths tested
- All error paths tested
- All response formatting paths tested

### Test Execution
```bash
# Run handler tests only
go test ./pkg/handler/contractorpayables -v

# Run with coverage
go test ./pkg/handler/contractorpayables -coverprofile=coverage.out
go tool cover -html=coverage.out
```
