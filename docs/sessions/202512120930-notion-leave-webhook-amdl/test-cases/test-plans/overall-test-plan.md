# Overall Test Plan: Notion Leave Webhook AM/DL Integration

**Version:** 1.0
**Date:** 2025-12-12
**Status:** Design Phase

## Overview

This test plan covers comprehensive unit testing for the Notion Leave Webhook AM/DL integration feature. The tests validate the automatic lookup and notification of Account Managers (AM) and Delivery Leads (DL) from the Deployment Tracker when leave requests are created.

## Scope

### In Scope - Unit Tests
1. GetActiveDeploymentsForContractor - Deployment query
2. ExtractStakeholdersFromDeployment - Property extraction and priority logic
3. GetDiscordUsernameFromContractor - Notion contractor page fetching
4. GetDiscordMentionFromUsername - Database lookup and mention formatting
5. UpdateApprovedRejectedBy - Approval relation update

### Out of Scope
- Integration tests (per user requirements)
- End-to-end webhook flow tests
- Performance/load testing
- Security testing (covered by existing patterns)

## Test Strategy

### Testing Approach
- **Pattern:** Table-driven tests following fortress-api conventions
- **Mocking:** Mock external dependencies (Notion API, database)
- **Coverage Target:** >90% code coverage for new functions
- **Isolation:** Each function tested independently with mocked dependencies

### Test Data Management
- Mock Notion API responses stored in test data structures
- Mock database records defined inline in test cases
- Property extraction uses realistic Notion property structures
- Edge cases covered with boundary value analysis

## Function Test Coverage Summary

### 1. GetActiveDeploymentsForContractor
**Total Test Cases:** 8
**Coverage Areas:**
- Happy path: Single and multiple deployments
- Empty results (no active deployments)
- API errors and network failures
- Context cancellation
- Pagination handling
- Input validation

**Critical Scenarios:**
- Filter construction correctness (AND condition with relation + status)
- Graceful handling of empty results
- Error propagation vs graceful degradation

---

### 2. ExtractStakeholdersFromDeployment
**Total Test Cases:** 14
**Coverage Areas:**
- Override AM/DL priority logic
- Rollup extraction for AM/DL
- Mixed scenarios (override + rollup)
- Deduplication of stakeholders
- Complex nested rollup structures
- Missing/invalid properties
- Property type mismatches

**Critical Scenarios:**
- Override takes precedence over rollup
- Deduplication when same stakeholder is both AM and DL
- Graceful degradation with partial property failures
- Flattening multi-level rollup arrays

---

### 3. GetDiscordUsernameFromContractor
**Total Test Cases:** 14
**Coverage Areas:**
- Simple rich text extraction
- Multi-part rich text concatenation
- Whitespace trimming
- Empty Discord field handling
- Missing property handling
- API errors (page not found, rate limit)
- Context cancellation
- Invalid page properties

**Critical Scenarios:**
- Rich text concatenation logic
- Empty field returns empty string (not error)
- API error propagation
- Input validation

---

### 4. GetDiscordMentionFromUsername
**Total Test Cases:** 14 (including store method tests)
**Coverage Areas:**
- Username to Discord ID lookup
- Mention format construction
- Database query errors
- Username not found
- Empty username handling
- Whitespace trimming
- Case sensitivity
- New store method: OneByUsername

**Critical Scenarios:**
- Correct mention format: `<@discord_id>`
- Graceful degradation on database errors
- Store method implementation (OneByUsername)
- Database index recommendation for performance

---

### 5. UpdateApprovedRejectedBy
**Total Test Cases:** 17
**Coverage Areas:**
- Full approval flow (Discord ID → username → contractor → update)
- Each step failure handling
- Input validation
- Database and Notion API errors
- Multiple contractors found
- Query filter construction
- Relation update format

**Critical Scenarios:**
- Multi-step lookup chain correctness
- Graceful degradation (don't fail approval if metadata update fails)
- Notion relation format correctness
- Helper function for contractor lookup

---

## Test Dependencies

### External Dependencies to Mock
1. **Notion API Client (`nt.Client`)**
   - Methods: `QueryDatabase`, `FindPageByID`, `UpdatePage`
   - Mock responses with realistic Notion property structures

2. **Database (GORM `*gorm.DB`)**
   - Tables: `discord_accounts`
   - Mock store methods: `OneByDiscordID`, `OneByUsername` (new)

3. **Logger (`logger.Logger`)**
   - Mock logger to verify log messages and levels
   - Test nil logger handling

### Configuration Dependencies
- `config.Notion.Databases.DeploymentTracker`
- `config.LeaveIntegration.Notion.ContractorDBID`
- Mock config structs with valid database IDs

## Error Handling Strategy

### Critical Errors (Return Error)
- Input validation failures (empty required parameters)
- Notion API authentication failures
- Database connection failures (when critical to operation)

### Graceful Degradation (Log Warning, Continue)
- No active deployments found
- Contractor not in Notion
- Discord username not in database
- Notion API temporary failures (rate limits)
- Metadata update failures (UpdateApprovedRejectedBy)

### Logging Levels
- **DEBUG:** Normal operation details, query parameters
- **INFO:** Successful operations, empty results
- **WARNING:** Expected failures, missing data, fallbacks
- **ERROR:** Unexpected failures, API errors, data inconsistencies

## Mock Structures

### Mock Notion Client
```go
type MockNotionClient struct {
    QueryDatabaseFunc func(ctx context.Context, dbID string, query *nt.DatabaseQuery) (*nt.DatabaseQueryResponse, error)
    FindPageByIDFunc  func(ctx context.Context, pageID string) (*nt.Page, error)
    UpdatePageFunc    func(ctx context.Context, pageID string, params nt.UpdatePageParams) (*nt.Page, error)
}
```

### Mock Discord Account Store
```go
type MockDiscordAccountStore struct {
    OneByDiscordIDFunc func(db *gorm.DB, discordID string) (*model.DiscordAccount, error)
    OneByUsernameFunc  func(db *gorm.DB, username string) (*model.DiscordAccount, error)
}
```

### Mock Database
```go
type MockDB struct {
    *gorm.DB
    queryResults map[string]interface{}
    queryErrors  map[string]error
}
```

## Test Data Patterns

### Notion Deployment Page Structure
```go
func createMockDeploymentPage(
    contractorID string,
    overrideAM []string,
    overrideDL []string,
    rollupAM [][]string,
    rollupDL [][]string,
) nt.Page
```

### Notion Contractor Page Structure
```go
func createMockContractorPage(
    pageID string,
    discordUsername string,
    teamEmail string,
) nt.Page
```

### Discord Account Record
```go
func createMockDiscordAccount(
    discordID string,
    discordUsername string,
) *model.DiscordAccount
```

## Property Extraction Test Patterns

### Relation Property
- Single relation: `Relation: [{ID: "page-id"}]`
- Multiple relations: `Relation: [{ID: "id-1"}, {ID: "id-2"}]`
- Empty relation: `Relation: []`
- Missing property: not in properties map

### Rollup Property (Array of Relations)
```go
Rollup: nt.RollupResult{
    Type: "array",
    Array: []nt.DatabasePageProperty{
        {Relation: [{ID: "id-1"}, {ID: "id-2"}]},
        {Relation: [{ID: "id-3"}]},
    },
}
```

### Rich Text Property
```go
RichText: []nt.RichText{
    {Type: "text", PlainText: "part1"},
    {Type: "text", PlainText: "part2"},
}
```

## Assertion Patterns

### API Call Assertions
```go
// Verify Notion API called with correct parameters
assert.Equal(t, expectedDBID, capturedDBID)
assert.Equal(t, expectedFilter, capturedFilter)
```

### Data Assertions
```go
// Verify stakeholder extraction
assert.ElementsMatch(t, expectedStakeholders, actualStakeholders)
assert.Len(t, actualStakeholders, expectedCount)
```

### Error Assertions
```go
// Verify error handling
assert.NoError(t, err) // Happy path
assert.Error(t, err)   // Error path
assert.Nil(t, result)  // Nil on error
```

### Logging Assertions
```go
// Verify logging behavior
mockLogger.AssertLogged(t, logger.LevelDebug, "expected message")
mockLogger.AssertLogged(t, logger.LevelError, "error occurred")
```

## Edge Cases and Boundary Conditions

### Empty/Null Values
- Empty strings: `""`
- Nil pointers: `nil`
- Empty arrays: `[]`
- Zero values: `0`, `false`

### Special Characters
- Discord usernames: `user_name-2024.test`
- Emails: `test+alias@domain.com`
- Notion page IDs: UUID format validation

### Concurrency and Context
- Context cancellation mid-operation
- Nil context handling
- Timeout scenarios

### Data Integrity
- Multiple records when expecting one (data issues)
- Mismatched data types (Notion schema changes)
- Partial property availability

## Performance Considerations

### Expected Query Counts
- Per leave request creation:
  - 1 DB query: Employee lookup (existing)
  - 1 Notion query: Contractor lookup by email
  - 1 Notion query: Active deployments
  - N Notion queries: Contractor pages (N = unique stakeholders, typically 1-5)
  - M DB queries: Discord accounts (M = unique usernames, typically 1-5)

### Performance Requirements
- Total time: < 5s per leave request
- Individual API calls: < 2s each
- Database queries: < 100ms each

### No Performance Tests in Unit Tests
- Performance validated in integration/load tests
- Unit tests focus on correctness and edge cases

## Database Schema Requirements

### New Store Method
**Method:** `DiscordAccount.OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)`

**Index Required:**
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

**Implementation Notes:**
- Return `nil, nil` for not found (graceful)
- Return `nil, error` for database errors
- Use `LIMIT 1` for performance

## Configuration Testing

### Required Config Values
- `NOTION_DEPLOYMENT_TRACKER_DB_ID`
- `NOTION_CONTRACTOR_DB_ID` (already exists)

### Config Validation Tests
- Missing config values return descriptive errors
- Invalid config values handled gracefully

## Integration Test Boundary

### What Unit Tests Don't Cover
- Real Notion API calls and responses
- Real database transactions
- Full webhook flow end-to-end
- Discord API integration
- Actual deployment data validation

### Transition to Integration Tests
- Unit tests validate individual functions
- Integration tests validate full workflow
- E2E tests validate entire system behavior

## Test Execution Plan

### Phase 1: Core Property Extraction (Week 1)
- [ ] ExtractStakeholdersFromDeployment tests
- [ ] Property extraction helper tests
- [ ] Mock Notion page structures

### Phase 2: Notion API Interactions (Week 1)
- [ ] GetActiveDeploymentsForContractor tests
- [ ] GetDiscordUsernameFromContractor tests
- [ ] Mock Notion client setup

### Phase 3: Database Interactions (Week 1-2)
- [ ] GetDiscordMentionFromUsername tests
- [ ] OneByUsername store method implementation
- [ ] Mock database setup

### Phase 4: Update Flow (Week 2)
- [ ] UpdateApprovedRejectedBy tests
- [ ] Multi-step flow validation
- [ ] Error handling chains

### Phase 5: Validation and Review (Week 2)
- [ ] Code coverage report (target >90%)
- [ ] Review all edge cases
- [ ] Documentation review

## Success Criteria

### Test Coverage
- [ ] All functions have >90% code coverage
- [ ] All happy paths tested
- [ ] All error paths tested
- [ ] All edge cases documented and tested

### Quality Metrics
- [ ] Zero flaky tests
- [ ] Tests run in <10s total
- [ ] All tests pass consistently
- [ ] Mock behavior matches real API closely

### Documentation
- [ ] All test cases documented with rationale
- [ ] Mock structures documented
- [ ] Test data patterns documented
- [ ] Edge cases and assumptions listed

## Risks and Mitigations

### Risk: Notion API Changes
- **Impact:** Tests may not match real API behavior
- **Mitigation:** Validate mocks against real API responses during integration testing
- **Mitigation:** Document Notion API version used

### Risk: Database Schema Changes
- **Impact:** Store method tests may break
- **Mitigation:** Use migrations to validate schema
- **Mitigation:** Integration tests catch schema mismatches

### Risk: Complex Property Extraction Logic
- **Impact:** Subtle bugs in rollup extraction
- **Mitigation:** Extensive edge case testing
- **Mitigation:** Real Notion data validation in integration tests

### Risk: Graceful Degradation Too Permissive
- **Impact:** Silent failures go unnoticed
- **Mitigation:** Comprehensive logging assertions
- **Mitigation:** Monitor logs in production

## Appendix

### Related Documents
- Requirements: `/docs/sessions/202512120930-notion-leave-webhook-amdl/requirements/requirements.md`
- Technical Spec: `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/specifications/am-dl-integration-spec.md`
- ADR-001: `/docs/sessions/202512120930-notion-leave-webhook-amdl/planning/ADRs/001-am-dl-lookup-strategy.md`

### Test Case Documents
1. `/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/01-get-active-deployments-for-contractor.md`
2. `/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/02-extract-stakeholders-from-deployment.md`
3. `/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/03-get-discord-username-from-contractor.md`
4. `/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/04-get-discord-mention-from-username.md`
5. `/docs/sessions/202512120930-notion-leave-webhook-amdl/test-cases/unit/05-update-approved-rejected-by.md`

### Notion Property Type Reference
- **Relation:** Array of `{ID: string}` objects
- **Rollup:** Nested structure with `Type` and `Array` or other value
- **Rich Text:** Array of `{Type, PlainText}` objects
- **Select:** Object with `{Name: string}`
- **Status:** Object with `{Name: string}`
- **Email:** Pointer to string value

### Discord Mention Format
- **Pattern:** `<@DISCORD_ID>`
- **Example:** `<@123456789012345678>`
- **Discord ID:** 18-digit numeric string (snowflake)
