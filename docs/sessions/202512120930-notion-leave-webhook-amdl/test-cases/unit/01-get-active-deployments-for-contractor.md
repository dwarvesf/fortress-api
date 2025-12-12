# Test Case Design: GetActiveDeploymentsForContractor

**Function Under Test:** `LeaveService.GetActiveDeploymentsForContractor(ctx context.Context, contractorPageID string) ([]nt.Page, error)`

**Location:** `pkg/service/notion/leave.go`

**Purpose:** Query Deployment Tracker database for active deployments where the contractor relation matches the given contractor page ID.

## Test Strategy

### Approach
- **Pattern:** Table-driven tests with multiple scenarios
- **Dependencies:** Mock Notion API client
- **Validation:** Assert returned deployment pages and error handling
- **Focus:** Query filter construction, API response handling, graceful degradation

### Test Data Requirements
- Mock Notion API responses for QueryDatabase
- Valid contractor page IDs
- Various deployment scenarios (active, inactive, multiple)

## Test Cases

### TC-1.1: Successfully Query Active Deployments - Single Result

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Notion API returns 1 active deployment page

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "contractor-abc-123")`

**Then:**
- Should call `client.QueryDatabase()` with correct parameters:
  - Database ID: `NOTION_DEPLOYMENT_TRACKER_DB_ID` from config
  - Filter: AND condition with:
    - Property: "Contractor", Relation Contains: "contractor-abc-123"
    - Property: "Deployment Status", Status Equals: "Active"
  - PageSize: 100 (or default pagination)
- Should return array with 1 deployment page
- Should return nil error
- Page should contain expected properties: Contractor, Deployment Status, Override AM, Override DL, Account Managers, Delivery Leads

**Test Data:**
```json
{
  "results": [
    {
      "id": "deployment-page-1",
      "properties": {
        "Contractor": {
          "type": "relation",
          "relation": [{"id": "contractor-abc-123"}]
        },
        "Deployment Status": {
          "type": "status",
          "status": {"name": "Active"}
        }
      }
    }
  ],
  "has_more": false
}
```

---

### TC-1.2: Successfully Query Active Deployments - Multiple Results

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Notion API returns 3 active deployment pages

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "contractor-abc-123")`

**Then:**
- Should call `client.QueryDatabase()` once
- Should return array with 3 deployment pages in order
- Should return nil error
- All pages should have matching contractor relation

**Test Data:**
```json
{
  "results": [
    {"id": "deployment-page-1", "properties": {...}},
    {"id": "deployment-page-2", "properties": {...}},
    {"id": "deployment-page-3", "properties": {...}}
  ],
  "has_more": false
}
```

---

### TC-1.3: No Active Deployments Found - Graceful Return

**Given:**
- Valid contractor page ID: `contractor-xyz-789`
- Notion API returns empty results array (no active deployments)

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "contractor-xyz-789")`

**Then:**
- Should call `client.QueryDatabase()` with correct filter
- Should return empty array `[]nt.Page{}` (not nil)
- Should return nil error (graceful handling, not an error condition)
- Should log at DEBUG or INFO level: "no active deployments found for contractor"

**Test Data:**
```json
{
  "results": [],
  "has_more": false
}
```

---

### TC-1.4: Notion API Returns Error - Propagate Error

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Notion API returns error (network failure, API error, rate limit, etc.)

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "contractor-abc-123")`

**Then:**
- Should call `client.QueryDatabase()`
- Should return nil for pages array
- Should return error with descriptive message: "failed to query deployment tracker: [original error]"
- Should log error at ERROR level with contractor page ID for debugging

**Mock Error:**
```go
errors.New("notion API: rate limit exceeded")
```

---

### TC-1.5: Context Cancellation - Respect Context

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Context is cancelled before API call completes

**When:**
- Call `GetActiveDeploymentsForContractor(cancelledCtx, "contractor-abc-123")`

**Then:**
- Should attempt to call `client.QueryDatabase()` with cancelled context
- Should return nil for pages array
- Should return context.Canceled error or wrapped error
- Should not hang or block indefinitely

**Mock Behavior:**
- Client should return context cancellation error

---

### TC-1.6: Pagination Handling - Multiple Pages of Results

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Notion API returns paginated results (has_more: true, next_cursor provided)
- First page: 2 deployments, Second page: 1 deployment

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "contractor-abc-123")`

**Then:**
- Should call `client.QueryDatabase()` twice:
  - First call: no cursor
  - Second call: with next_cursor from first response
- Should return array with all 3 deployment pages (combined from both pages)
- Should return nil error

**Test Data (Page 1):**
```json
{
  "results": [
    {"id": "deployment-page-1", "properties": {...}},
    {"id": "deployment-page-2", "properties": {...}}
  ],
  "has_more": true,
  "next_cursor": "cursor-abc"
}
```

**Test Data (Page 2):**
```json
{
  "results": [
    {"id": "deployment-page-3", "properties": {...}}
  ],
  "has_more": false
}
```

---

### TC-1.7: Empty Contractor Page ID - Input Validation

**Given:**
- Empty contractor page ID: `""`

**When:**
- Call `GetActiveDeploymentsForContractor(ctx, "")`

**Then:**
- Should return empty array or nil
- Should return error: "contractor page ID is required"
- Should NOT call Notion API
- Should log warning about invalid input

**Rationale:** Prevent unnecessary API calls with invalid input

---

### TC-1.8: Nil Context - Handle Gracefully

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Context is nil

**When:**
- Call `GetActiveDeploymentsForContractor(nil, "contractor-abc-123")`

**Then:**
- Should panic or return error (depending on implementation strategy)
- Recommended: Return error "context is required"
- Should NOT proceed with API call

**Note:** This is an edge case for defensive programming

---

## Mock Setup Requirements

### Mock Notion Client
```go
type MockNotionClient struct {
    QueryDatabaseFunc func(ctx context.Context, dbID string, query *nt.DatabaseQuery) (*nt.DatabaseQueryResponse, error)
    callCount         int
    capturedFilters   []*nt.DatabaseQueryFilter
}
```

### Mock Assertions
- Verify `QueryDatabase` called with correct database ID from config
- Verify filter structure matches expected AND condition
- Verify contractor page ID is correctly embedded in relation filter
- Verify deployment status filter checks for "Active"
- Verify pagination cursor is passed correctly on subsequent calls

## Error Handling Strategy

### Expected Error Cases
1. **Notion API Errors:** Wrap and return with context
2. **Network Errors:** Wrap and return with context
3. **Rate Limiting:** Return error for retry handling at caller level
4. **Invalid Response Format:** Log and return parsing error

### Graceful Degradation Cases
1. **No Results:** Return empty array, nil error (not an error condition)
2. **Partial Results:** Return what was received before error occurred (implementation choice)

## Logging Assertions

### DEBUG Level
- "querying deployment tracker: contractor_id=%s db_id=%s"
- "deployment tracker query returned %d results"
- "pagination: has_more=%t next_cursor=%s"

### ERROR Level
- "failed to query deployment tracker: contractor_id=%s error=%v"

## Configuration Dependencies

### Required Config Values
- `cfg.Notion.Databases.DeploymentTracker` - Database ID for Deployment Tracker

### Config Validation
- Test should fail gracefully if deployment tracker DB ID is not configured
- Return error: "deployment tracker database not configured"

## Performance Considerations

### Expected Behavior
- Single API call for most cases (< 100 active deployments per contractor)
- Pagination for contractors with many deployments (rare)
- Response time: < 2s for typical case (depends on Notion API)

### No Optimization Required
- Batching not needed (single contractor lookup)
- Caching not in scope for this function

## Integration Notes

### Caller Expectations
- Caller: `handleNotionLeaveCreated` in webhook handler
- Caller will pass contractor page ID obtained from `GetContractorPageIDByEmail`
- Caller expects empty array as valid response (no deployments is acceptable)
- Caller will log warnings if no deployments found

### Downstream Usage
- Returned deployment pages will be passed to `ExtractStakeholdersFromDeployment`
- Each deployment page must have properties accessible via `nt.DatabasePageProperties`

## Test Implementation Checklist

- [ ] Create mock Notion client with configurable responses
- [ ] Test happy path with single deployment
- [ ] Test happy path with multiple deployments
- [ ] Test empty results (no active deployments)
- [ ] Test Notion API errors
- [ ] Test context cancellation
- [ ] Test pagination (if implemented)
- [ ] Test input validation (empty contractor ID)
- [ ] Verify filter construction correctness
- [ ] Verify logging at appropriate levels
- [ ] Verify error messages are descriptive
- [ ] Test with real Notion API response structure (integration test boundary)
