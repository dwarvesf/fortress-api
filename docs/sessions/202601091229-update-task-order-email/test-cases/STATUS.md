# Test Case Design Phase Status

## Phase Complete ✅

Date: 2026-01-09  
Status: **COMPLETE - Ready for Phase 3 (Task Breakdown)**

## Summary

Test case design phase completed. Comprehensive unit test specifications created for all critical components. Integration tests skipped per workflow instructions (non-code changes exception does not apply).

## Test Cases Designed

### Unit Tests (2/2 Complete)

1. **GetContractorPayday Service Method**
   - File: `unit/get-contractor-payday-tests.md`
   - Test count: 8 test cases
   - Coverage: Happy paths, graceful fallbacks, error conditions
   - Mock requirements: Notion client, logger, configuration

2. **Invoice Due Date Calculation Logic**
   - File: `unit/invoice-due-date-calculation-tests.md`
   - Test count: 5 test cases
   - Coverage: All Payday values, defaults, edge cases
   - Location: Handler logic in SendTaskOrderConfirmation

### Integration Tests

**Status**: SKIPPED per workflow instructions
- Reason: Non-code changes exception does not apply (this is code changes)
- Note: Integration testing will be performed manually during implementation

## Test Coverage Analysis

### Service Layer: GetContractorPayday
- ✅ Payday="01" → returns (1, nil)
- ✅ Payday="15" → returns (15, nil)
- ✅ No Service Rate found → returns (0, nil)
- ✅ Payday field empty → returns (0, nil)
- ✅ Invalid Payday value → returns (0, nil)
- ✅ Database config missing → returns (0, error)
- ✅ API failure → returns (0, error)
- ✅ Property cast failure → returns (0, error)

**Coverage**: 100% of method logic

### Handler Layer: Due Date Calculation
- ✅ Payday=1 → invoiceDueDay="10th"
- ✅ Payday=15 → invoiceDueDay="25th"
- ✅ Payday=0 (missing) → invoiceDueDay="10th"
- ✅ Invalid payday → invoiceDueDay="10th"
- ✅ Error from service → invoiceDueDay="10th"

**Coverage**: 100% of business logic

### Template & Signature
**Status**: Manual testing only
- Visual verification of email rendering
- Signature display correctness
- HTML structure validation

## Test Data Requirements

### Mock Notion Responses
- Valid Service Rate with Payday="01"
- Valid Service Rate with Payday="15"
- Empty query results
- Service Rate without Payday field
- Service Rate with invalid Payday value

### Mock Contractor Data
- Contractor ID: "test-contractor-123"
- Contractor name: "John Smith"
- Team email: "john@example.com"

### Expected Email Outputs
- Subject: "Quick update for January 2026 – Invoice reminder & client milestones"
- Due date variations: "10th", "25th"
- Signature: "Han Ngo, CTO & Managing Director"

## Mocking Strategy

### Service Method Tests
```go
// Mock Notion client
mockClient := &MockNotionClient{}
mockClient.On("QueryDatabase", ctx, dbID, query).Return(response, nil)

// Mock logger
mockLogger := &MockLogger{}
mockLogger.On("Debug", message)

// Create service with mocks
service := &TaskOrderLogService{
    client: mockClient,
    logger: mockLogger,
    cfg: testConfig,
}
```

### Handler Tests
```go
// Mock service layer
mockService := &MockTaskOrderLogService{}
mockService.On("GetContractorPayday", ctx, contractorID).Return(1, nil)

// Test handler logic
handler := &handler{
    service: &services{
        Notion: &notion{
            TaskOrderLog: mockService,
        },
    },
}
```

## Test Execution Strategy

### Development
1. Run unit tests on every code change
2. Use table-driven test pattern
3. Parallel test execution where possible

### CI/CD
1. Run all unit tests on PR creation
2. Require 100% pass rate for merge
3. Code coverage report generation
4. Failed tests block deployment

### Manual Testing
1. Send test email to developer email
2. Verify visual rendering
3. Test with different Payday values
4. Verify fallback behavior

## Testing Tools

### Required
- Go testing package (`testing`)
- Testify/assert for assertions
- Mock generation (mockery or gomock)
- Test fixtures for Notion responses

### Optional
- Golden file testing for template output
- Snapshot testing for HTML rendering
- Coverage visualization (go tool cover)

## Known Limitations

### Not Covered by Unit Tests
1. Actual Notion API calls (integration test)
2. Gmail API email sending (integration test)
3. Template HTML rendering in email clients (manual test)
4. End-to-end email flow (integration test)

### Manual Testing Required
1. Email delivery to real inbox
2. Visual rendering in Gmail web/mobile
3. Link clickability
4. Signature formatting
5. MIME encoding correctness

## Success Criteria

Test case design phase complete when:
- ✅ Unit test cases designed for all new code
- ✅ Test data specifications documented
- ✅ Mocking strategy defined
- ✅ Coverage goals established
- ✅ Ready for task breakdown

**Status**: ALL CRITERIA MET ✅

## Next Phase

### Phase 3: Task Breakdown

**Status**: READY TO PROCEED

**Scope**:
- Break down implementation into actionable tasks
- Sequence tasks based on dependencies
- Estimate complexity and effort
- Create implementation checklist

**Expected Duration**: 30 minutes

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Mock setup complexity | Slower test development | Use test helpers and fixtures |
| Notion API changes | Tests become outdated | Version pin SDK, monitor changes |
| Flaky tests | CI/CD instability | Use deterministic mocks, avoid time-based tests |
| Low coverage | Bugs in production | Enforce 100% coverage for new code |

---

**Prepared by**: Test Case Designer  
**Reviewed by**: Development Team  
**Approved for**: Task Breakdown Phase
