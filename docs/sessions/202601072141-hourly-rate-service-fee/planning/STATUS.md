# Planning Phase Status

**Session**: 202601072141-hourly-rate-service-fee
**Feature**: Hourly Rate-Based Service Fee Display in Contractor Invoices
**Date**: 2026-01-07
**Phase**: Planning Complete
**Status**: Ready for Test Case Design

---

## Executive Summary

Planning phase completed successfully. All architectural decisions documented in ADRs, detailed specifications created for implementation team. Feature is well-scoped, risk-assessed, and ready for test case design phase.

**Confidence Level**: High - All major design decisions made with clear rationale

---

## Planning Deliverables

### Architecture Decision Records (ADRs)

Created 4 ADRs documenting key architectural decisions:

1. **ADR-001: Data Fetching Strategy**
   - Decision: Sequential, lazy-loading data fetching
   - Rationale: Low volume (1-3 Service Fees per invoice), simpler error handling
   - Trade-offs: Slightly slower vs simpler implementation
   - Status: Proposed

2. **ADR-002: Aggregation Approach**
   - Decision: Post-processing aggregation after line item creation
   - Rationale: Minimal disruption to existing flow, clear separation of concerns
   - Trade-offs: Two-pass processing vs maintainability
   - Status: Proposed

3. **ADR-003: Error Handling and Fallback Strategy**
   - Decision: Defensive programming with graceful degradation at every step
   - Rationale: NFR-2 requirement - no invoice generation failures
   - Key principle: Always fallback to default display (Qty=1)
   - Status: Proposed

4. **ADR-004: Code Organization**
   - Decision: Modify existing files, add helper functions
   - Rationale: Follow existing patterns, minimal new abstractions
   - File impact: 4 files modified, +285 lines total
   - Status: Proposed

### Detailed Specifications

Created 4 specifications with implementation details:

1. **spec-001: Data Structures**
   - Modified: PayoutEntry, ContractorInvoiceLineItem (3 new fields)
   - New: hourlyRateData, hourlyRateAggregation
   - All structures documented with examples and edge cases
   - Testing data provided

2. **spec-002: Service Methods**
   - FetchContractorRateByPageID (ContractorRatesService)
   - FetchTaskOrderHoursByPageID (TaskOrderLogService)
   - Full implementation code provided
   - Error handling and logging specified
   - Unit test cases defined

3. **spec-003: Detection and Aggregation Logic**
   - fetchHourlyRateData algorithm (step-by-step)
   - aggregateHourlyServiceFees algorithm
   - Helper functions: generateServiceFeeTitle, concatenateDescriptions
   - Decision trees and flowcharts included
   - Edge cases documented

4. **spec-004: Integration**
   - Integration points in GenerateContractorInvoice
   - Modified flow diagrams
   - Backward compatibility analysis
   - Performance impact assessment
   - Rollback plan

---

## Architectural Approach

### High-Level Design

```
┌──────────────────────────────────────────────────────────┐
│ Invoice Generation Flow                                  │
└──────────────────┬───────────────────────────────────────┘
                   │
                   v
┌──────────────────────────────────────────────────────────┐
│ 1. Query Rates, Payouts, Bank Account                   │
└──────────────────┬───────────────────────────────────────┘
                   │
                   v
┌──────────────────────────────────────────────────────────┐
│ 2. Build Line Items (per payout)                        │
│    ├─ For Service Fees with ServiceRateID:              │
│    │  ├─ Fetch Contractor Rate (BillingType, Rate)      │
│    │  ├─ Check BillingType = "Hourly Rate"?             │
│    │  ├─ Fetch Task Order Hours                         │
│    │  └─ Create hourly line item OR fallback            │
│    └─ For others: Create default line item              │
└──────────────────┬───────────────────────────────────────┘
                   │
                   v
┌──────────────────────────────────────────────────────────┐
│ 3. Aggregate Hourly Service Fees [NEW]                  │
│    ├─ Identify items with IsHourlyRate=true             │
│    ├─ Sum hours, amounts                                │
│    ├─ Concatenate descriptions                          │
│    └─ Replace with single aggregated item               │
└──────────────────┬───────────────────────────────────────┘
                   │
                   v
┌──────────────────────────────────────────────────────────┐
│ 4. Group Commissions, Sort, Calculate, Return           │
└──────────────────────────────────────────────────────────┘
```

### Key Design Patterns

1. **Sequential Data Fetching**: Fetch data only when needed (lazy loading)
2. **Post-Processing Aggregation**: Build all items first, then aggregate
3. **Graceful Degradation**: Fallback to default display on any error
4. **Defensive Programming**: Every decision point has error handling
5. **Separation of Concerns**: Services fetch data, controller orchestrates logic

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin HTTP framework
- **ORM**: GORM
- **External API**: Notion API (via go-notion client)
- **Logging**: Structured logger with DEBUG/ERROR/WARN levels

---

## Risk Assessment

### Low Risk

1. **Service layer changes**: Simple data fetching methods
   - New methods follow existing patterns
   - No breaking changes to existing methods
   - Mitigation: Unit tests for all scenarios

2. **Error handling**: Graceful degradation strategy
   - Fallback to default display on any error
   - No invoice generation failures
   - Mitigation: Comprehensive error logging

### Medium Risk

1. **Controller aggregation logic**: New algorithm
   - Multiple line items consolidated into one
   - Complex logic with edge cases
   - Mitigation:
     - Detailed specifications with step-by-step algorithms
     - Extensive unit tests (90%+ coverage target)
     - Golden file comparison for integration tests
     - Peer code review

2. **Integration point**: Modifying existing invoice generation flow
   - Changes to critical business logic
   - Risk of regression in existing functionality
   - Mitigation:
     - All changes additive (no deletion of existing code)
     - Existing tests must pass
     - Integration tests verify end-to-end flow
     - Feature can be disabled by commenting out aggregation call

### Risk Mitigation Summary

- **Backward Compatibility**: All changes backward compatible, fallback to current behavior
- **Testing**: Unit + integration tests with >80% coverage
- **Logging**: DEBUG logs at every decision point for troubleshooting
- **Rollback**: Quick rollback by commenting out aggregation call
- **Monitoring**: Track fallback rate, error types, zero hours count

---

## Implementation Estimate

### Code Changes

```
SERVICE LAYER (pkg/service/notion/):
  contractor_payouts.go:    +10 lines
  contractor_rates.go:      +40 lines
  task_order_log.go:        +35 lines

CONTROLLER LAYER (pkg/controller/invoice/):
  contractor_invoice.go:    +200 lines
    - New structs:           +20 lines
    - Modified struct:       +3 fields
    - Helper functions:      +150 lines
    - Integration point:     +40 lines

TOTAL NEW CODE:             ~285 lines across 4 files
```

### Testing Code (Estimated)

```
UNIT TESTS:
  Service methods:          +150 lines
  Controller helpers:       +300 lines
  Integration tests:        +200 lines

TOTAL TEST CODE:            ~650 lines
```

### Complexity

- **Service methods**: Low complexity (simple data fetching)
- **Aggregation logic**: Medium complexity (multiple steps, edge cases)
- **Integration**: Medium complexity (modify existing critical flow)

**Overall**: Medium complexity feature

---

## Design Patterns Used

### 1. Repository Pattern
- Services act as repositories for Notion data
- Encapsulate data access logic
- Reusable methods across features

### 2. Helper Function Pattern
- Private helper functions for complex logic
- Single responsibility per function
- Testable in isolation

### 3. Fail-Safe / Graceful Degradation
- Every fetch can fail without breaking invoice generation
- Fallback to safe default behavior
- Comprehensive error logging

### 4. Post-Processing Pipeline
- Build → Identify → Aggregate → Display
- Clear boundaries between phases
- Easy to test each phase independently

### 5. Lazy Loading
- Fetch data only when needed
- Avoid unnecessary API calls
- Conditional fetching based on early detection

---

## Success Criteria

### Functional Requirements (From requirements.md)

- [x] FR-1: Hourly rate detection via ServiceRateID → BillingType = "Hourly Rate"
- [x] FR-2: Data retrieval from Contractor Payouts, Contractor Rates, Task Order Log
- [x] FR-3: Line item display with hours, rate, amount
- [x] FR-4: Aggregation of ALL hourly Service Fees into single item
- [x] FR-5: Multi-currency support (USD and VND)
- [x] FR-6: Fallback behavior on missing data or errors

### Non-Functional Requirements

- [x] NFR-1: Backward compatibility (no breaking changes)
- [x] NFR-2: Graceful error handling (no invoice generation failures)
- [x] NFR-3: Performance (2-6 API calls per invoice, <500ms overhead)
- [x] NFR-4: Code quality (follows existing patterns, DEBUG logging)

### Planning Phase Success Criteria

- [x] All ADRs created with clear rationale
- [x] All specifications completed with implementation details
- [x] Architectural decisions documented and justified
- [x] Risk assessment completed
- [x] Edge cases identified and handled
- [x] Backward compatibility verified
- [x] Testing strategy defined
- [x] Rollback plan documented

---

## Key Decision Summary

| Decision Area | Decision Made | Rationale |
|--------------|---------------|-----------|
| Data Fetching | Sequential, lazy loading | Low volume, simpler error handling |
| Aggregation | Post-processing after line item creation | Minimal disruption, clear separation |
| Error Handling | Graceful degradation with fallback | NFR-2 requirement, no invoice failures |
| Code Organization | Modify existing files, add helpers | Follow existing patterns, minimal abstractions |
| Service Methods | Add to existing services | Methods belong naturally to services |
| Title Format | "Service Fee (Development work from {date} to {date})" | Requirements FR-3 |
| Description | Concatenated with double line breaks | Clear separation between items |
| Hours Fallback | 0 hours (continue processing) | Graceful degradation, show actual amount |
| Currency Validation | Log warning, use first currency | Shouldn't happen per requirements |

---

## Assumptions and Constraints

### Assumptions

1. All hourly Service Fees for same contractor use same hourly rate
2. Contractor Rate data available for current month
3. Task Order Log has "Final Hours Worked" formula field
4. Service Fee payouts have TaskOrderID populated
5. Notion API available and responsive (<500ms per call)
6. Low volume: 1-3 Service Fees per invoice

### Constraints

1. No Notion database schema changes allowed
2. No changes to PDF template (uses existing fields)
3. Must maintain multi-currency support (USD, VND)
4. Must follow existing DEBUG logging pattern
5. Must reuse existing service methods where possible
6. No breaking changes to existing invoice generation

### Technical Constraints

1. Sequential processing acceptable (per NFR-3)
2. No real-time aggregation (post-processing only)
3. No database persistence of hourly rate data
4. All data fetched per-invoice (no caching)

---

## Open Questions / Decisions Deferred

### For Implementation Phase

1. **AmountUSD calculation**: Line item AmountUSD field usage in aggregation
   - Current approach: Use amountUSD from parallel conversion
   - Verify: Does aggregated item need separate USD conversion?

2. **Currency validation**: What if hourly items have different currencies?
   - Current approach: Log warning, use first currency
   - Per requirements: Shouldn't happen (all Service Fees same currency)
   - Decision: Leave as warning for now, monitor in production

3. **Rate validation**: What if hourly items have different rates?
   - Current approach: Log warning, use first rate
   - Per requirements: Assumes same rate for all hourly items
   - Decision: Leave as warning, can enhance later if needed

### For Future Enhancement

1. **Dynamic FX support fee**: Currently hardcoded at $8
   - Deferred: Out of scope per requirements
2. **Section grouping**: Group development work items by project
   - Deferred: Separate feature
3. **Rate caching**: Cache contractor rates for invoice month
   - Deferred: Optimization not needed for current volume

---

## Next Steps

### Immediate Next Actions

1. **Review and Approval**
   - [ ] Review all ADRs with stakeholders
   - [ ] Review all specifications with development team
   - [ ] Approve architectural decisions

2. **Test Case Design Phase**
   - [ ] Create unit test specifications for all functions
   - [ ] Create integration test scenarios
   - [ ] Design golden file test cases
   - [ ] Document edge case tests

3. **Implementation Preparation**
   - [ ] Set up feature branch
   - [ ] Review existing codebase patterns
   - [ ] Prepare test databases in Notion

### Implementation Phase Handoff

**Ready for Implementation When**:
- All ADRs approved by stakeholders
- All specifications reviewed by development team
- Test cases designed and approved
- Test databases prepared

**Implementation Order**:
1. Service layer changes (PayoutEntry, service methods)
2. Controller helper functions (fetchHourlyRateData, aggregation)
3. Integration into GenerateContractorInvoice
4. Unit tests for all new code
5. Integration tests end-to-end
6. Code review and testing
7. Deployment to staging
8. Production deployment

---

## Document Index

### Planning Documents

```
docs/sessions/202601072141-hourly-rate-service-fee/planning/
├── ADRs/
│   ├── ADR-001-data-fetching-strategy.md
│   ├── ADR-002-aggregation-approach.md
│   ├── ADR-003-error-handling-strategy.md
│   └── ADR-004-code-organization.md
├── specifications/
│   ├── spec-001-data-structures.md
│   ├── spec-002-service-methods.md
│   ├── spec-003-detection-logic.md
│   └── spec-004-integration.md
└── STATUS.md (this file)
```

### Related Documents

- Requirements: `docs/sessions/202601072141-hourly-rate-service-fee/requirements/requirements.md`
- Research: `docs/sessions/202601072141-hourly-rate-service-fee/research/STATUS.md`
- Codebase patterns: `/Users/quang/workspace/dwarvesf/fortress-api/CLAUDE.md`

---

## Approval Signatures

| Role | Name | Date | Approved |
|------|------|------|----------|
| Product Owner | | | [ ] |
| Tech Lead | | | [ ] |
| Senior Developer | | | [ ] |

---

## Change Log

| Date | Author | Changes |
|------|--------|---------|
| 2026-01-07 | Claude (AI Planner) | Initial planning phase completed |

---

**Status**: Planning phase complete, ready for test case design phase.

**Confidence**: High - All major design decisions made with clear rationale and risk mitigation.

**Next Phase**: Test Case Design (to be done by test-case-designer agent)
