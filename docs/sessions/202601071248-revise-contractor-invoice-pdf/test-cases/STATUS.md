# Test Case Design Status

**Session:** 202601071248-revise-contractor-invoice-pdf
**Component:** Contractor Invoice Multi-Currency Feature
**Phase:** Test Design
**Status:** COMPLETE
**Last Updated:** 2026-01-07

## Overview

Comprehensive test case designs have been created for the contractor invoice PDF multi-currency feature. All test specifications are ready for implementation by the feature-implementer agent.

## Deliverables

### Unit Test Specifications

| Test Case | Component | Test Count | Priority | Status |
|-----------|-----------|-----------|----------|--------|
| TC-001 | formatVND Function | 14 | P0 | COMPLETE |
| TC-002 | formatUSD Function | 17 | P0 | COMPLETE |
| TC-003 | formatCurrency Function | 12 | P0 | COMPLETE |
| TC-004 | formatExchangeRate Function | 14 | P1 | COMPLETE |
| TC-005 | Calculation Logic | 12 + 5 validation | P0 | COMPLETE |
| TC-006 | Data Structure Population | 12 | P1 | COMPLETE |
| TC-007 | Section Helper Functions | 20 | P2 | COMPLETE |

**Total Unit Test Specifications:** 106 test cases

### Test Plans

| Document | Description | Status |
|----------|-------------|--------|
| MASTER-TEST-PLAN.md | Overall test strategy and execution plan | COMPLETE |
| TEST-PLAN-001-edge-cases.md | Edge cases and boundary conditions (36 scenarios) | COMPLETE |
| TEST-DATA-REQUIREMENTS.md | Test data fixtures and mock requirements | COMPLETE |

### Document Structure

```
test-cases/
├── unit/
│   ├── TC-001-format-vnd-spec.md              [COMPLETE]
│   ├── TC-002-format-usd-spec.md              [COMPLETE]
│   ├── TC-003-format-currency-spec.md         [COMPLETE]
│   ├── TC-004-format-exchange-rate-spec.md    [COMPLETE]
│   ├── TC-005-calculation-logic-spec.md       [COMPLETE]
│   ├── TC-006-data-structure-spec.md          [COMPLETE]
│   └── TC-007-section-grouping-spec.md        [COMPLETE]
├── test-plans/
│   ├── MASTER-TEST-PLAN.md                    [COMPLETE]
│   ├── TEST-PLAN-001-edge-cases.md            [COMPLETE]
│   └── TEST-DATA-REQUIREMENTS.md              [COMPLETE]
└── STATUS.md                                   [COMPLETE]
```

## Test Coverage Summary

### By Component

1. **Currency Formatters (TC-001, TC-002, TC-003)**
   - VND formatting: 14 test cases
   - USD formatting: 17 test cases
   - Currency dispatcher: 12 test cases
   - Coverage: 100% of formatting requirements

2. **Calculation Logic (TC-005)**
   - Happy path scenarios: 7 test cases
   - Edge cases: 5 test cases
   - Validation scenarios: 5 test cases
   - Coverage: All calculation paths

3. **Data Structures (TC-006)**
   - LineItem population: 6 test cases
   - InvoiceData population: 6 test cases
   - Coverage: All new fields

4. **Template Helpers (TC-004, TC-007)**
   - Exchange rate formatting: 14 test cases
   - Section identification: 20 test cases
   - Coverage: All template functions

### By Priority

- **P0 (Critical):** ~50 test cases
- **P1 (High):** ~40 test cases
- **P2 (Medium):** ~30 test cases
- **P3 (Low):** ~20 test cases

**Total:** 142+ test scenarios

### Edge Case Coverage

| Category | Scenarios | Status |
|----------|-----------|--------|
| Numeric Boundaries | 8 | COMPLETE |
| Currency Scenarios | 8 | COMPLETE |
| Exchange Rate | 4 | COMPLETE |
| Data Validation | 4 | COMPLETE |
| Collections | 3 | COMPLETE |
| Formatting | 3 | COMPLETE |
| Calculations | 3 | COMPLETE |
| Template Rendering | 3 | COMPLETE |

## Test Strategy

### Approach

1. **TDD-First:** Test specifications created before implementation
2. **Comprehensive:** All requirements covered
3. **Prioritized:** P0/P1 tests identified for initial implementation
4. **Data-Driven:** Extensive test fixtures and scenarios provided
5. **Edge-Case Focused:** Boundary conditions explicitly tested

### Testing Principles Applied

- **Isolation:** Each test case is independent
- **Clarity:** Clear input/output specifications
- **Completeness:** Happy path, edge cases, and error conditions
- **Repeatability:** Deterministic test data and expectations
- **Documentation:** Rationale provided for each test case

## Key Test Scenarios

### Critical Paths (P0)

1. **All VND Invoice** - 100% VND items, verify conversion and formatting
2. **All USD Invoice** - 100% USD items, no conversion needed
3. **Mixed Currency** - VND + USD items, verify both subtotals
4. **Empty Invoice** - No line items, handle gracefully
5. **API Failure** - Wise API error, proper error handling
6. **Invalid Currency** - Non VND/USD currency, validation error

### Edge Cases Highlighted

1. **Zero Amounts** - All items have 0 value
2. **Negative Amounts** - Validation should reject
3. **Very Large Amounts** - Billions in VND, millions in USD
4. **Very Small Amounts** - Sub-cent USD, sub-dong VND
5. **Rounding Boundaries** - Exact half values (X.5)
6. **Exchange Rate = 0** - Invalid rate handling

## Implementation Guidance

### For Feature Implementer

1. **Start with P0 Tests**
   - Implement formatVND, formatUSD, formatCurrency first
   - Implement calculation logic next
   - These are blocking for core functionality

2. **Follow Test Specifications**
   - Each TC document has clear test structure
   - Copy test table structure for test implementation
   - Use provided assertion strategies

3. **Test File Organization**
   - Create separate test files as specified in MASTER-TEST-PLAN.md
   - Group related tests together
   - Use table-driven test pattern (Go best practice)

4. **Test Data**
   - Reference TEST-DATA-REQUIREMENTS.md for fixtures
   - Use pre-defined mock data
   - Add custom test data as needed

5. **Validation**
   - All P0 tests must pass before moving to implementation
   - All P1 tests should pass before code review
   - P2/P3 tests can be implemented post-launch

### Test Implementation Template

Each test file should follow this structure:

```go
func TestComponentName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        // Test cases from specification
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionUnderTest(tt.input)
            // Assertions
        })
    }
}
```

## Dependencies for Implementation

### Required Packages
- `math` - For rounding functions
- `fmt` - For string formatting
- `strings` - For string manipulation (if case-insensitive)
- `testing` - Standard Go testing
- `github.com/stretchr/testify/require` - Assertions (project standard)

### External Dependencies
- Mock Wise API service
- Mock Notion API service (if testing end-to-end)

### Test Data Files
- Create `testdata/` directory in pkg/controller/invoice/
- Add JSON fixtures for complex scenarios (optional)

## Acceptance Criteria

### Test Design Phase (COMPLETE)

- [x] All unit test specifications created
- [x] Edge case scenarios documented
- [x] Test data requirements specified
- [x] Master test plan created
- [x] Test priorities assigned
- [x] Implementation guidance provided

### Test Implementation Phase (PENDING)

- [ ] All P0 test cases implemented
- [ ] All P1 test cases implemented
- [ ] P2/P3 test cases implemented or documented
- [ ] Code coverage > 85%
- [ ] All tests passing

### Feature Implementation Phase (PENDING)

- [ ] Implementation code written
- [ ] All tests passing
- [ ] Code review complete
- [ ] Manual testing complete

## Risks and Mitigations

### Risk: Floating Point Precision

**Mitigation:**
- Explicit rounding at each step
- Use tolerance-based assertions (0.01 for USD, 1.0 for VND)
- Test specifications include precision expectations

### Risk: Exchange Rate API Dependency

**Mitigation:**
- Comprehensive error handling tests
- Mock service for all test cases
- Validation of rate values

### Risk: Test Coverage Gaps

**Mitigation:**
- 142+ test scenarios defined
- Edge cases explicitly documented
- Code coverage tools to verify

## Next Steps

### Immediate (For Feature Implementer)

1. Read MASTER-TEST-PLAN.md for execution strategy
2. Review TC-001 through TC-007 specifications
3. Start implementing P0 tests (formatters and calculations)
4. Create test fixtures from TEST-DATA-REQUIREMENTS.md
5. Write implementation code to pass tests

### Follow-up

1. Implement P1 tests after P0 complete
2. Run full test suite with coverage
3. Address any gaps identified by coverage report
4. Manual testing with real Notion/Wise data
5. Update STATUS.md when tests implemented

## Questions or Clarifications

If the feature-implementer agent has questions about test specifications:

1. Refer to related specification docs in `planning/specifications/`
2. Check ADRs in `planning/ADRs/` for design rationale
3. Review MASTER-TEST-PLAN.md for overall strategy
4. Test specifications are intentionally detailed - follow closely

## Sign-off

**Test Designer:** Test Planner Agent
**Date:** 2026-01-07
**Status:** Test design complete and ready for implementation

---

**Note:** This document should be updated by the feature-implementer agent as tests are implemented, and by the QA agent during verification.
