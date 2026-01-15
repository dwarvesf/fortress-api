# Task Summary: Generate Invoice Splits Endpoint

## Quick Overview

**Objective**: Create REST API endpoint to generate invoice splits by Legacy Number

**Endpoint**: `POST /api/v1/invoices/generate-splits`

**Total Tasks**: 7

**Total Time**: ~2 hours

**Status**: Ready for Implementation ✓

---

## Task Breakdown

### TASK-001: Add Request Model
- **File**: `pkg/handler/invoice/request/request.go`
- **Time**: 15 min
- **Priority**: P0
- **Action**: Add `GenerateSplitsRequest` struct with validation
- **Test**: `go test ./pkg/handler/invoice/request/... -v`

---

### TASK-002: Create Controller Method
- **File**: `pkg/controller/invoice/generate_splits.go` (NEW)
- **Time**: 30 min
- **Priority**: P0
- **Action**: Implement controller logic to query Notion and enqueue worker job
- **Test**: `go test ./pkg/controller/invoice/... -run TestController_GenerateInvoiceSplitsByLegacyNumber -v`

---

### TASK-003: Add Handler Interface
- **File**: `pkg/handler/invoice/interface.go`
- **Time**: 5 min
- **Priority**: P1
- **Action**: Add `GenerateSplits(c *gin.Context)` to interface
- **Test**: `go build ./pkg/handler/invoice/...`

---

### TASK-004: Implement Handler Method
- **File**: `pkg/handler/invoice/invoice.go`
- **Time**: 30 min
- **Priority**: P1
- **Action**: Implement HTTP handler with Swagger annotations
- **Test**: `go test ./pkg/handler/invoice/... -run TestHandler_GenerateSplits -v`

---

### TASK-005: Register Route
- **File**: `pkg/routes/v1.go`
- **Time**: 10 min
- **Priority**: P2
- **Action**: Register route with auth and permission middleware
- **Test**: `make dev` then `curl http://localhost:8080/api/v1/invoices/generate-splits`

---

### TASK-006: Update Swagger Documentation
- **Command**: `make gen-swagger`
- **Time**: 5 min
- **Priority**: P3
- **Action**: Generate API documentation
- **Test**: Check `http://localhost:8080/swagger/index.html`

---

### TASK-007: Integration Testing
- **File**: `scripts/test/integration_generate_splits.sh` (NEW)
- **Time**: 20 min
- **Priority**: P3
- **Action**: Create and run integration tests
- **Test**: `./scripts/test/integration_generate_splits.sh`

---

## Execution Order

```
001 → 002 → 003 → 004 → 005 → 006 → 007
```

**Must execute in order due to dependencies.**

---

## Quick Commands

### Development
```bash
# Start app
make dev

# Run tests
go test ./...

# Generate Swagger
make gen-swagger

# Run linter
make lint
```

### Testing
```bash
# Test request model
go test ./pkg/handler/invoice/request/... -v

# Test controller
go test ./pkg/controller/invoice/... -v

# Test handler
go test ./pkg/handler/invoice/... -v

# Integration test
./scripts/test/integration_generate_splits.sh
```

### Manual API Test
```bash
# Success case
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}' | jq .

# Validation error
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": ""}' | jq .

# Not found
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-NOTFOUND"}' | jq .
```

---

## Files to Create

1. `pkg/controller/invoice/generate_splits.go`
2. `pkg/controller/invoice/generate_splits_test.go`
3. `scripts/test/integration_generate_splits.sh`

---

## Files to Modify

1. `pkg/handler/invoice/request/request.go`
2. `pkg/handler/invoice/interface.go`
3. `pkg/handler/invoice/invoice.go`
4. `pkg/routes/v1.go`

---

## Expected API Behavior

### Request
```json
POST /api/v1/invoices/generate-splits
{
  "legacy_number": "INV-2024-001"
}
```

### Success Response (200)
```json
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

### Error Responses
- **400**: Invalid/empty legacy number
- **404**: Invoice not found
- **500**: Server error

---

## Validation Checklist

### Before Starting
- [ ] Development environment working
- [ ] Dependencies installed
- [ ] Notion API configured
- [ ] Worker infrastructure functional

### During Implementation
- [ ] Follow task order (001 → 007)
- [ ] Run verification commands after each task
- [ ] Check acceptance criteria
- [ ] Write unit tests

### After Completion
- [ ] All tests pass
- [ ] Endpoint returns correct responses
- [ ] Worker processes jobs
- [ ] Swagger docs updated
- [ ] Code follows conventions
- [ ] No linter warnings

---

## Key Patterns

### Error Handling
```go
if strings.Contains(err.Error(), "not found") {
    c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
    return
}
c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
```

### Logging
```go
l := h.logger.Fields(logger.Fields{
    "handler": "invoice",
    "method":  "GenerateSplits",
})
l.Debug("message")
l.Infof("message: %s", value)
```

### Worker Enqueue
```go
c.worker.Enqueue(worker.GenerateInvoiceSplitsMsg, worker.GenerateInvoiceSplitsPayload{
    InvoicePageID: notionPage.ID,
})
```

---

## Reference Files

- Similar handler: `pkg/handler/invoice/invoice.go:527` (MarkPaid)
- Similar controller: `pkg/controller/invoice/mark_paid.go`
- Worker implementation: `pkg/worker/worker.go:104`
- Notion service: `pkg/service/notion/invoice.go`

---

## Documentation

- **Full Guide**: `IMPLEMENTATION_GUIDE.md`
- **ADR**: `ADRs/001-api-endpoint-design.md`
- **Status**: `STATUS.md`
- **Task Details**: `specifications/TASK-*.md`

---

## Success Criteria

✓ All 7 tasks completed
✓ All unit tests passing
✓ Integration tests passing
✓ Endpoint functional
✓ Worker processes jobs
✓ Swagger docs updated
✓ Code reviewed and merged

---

**Quick Start**: Begin with TASK-001 and follow the execution order.

**Estimated Completion**: 2-3 hours with testing and review.

**Need Help?**: Refer to detailed task specifications in `specifications/` folder.
