# Test Cases: Contractor Invoice Multi-Currency Feature

**Session:** 202601071248-revise-contractor-invoice-pdf
**Status:** Design Complete - Ready for Implementation
**Last Updated:** 2026-01-07

## Overview

This directory contains comprehensive test case designs for the contractor invoice PDF multi-currency feature. All test specifications are design documents only - no actual test code has been written yet.

## Quick Start

### For Feature Implementer

1. **Start Here:** Read [MASTER-TEST-PLAN.md](test-plans/MASTER-TEST-PLAN.md)
2. **Test Data:** Review [TEST-DATA-REQUIREMENTS.md](test-plans/TEST-DATA-REQUIREMENTS.md)
3. **Implementation Order:**
   - Implement TC-001 (formatVND)
   - Implement TC-002 (formatUSD)
   - Implement TC-003 (formatCurrency)
   - Implement TC-005 (Calculation Logic)
   - Implement TC-006 (Data Structures)
   - Implement TC-004 (formatExchangeRate)
   - Implement TC-007 (Section Helpers)
   - Implement Edge Cases

## Directory Structure

```
test-cases/
├── README.md                           (this file)
├── STATUS.md                           (current status and progress)
│
├── unit/                               (unit test specifications)
│   ├── TC-001-format-vnd-spec.md       (14 test cases - VND formatting)
│   ├── TC-002-format-usd-spec.md       (17 test cases - USD formatting)
│   ├── TC-003-format-currency-spec.md  (12 test cases - currency dispatcher)
│   ├── TC-004-format-exchange-rate-spec.md (14 test cases - exchange rate)
│   ├── TC-005-calculation-logic-spec.md    (17 test cases - calculations)
│   ├── TC-006-data-structure-spec.md       (12 test cases - data population)
│   └── TC-007-section-grouping-spec.md     (20 test cases - template helpers)
│
└── test-plans/                         (test planning documents)
    ├── MASTER-TEST-PLAN.md             (overall test strategy)
    ├── TEST-PLAN-001-edge-cases.md     (36 edge case scenarios)
    └── TEST-DATA-REQUIREMENTS.md       (test fixtures and mock data)
```

## Test Coverage Summary

### Total Test Specifications
- **Unit Tests:** 106 test cases across 7 specifications
- **Edge Cases:** 36 scenarios across 8 categories
- **Total:** 142+ test scenarios

### By Priority
- **P0 (Critical):** ~50 test cases - Must pass before deployment
- **P1 (High):** ~40 test cases - Should pass before deployment
- **P2 (Medium):** ~30 test cases - Nice to have
- **P3 (Low):** ~20 test cases - Optional enhancements

## Test Specifications

### TC-001: formatVND Function
**File:** [unit/TC-001-format-vnd-spec.md](unit/TC-001-format-vnd-spec.md)
**Test Count:** 14
**Priority:** P0
**Focus:** VND currency formatting with period separators and dong symbol

### TC-002: formatUSD Function
**File:** [unit/TC-002-format-usd-spec.md](unit/TC-002-format-usd-spec.md)
**Test Count:** 17
**Priority:** P0
**Focus:** USD currency formatting with comma separators and dollar symbol

### TC-003: formatCurrency Function
**File:** [unit/TC-003-format-currency-spec.md](unit/TC-003-format-currency-spec.md)
**Test Count:** 12
**Priority:** P0
**Focus:** Currency dispatcher routing to correct formatter

### TC-004: formatExchangeRate Function
**File:** [unit/TC-004-format-exchange-rate-spec.md](unit/TC-004-format-exchange-rate-spec.md)
**Test Count:** 14
**Priority:** P1
**Focus:** Exchange rate display formatting (1 USD = X VND)

### TC-005: Calculation Logic
**File:** [unit/TC-005-calculation-logic-spec.md](unit/TC-005-calculation-logic-spec.md)
**Test Count:** 12 calculation + 5 validation
**Priority:** P0
**Focus:** Subtotal calculations, VND to USD conversion, totals

### TC-006: Data Structure Population
**File:** [unit/TC-006-data-structure-spec.md](unit/TC-006-data-structure-spec.md)
**Test Count:** 12
**Priority:** P1
**Focus:** Populating new fields (OriginalAmount, OriginalCurrency, subtotals)

### TC-007: Section Grouping Helpers
**File:** [unit/TC-007-section-grouping-spec.md](unit/TC-007-section-grouping-spec.md)
**Test Count:** 20
**Priority:** P2
**Focus:** Template helper functions (isSectionHeader, isServiceFee)

## Test Plans

### Master Test Plan
**File:** [test-plans/MASTER-TEST-PLAN.md](test-plans/MASTER-TEST-PLAN.md)
**Purpose:** Overall test strategy, execution plan, priorities, and acceptance criteria

**Key Sections:**
- Test coverage overview
- Execution plan (4 phases)
- Success metrics (85% code coverage target)
- Risk assessment
- Sign-off criteria

### Edge Cases Test Plan
**File:** [test-plans/TEST-PLAN-001-edge-cases.md](test-plans/TEST-PLAN-001-edge-cases.md)
**Purpose:** Document edge cases and boundary conditions

**Categories:**
1. Numeric Boundaries (8 scenarios)
2. Currency Edge Cases (8 scenarios)
3. Exchange Rate Edge Cases (4 scenarios)
4. Data Validation (4 scenarios)
5. Collection Edge Cases (3 scenarios)
6. Formatting Edge Cases (3 scenarios)
7. Calculation Edge Cases (3 scenarios)
8. Template Rendering (3 scenarios)

### Test Data Requirements
**File:** [test-plans/TEST-DATA-REQUIREMENTS.md](test-plans/TEST-DATA-REQUIREMENTS.md)
**Purpose:** Define test fixtures, mock data, and helper functions

**Contents:**
- Mock data structures
- Pre-defined test payouts (VND/USD)
- Exchange rate fixtures
- Invoice scenario fixtures
- Formatting test data
- Helper function specifications

## Implementation Guidelines

### Test File Organization

Create these test files in `pkg/controller/invoice/`:

```
pkg/controller/invoice/
├── contractor_invoice_formatters_test.go    (TC-001, TC-002, TC-003, TC-004)
├── contractor_invoice_calculations_test.go  (TC-005)
├── contractor_invoice_data_test.go          (TC-006)
├── contractor_invoice_helpers_test.go       (TC-007)
└── contractor_invoice_edge_cases_test.go    (Edge cases from TEST-PLAN-001)
```

### Test Implementation Pattern

Each test specification follows this structure:

1. **Purpose** - What the test validates
2. **Function Signature** - Interface being tested
3. **Test Data Requirements** - Input/output specifications
4. **Test Cases** - Detailed test scenarios with expected results
5. **Assertion Strategy** - How to verify correctness
6. **Error Conditions** - What errors to handle
7. **Test Implementation Notes** - Go code structure guidance

### Recommended Order

1. **Phase 1 (Week 1):** TC-001, TC-002, TC-003, TC-005 (Core functionality)
2. **Phase 2 (Week 1):** TC-006, TC-007, TC-004 (Data & integration)
3. **Phase 3 (Week 2):** P0 and P1 edge cases
4. **Phase 4 (Week 2):** P2/P3 edge cases, final validation

## Key Testing Principles

### 1. Test-Driven Development (TDD)
- Write tests based on these specifications FIRST
- Then implement code to pass tests
- No implementation code without test

### 2. Table-Driven Tests
All test specifications are designed for Go's table-driven test pattern:

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    Type
        expected Type
    }{
        // Test cases from specification
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 3. Clear Assertions
Use testify/require for clear assertions:

```go
require.Equal(t, expected, actual, "description")
require.NoError(t, err)
require.True(t, condition, "description")
```

### 4. Comprehensive Coverage
- Happy path tests
- Edge case tests
- Error condition tests
- Boundary condition tests

## Success Criteria

### Test Design Phase (COMPLETE)
- [x] All unit test specifications created
- [x] Edge case scenarios documented
- [x] Test data requirements defined
- [x] Master test plan written
- [x] Implementation guidance provided

### Test Implementation Phase (NEXT)
- [ ] All P0 tests implemented and passing
- [ ] All P1 tests implemented and passing
- [ ] Code coverage > 85%
- [ ] All edge cases handled

### Feature Complete Phase (FUTURE)
- [ ] All tests passing
- [ ] Implementation code complete
- [ ] Code review approved
- [ ] Manual testing verified

## Common Questions

### Q: Are these actual test implementations?
**A:** No, these are test specifications/designs. The actual test code needs to be written based on these specs.

### Q: Do I need to implement all 142 test cases?
**A:** Start with P0 tests (~50 cases), then P1 (~40 cases). P2/P3 can be deferred.

### Q: What testing framework should I use?
**A:** Go's standard `testing` package with `testify/require` for assertions (project standard).

### Q: Can I modify the test specifications?
**A:** Test specs are design documents. If requirements change, update the specs first, then tests.

### Q: Where do I put test fixtures?
**A:** Create `testdata/` directory in `pkg/controller/invoice/` for fixtures.

## Related Documentation

### Planning Documents
- [planning/specifications/spec-001-data-structure-changes.md](../planning/specifications/spec-001-data-structure-changes.md)
- [planning/specifications/spec-002-template-functions.md](../planning/specifications/spec-002-template-functions.md)
- [planning/specifications/spec-004-calculation-logic.md](../planning/specifications/spec-004-calculation-logic.md)

### Architecture Decisions
- [planning/ADRs/ADR-001-data-structure-multi-currency.md](../planning/ADRs/ADR-001-data-structure-multi-currency.md)
- [planning/ADRs/ADR-002-currency-formatting-approach.md](../planning/ADRs/ADR-002-currency-formatting-approach.md)

### Requirements
- [requirements/requirements.md](../requirements/requirements.md)

## Contact

For questions or clarifications about test specifications, refer to:
1. This README
2. MASTER-TEST-PLAN.md for strategy
3. Individual TC specifications for details
4. Planning documents for requirements context

---

**Last Updated:** 2026-01-07
**Test Designer:** Test Planner Agent
**Status:** Ready for Implementation
