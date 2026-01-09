# SPEC-002: Payday Fetching Service

**Status**: Ready for Implementation
**Priority**: High
**Estimated Effort**: 2 hours
**Dependencies**: None

## Overview

Implement service methods to fetch contractor Payday configuration from the Notion Service Rate database, which will be used to calculate personalized invoice due dates.

## Objectives

1. Query Notion Service Rate database for contractor Payday value
2. Extract Payday select field from Notion properties
3. Provide graceful fallback for missing or invalid data
4. Enable invoice due date calculation in handler layer

## Technical Details

### File to Modify

`pkg/service/notion/task_order_log.go`

### Data Source

**Notion Database**: Service Rate
- **Config Key**: `config.Notion.Databases.ServiceRate`
- **Filter Criteria**:
  - Contractor relation matches contractor page ID
  - Status = "Active"
- **Field**: "Payday" (Select field)
- **Valid Values**: "01" or "15"

### Required Methods

#### 1. GetContractorPayday

Fetches the Payday value for a specific contractor from Service Rate database.

```go
// GetContractorPayday fetches the Payday value from Service Rate database
// Returns:
//   - 1 if Payday = "01"
//   - 15 if Payday = "15"
//   - 0 if not found or invalid (caller should use default)
//   - nil error on success, error object on failure
func (s *impl) GetContractorPayday(ctx context.Context, contractorPageID string) (int, error) {
    // Implementation details below
}
```

**Method Signature**:
- **Receiver**: `(s *impl)`
- **Parameters**:
  - `ctx context.Context` - Request context for cancellation
  - `contractorPageID string` - Notion page ID of the contractor
- **Returns**:
  - `int` - Payday value (1, 15, or 0 for fallback)
  - `error` - Error object or nil

**Implementation Logic**:

1. **Query Service Rate Database**:
   ```go
   filter := notionapi.PropertyFilter{
       Property: "Contractor",
       Relation: &notionapi.RelationFilter{
           Contains: contractorPageID,
       },
   }

   statusFilter := notionapi.PropertyFilter{
       Property: "Status",
       Status: &notionapi.StatusFilter{
           Equals: "Active",
       },
   }

   query := &notionapi.DatabaseQuery{
       Filter: notionapi.CompoundFilter{
           And: []notionapi.Filter{filter, statusFilter},
       },
       PageSize: 1, // We only need one active record
   }
   ```

2. **Execute Query**:
   ```go
   resp, err := s.client.Database.Query(ctx, s.cfg.Notion.Databases.ServiceRate, query)
   if err != nil {
       s.logger.Debug(ctx, "failed to query service rate database", "contractor_id", contractorPageID, "error", err)
       return 0, nil // Graceful fallback
   }
   ```

3. **Extract Payday Value**:
   ```go
   if len(resp.Results) == 0 {
       s.logger.Debug(ctx, "no active service rate found for contractor", "contractor_id", contractorPageID)
       return 0, nil
   }

   paydayStr := extractSelect(resp.Results[0].Properties, "Payday")
   if paydayStr == "" {
       s.logger.Debug(ctx, "payday field is empty", "contractor_id", contractorPageID)
       return 0, nil
   }
   ```

4. **Parse and Validate**:
   ```go
   switch paydayStr {
   case "01":
       s.logger.Debug(ctx, "contractor payday found", "contractor_id", contractorPageID, "payday", 1)
       return 1, nil
   case "15":
       s.logger.Debug(ctx, "contractor payday found", "contractor_id", contractorPageID, "payday", 15)
       return 15, nil
   default:
       s.logger.Debug(ctx, "invalid payday value", "contractor_id", contractorPageID, "payday", paydayStr)
       return 0, nil
   }
   ```

#### 2. extractSelect (Helper Method)

Extracts a Select field value from Notion property map.

```go
// extractSelect extracts the value from a Select property
// Returns empty string if property not found or not a Select type
func extractSelect(props notionapi.Properties, propName string) string {
    prop, exists := props[propName]
    if !exists {
        return ""
    }

    selectProp, ok := prop.(*notionapi.SelectProperty)
    if !ok || selectProp.Select.Name == "" {
        return ""
    }

    return selectProp.Select.Name
}
```

**Method Signature**:
- **Parameters**:
  - `props notionapi.Properties` - Property map from Notion page
  - `propName string` - Name of the Select property to extract
- **Returns**:
  - `string` - Select value or empty string

**Implementation Logic**:
1. Check if property exists in map
2. Type assert to `*notionapi.SelectProperty`
3. Return `Select.Name` or empty string if not found/invalid

### Error Handling Strategy

**Principle**: Never block email sending due to Payday fetch failures.

| Scenario | Return Value | Behavior |
|----------|--------------|----------|
| Query fails (network/API error) | `(0, nil)` | Use default "10th" |
| No active Service Rate found | `(0, nil)` | Use default "10th" |
| Payday field is empty | `(0, nil)` | Use default "10th" |
| Invalid Payday value (not "01" or "15") | `(0, nil)` | Use default "10th" |
| Valid Payday = "01" | `(1, nil)` | Use "10th" |
| Valid Payday = "15" | `(15, nil)` | Use "25th" |

**Error Logging**:
- Use `Debug` level for all scenarios (not `Error`)
- Log sufficient context for troubleshooting
- Include contractor ID in all logs

### Integration Points

**Called By**: `pkg/handler/notion/task_order_log.go`
- In `SendTaskOrderConfirmation` method
- Before building `TaskOrderConfirmationEmail` struct

**Dependencies**:
- Notion client (`s.client.Database.Query`)
- Service Rate database ID (`s.cfg.Notion.Databases.ServiceRate`)
- Logger (`s.logger.Debug`)

## Testing Requirements

### Unit Tests

**Test File**: `pkg/service/notion/task_order_log_test.go`

**Test Cases**:

1. **TestGetContractorPayday_ActivePayday01**
   - Setup: Mock Service Rate with Active status, Payday="01"
   - Expected: Returns `(1, nil)`
   - Verifies: Correct parsing of "01" value

2. **TestGetContractorPayday_ActivePayday15**
   - Setup: Mock Service Rate with Active status, Payday="15"
   - Expected: Returns `(15, nil)`
   - Verifies: Correct parsing of "15" value

3. **TestGetContractorPayday_NoServiceRate**
   - Setup: Mock empty query result
   - Expected: Returns `(0, nil)`
   - Verifies: Graceful fallback for missing data

4. **TestGetContractorPayday_EmptyPayday**
   - Setup: Mock Service Rate with empty Payday field
   - Expected: Returns `(0, nil)`
   - Verifies: Graceful fallback for empty field

5. **TestGetContractorPayday_InvalidPayday**
   - Setup: Mock Service Rate with Payday="invalid"
   - Expected: Returns `(0, nil)`
   - Verifies: Graceful fallback for invalid value

6. **TestGetContractorPayday_QueryError**
   - Setup: Mock Notion API error
   - Expected: Returns `(0, nil)`
   - Verifies: Graceful fallback on API failure

7. **TestExtractSelect_ValidSelect**
   - Setup: Mock properties with valid Select field
   - Expected: Returns select value string
   - Verifies: Correct extraction from properties

8. **TestExtractSelect_MissingProperty**
   - Setup: Mock properties without the field
   - Expected: Returns empty string
   - Verifies: Handles missing property

9. **TestExtractSelect_WrongType**
   - Setup: Mock properties with non-Select field
   - Expected: Returns empty string
   - Verifies: Type safety

### Integration Tests

**Manual Testing**:

1. **Test with Real Contractor**:
   ```bash
   # In handler test, call GetContractorPayday
   payday, err := s.notionSvc.GetContractorPayday(ctx, contractorPageID)
   // Verify correct value returned
   ```

2. **Test Database Query**:
   - Verify filter correctly finds Active Service Rate
   - Verify only one record returned (PageSize: 1)
   - Verify correct contractor matched

3. **Test Logging**:
   - Verify debug logs appear for each scenario
   - Verify contractor ID included in logs
   - Verify payday value logged when found

### Mock Data Requirements

For unit tests, mock the following Notion API responses:

```go
// Active Service Rate with Payday="01"
mockResponse := &notionapi.DatabaseQueryResponse{
    Results: []notionapi.Page{
        {
            Properties: notionapi.Properties{
                "Payday": &notionapi.SelectProperty{
                    Select: notionapi.Select{Name: "01"},
                },
            },
        },
    },
}
```

## Acceptance Criteria

- [ ] `GetContractorPayday` method implemented in `task_order_log.go`
- [ ] `extractSelect` helper method implemented
- [ ] Method queries Service Rate database with correct filters
- [ ] Method returns correct Payday values (1, 15, or 0)
- [ ] Graceful fallback for all error scenarios
- [ ] Debug logging for all code paths
- [ ] All unit tests pass
- [ ] Integration test with real contractor data succeeds
- [ ] Code review approved

## Implementation Notes

### Database Configuration

Verify configuration exists:
```go
// In config/config.go, ensure ServiceRate database ID is configured
type NotionDatabases struct {
    ServiceRate string `env:"SERVICE_RATE_DATABASE_ID"`
    // ... other databases
}
```

### Notion API Reference

Service Rate database structure (already exists):
- **Contractor**: Relation to Contractors database
- **Status**: Status field with "Active" option
- **Payday**: Select field with "01" and "15" options

### Code Location

- **File**: `pkg/service/notion/task_order_log.go`
- **Insert After**: Existing helper methods (around line 1800+)
- **Pattern**: Follow existing method patterns in the file

### Example Usage in Handler

```go
// In SendTaskOrderConfirmation handler
payday, err := s.notionSvc.GetContractorPayday(ctx, contractorPageID)
if err != nil {
    // Already logged in service layer, continue with default
    payday = 0
}

invoiceDueDay := "10th" // Default
if payday == 15 {
    invoiceDueDay = "25th"
}
```

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Service Rate database not configured | High | Graceful fallback to default |
| Network timeout querying Notion | Medium | Graceful fallback, debug logging |
| Data quality issues (wrong Payday values) | Low | Validation logic, fallback |
| Performance impact of extra API call | Low | Query limited to 1 result, cached in future |

## Future Enhancements

1. **Caching**: Add Redis cache for Payday values (30-day TTL)
2. **Batch Fetching**: Fetch all Paydays in one query for bulk emails
3. **Monitoring**: Add metrics for fallback rate
4. **Validation**: Add admin endpoint to check Payday data quality

## Documentation Updates

Add inline comments:
- Explain Payday value to due date mapping
- Document graceful fallback strategy
- Reference ADR-001 for data source decision

## Related Specifications

- **SPEC-001**: Data Model Update (uses this data)
- **SPEC-003**: Handler Logic Update (calls this service)
- **ADR-001**: Payday Data Source Selection
- **ADR-002**: Default Fallback Strategy

## Sign-off

- [ ] Implementation completed
- [ ] Unit tests pass
- [ ] Integration test successful
- [ ] Code review approved
- [ ] Documentation updated
