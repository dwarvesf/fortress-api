# Master Test Plan: Contractor Invoice Multi-Currency Feature

**Version:** 1.0
**Date:** 2026-01-07
**Feature:** Multi-Currency Contractor Invoice PDF Generation
**Status:** Test Design Complete - Ready for Implementation

## Executive Summary

This master test plan provides comprehensive test coverage for the contractor invoice PDF multi-currency feature, including:
- Currency formatting functions (VND, USD)
- Multi-currency calculation logic
- Data structure population
- Template helper functions
- Edge cases and boundary conditions

## Test Coverage Overview

### Unit Tests

| Component | Test Cases | Priority | Status |
|-----------|-----------|----------|--------|
| formatVND | 14 | P0 | Design Complete |
| formatUSD | 17 | P0 | Design Complete |
| formatCurrency | 12 | P0 | Design Complete |
| formatExchangeRate | 14 | P1 | Design Complete |
| Calculation Logic | 12 + 5 validation | P0 | Design Complete |
| Data Structure | 12 | P1 | Design Complete |
| Section Helpers | 20 | P2 | Design Complete |

**Total Unit Test Cases:** 106 test case specifications

### Edge Case Tests

| Category | Test Cases | Priority | Status |
|----------|-----------|----------|--------|
| Numeric Boundaries | 8 | P0-P1 | Design Complete |
| Currency Scenarios | 8 | P0-P1 | Design Complete |
| Exchange Rate | 4 | P0 | Design Complete |
| Data Validation | 4 | P0 | Design Complete |
| Collections | 3 | P1-P2 | Design Complete |
| Formatting | 3 | P2 | Design Complete |
| Calculations | 3 | P1-P2 | Design Complete |
| Template Rendering | 3 | P2-P3 | Design Complete |

**Total Edge Case Scenarios:** 36 edge case specifications

### Total Test Coverage

- **Unit Test Specifications:** 106
- **Edge Case Specifications:** 36
- **Test Data Fixtures:** 20+ scenarios
- **Total Test Scenarios:** 142+

## Test Organization

### Test Files Structure

```
pkg/controller/invoice/
├── contractor_invoice_formatters_test.go    (TC-001, TC-002, TC-003, TC-004)
├── contractor_invoice_calculations_test.go  (TC-005)
├── contractor_invoice_data_test.go          (TC-006)
├── contractor_invoice_helpers_test.go       (TC-007)
└── contractor_invoice_edge_cases_test.go    (Edge cases)
```

### Test Documentation Structure

```
docs/sessions/202601071248-revise-contractor-invoice-pdf/test-cases/
├── unit/
│   ├── TC-001-format-vnd-spec.md
│   ├── TC-002-format-usd-spec.md
│   ├── TC-003-format-currency-spec.md
│   ├── TC-004-format-exchange-rate-spec.md
│   ├── TC-005-calculation-logic-spec.md
│   ├── TC-006-data-structure-spec.md
│   └── TC-007-section-grouping-spec.md
├── test-plans/
│   ├── MASTER-TEST-PLAN.md (this file)
│   ├── TEST-PLAN-001-edge-cases.md
│   └── TEST-DATA-REQUIREMENTS.md
└── STATUS.md
```

## Test Priorities

### P0 - Critical (Must Pass)

Must pass before feature can be deployed to production:

1. **formatVND** - Core VND formatting
2. **formatUSD** - Core USD formatting
3. **formatCurrency** - Currency dispatcher
4. **Calculation Logic** - All subtotal/total calculations
5. **Data Structure Population** - All new fields populated
6. **Empty Invoice** - Handle zero items
7. **All VND Invoice** - VND-only scenario
8. **All USD Invoice** - USD-only scenario
9. **Exchange Rate API Failure** - Error handling
10. **Invalid Currency Validation** - Input validation

**P0 Test Count:** ~50 test cases

### P1 - High (Should Pass)

Should pass before deployment, may require fixes:

1. **formatExchangeRate** - Exchange rate display
2. **Section Helpers** - Template helpers
3. **Mixed Currency Invoice** - Combined scenario
4. **Negative Amount Validation** - Input validation
5. **Large Amounts** - Boundary conditions
6. **Small Amounts** - Boundary conditions

**P1 Test Count:** ~40 test cases

### P2 - Medium (Can Defer)

Nice to have, can be deferred to post-launch:

1. **Rounding Edge Cases** - Exact half rounding
2. **Large Item Count** - Performance testing
3. **Formatting Boundaries** - Separator edge cases

**P2 Test Count:** ~30 test cases

### P3 - Low (Optional)

Enhancement testing, not blocking:

1. **Very Long Descriptions** - Layout testing
2. **Special Characters** - Character encoding
3. **Extreme Exchange Rates** - Unrealistic scenarios

**P3 Test Count:** ~20 test cases

## Test Execution Plan

### Phase 1: Core Functionality (Week 1)

**Goal:** Verify core formatting and calculation logic

**Tasks:**
1. Implement formatVND tests (TC-001)
2. Implement formatUSD tests (TC-002)
3. Implement formatCurrency tests (TC-003)
4. Implement calculation logic tests (TC-005)

**Success Criteria:**
- All P0 formatting tests pass
- All P0 calculation tests pass
- Code coverage > 80% for new functions

### Phase 2: Data & Integration (Week 1)

**Goal:** Verify data structure population and template helpers

**Tasks:**
1. Implement data structure tests (TC-006)
2. Implement section helper tests (TC-007)
3. Implement formatExchangeRate tests (TC-004)

**Success Criteria:**
- All data population tests pass
- Template helpers work correctly
- Exchange rate formatting accurate

### Phase 3: Edge Cases (Week 2)

**Goal:** Verify robust handling of edge cases

**Tasks:**
1. Implement P0 edge case tests
2. Implement P1 edge case tests
3. Implement error handling tests

**Success Criteria:**
- No panics for any edge case
- All validation errors return clear messages
- API failure handling works

### Phase 4: Final Validation (Week 2)

**Goal:** Complete test coverage and validation

**Tasks:**
1. Implement P2 edge case tests
2. Run full test suite
3. Verify code coverage
4. Document any deferred tests

**Success Criteria:**
- All P0 and P1 tests pass
- Code coverage > 85%
- P2/P3 tests documented for future

## Test Execution Commands

### Run All Tests
```bash
go test ./pkg/controller/invoice -v
```

### Run Specific Test Suites
```bash
# Formatting tests
go test ./pkg/controller/invoice -run TestFormat

# Calculation tests
go test ./pkg/controller/invoice -run TestCalculate

# Data structure tests
go test ./pkg/controller/invoice -run TestLineItem
go test ./pkg/controller/invoice -run TestInvoiceData

# Edge case tests
go test ./pkg/controller/invoice -run TestEdgeCase
```

### Run with Coverage
```bash
go test ./pkg/controller/invoice -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Data Requirements

See **TEST-DATA-REQUIREMENTS.md** for:
- Mock data structures
- Pre-defined test fixtures
- Test scenarios
- Helper functions

## Success Metrics

### Code Coverage Targets

- **Overall:** > 85%
- **New Functions:** 100%
- **Calculation Logic:** > 95%
- **Formatting Functions:** 100%
- **Error Handling:** 100%

### Quality Gates

- [ ] All P0 tests pass
- [ ] All P1 tests pass
- [ ] No panics in any test
- [ ] All error cases return clear messages
- [ ] Code coverage meets targets
- [ ] All test documentation complete

## Risk Assessment

### High Risk Areas

1. **Exchange Rate API Dependency**
   - **Risk:** Wise API failure during invoice generation
   - **Mitigation:** Comprehensive error handling tests
   - **Test Coverage:** TC-005-V3, EC-009

2. **Floating Point Precision**
   - **Risk:** Rounding errors in calculations
   - **Mitigation:** Consistent rounding strategy, precision tests
   - **Test Coverage:** TC-005-08, TC-005-09, EC-023

3. **Currency Validation**
   - **Risk:** Invalid currency codes from Notion
   - **Mitigation:** Strict validation, clear errors
   - **Test Coverage:** TC-005-V1, EC-013

### Medium Risk Areas

1. **Large Amounts**
   - **Risk:** Overflow or formatting issues
   - **Mitigation:** Boundary condition tests
   - **Test Coverage:** EC-001

2. **Empty/Zero Data**
   - **Risk:** Division by zero or empty array errors
   - **Mitigation:** Edge case handling
   - **Test Coverage:** EC-003, EC-017

## Dependencies

### External Systems
- Wise API (for currency conversion)
- Notion API (for payout data)

### Internal Dependencies
- pkg/service/wise
- pkg/service/notion
- Template rendering engine

### Testing Dependencies
- testify/require (assertions)
- gomock (mocking, if needed)
- Test fixtures in testdata/

## Acceptance Criteria

### Feature Implementation Complete When:

- [ ] All P0 test cases implemented and passing
- [ ] All P1 test cases implemented and passing
- [ ] Code coverage > 85%
- [ ] No critical bugs found
- [ ] All edge cases handled gracefully
- [ ] Error messages are clear and actionable
- [ ] Test documentation complete
- [ ] Code review approved

### Ready for Production When:

- [ ] All acceptance criteria met
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing complete
- [ ] Performance acceptable (<2s for invoice generation)
- [ ] Security review complete
- [ ] Documentation updated

## Related Documents

### Specifications
- spec-001-data-structure-changes.md
- spec-002-template-functions.md
- spec-004-calculation-logic.md

### Architecture Decisions
- ADR-001-data-structure-multi-currency.md
- ADR-002-currency-formatting-approach.md

### Test Cases
- TC-001 through TC-007 (Unit test specifications)
- TEST-PLAN-001-edge-cases.md
- TEST-DATA-REQUIREMENTS.md

## Sign-off

### Test Plan Approval

- [ ] Test Designer: _____________________ Date: _____
- [ ] Project Manager: ___________________ Date: _____
- [ ] Tech Lead: _________________________ Date: _____

### Test Execution Sign-off

- [ ] Test Implementer: __________________ Date: _____
- [ ] QA Reviewer: ______________________ Date: _____
- [ ] Feature Approved for Release: ______ Date: _____
