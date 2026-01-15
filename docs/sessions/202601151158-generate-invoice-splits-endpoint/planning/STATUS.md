# Planning Status: Generate Invoice Splits Endpoint

## Session Information
- **Session ID**: 202601151158-generate-invoice-splits-endpoint
- **Created**: 2026-01-15
- **Status**: Planning Complete
- **Phase**: Ready for Implementation

## Overview

Planning phase completed for creating a REST API endpoint to generate/regenerate invoice splits by invoice Legacy Number. The endpoint will query Notion Client Invoices database and enqueue a worker job for async processing.

## Requirements Summary

Create API endpoint that:
- Accepts invoice Legacy Number as input
- Queries Notion Client Invoices database
- Enqueues worker job to generate invoice splits
- Returns immediate success response (async processing)
- Follows existing codebase patterns

## Planning Deliverables

### Architecture Decision Records (ADRs)

1. **ADR-001: API Endpoint Design** ✓ Complete
   - Endpoint specification
   - Technical approach
   - File structure
   - Consequences and alternatives
   - Implementation notes

### Task Specifications

All 7 implementation tasks have been documented with:
- Clear objectives and scope
- Detailed implementation instructions
- Code examples and patterns
- Test requirements
- Acceptance criteria
- Verification commands

| Task ID | Description | Priority | Estimated Effort | Status |
|---------|-------------|----------|------------------|--------|
| TASK-001 | Add Request Model | P0 | 15 min | Specified |
| TASK-002 | Create Controller Method | P0 | 30 min | Specified |
| TASK-003 | Add Handler Interface | P1 | 5 min | Specified |
| TASK-004 | Implement Handler Method | P1 | 30 min | Specified |
| TASK-005 | Register Route | P2 | 10 min | Specified |
| TASK-006 | Update Swagger Documentation | P3 | 5 min | Specified |
| TASK-007 | Integration Testing | P3 | 20 min | Specified |

**Total Estimated Effort**: ~115 minutes (~2 hours)

## Technical Design Summary

### Endpoint Specification
```
POST /api/v1/invoices/generate-splits
```

**Request:**
```json
{
  "legacy_number": "INV-2024-001"
}
```

**Response:**
```json
{
  "data": {
    "legacy_number": "INV-2024-001",
    "invoice_page_id": "2bf64b29-...",
    "job_enqueued": true,
    "message": "Invoice splits generation job enqueued successfully"
  }
}
```

### Architecture Flow

```
HTTP Request
    ↓
Handler (pkg/handler/invoice/invoice.go)
    ↓ Parse & Validate
Controller (pkg/controller/invoice/generate_splits.go)
    ↓ Query Notion
Notion Service (pkg/service/notion/)
    ↓ Enqueue Job
Worker (pkg/worker/worker.go)
    ↓ Async Processing
Invoice Splits Generated
```

### Files to Create/Modify

**New Files:**
- `pkg/controller/invoice/generate_splits.go` - Controller logic

**Modified Files:**
- `pkg/handler/invoice/request/request.go` - Add request model
- `pkg/handler/invoice/interface.go` - Add method to interface
- `pkg/handler/invoice/invoice.go` - Add handler implementation
- `pkg/routes/v1.go` - Register route

## Dependencies

### External Services
- Notion API (Client Invoices database: 2bf64b29-b84c-80e2-8cc7-000bfe534203)
- Worker infrastructure (existing)

### Internal Dependencies
- Existing Notion service methods
- Existing worker infrastructure
- Permission system (PermissionInvoiceEdit)

### Reusable Components
- `service.Notion.QueryClientInvoiceByNumber()` - Already exists
- `worker.GenerateInvoiceSplitsMsg` - Already defined
- `worker.GenerateInvoiceSplitsPayload` - Already defined
- `handleGenerateInvoiceSplits` worker - Already implemented

## Risk Assessment

### Low Risk
- ✓ Worker infrastructure already exists and is tested
- ✓ Notion query method already exists
- ✓ Similar pattern exists (mark-paid endpoint)
- ✓ Idempotency handled by worker (Splits Generated checkbox)

### Medium Risk
- ⚠ Async processing means no immediate error feedback to API caller
- ⚠ Notion API rate limits could affect performance
- **Mitigation**: Worker logs errors, rate limiting already handled in service layer

### No Blocking Issues
All dependencies exist and are functional.

## Testing Strategy

### Unit Tests
- Request model validation
- Controller logic (with mocks)
- Handler method (with mocks)

### Integration Tests
- End-to-end API call to worker enqueue
- Actual Notion query (test database)
- Worker job processing

### Test Coverage Target
- Minimum 80% code coverage
- All error paths tested
- Golden file comparison for handler responses

## Implementation Approach

### Phase 1: Foundation (P0 - ~45 min)
1. TASK-001: Add request model
2. TASK-002: Create controller method

### Phase 2: Integration (P1 - ~35 min)
3. TASK-003: Add handler interface
4. TASK-004: Implement handler method

### Phase 3: Deployment (P2-P3 - ~35 min)
5. TASK-005: Register route
6. TASK-006: Update Swagger docs
7. TASK-007: Integration testing

### Recommended Order
Execute tasks in order TASK-001 through TASK-007 due to dependencies.

## Success Criteria

Planning phase is complete when:
- [x] All ADRs documented
- [x] All task specifications written
- [x] Technical approach validated
- [x] Dependencies identified
- [x] Risk assessment completed
- [x] Testing strategy defined

Implementation will be successful when:
- [ ] All 7 tasks completed
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Swagger documentation updated
- [ ] Endpoint accessible and functional
- [ ] Code reviewed and merged

## Next Steps

1. **Handoff to Developer**
   - Review all task specifications
   - Set up development environment
   - Begin with TASK-001

2. **During Implementation**
   - Follow task order (TASK-001 through TASK-007)
   - Run verification commands after each task
   - Check acceptance criteria before moving to next task

3. **Quality Assurance**
   - Run full test suite
   - Execute integration test script
   - Manual testing in local environment
   - Code review before merge

4. **Deployment**
   - Deploy to staging for validation
   - Monitor worker logs
   - Validate with real Notion data
   - Production deployment with monitoring

## Documentation References

### Planning Documents
- ADR: `docs/sessions/202601151158-generate-invoice-splits-endpoint/planning/ADRs/001-api-endpoint-design.md`
- Task Specs: `docs/sessions/202601151158-generate-invoice-splits-endpoint/planning/specifications/TASK-*.md`

### Codebase References
- Similar endpoint: `pkg/handler/invoice/invoice.go:527` (MarkPaid)
- Worker: `pkg/worker/worker.go:104` (handleGenerateInvoiceSplits)
- Notion service: `pkg/service/notion/invoice.go` (QueryClientInvoiceByNumber)

### External References
- Notion Client Invoices DB: 2bf64b29-b84c-80e2-8cc7-000bfe534203
- API Documentation: Will be at `/swagger/index.html` after TASK-006

## Notes

- This endpoint follows established patterns in the codebase
- All required infrastructure already exists
- Implementation should be straightforward
- Estimated total implementation time: ~2 hours
- No breaking changes to existing code

---

**Planning Completed By**: Claude Code (Project Manager Agent)
**Date**: 2026-01-15
**Ready for Implementation**: Yes
