# Unit Test Cases: Invoice Due Date Calculation

**Component**: `pkg/handler/notion/task_order_log.go`
**Method**: `SendTaskOrderConfirmation` (invoice due date calculation logic)
**Specification**: SPEC-003
**Test Framework**: Go testing, testify/require, gomock

## Test Overview

This test suite validates the invoice due date calculation logic in the handler layer, ensuring correct mapping from Payday values to invoice due day strings.

## Test Structure

### Test File Location
`pkg/handler/notion/task_order_log_test.go`

### Mock Dependencies
- Notion service (`service.Notion` with `GetContractorPayday` method)
- GoogleMail service (`service.GoogleMail`)
- Logger (`logger.Logger`)
- Config (`config.Config`)

## Test Cases

### TC-013: Payday 1 Calculates to "10th"

**Description**: Verify that Payday value 1 correctly calculates to invoice due day "10th"

**Setup**:
```go
// Mock service returns Payday=1
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), "contractor-page-123").
    Return(1, nil)

// Mock email sending to capture email data
mockGoogleMail.EXPECT().
    SendTaskOrderConfirmation(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, data model.TaskOrderConfirmationEmail) error {
        // Capture and verify email data
        capturedEmailData = data
        return nil
    })
```

**Input**:
- HTTP request to `/api/v1/cronjobs/send-task-order-confirmation`
- Query params: `month=2026-01&discord=test-contractor`
- Mock Notion task order log with contractor relation

**Expected Behavior**:
- Handler calls `GetContractorPayday` with contractor page ID
- Receives payday=1
- Calculates `invoiceDueDay = "10th"`
- Populates email struct with `InvoiceDueDay: "10th"`
- Logs debug message: "calculated invoice due day" with payday=1, invoice_due_day="10th"

**Assertions**:
```go
require.Equal(t, "10th", capturedEmailData.InvoiceDueDay)
require.Equal(t, http.StatusOK, response.Code)
```

---

### TC-014: Payday 15 Calculates to "25th"

**Description**: Verify that Payday value 15 correctly calculates to invoice due day "25th"

**Setup**:
```go
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), "contractor-page-456").
    Return(15, nil)

mockGoogleMail.EXPECT().
    SendTaskOrderConfirmation(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, data model.TaskOrderConfirmationEmail) error {
        capturedEmailData = data
        return nil
    })
```

**Input**:
- HTTP request with contractor having Payday=15
- Mock task order log data

**Expected Behavior**:
- Handler receives payday=15
- Calculates `invoiceDueDay = "25th"`
- Populates email struct with `InvoiceDueDay: "25th"`
- Logs debug message with payday=15, invoice_due_day="25th"

**Assertions**:
```go
require.Equal(t, "25th", capturedEmailData.InvoiceDueDay)
require.Equal(t, http.StatusOK, response.Code)
```

---

### TC-015: Payday 0 (Missing) Defaults to "10th"

**Description**: Verify that Payday value 0 (fallback signal) correctly defaults to "10th"

**Setup**:
```go
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), "contractor-page-no-payday").
    Return(0, nil)

mockGoogleMail.EXPECT().
    SendTaskOrderConfirmation(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, data model.TaskOrderConfirmationEmail) error {
        capturedEmailData = data
        return nil
    })
```

**Input**:
- HTTP request with contractor having no Payday configured
- Service returns 0 to signal fallback

**Expected Behavior**:
- Handler receives payday=0
- Applies default: `invoiceDueDay = "10th"`
- Populates email struct with `InvoiceDueDay: "10th"`
- Email sends successfully

**Assertions**:
```go
require.Equal(t, "10th", capturedEmailData.InvoiceDueDay)
require.Equal(t, http.StatusOK, response.Code)
```

**Rationale**: Per ADR-002, missing Payday should default to "10th" (Payday 1 behavior).

---

### TC-016: Service Error Falls Back to "10th"

**Description**: Verify graceful fallback when GetContractorPayday returns error

**Setup**:
```go
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), gomock.Any()).
    Return(0, errors.New("notion api timeout"))

mockGoogleMail.EXPECT().
    SendTaskOrderConfirmation(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, data model.TaskOrderConfirmationEmail) error {
        capturedEmailData = data
        return nil
    })
```

**Input**:
- HTTP request triggering Payday fetch
- Service returns error

**Expected Behavior**:
- Handler catches error from GetContractorPayday
- Sets `payday = 0` for fallback
- Calculates `invoiceDueDay = "10th"`
- Email sends successfully (error doesn't block)
- Logs debug message about error

**Assertions**:
```go
require.Equal(t, "10th", capturedEmailData.InvoiceDueDay)
require.Equal(t, http.StatusOK, response.Code)
```

---

### TC-017: Invalid Payday Value Falls Back to "10th"

**Description**: Verify handler gracefully handles unexpected Payday values

**Setup**:
```go
// Mock service returns unexpected value (should not happen, but defensive)
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), gomock.Any()).
    Return(99, nil) // Invalid value
```

**Input**:
- Service returns invalid payday value (not 0, 1, or 15)

**Expected Behavior**:
- Handler logic: `if payday == 15 then "25th" else "10th"`
- Invalid value doesn't match 15, so defaults to "10th"
- Email sends successfully

**Assertions**:
```go
require.Equal(t, "10th", capturedEmailData.InvoiceDueDay)
require.Equal(t, http.StatusOK, response.Code)
```

**Note**: This should not occur if SPEC-002 is implemented correctly, but handler should be defensive.

---

## Integration with Email Struct

### TC-018: Email Struct Population

**Description**: Verify complete email struct includes new InvoiceDueDay field

**Setup**:
```go
mockNotionService.EXPECT().
    GetContractorPayday(gomock.Any(), gomock.Any()).
    Return(15, nil)

mockGoogleMail.EXPECT().
    SendTaskOrderConfirmation(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, data model.TaskOrderConfirmationEmail) error {
        capturedEmailData = data
        return nil
    })
```

**Expected Email Struct Fields**:
```go
capturedEmailData = model.TaskOrderConfirmationEmail{
    ContractorName:     "John Smith",       // Existing
    ContractorLastName: "Smith",            // Existing
    TeamEmail:          "john@example.com", // Existing
    Month:              "2026-01",          // Existing
    Clients:            []model.TaskOrderClient{...}, // Existing
    InvoiceDueDay:      "25th",             // NEW
    Milestones:         []string{...},      // NEW
}
```

**Assertions**:
```go
require.NotEmpty(t, capturedEmailData.ContractorName)
require.NotEmpty(t, capturedEmailData.TeamEmail)
require.NotEmpty(t, capturedEmailData.Month)
require.Equal(t, "25th", capturedEmailData.InvoiceDueDay)
require.NotNil(t, capturedEmailData.Milestones)
```

---

## Logging Verification

### TC-019: Debug Logging for Calculation

**Description**: Verify appropriate debug logs are generated during calculation

**Setup**:
- Use mock logger to capture log calls
- Verify log messages and structured fields

**Expected Logs**:

1. **After Payday Fetch**:
```go
"fetched contractor payday"
Fields: contractor_id, payday
```

2. **After Calculation**:
```go
"calculated invoice due day"
Fields: payday, invoice_due_day
```

**Verification**:
```go
logs := mockLogger.GetDebugLogs()
require.Contains(t, logs, "fetched contractor payday")
require.Contains(t, logs, "calculated invoice due day")
require.Contains(t, logs[1].Fields, "invoice_due_day")
```

---

## Edge Cases

### TC-020: Contractor Page ID Extraction Failure

**Description**: Verify fallback when contractor page ID cannot be extracted from task order log

**Setup**:
```go
// Mock task order log with missing contractor relation
mockTaskOrderLog := &notionapi.Page{
    Properties: notionapi.Properties{
        "Contractor": &notionapi.RelationProperty{
            Relation: []notionapi.Relation{}, // Empty
        },
    },
}
```

**Expected Behavior**:
- Handler detects empty contractor relation
- Returns error before calling GetContractorPayday
- OR: Uses default payday=0, calculates "10th"

**Assertions**:
```go
// Option 1: Returns error
require.Equal(t, http.StatusBadRequest, response.Code)

// Option 2: Falls back gracefully (preferred per ADR-002)
require.Equal(t, http.StatusOK, response.Code)
require.Equal(t, "10th", capturedEmailData.InvoiceDueDay)
```

**Note**: Check existing handler error handling for contractor relation.

---

## Calculation Logic Table

| Payday Value | Condition | Invoice Due Day | Scenario |
|--------------|-----------|-----------------|----------|
| 0 | Default branch | "10th" | Missing/invalid/error |
| 1 | Default branch | "10th" | Valid Payday "01" |
| 15 | `if payday == 15` | "25th" | Valid Payday "15" |
| 99 | Default branch | "10th" | Invalid (defensive) |

**Implementation Logic**:
```go
invoiceDueDay := "10th" // Default for Payday 1 or fallback
if payday == 15 {
    invoiceDueDay = "25th"
}
```

## Test Data Requirements

### Mock Contractor Data

```go
// Contractor with Payday 1
contractorWithPayday1 := createMockContractor("contractor-page-123", "John Smith", "john@example.com")

// Contractor with Payday 15
contractorWithPayday15 := createMockContractor("contractor-page-456", "Jane Doe", "jane@example.com")

// Contractor without Payday
contractorNoPayday := createMockContractor("contractor-page-no-payday", "Bob Lee", "bob@example.com")
```

### Mock Task Order Log Data

```go
func createMockTaskOrderLog(contractorPageID string) *notionapi.Page {
    return &notionapi.Page{
        ID: "task-order-log-id",
        Properties: notionapi.Properties{
            "Contractor": &notionapi.RelationProperty{
                Relation: []notionapi.Relation{
                    {ID: contractorPageID},
                },
            },
            "Month": &notionapi.TitleProperty{
                Title: []notionapi.RichText{
                    {PlainText: "2026-01"},
                },
            },
            // ... other properties
        },
    }
}
```

## Mock Setup Example

```go
func setupHandlerTest(t *testing.T) (*handler, *httptest.ResponseRecorder, *gin.Context) {
    gin.SetMode(gin.TestMode)

    mockNotionService := mock.NewMockNotionService(t)
    mockGoogleMail := mock.NewMockGoogleMailService(t)
    mockLogger := logger.NewLogrusLogger("debug")

    h := &handler{
        service: &service.Service{
            Notion:     mockNotionService,
            GoogleMail: mockGoogleMail,
        },
        logger: mockLogger,
        config: &config.Config{},
    }

    recorder := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(recorder)

    return h, recorder, c
}
```

## Test Execution

### Run All Calculation Tests
```bash
go test -v ./pkg/handler/notion -run TestSendTaskOrderConfirmation.*DueDay
```

### Run Single Test
```bash
go test -v ./pkg/handler/notion -run TestSendTaskOrderConfirmation_Payday01
```

### Run with Coverage
```bash
go test -cover ./pkg/handler/notion -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Expected Coverage

- **Line Coverage**: 100% of due day calculation logic
- **Branch Coverage**: Both branches (payday==15, else)
- **Error Paths**: Service error handling
- **Integration**: Full email struct population

## Success Criteria

- All test cases pass
- Calculation logic correct for all Payday values
- Graceful fallbacks work as expected
- Email sends successfully in all scenarios
- Debug logs verified
- No blocking on Payday fetch failures

## Related Specifications

- **SPEC-002**: Payday Fetching Service (provides payday value)
- **SPEC-003**: Handler Logic Update (implements calculation)
- **SPEC-001**: Data Model Extension (InvoiceDueDay field)
- **ADR-002**: Default Fallback Strategy (10th default)
