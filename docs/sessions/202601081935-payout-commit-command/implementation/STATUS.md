# Implementation Status: Payout Commit Command

## Overview

**Feature**: `?payout commit` Discord command for committing contractor payables
**Session**: 202601081935-payout-commit-command
**Status**: IN PROGRESS
**Last Updated**: 2026-01-08

---

## Phase Progress

| Phase | Status | Tasks Complete | Total Tasks | Progress |
|-------|--------|----------------|-------------|----------|
| Phase 1: fortress-api Notion Services | Complete | 7 | 7 | 100% |
| Phase 2: fortress-api Handler/Controller | Complete | 7 | 7 | 100% |
| Phase 3: fortress-discord Command | Not Started | 0 | 7 | 0% |
| Phase 4: Integration & Testing | Not Started | 0 | 3 | 0% |
| **TOTAL** | **In Progress** | **14** | **24** | **58%** |

---

## Detailed Task Status

### Phase 1: fortress-api Notion Services (7/7) ✅

- [x] **Task 1.1**: QueryPendingPayablesByPeriod ✅
- [x] **Task 1.2**: UpdatePayableStatus ✅
- [x] **Task 1.3**: GetContractorPayDay ✅
- [x] **Task 1.4**: GetPayoutWithRelations ✅
- [x] **Task 1.5**: UpdatePayoutStatus ✅
- [x] **Task 1.6**: UpdateInvoiceSplitStatus (Select type!) ✅
- [x] **Task 1.7**: UpdateRefundRequestStatus ✅

### Phase 2: fortress-api Handler/Controller (7/7) ✅

- [x] **Task 2.1**: Controller interface ✅
- [x] **Task 2.2**: Controller implementation ✅
- [x] **Task 2.3**: Controller registration ✅
- [x] **Task 2.4**: Handler request/response types ✅
- [x] **Task 2.5**: Handler interface ✅
- [x] **Task 2.6**: Handler implementation ✅
- [x] **Task 2.7**: Routes registration ✅

### Phase 3: fortress-discord Command (0/7)

- [ ] **Task 3.1**: Discord model types
- [ ] **Task 3.2**: Adapter implementation
- [ ] **Task 3.3**: Service implementation
- [ ] **Task 3.4**: Discord view
- [ ] **Task 3.5**: Command implementation
- [ ] **Task 3.6**: Command registration
- [ ] **Task 3.7**: Button interaction handler

### Phase 4: Integration & Testing (0/3)

- [ ] **Task 4.1**: fortress-api unit tests
- [ ] **Task 4.2**: fortress-discord unit tests
- [ ] **Task 4.3**: Integration testing and documentation

---

## Key Milestones

- [x] **Milestone 1**: Notion services complete (Phase 1) ✅
- [x] **Milestone 2**: API endpoints functional (Phase 2) ✅
- [ ] **Milestone 3**: Discord command working (Phase 3)
- [ ] **Milestone 4**: All tests passing (Phase 4)
- [ ] **Milestone 5**: Production ready

---

## Known Blockers

None currently identified.

---

## Notes

### Critical Implementation Points
1. **Property Type Difference**: Invoice Split uses `Select` type, others use `Status` type
2. **Best-Effort Updates**: Continue processing on individual failures, track counts
3. **Idempotency**: Re-running commit must be safe
4. **Pagination**: Handle large result sets (>100 payables)

### Testing Priorities
1. Property type tests (prevent API rejections)
2. Pagination handling (prevent data loss)
3. Cascade update sequence (ensure correct order)
4. Error handling (best-effort approach)

### Environment Variables Required

**fortress-api**:
```
NOTION_DATABASE_CONTRACTOR_PAYABLES=2c264b29-b84c-8037-807c-000bf6d0792c
NOTION_DATABASE_CONTRACTOR_PAYOUTS=2c564b29-b84c-8045-80ee-000bee2e3669
NOTION_DATABASE_INVOICE_SPLIT=2c364b29-b84c-804f-9856-000b58702dea
NOTION_DATABASE_REFUND_REQUEST=2cc64b29-b84c-8066-adf2-cc56171cedf4
NOTION_DATABASE_SERVICE_RATE=2c464b29-b84c-80cf-bef6-000b42bce15e
```

**fortress-discord**:
```
FORTRESS_API_URL=https://api.fortress.example.com
FORTRESS_API_KEY=<api-key>
```

---

## Files Created/Modified in Phase 2

### Created
- `pkg/handler/contractorpayables/interface.go` - Handler interface
- `pkg/handler/contractorpayables/request.go` - Request types (PreviewCommitRequest, CommitRequest)
- `pkg/handler/contractorpayables/contractorpayables.go` - Handler implementation
- `pkg/controller/contractorpayables/interface.go` - Controller interface and response types
- `pkg/controller/contractorpayables/contractorpayables.go` - Controller implementation

### Modified
- `pkg/controller/controller.go` - Added ContractorPayables controller registration
- `pkg/handler/handler.go` - Added ContractorPayables handler registration
- `pkg/routes/v1.go` - Added /contractor-payables routes

### API Endpoints
- `GET /api/v1/contractor-payables/preview-commit?month=YYYY-MM&batch=1|15`
- `POST /api/v1/contractor-payables/commit` with body `{"month": "YYYY-MM", "batch": 1|15}`

---

## References

- [Task Breakdown](./tasks.md)
- [Requirements](../requirements/requirements.md)
- [Full Specification](../../specs/payout-commit-command.md)
- [ADR-001: Cascade Status Update](../planning/ADRs/ADR-001-cascade-status-update.md)

---

## Change Log

| Date | Changes | Updated By |
|------|---------|------------|
| 2026-01-08 | Initial task breakdown created | Claude Code |
| 2026-01-08 | Completed Tasks 1.1-1.3 (QueryPendingPayablesByPeriod, UpdatePayableStatus, GetContractorPayDay) | Claude Code |
| 2026-01-08 | Completed Phase 1 (Tasks 1.4-1.7) - All Notion service methods implemented | Claude Code |
| 2026-01-08 | Completed Phase 2 - Handler/Controller layer complete, API endpoints registered | Claude Code |
