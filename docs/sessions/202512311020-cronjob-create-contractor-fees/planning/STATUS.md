# Planning Status: Contractor Fees Cronjob

**Session ID**: `202512311020-cronjob-create-contractor-fees`

**Date**: 2025-12-31

**Status**: PLANNING COMPLETE - READY FOR IMPLEMENTATION

---

## Summary

Planning is complete for the automated cronjob endpoint that creates Contractor Fee entries from approved Task Order Logs. All architectural decisions have been documented, technical specifications defined, and implementation tasks identified.

## Deliverables

### 1. Architecture Decision Record (ADR)

**File**: `docs/sessions/202512311020-cronjob-create-contractor-fees/planning/ADRs/001-automated-contractor-fee-creation.md`

**Status**: Complete

**Key Decisions**:
- Service-only pattern (extend existing Notion services)
- Query-based idempotency checks
- Continue-on-error handling for batch processing
- Create-then-update timing for status updates
- Application-level date range filtering
- Contractor extraction via rollup property

### 2. Technical Specification

**File**: `docs/sessions/202512311020-cronjob-create-contractor-fees/planning/specifications/contractor-fees-cronjob.md`

**Status**: Complete

**Includes**:
- API contract (request/response)
- Data flow diagram
- Service method signatures
- Handler implementation logic
- Error scenarios matrix
- Test scenarios
- Notion property reference
- Performance considerations

### 3. Planning Document

**File**: This file (`STATUS.md`)

**Status**: In progress

---

## Implementation Breakdown

### Service Layer Extensions

| File | Methods to Add | Status |
|------|----------------|--------|
| `pkg/service/notion/task_order_log.go` | QueryApprovedOrders, UpdateOrderStatus | Not started |
| `pkg/service/notion/contractor_rates.go` | FindActiveRateByContractor | Not started |
| `pkg/service/notion/contractor_fees.go` | CheckFeeExistsByTaskOrder, CreateContractorFee | Not started |

**Data Structures**:
- `ApprovedOrderData` struct (in task_order_log.go)

### Handler Layer

| File | Action | Status |
|------|--------|--------|
| `pkg/handler/notion/contractor_fees.go` | Create new file with CreateContractorFees handler | Not started |
| `pkg/handler/notion/interface.go` | Add CreateContractorFees method to IHandler | Not started |

### Routes

| File | Action | Status |
|------|--------|--------|
| `pkg/routes/v1.go` | Add POST /cronjobs/contractor-fees route | Not started |

### Testing

| Test File | Scope | Status |
|-----------|-------|--------|
| `pkg/service/notion/task_order_log_test.go` | Service method tests | Not started |
| `pkg/service/notion/contractor_rates_test.go` | Service method tests | Not started |
| `pkg/service/notion/contractor_fees_test.go` | Service method tests | Not started |
| `pkg/handler/notion/contractor_fees_test.go` | Handler integration tests | Not started |

---

## Task Dependencies

```
Phase 1: Service Implementation (Parallel)
├── Task 1.1: Extend TaskOrderLogService
├── Task 1.2: Extend ContractorRatesService
└── Task 1.3: Extend ContractorFeesService
    ↓
Phase 2: Handler Implementation (Sequential)
├── Task 2.1: Create contractor_fees.go handler
└── Task 2.2: Update interface.go
    ↓
Phase 3: Route Registration (Sequential)
└── Task 3.1: Add route to v1.go
    ↓
Phase 4: Testing (Parallel with implementation)
├── Task 4.1: Service unit tests
├── Task 4.2: Handler integration tests
└── Task 4.3: Manual testing with Notion
```

---

## Next Steps

### For Implementation Team

1. **Review Planning Documents**
   - Read ADR to understand architectural decisions
   - Review technical specification for implementation details
   - Clarify any questions before starting

2. **Verify Notion Configuration**
   - Confirm database IDs in config match requirements
   - Test Notion API access with configured secret
   - Verify property names and IDs in actual Notion databases

3. **Implement Service Methods** (Tasks 1.1-1.3)
   - Start with TaskOrderLogService (QueryApprovedOrders, UpdateOrderStatus)
   - Continue with ContractorRatesService (FindActiveRateByContractor)
   - Finish with ContractorFeesService (CheckFeeExistsByTaskOrder, CreateContractorFee)
   - Follow existing service patterns in the codebase

4. **Implement Handler** (Tasks 2.1-2.2)
   - Create new handler file following SyncTaskOrderLogs pattern
   - Add method to interface
   - Implement orchestration logic with proper logging

5. **Register Route** (Task 3.1)
   - Add route to cronjob group in v1.go
   - Apply authentication and permission middleware

6. **Write Tests** (Tasks 4.1-4.3)
   - Unit tests for each service method
   - Integration tests for handler
   - Manual testing with real Notion data

7. **Deploy and Monitor**
   - Deploy to staging environment
   - Run cronjob manually to verify
   - Monitor logs for errors
   - Deploy to production
   - Schedule cronjob execution

### For Stakeholders

1. **Review Planning Documents**
   - Validate requirements are met
   - Approve architectural decisions
   - Identify any missing requirements

2. **Prepare Test Data**
   - Create test Task Order Logs in Notion with Status="Approved"
   - Ensure corresponding Contractor Rates exist
   - Prepare expected outcomes for validation

3. **Define Cronjob Schedule**
   - Determine execution frequency (daily, weekly, etc.)
   - Define execution time (e.g., end of business day)
   - Set up monitoring/alerting

---

## Risk Assessment

### High Priority Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Notion property names don't match schemas | High | Medium | Verify against actual databases before implementation |
| No matching contractor rate found | Medium | Medium | Skip with error, continue processing others |
| Contractor rollup returns empty array | Medium | Low | Skip with warning, log for investigation |

### Medium Priority Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Date range edge cases (nil end date) | Medium | Medium | Handle explicitly in code, add tests |
| Notion API rate limiting | Medium | Low | Use pagination, sequential processing |
| Status update fails after fee creation | Low | Low | Log error, don't fail (fee already created) |

### Mitigation Strategies

1. **Extensive Logging**: DEBUG-level logs for all property extractions
2. **Idempotency**: Check before creating to allow safe retries
3. **Continue-on-Error**: Process all orders, don't fail fast
4. **Manual Verification**: Test with real data before production deployment

---

## Configuration Requirements

### Notion Database IDs

These should already exist in `pkg/config/config.go`:

```go
cfg.Notion.Databases.TaskOrderLog     // "2b964b29-b84c-801e-ab9e-000b0662b987"
cfg.Notion.Databases.ContractorRates  // "2c464b29-b84c-80cf-bef6-000b42bce15e"
cfg.Notion.Databases.ContractorFees   // "2c264b29-b84c-8037-807c-000bf6d0792c"
```

**Action**: Verify these are set correctly in environment config.

### Notion API Secret

```go
cfg.Notion.Secret
```

**Action**: Ensure valid API token with access to all three databases.

---

## Testing Checklist

### Pre-Implementation Testing

- [ ] Verify Notion database IDs are correct
- [ ] Confirm API secret has necessary permissions
- [ ] Test Notion API queries manually (optional)

### Unit Testing

- [ ] TaskOrderLogService.QueryApprovedOrders
- [ ] TaskOrderLogService.UpdateOrderStatus
- [ ] ContractorRatesService.FindActiveRateByContractor
- [ ] ContractorFeesService.CheckFeeExistsByTaskOrder
- [ ] ContractorFeesService.CreateContractorFee

### Integration Testing

- [ ] Handler with mocked services (happy path)
- [ ] Handler with mocked services (error scenarios)
- [ ] Handler with mocked services (idempotency)

### Manual Testing

- [ ] Single approved order → fee created
- [ ] Multiple approved orders → all processed
- [ ] Existing fee → skipped (idempotent)
- [ ] Missing contractor → skipped with warning
- [ ] No matching rate → skipped with error
- [ ] Status update after fee creation

### Production Testing

- [ ] Deploy to staging
- [ ] Run cronjob manually
- [ ] Verify fees created in Notion
- [ ] Check logs for errors
- [ ] Deploy to production
- [ ] Monitor first production run

---

## Success Criteria

### Functional Requirements Met

- [x] FR1: Query approved orders (Type=Order, Status=Approved)
- [x] FR2: Find matching contractor rate (by contractor, status, date)
- [x] FR3: Create contractor fee (with relations)
- [x] FR4: Update order status to "Completed"

### Non-Functional Requirements Met

- [x] NFR1: Idempotency (check before creating)
- [x] NFR2: Logging (DEBUG/INFO/WARNING/ERROR levels)
- [x] NFR3: Error handling (continue-on-error)

### Implementation Ready

- [x] Architecture decisions documented
- [x] Technical specification complete
- [x] Service methods defined
- [x] Handler logic defined
- [x] Test scenarios identified
- [x] Error scenarios covered
- [x] Configuration requirements clear

---

## Handoff Notes

### For Developers

**Key Files to Reference**:
- ADR: Architecture decisions and rationale
- Spec: Implementation details and code examples
- Existing pattern: `pkg/handler/notion/task_order_log.go` (SyncTaskOrderLogs)

**Development Tips**:
1. Follow existing service patterns (see task_order_log.go, contractor_rates.go)
2. Use structured logging with logger.Fields
3. Handle pagination for Notion queries (PageSize: 100)
4. Extract rollup properties carefully (array handling)
5. Test date range logic thoroughly (nil end dates)

**Common Pitfalls to Avoid**:
- Don't assume rollup arrays have values (check length)
- Don't fail entire batch on individual errors
- Don't update status before creating fee
- Don't forget to handle nil End Date in rate matching

### For Test-Case Designers

**Test Scenarios Defined**:
- Happy path: All orders processed successfully
- Idempotency: Existing fees skipped
- Missing data: Orders skipped with warnings
- No matching rate: Orders skipped with errors
- Batch processing: Mixed success/failure results

**Test Data Requirements**:
- Approved orders with valid contractor rollup
- Matching active contractor rates
- Existing contractor fees (for idempotency tests)
- Orders with missing contractors
- Orders with no matching rates

### For QA Team

**Manual Test Plan**:
1. Create test approved orders in Notion
2. Run cronjob endpoint via Postman/curl
3. Verify contractor fees created in Notion
4. Check order status updated to "Completed"
5. Re-run cronjob, verify no duplicates
6. Check logs for expected messages

**Validation Points**:
- Correct number of fees created
- All relations properly set
- Payment status = "New"
- Order status = "Completed"
- No duplicate fees on re-run

---

## Implementation Estimate

**Total Tasks**: 8

**Breakdown**:
- Service methods: 3 files, 5 methods
- Handler: 1 new file, 1 interface update
- Routes: 1 route addition
- Testing: 4 test files

**Complexity**: Medium

**Dependencies**: None (all using existing infrastructure)

---

## References

### Planning Documents

- **ADR**: `docs/sessions/202512311020-cronjob-create-contractor-fees/planning/ADRs/001-automated-contractor-fee-creation.md`
- **Specification**: `docs/sessions/202512311020-cronjob-create-contractor-fees/planning/specifications/contractor-fees-cronjob.md`
- **Requirements**: `docs/sessions/202512311020-cronjob-create-contractor-fees/requirements/README.md`

### Notion Database Schemas

- **Task Order Log**: `docs/specs/notion/schema/task-order-log.md`
- **Contractor Fees**: `docs/specs/notion/contractor-fees.md`

### Codebase References

- **Existing Cronjob Pattern**: `pkg/handler/notion/task_order_log.go` (SyncTaskOrderLogs)
- **Service Examples**: `pkg/service/notion/task_order_log.go`, `pkg/service/notion/contractor_rates.go`, `pkg/service/notion/contractor_fees.go`
- **Routes Pattern**: `pkg/routes/v1.go` (cronjob group)

---

## Approval Status

- [ ] Planning reviewed by stakeholders
- [ ] Architecture decisions approved
- [ ] Technical specification approved
- [ ] Ready for implementation

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2025-12-31 | Planning Agent | Initial planning complete |

---

**Next Action**: Implementation team to begin Phase 1 (Service Extensions) after planning approval.
