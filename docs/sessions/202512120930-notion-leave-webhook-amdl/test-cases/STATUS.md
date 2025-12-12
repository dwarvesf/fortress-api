# Test Cases Status

**Session:** 202512120930-notion-leave-webhook-amdl
**Feature:** Notion Leave Webhook AM/DL Integration
**Phase:** Test Case Design - UNIT TESTS ONLY
**Last Updated:** 2025-12-12

## Overview

Comprehensive unit test case designs for the Notion Leave Webhook AM/DL integration feature. These test designs cover all new service layer functions and handler functions required to automatically fetch Account Managers and Delivery Leads from the Deployment Tracker.

**Note:** Per user requirements, this phase covers UNIT TEST DESIGNS ONLY. Integration tests are explicitly excluded from this scope.

## Test Case Design Documents

### Unit Test Designs

| # | Function | Test Cases | Document | Status |
|---|----------|------------|----------|--------|
| 1 | GetActiveDeploymentsForContractor | 8 | [01-get-active-deployments-for-contractor.md](unit/01-get-active-deployments-for-contractor.md) | Complete |
| 2 | ExtractStakeholdersFromDeployment | 14 | [02-extract-stakeholders-from-deployment.md](unit/02-extract-stakeholders-from-deployment.md) | Complete |
| 3 | GetDiscordUsernameFromContractor | 14 | [03-get-discord-username-from-contractor.md](unit/03-get-discord-username-from-contractor.md) | Complete |
| 4 | GetDiscordMentionFromUsername | 14 | [04-get-discord-mention-from-username.md](unit/04-get-discord-mention-from-username.md) | Complete |
| 5 | UpdateApprovedRejectedBy | 17 | [05-update-approved-rejected-by.md](unit/05-update-approved-rejected-by.md) | Complete |

**Total Unit Test Cases Designed:** 67

### Integration Tests

**Status:** Out of scope per user requirements - SKIPPED

Integration tests will be designed and implemented in a separate phase if needed.

## Test Coverage Summary

### Functions Covered

#### Service Layer (`pkg/service/notion/leave.go`)
- [x] `GetActiveDeploymentsForContractor` - Query Deployment Tracker for active deployments
- [x] `ExtractStakeholdersFromDeployment` - Extract AM/DL with override priority logic
- [x] `GetDiscordUsernameFromContractor` - Fetch Discord username from Notion contractor page

#### Handler Layer (`pkg/handler/webhook/notion_leave.go`)
- [x] `getDiscordMentionFromUsername` - Convert username to Discord mention format
- [x] `updateApprovedRejectedBy` - Update Notion relation on approval/rejection

#### Store Layer (`pkg/store/discordaccount/`)
- [x] `OneByUsername` - NEW METHOD - Database lookup by Discord username

### Test Categories

| Category | Count | Coverage |
|----------|-------|----------|
| Happy Path | 10 | All primary workflows |
| Error Handling | 22 | API errors, database errors, network failures |
| Edge Cases | 18 | Empty values, nil checks, boundary conditions |
| Input Validation | 7 | Empty inputs, invalid formats |
| Property Extraction | 10 | Notion property types and structures |

### Coverage Areas

#### Notion API Interactions
- [x] QueryDatabase with compound filters
- [x] FindPageByID for contractor pages
- [x] UpdatePage with relation properties
- [x] Pagination handling
- [x] API error responses (404, 429, 500)
- [x] Context cancellation

#### Property Extraction
- [x] Relation properties (single, multiple, empty)
- [x] Rollup properties (array of relations, nested structures)
- [x] Rich text properties (single part, multi-part, concatenation)
- [x] Email properties
- [x] Select/Status properties
- [x] Missing properties
- [x] Invalid property types

#### Database Operations
- [x] Discord account lookup by ID
- [x] Discord account lookup by username (NEW)
- [x] Record not found handling
- [x] Database errors
- [x] Transaction handling (via testhelper patterns)

#### Business Logic
- [x] Override AM/DL priority over rollup
- [x] Stakeholder deduplication
- [x] Multi-deployment aggregation
- [x] Graceful degradation on failures
- [x] Discord mention formatting
- [x] Relation update correctness

## Test Design Patterns

### Mock Structures
- **Mock Notion Client:** QueryDatabase, FindPageByID, UpdatePage
- **Mock Database:** GORM with test fixtures
- **Mock Store:** DiscordAccount store with configurable responses
- **Mock Logger:** Verify logging behavior and levels

### Assertion Strategies
- Table-driven tests for multiple scenarios
- API call verification (parameters, call count)
- Data structure validation
- Error type and message verification
- Logging level and content verification

### Test Data Patterns
- Realistic Notion property structures
- Edge case property configurations
- Database records with various states
- Discord ID and username formats

## Key Test Scenarios

### Critical Path Testing
1. **Full AM/DL Lookup Flow**
   - Email → Contractor → Deployments → Stakeholders → Usernames → Mentions
   - Tests: TC-1.1, TC-2.1, TC-3.1, TC-4.1

2. **Override Priority Logic**
   - Override AM/DL takes precedence over rollup
   - Tests: TC-2.1, TC-2.2, TC-2.3, TC-2.4

3. **Graceful Degradation**
   - Empty deployments → Empty mentions (no failure)
   - Contractor not found → Skip notification (log warning)
   - Tests: TC-1.3, TC-4.2, TC-5.4

4. **Approval/Rejection Metadata**
   - Discord ID → Username → Contractor → Update relation
   - Tests: TC-5.1, TC-5.2, TC-5.3, TC-5.4

### Edge Case Testing
- Empty/nil properties at every extraction point
- Missing properties in Notion responses
- Multiple stakeholders in same deployment
- Same stakeholder as both AM and DL (deduplication)
- Nested rollup arrays with multiple relations
- Context cancellation during API calls

### Error Handling Testing
- Notion API errors (network, rate limit, not found)
- Database query errors
- Invalid property types (schema mismatches)
- Input validation (empty required parameters)
- Partial failure scenarios (some stakeholders succeed, others fail)

## New Store Method Requirements

### OneByUsername Implementation

**File:** `pkg/store/discordaccount/discordaccount.go`

**Signature:**
```go
func (s *store) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
```

**Test Coverage:**
- Find existing account by username
- Return nil, nil for not found (graceful)
- Return error for database failures
- Handle empty username input

**Database Index:**
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

**Test Cases:** TC-4.11, TC-4.12, TC-4.13, TC-4.14

## Logging Verification

### Debug Level (30 assertions)
- Query parameters and filters
- Extracted property values
- Lookup results at each step

### Info Level (8 assertions)
- Successful operations
- Empty results (expected condition)

### Warning Level (15 assertions)
- Contractor not found
- Discord username not found
- Property not found (expected to exist)
- Multiple records found (data issue)

### Error Level (14 assertions)
- API call failures
- Database query errors
- Invalid property types
- Unexpected failures

## Configuration Requirements

### Environment Variables
- `NOTION_DEPLOYMENT_TRACKER_DB_ID` - NEW
- `NOTION_CONTRACTOR_DB_ID` - Existing

### Config Validation
- Tests verify missing config values return errors
- Tests verify config values used correctly in queries

## Dependencies and Prerequisites

### External Libraries
- `github.com/dstotijn/go-notion` - Notion API client
- `gorm.io/gorm` - Database ORM
- `github.com/stretchr/testify` - Test assertions

### Internal Packages
- `pkg/model` - Data models
- `pkg/store` - Repository layer
- `pkg/service/notion` - Notion service layer
- `pkg/handler/webhook` - Webhook handlers
- `pkg/logger` - Logging interface

### Test Helpers
- `testhelper.TestWithTxDB()` - Database transaction wrapper (existing pattern)
- Mock Notion client implementation (to be created)
- Mock store implementations (to be created)

## Implementation Readiness

### Ready for Implementation
All test case designs are complete and ready for implementation. Each test case includes:
- Clear Given/When/Then structure
- Expected inputs and outputs
- Mock data structures
- Assertion requirements
- Rationale and context

### Implementation Order (Recommended)
1. **Phase 1:** Property extraction (TC-2.x) - No external dependencies
2. **Phase 2:** Notion API interactions (TC-1.x, TC-3.x) - Mock Notion client
3. **Phase 3:** Database interactions (TC-4.x) - New store method
4. **Phase 4:** Multi-step flows (TC-5.x) - Combine all mocks
5. **Phase 5:** Coverage validation and refinement

## Metrics and Goals

### Coverage Goals
- **Code Coverage:** >90% for all new functions
- **Branch Coverage:** >85% for conditional logic
- **Error Path Coverage:** 100% for error handling

### Quality Metrics
- **Test Execution Time:** <10 seconds for full unit test suite
- **Flaky Tests:** Zero tolerance
- **Mock Accuracy:** Mocks validated against real API during integration testing

## Risks and Assumptions

### Assumptions
1. Notion API property structures remain stable
2. Discord account username field is populated and accurate
3. Deployment Tracker data is maintained and up-to-date
4. Contractors database has Discord usernames for key stakeholders

### Risks
1. **Notion API Changes:** Property types or structures change
   - **Mitigation:** Comprehensive property extraction tests, graceful degradation
2. **Data Quality:** Incomplete or inconsistent data in Notion
   - **Mitigation:** Extensive edge case testing, logging for manual follow-up
3. **Mock Divergence:** Mocks don't match real Notion API behavior
   - **Mitigation:** Validate mocks with real API responses in integration tests

## Next Steps

### For Feature Implementer
1. Review test case designs for completeness and accuracy
2. Implement mock structures (Notion client, database)
3. Implement test fixtures and helpers
4. Implement unit tests following test case designs
5. Verify >90% code coverage
6. Document any deviations from test designs

### For QA Agent (Future)
1. Execute implemented unit tests
2. Validate test coverage metrics
3. Verify logging behavior
4. Review error handling completeness
5. Integration test design (separate phase)

## Related Documents

### Requirements and Planning
- [Requirements](../requirements/requirements.md)
- [Technical Specification](../planning/specifications/am-dl-integration-spec.md)
- [ADR-001: AM/DL Lookup Strategy](../planning/ADRs/001-am-dl-lookup-strategy.md)

### Test Plans
- [Overall Test Plan](test-plans/overall-test-plan.md)

### Unit Test Case Designs
1. [GetActiveDeploymentsForContractor](unit/01-get-active-deployments-for-contractor.md)
2. [ExtractStakeholdersFromDeployment](unit/02-extract-stakeholders-from-deployment.md)
3. [GetDiscordUsernameFromContractor](unit/03-get-discord-username-from-contractor.md)
4. [GetDiscordMentionFromUsername](unit/04-get-discord-mention-from-username.md)
5. [UpdateApprovedRejectedBy](unit/05-update-approved-rejected-by.md)

## Approval and Sign-off

### Test Case Design Review
- [ ] Test architect approval
- [ ] Development team review
- [ ] QA team review

### Ready for Implementation
- [x] All test cases documented
- [x] Mock structures defined
- [x] Test data patterns established
- [x] Assertion strategies documented
- [x] Edge cases identified
- [ ] Implementation team notified

---

**Status Summary:** Test case designs COMPLETE. Ready for implementation phase.

**Total Test Cases:** 67 unit test cases designed across 5 functions.

**Coverage:** Comprehensive coverage of happy paths, error paths, edge cases, and input validation.

**Next Phase:** Implementation of unit tests by feature-implementer agent.
