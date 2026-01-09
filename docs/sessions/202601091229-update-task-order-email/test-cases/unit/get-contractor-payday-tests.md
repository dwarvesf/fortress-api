# Unit Test Cases: GetContractorPayday Service Method

**Component**: `pkg/service/notion/task_order_log.go`
**Method**: `GetContractorPayday(ctx context.Context, contractorPageID string) (int, error)`
**Specification**: SPEC-002
**Test Framework**: Go testing, testify/require, gomock

## Test Overview

This test suite validates the Payday fetching logic from the Notion Service Rate database, ensuring graceful fallbacks for all error scenarios and correct parsing of valid Payday values.

## Test Structure

### Test File Location
`pkg/service/notion/task_order_log_test.go` (if exists) or create new file

### Mock Dependencies
- Notion client (`notionapi.Client`)
- Logger (`logger.Logger`)
- Config (`config.Config`)

## Test Cases

### TC-001: Happy Path - Payday "01"

**Description**: Verify correct parsing when Service Rate has active status and Payday="01"

**Setup**:
```go
// Mock Notion API response
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{
        {
            ID: "test-service-rate-id",
            Properties: notionapi.Properties{
                "Payday": &notionapi.SelectProperty{
                    Select: notionapi.Select{Name: "01"},
                },
                "Status": &notionapi.StatusProperty{
                    Status: notionapi.Status{Name: "Active"},
                },
            },
        },
    },
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-123"

**Expected Behavior**:
- Queries Service Rate database with:
  - Contractor relation filter: `Contains: "test-contractor-123"`
  - Status filter: `Equals: "Active"`
  - PageSize: 1
- Returns `(1, nil)`
- Logs debug message: "contractor payday found" with payday=1

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-123")
require.NoError(t, err)
require.Equal(t, 1, payday)
```

---

### TC-002: Happy Path - Payday "15"

**Description**: Verify correct parsing when Service Rate has active status and Payday="15"

**Setup**:
```go
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{
        {
            ID: "test-service-rate-id",
            Properties: notionapi.Properties{
                "Payday": &notionapi.SelectProperty{
                    Select: notionapi.Select{Name: "15"},
                },
                "Status": &notionapi.StatusProperty{
                    Status: notionapi.Status{Name: "Active"},
                },
            },
        },
    },
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-456"

**Expected Behavior**:
- Queries Service Rate database
- Returns `(15, nil)`
- Logs debug message: "contractor payday found" with payday=15

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-456")
require.NoError(t, err)
require.Equal(t, 15, payday)
```

---

### TC-003: Graceful Fallback - No Service Rate Found

**Description**: Verify graceful fallback when no active Service Rate record exists for contractor

**Setup**:
```go
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{}, // Empty results
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-no-rate"

**Expected Behavior**:
- Queries database successfully
- No records found
- Returns `(0, nil)` for fallback
- Logs debug message: "no active service rate found for contractor"

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-no-rate")
require.NoError(t, err)
require.Equal(t, 0, payday)
```

**Rationale**: Per ADR-002, missing data should not block email sending. Return 0 to signal handler to use default "10th".

---

### TC-004: Graceful Fallback - Empty Payday Field

**Description**: Verify graceful fallback when Payday field exists but is empty/null

**Setup**:
```go
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{
        {
            Properties: notionapi.Properties{
                "Payday": &notionapi.SelectProperty{
                    Select: notionapi.Select{Name: ""}, // Empty string
                },
            },
        },
    },
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-empty-payday"

**Expected Behavior**:
- Extracts empty string from Payday field
- Returns `(0, nil)` for fallback
- Logs debug message: "payday field is empty"

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-empty-payday")
require.NoError(t, err)
require.Equal(t, 0, payday)
```

---

### TC-005: Graceful Fallback - Invalid Payday Value

**Description**: Verify graceful fallback when Payday field contains invalid value (not "01" or "15")

**Setup**:
```go
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{
        {
            Properties: notionapi.Properties{
                "Payday": &notionapi.SelectProperty{
                    Select: notionapi.Select{Name: "invalid-value"},
                },
            },
        },
    },
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-invalid"

**Expected Behavior**:
- Extracts "invalid-value" from Payday field
- Switch statement doesn't match "01" or "15"
- Returns `(0, nil)` for fallback
- Logs debug message: "invalid payday value" with payday="invalid-value"

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-invalid")
require.NoError(t, err)
require.Equal(t, 0, payday)
```

**Test Variations**:
- Test with: "1" (missing leading zero)
- Test with: "10" (wrong value)
- Test with: "30" (invalid day)
- Test with: "abc" (non-numeric)

---

### TC-006: Error Handling - Database Query Failure

**Description**: Verify graceful fallback when Notion API returns error

**Setup**:
```go
// Mock Notion client to return error
mockClient.EXPECT().
    Database.Query(gomock.Any(), gomock.Any(), gomock.Any()).
    Return(nil, errors.New("notion api timeout"))
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor-api-error"

**Expected Behavior**:
- Query fails with error
- Returns `(0, nil)` for graceful fallback (NOT returning error)
- Logs debug message: "failed to query service rate database" with error

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor-api-error")
require.NoError(t, err) // No error returned to caller
require.Equal(t, 0, payday)
```

**Rationale**: Per SPEC-002, API failures should not block email sending. Log for monitoring but return graceful fallback.

---

### TC-007: Error Handling - Database Not Configured

**Description**: Verify behavior when Service Rate database ID is missing from config

**Setup**:
```go
// Config with empty ServiceRate database ID
cfg := &config.Config{
    Notion: config.NotionConfig{
        Databases: config.NotionDatabases{
            ServiceRate: "", // Empty database ID
        },
    },
}
```

**Input**:
- `ctx`: Valid context
- `contractorPageID`: "test-contractor"

**Expected Behavior**:
- Query attempt with empty database ID
- Notion API returns error (invalid database ID)
- Returns `(0, nil)` for graceful fallback
- Logs debug message with error

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor")
require.NoError(t, err)
require.Equal(t, 0, payday)
```

---

### TC-008: Error Handling - Context Cancellation

**Description**: Verify behavior when context is cancelled during query

**Setup**:
```go
ctx, cancel := context.WithCancel(context.Background())
cancel() // Cancel immediately

mockClient.EXPECT().
    Database.Query(gomock.Any(), gomock.Any(), gomock.Any()).
    Return(nil, context.Canceled)
```

**Input**:
- `ctx`: Cancelled context
- `contractorPageID`: "test-contractor"

**Expected Behavior**:
- Query returns context.Canceled error
- Returns `(0, nil)` for graceful fallback
- Logs debug message with error

**Assertions**:
```go
payday, err := service.GetContractorPayday(ctx, "test-contractor")
require.NoError(t, err)
require.Equal(t, 0, payday)
```

---

## Helper Method Tests

### TC-009: extractSelect - Valid Select Property

**Description**: Verify correct extraction of Select field value

**Input**:
```go
props := notionapi.Properties{
    "Payday": &notionapi.SelectProperty{
        Select: notionapi.Select{Name: "01"},
    },
}
propName := "Payday"
```

**Expected Output**: `"01"`

**Assertions**:
```go
result := extractSelect(props, "Payday")
require.Equal(t, "01", result)
```

---

### TC-010: extractSelect - Property Not Found

**Description**: Verify empty string returned when property doesn't exist

**Input**:
```go
props := notionapi.Properties{
    "OtherField": &notionapi.SelectProperty{
        Select: notionapi.Select{Name: "value"},
    },
}
propName := "Payday" // Doesn't exist
```

**Expected Output**: `""`

**Assertions**:
```go
result := extractSelect(props, "Payday")
require.Equal(t, "", result)
```

---

### TC-011: extractSelect - Wrong Property Type

**Description**: Verify empty string returned when property is not Select type

**Input**:
```go
props := notionapi.Properties{
    "Payday": &notionapi.TextProperty{ // Wrong type
        Text: []notionapi.RichText{
            {PlainText: "01"},
        },
    },
}
propName := "Payday"
```

**Expected Output**: `""`

**Assertions**:
```go
result := extractSelect(props, "Payday")
require.Equal(t, "", result)
```

---

### TC-012: extractSelect - Empty Select Value

**Description**: Verify empty string returned when Select.Name is empty

**Input**:
```go
props := notionapi.Properties{
    "Payday": &notionapi.SelectProperty{
        Select: notionapi.Select{Name: ""}, // Empty
    },
}
propName := "Payday"
```

**Expected Output**: `""`

**Assertions**:
```go
result := extractSelect(props, "Payday")
require.Equal(t, "", result)
```

---

## Test Data Requirements

### Mock Contractor IDs
- `test-contractor-123` - Has Payday "01"
- `test-contractor-456` - Has Payday "15"
- `test-contractor-no-rate` - No Service Rate record
- `test-contractor-empty-payday` - Service Rate with empty Payday
- `test-contractor-invalid` - Service Rate with invalid Payday

### Mock Service Rate Database ID
- `sr-database-123` - Valid database ID for testing

### Mock Notion API Filters
```go
expectedFilter := notionapi.CompoundFilter{
    And: []notionapi.Filter{
        notionapi.PropertyFilter{
            Property: "Contractor",
            Relation: &notionapi.RelationFilter{
                Contains: contractorPageID,
            },
        },
        notionapi.PropertyFilter{
            Property: "Status",
            Status: &notionapi.StatusFilter{
                Equals: "Active",
            },
        },
    },
}
```

## Logging Verification

All test cases should verify appropriate debug logs are generated:

### Success Logs
```go
// Verify log contains
"contractor payday found"
"contractor_id": contractorPageID
"payday": 1 or 15
```

### Fallback Logs
```go
// Verify one of these logs
"no active service rate found for contractor"
"payday field is empty"
"invalid payday value"
"failed to query service rate database"
```

## Mock Setup Example

```go
type mockNotionClient struct {
    mock.Mock
}

func (m *mockNotionClient) DatabaseQuery(
    ctx context.Context,
    databaseID string,
    query *notionapi.DatabaseQuery,
) (*notionapi.DatabaseQueryResponse, error) {
    args := m.Called(ctx, databaseID, query)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*notionapi.DatabaseQueryResponse), args.Error(1)
}

func setupMockService(t *testing.T) (*impl, *mockNotionClient) {
    mockClient := new(mockNotionClient)
    mockLogger := logger.NewLogrusLogger("debug")
    cfg := &config.Config{
        Notion: config.NotionConfig{
            Databases: config.NotionDatabases{
                ServiceRate: "sr-database-123",
            },
        },
    }

    service := &impl{
        client: mockClient,
        logger: mockLogger,
        cfg:    cfg,
    }

    return service, mockClient
}
```

## Test Execution

### Run All Tests
```bash
go test -v ./pkg/service/notion -run TestGetContractorPayday
```

### Run Single Test
```bash
go test -v ./pkg/service/notion -run TestGetContractorPayday_ActivePayday01
```

### Run with Coverage
```bash
go test -cover ./pkg/service/notion -run TestGetContractorPayday
```

## Expected Coverage

- **Line Coverage**: 100% of GetContractorPayday method
- **Branch Coverage**: All switch cases (01, 15, default)
- **Error Paths**: All error handling branches
- **Helper Method**: 100% of extractSelect method

## Success Criteria

- All 12 test cases pass
- All assertions validate
- Debug logs verified for all scenarios
- Mock expectations satisfied
- No test flakiness
- Tests run in <1 second total

## Related Specifications

- **SPEC-002**: Payday Fetching Service
- **ADR-001**: Payday Data Source Selection
- **ADR-002**: Default Fallback Strategy
