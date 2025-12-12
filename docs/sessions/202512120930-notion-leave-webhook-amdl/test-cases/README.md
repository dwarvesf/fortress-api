# Test Case Designs - Notion Leave Webhook AM/DL Integration

**Session:** 202512120930-notion-leave-webhook-amdl
**Date:** 2025-12-12
**Type:** Unit Test Designs Only

## Quick Navigation

### Test Case Documents
1. [GetActiveDeploymentsForContractor](unit/01-get-active-deployments-for-contractor.md) - 8 test cases
2. [ExtractStakeholdersFromDeployment](unit/02-extract-stakeholders-from-deployment.md) - 14 test cases
3. [GetDiscordUsernameFromContractor](unit/03-get-discord-username-from-contractor.md) - 14 test cases
4. [GetDiscordMentionFromUsername](unit/04-get-discord-mention-from-username.md) - 14 test cases
5. [UpdateApprovedRejectedBy](unit/05-update-approved-rejected-by.md) - 17 test cases

### Planning Documents
- [Overall Test Plan](test-plans/overall-test-plan.md) - Comprehensive test strategy and execution plan
- [STATUS.md](STATUS.md) - Current status, coverage summary, and next steps

## What's Inside

### Test Case Designs (Not Implementations)

This directory contains **detailed test case designs** for unit testing the AM/DL integration feature. These are NOT test implementations - they are specifications that describe:

- **What to test:** Clear test scenarios with Given/When/Then structure
- **How to test:** Mock setup, assertions, test data
- **Why to test:** Rationale for each test case
- **Expected results:** Precise outputs and error handling

### Scope: Unit Tests Only

Per user requirements, this phase focuses exclusively on **unit test designs**. Integration tests are explicitly out of scope.

## Summary Statistics

| Metric | Value |
|--------|-------|
| Total Test Cases | 67 |
| Functions Covered | 5 (+ 1 new store method) |
| Happy Path Tests | 10 |
| Error Handling Tests | 22 |
| Edge Case Tests | 18 |
| Input Validation Tests | 7 |
| Property Extraction Tests | 10 |

## Test Coverage by Function

### 1. GetActiveDeploymentsForContractor (8 tests)

**Purpose:** Query Deployment Tracker for active deployments by contractor

**Key Test Areas:**
- Notion API QueryDatabase with compound filters (Contractor relation + Active status)
- Empty results (no active deployments) → graceful return
- Pagination handling for multiple deployment pages
- API errors and network failures → error propagation
- Context cancellation and timeout handling

**Critical Scenarios:**
- Filter construction: AND(relation contains X, status equals "Active")
- Empty array return for no deployments (not an error)
- Graceful degradation vs error propagation strategy

---

### 2. ExtractStakeholdersFromDeployment (14 tests)

**Purpose:** Extract AM/DL page IDs with override priority logic

**Key Test Areas:**
- Override AM/DL takes precedence over rollup values
- Rollup extraction from nested array structures
- Mixed scenarios (override for AM, rollup for DL)
- Stakeholder deduplication (same person as AM and DL)
- Complex rollup structures (multiple arrays, multiple relations per array)
- Missing or invalid property types → graceful skip

**Critical Scenarios:**
- Priority logic: Override > Rollup > Empty
- Deduplication of same stakeholder appearing as both AM and DL
- Flattening nested rollup arrays with multiple relation objects
- Partial extraction when some properties fail

---

### 3. GetDiscordUsernameFromContractor (14 tests)

**Purpose:** Fetch contractor page from Notion and extract Discord username

**Key Test Areas:**
- Rich text property extraction and concatenation
- Multi-part rich text joining (no separators)
- Whitespace trimming on final result
- Empty Discord field → empty string (not error)
- Missing property → empty string (not error)
- Notion API errors (404, rate limit) → error propagation

**Critical Scenarios:**
- Concatenate multiple rich text parts: ["user", "name"] → "username"
- Trim whitespace: "  username  " → "username"
- Empty field is valid (not all contractors have Discord)
- API errors vs missing data (different error handling)

---

### 4. GetDiscordMentionFromUsername (14 tests)

**Purpose:** Convert Discord username to mention format via database lookup

**Key Test Areas:**
- Database query by username → Discord ID retrieval
- Mention formatting: `<@discord_id>`
- Username not found → empty string (graceful)
- Database errors → empty string (graceful)
- Whitespace trimming on username before query
- NEW: OneByUsername store method implementation

**Critical Scenarios:**
- Mention format validation: `<@123456789012345678>`
- Graceful degradation on all failures (don't block webhook)
- New store method: Return nil,nil for not found (not error)
- Database index recommendation: discord_username

---

### 5. UpdateApprovedRejectedBy (17 tests)

**Purpose:** Update Notion leave request with approver contractor relation

**Key Test Areas:**
- Multi-step lookup: Discord ID → username → contractor page → update
- Each step failure handling (graceful degradation)
- Notion query filter for rich text property
- Relation update format correctness
- Input validation (empty parameters)
- Multiple contractors found → use first (data integrity issue)

**Critical Scenarios:**
- Full lookup chain: DiscordAccount(by ID) → username → Contractors(by Discord) → UpdatePage(relation)
- All failures are graceful (don't fail approval/rejection)
- Relation format: `{Relation: [{ID: "contractor-page-id"}]}`
- Helper function for Notion contractor lookup by Discord username

---

## New Implementation Requirements

### Store Method: OneByUsername

**File:** `pkg/store/discordaccount/discordaccount.go`

**Interface Update:**
```go
// pkg/store/discordaccount/interface.go
type IStore interface {
    // ... existing methods ...
    OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
}
```

**Implementation:**
```go
func (s *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error) {
    var discordAccount model.DiscordAccount
    err := db.Where("discord_username = ?", username).First(&discordAccount).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // Graceful not found
        }
        return nil, err
    }
    return &discordAccount, nil
}
```

**Database Index:**
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

## Test Design Patterns

### Mock Structures Required

#### Mock Notion Client
```go
type MockNotionClient struct {
    QueryDatabaseFunc func(ctx context.Context, dbID string, query *nt.DatabaseQuery) (*nt.DatabaseQueryResponse, error)
    FindPageByIDFunc  func(ctx context.Context, pageID string) (*nt.Page, error)
    UpdatePageFunc    func(ctx context.Context, pageID string, params nt.UpdatePageParams) (*nt.Page, error)
}
```

#### Mock Discord Account Store
```go
type MockDiscordAccountStore struct {
    OneByDiscordIDFunc func(db *gorm.DB, discordID string) (*model.DiscordAccount, error)
    OneByUsernameFunc  func(db *gorm.DB, username string) (*model.DiscordAccount, error)
}
```

### Table-Driven Test Pattern
```go
func TestExtractStakeholdersFromDeployment(t *testing.T) {
    tests := []struct {
        name           string
        deploymentPage nt.Page
        expected       []string
        expectError    bool
    }{
        {
            name:           "Override AM and DL present",
            deploymentPage: createMockPage(...),
            expected:       []string{"am-override-123", "dl-override-456"},
            expectError:    false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Error Handling Strategy

### Critical Errors (Return Error)
- Input validation failures (empty required parameters)
- Cannot cast Notion page properties
- Nil context (defensive check)

### Graceful Degradation (Log Warning, Return Empty/Nil)
- No active deployments found
- Contractor not in Notion
- Discord username not in database
- Notion API temporary failures
- Metadata update failures (UpdateApprovedRejectedBy)

**Principle:** Don't fail the webhook for non-critical failures. Log warnings for manual follow-up.

## Logging Verification Matrix

| Level | Count | Examples |
|-------|-------|----------|
| DEBUG | 30 | Query parameters, extracted values, lookup results |
| INFO | 8 | Successful operations, empty results (expected) |
| WARNING | 15 | Data not found, multiple records found, skipped operations |
| ERROR | 14 | API failures, database errors, unexpected failures |

**Test Requirement:** Each test case should verify logging at the appropriate level.

## Property Extraction Reference

### Notion Property Types Tested

| Type | Extraction Method | Test Cases |
|------|------------------|------------|
| Relation | `extractFirstRelationID` | TC-2.1, TC-2.2, TC-2.3 |
| Rollup (array) | `extractRollupRelations` | TC-2.2, TC-2.6, TC-2.12 |
| Rich Text | `extractRichText` | TC-3.1, TC-3.3, TC-3.4 |
| Email | `extractEmail` | Existing tests |
| Select/Status | `extractSelect` | Existing tests |

### Helper Functions to Test
- `extractFirstRelationID(props, "Override AM")` → string or ""
- `extractRollupRelations(props, "Account Managers")` → []string or []
- `extractRichText(props, "Discord")` → string (trimmed) or ""

## Configuration Checklist

### Environment Variables
- [x] `NOTION_DEPLOYMENT_TRACKER_DB_ID` - New, required
- [x] `NOTION_CONTRACTOR_DB_ID` - Existing, required

### Config Validation Tests
- [ ] Missing deployment tracker DB ID → error
- [ ] Missing contractor DB ID → error
- [ ] Valid config values → successful operations

## Implementation Checklist

### Phase 1: Setup
- [ ] Create mock Notion client structure
- [ ] Create mock Discord account store
- [ ] Set up test fixtures and helpers
- [ ] Define test data creation functions

### Phase 2: Property Extraction (TC-2.x)
- [ ] Implement ExtractStakeholdersFromDeployment tests
- [ ] Test override priority logic
- [ ] Test rollup extraction
- [ ] Test deduplication
- [ ] Test edge cases (empty, missing, invalid properties)

### Phase 3: Notion API Interactions (TC-1.x, TC-3.x)
- [ ] Implement GetActiveDeploymentsForContractor tests
- [ ] Implement GetDiscordUsernameFromContractor tests
- [ ] Test API error handling
- [ ] Test context cancellation
- [ ] Test pagination (if applicable)

### Phase 4: Database Interactions (TC-4.x)
- [ ] Implement OneByUsername store method
- [ ] Implement GetDiscordMentionFromUsername tests
- [ ] Test database error handling
- [ ] Test mention formatting
- [ ] Verify index exists/recommended

### Phase 5: Multi-Step Flow (TC-5.x)
- [ ] Implement UpdateApprovedRejectedBy tests
- [ ] Test full lookup chain
- [ ] Test graceful degradation at each step
- [ ] Test relation update format
- [ ] Test helper functions

### Phase 6: Validation
- [ ] Run all tests, verify they pass
- [ ] Check code coverage (>90% target)
- [ ] Verify logging assertions
- [ ] Review error handling completeness
- [ ] Document any deviations from designs

## Quick Reference: Test Case Locations

| Function | File | Lines | Key Scenarios |
|----------|------|-------|---------------|
| GetActiveDeploymentsForContractor | unit/01-* | - | Filter construction, empty results, pagination |
| ExtractStakeholdersFromDeployment | unit/02-* | - | Override priority, rollup extraction, deduplication |
| GetDiscordUsernameFromContractor | unit/03-* | - | Rich text concat, empty field, API errors |
| GetDiscordMentionFromUsername | unit/04-* | - | DB lookup, mention format, graceful degradation |
| UpdateApprovedRejectedBy | unit/05-* | - | Multi-step chain, relation update, all-graceful |

## Expected Code Coverage

| Package | Target | Focus Areas |
|---------|--------|-------------|
| `pkg/service/notion/leave.go` | >90% | New AM/DL functions |
| `pkg/handler/webhook/notion_leave.go` | >90% | New helper functions |
| `pkg/store/discordaccount/` | >90% | New OneByUsername method |

**Coverage Report Command:**
```bash
go test -coverprofile=coverage.out ./pkg/service/notion ./pkg/handler/webhook ./pkg/store/discordaccount
go tool cover -html=coverage.out
```

## Dependencies for Testing

### External Libraries
- `github.com/dstotijn/go-notion` - Notion API types
- `gorm.io/gorm` - Database ORM
- `github.com/stretchr/testify/assert` - Test assertions
- `github.com/stretchr/testify/require` - Test requirements

### Internal Packages
- `pkg/model` - DiscordAccount model
- `pkg/store` - Repository interfaces
- `pkg/logger` - Logging interface
- `pkg/config` - Configuration structs

### Test Helpers (Existing Patterns)
- `testhelper.TestWithTxDB()` - Transaction-wrapped tests
- SQL fixtures in `testdata/` directories
- Table-driven test pattern

## Common Pitfalls to Avoid

1. **Mock Divergence:** Ensure mocks match real Notion API responses
2. **Over-Mocking:** Don't mock pure functions (property extraction)
3. **Under-Testing Errors:** Test every error path
4. **Ignoring Logs:** Verify logging in tests (assertions)
5. **Hard-Coded IDs:** Use constants or test data builders
6. **Missing Edge Cases:** Empty strings, nil values, whitespace
7. **Context Ignored:** Pass and test context cancellation

## Success Criteria

- [ ] All 67 test cases implemented and passing
- [ ] Code coverage >90% for new functions
- [ ] Zero flaky tests (consistent pass)
- [ ] Logging verified at all levels
- [ ] Error handling comprehensive
- [ ] Mock accuracy validated
- [ ] Documentation complete

## Questions or Issues?

### For Clarification
- Review [Technical Specification](../planning/specifications/am-dl-integration-spec.md)
- Review [ADR-001](../planning/ADRs/001-am-dl-lookup-strategy.md)
- Check [Overall Test Plan](test-plans/overall-test-plan.md)

### For Implementation Guidance
- Follow fortress-api test patterns in existing webhook tests
- Reference `pkg/handler/webhook/nocodb_expense_test.go` for mock patterns
- Use table-driven tests for multiple scenarios

### For Test Data
- See individual test case documents for mock data structures
- Each test case includes specific test data examples
- Property extraction tests include complete Notion property structures

---

**Ready for Implementation:** All test case designs are complete and ready for the feature-implementer agent to create actual test code.

**Next Step:** Implement unit tests following these designs, then proceed to feature implementation following TDD principles (tests first, implementation second).
