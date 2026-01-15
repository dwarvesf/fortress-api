# Planning Documentation: Generate Invoice Splits Endpoint

## Overview

This directory contains comprehensive planning documentation for implementing a REST API endpoint to generate/regenerate invoice splits by invoice Legacy Number.

## Session Information

- **Session ID**: 202601151158-generate-invoice-splits-endpoint
- **Created**: 2026-01-15
- **Phase**: Planning Complete - Ready for Implementation
- **Estimated Implementation Time**: 2-3 hours

## Requirement Summary

Create an API endpoint that:
- Accepts invoice Legacy Number as input
- Queries Notion Client Invoices database
- Enqueues worker job for async invoice splits generation
- Returns immediate success response

**Technical Approach**: Follow existing patterns (similar to mark-paid endpoint), leverage existing worker infrastructure, reuse Notion service methods.

## Documentation Structure

This planning phase produced the following documentation:

### 📋 Quick Start

**Start here for rapid implementation:**
- **[TASK_SUMMARY.md](./TASK_SUMMARY.md)** - Quick reference guide with all tasks, commands, and validation checklist

### 📖 Comprehensive Guide

**Read this for detailed implementation instructions:**
- **[IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)** - Complete step-by-step guide with code patterns, troubleshooting, and testing strategy

### 🏗️ Architecture & Decisions

**Understand the technical design:**
- **[ADRs/001-api-endpoint-design.md](./ADRs/001-api-endpoint-design.md)** - Architecture Decision Record documenting endpoint design, technical approach, alternatives considered, and consequences

### 📝 Task Specifications

**Detailed specifications for each implementation task:**

| Task | File | Description | Priority | Time |
|------|------|-------------|----------|------|
| TASK-001 | [specifications/TASK-001-add-request-model.md](./specifications/TASK-001-add-request-model.md) | Add request model with validation | P0 | 15 min |
| TASK-002 | [specifications/TASK-002-create-controller-method.md](./specifications/TASK-002-create-controller-method.md) | Create controller method | P0 | 30 min |
| TASK-003 | [specifications/TASK-003-add-handler-interface.md](./specifications/TASK-003-add-handler-interface.md) | Add handler interface method | P1 | 5 min |
| TASK-004 | [specifications/TASK-004-implement-handler-method.md](./specifications/TASK-004-implement-handler-method.md) | Implement handler method | P1 | 30 min |
| TASK-005 | [specifications/TASK-005-register-route.md](./specifications/TASK-005-register-route.md) | Register route with middleware | P2 | 10 min |
| TASK-006 | [specifications/TASK-006-update-swagger-docs.md](./specifications/TASK-006-update-swagger-docs.md) | Update Swagger documentation | P3 | 5 min |
| TASK-007 | [specifications/TASK-007-integration-testing.md](./specifications/TASK-007-integration-testing.md) | Integration testing | P3 | 20 min |

### 📊 Status & Tracking

**Monitor planning and implementation progress:**
- **[STATUS.md](./STATUS.md)** - Current status, deliverables checklist, risk assessment, and next steps

## How to Use This Documentation

### For Project Managers
1. Review [STATUS.md](./STATUS.md) for planning completion status
2. Review [ADRs/001-api-endpoint-design.md](./ADRs/001-api-endpoint-design.md) for technical decisions
3. Use task estimates for sprint planning

### For Developers
1. **Quick Implementation**: Start with [TASK_SUMMARY.md](./TASK_SUMMARY.md)
2. **Detailed Implementation**: Follow [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
3. **Specific Tasks**: Refer to individual task specifications in `specifications/`
4. **Technical Context**: Read ADR for understanding design decisions

### For QA/Testers
1. Review [TASK-007-integration-testing.md](./specifications/TASK-007-integration-testing.md) for test cases
2. Use test scripts provided in implementation guide
3. Validate against acceptance criteria in each task specification

### For Reviewers
1. Check [ADRs/001-api-endpoint-design.md](./ADRs/001-api-endpoint-design.md) for technical rationale
2. Verify implementation follows patterns documented in task specifications
3. Ensure all acceptance criteria are met

## Implementation Approach

### Recommended Execution Order

Execute tasks sequentially due to dependencies:

```
Phase 1: Foundation (P0)
  ├─ TASK-001: Add Request Model (15 min)
  └─ TASK-002: Create Controller Method (30 min)

Phase 2: Integration (P1)
  ├─ TASK-003: Add Handler Interface (5 min)
  └─ TASK-004: Implement Handler Method (30 min)

Phase 3: Deployment (P2-P3)
  ├─ TASK-005: Register Route (10 min)
  ├─ TASK-006: Update Swagger Docs (5 min)
  └─ TASK-007: Integration Testing (20 min)
```

### Quick Start Commands

```bash
# 1. Start development environment
make dev

# 2. Run tests after each implementation
go test ./pkg/handler/invoice/request/... -v    # After TASK-001
go test ./pkg/controller/invoice/... -v         # After TASK-002
go test ./pkg/handler/invoice/... -v            # After TASK-004

# 3. Generate Swagger documentation
make gen-swagger                                # After TASK-006

# 4. Run integration tests
./scripts/test/integration_generate_splits.sh   # After TASK-007
```

## Technical Summary

### Endpoint Specification

```
POST /api/v1/invoices/generate-splits
Content-Type: application/json

Request:
{
  "legacy_number": "INV-2024-001"
}

Response (200 OK):
{
  "data": {
    "legacy_number": "INV-2024-001",
    "invoice_page_id": "abc-123-def-456",
    "job_enqueued": true,
    "message": "Invoice splits generation job enqueued successfully"
  },
  "error": null
}
```

### Architecture Flow

```
HTTP Request
    ↓
Handler (parse & validate)
    ↓
Controller (query Notion & enqueue)
    ↓
Notion Service (find invoice)
    ↓
Worker (async processing)
    ↓
Invoice Splits Generated
```

### Files to Create/Modify

**New Files (3):**
- `pkg/controller/invoice/generate_splits.go`
- `pkg/controller/invoice/generate_splits_test.go`
- `scripts/test/integration_generate_splits.sh`

**Modified Files (4):**
- `pkg/handler/invoice/request/request.go`
- `pkg/handler/invoice/interface.go`
- `pkg/handler/invoice/invoice.go`
- `pkg/routes/v1.go`

## Success Criteria

### Planning Phase (Current)
- [x] All ADRs documented
- [x] All task specifications written
- [x] Technical approach validated
- [x] Dependencies identified
- [x] Risk assessment completed
- [x] Testing strategy defined

### Implementation Phase (Next)
- [ ] All 7 tasks completed in order
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] Swagger documentation updated
- [ ] Endpoint functional and accessible
- [ ] Worker processes jobs correctly
- [ ] Code reviewed and approved
- [ ] Changes merged to main branch

## Key Decisions & Rationale

### Why This Approach?
1. **Consistency**: Follows existing patterns (mark-paid endpoint)
2. **Reusability**: Leverages existing infrastructure (worker, Notion service)
3. **Async Processing**: Non-blocking operation via worker pattern
4. **Idempotency**: Worker checks "Splits Generated" flag
5. **Maintainability**: Clear separation of concerns across layers

### Why These Technologies?
- **Gin**: Existing HTTP framework in codebase
- **Notion API**: Invoice data source
- **Worker Pattern**: Established async processing mechanism
- **GORM**: Database ORM already in use

### Alternative Approaches Considered
- **Synchronous Processing**: Rejected - would block HTTP requests
- **Direct Page ID Input**: Rejected - Legacy Number is more user-friendly
- **Webhook-Only**: Rejected - API provides more flexibility

See [ADR-001](./ADRs/001-api-endpoint-design.md) for complete analysis.

## Dependencies & Prerequisites

### External Services
- Notion API (Client Invoices DB: `2bf64b29-b84c-80e2-8cc7-000bfe534203`)
- Worker infrastructure (already exists)

### Existing Code Components
- `service.Notion.QueryClientInvoiceByNumber()` - Query method
- `worker.GenerateInvoiceSplitsMsg` - Message type
- `worker.GenerateInvoiceSplitsPayload` - Payload struct
- `handleGenerateInvoiceSplits` - Worker handler (pkg/worker/worker.go:104)

### Required Permissions
- `model.PermissionInvoiceEdit` - For endpoint access

## Risk Assessment

### Low Risk ✓
- Worker infrastructure exists and is tested
- Notion query method exists
- Similar pattern exists (mark-paid endpoint)
- Idempotency handled by worker

### Medium Risk ⚠
- Async processing (no immediate error feedback)
- Notion API rate limits
- **Mitigation**: Worker logs errors; rate limiting already handled

### No Blocking Issues ✓
All dependencies exist and are functional.

## Testing Strategy

### Unit Tests
- Request validation (TASK-001)
- Controller logic with mocks (TASK-002)
- Handler method with mocks (TASK-004)

### Integration Tests
- End-to-end API call (TASK-007)
- Worker job processing (TASK-007)
- Notion database integration (TASK-007)

### Test Coverage Target
- Minimum 80% code coverage
- All error paths tested
- Golden file comparison for handler responses

## Timeline

### Estimated Effort
- **Phase 1 (Foundation)**: 45 minutes
- **Phase 2 (Integration)**: 35 minutes
- **Phase 3 (Deployment)**: 35 minutes
- **Total Core Development**: ~2 hours

### Realistic Timeline
Including overhead for code review, bug fixes, QA:
- **Development**: 2-3 hours
- **Code Review**: 30-60 minutes
- **QA/Testing**: 30-60 minutes
- **Total**: 3-5 hours

## Next Steps

### Immediate (Development Phase)
1. Review all task specifications
2. Set up development environment
3. Begin implementation with TASK-001
4. Execute tasks in order (001 → 007)

### After Implementation
1. Create pull request with changes
2. Request code review
3. Address review feedback
4. Deploy to staging for QA
5. Merge to main branch
6. Deploy to production
7. Monitor logs and metrics

## Support & References

### Planning Documentation
- This directory contains all planning artifacts
- Each task has detailed specification
- ADR provides technical rationale

### Codebase References
- Similar endpoint: `pkg/handler/invoice/invoice.go:527` (MarkPaid)
- Worker: `pkg/worker/worker.go:104` (handleGenerateInvoiceSplits)
- Notion service: `pkg/service/notion/invoice.go`

### External Resources
- Notion API Documentation
- Gin Framework Documentation
- Go Testing Best Practices

## Questions or Issues?

If you encounter issues during implementation:

1. **Check Task Specification**: Each task has detailed instructions and troubleshooting
2. **Review Implementation Guide**: Comprehensive guide with common patterns
3. **Reference Similar Code**: Check mark-paid endpoint for similar patterns
4. **Check ADR**: Technical decisions and alternatives are documented

## Document Maintenance

### Version History
- **v1.0** (2026-01-15): Initial planning documentation completed

### Future Updates
- Update STATUS.md as implementation progresses
- Document any deviations from plan
- Add lessons learned after implementation

---

**Planning Completed By**: Claude Code (Project Manager Agent)

**Date**: 2026-01-15

**Status**: ✓ Ready for Implementation

**Estimated Implementation Time**: 2-3 hours

**Next Phase**: Development (TASK-001 through TASK-007)
