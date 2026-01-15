# Implementation Guide: Generate Invoice Splits Endpoint

## Quick Reference

**Endpoint**: `POST /api/v1/invoices/generate-splits`

**Total Tasks**: 7 (TASK-001 through TASK-007)

**Estimated Time**: ~2 hours

**Status**: Ready for Implementation

## Executive Summary

This guide provides a complete breakdown of tasks needed to implement a REST API endpoint that generates invoice splits by querying Notion Client Invoices using Legacy Number and enqueuing a worker job for async processing.

### What This Endpoint Does

1. Accepts invoice Legacy Number in JSON request body
2. Queries Notion Client Invoices database to find the invoice
3. Enqueues a worker job to generate invoice splits (async)
4. Returns immediate success response with job status

### Why This Approach

- Follows existing codebase patterns (similar to mark-paid endpoint)
- Leverages existing worker infrastructure (handleGenerateInvoiceSplits)
- Reuses Notion service methods (QueryClientInvoiceByNumber)
- Provides idempotency through worker's Splits Generated checkbox

## Prerequisites

Before starting implementation:

- [ ] Go development environment set up
- [ ] Fortress API codebase cloned
- [ ] Dependencies installed: `go mod download`
- [ ] Development environment working: `make dev`
- [ ] Notion API access configured
- [ ] Understand existing invoice handler patterns
- [ ] Review worker implementation at `pkg/worker/worker.go:104`

## Task Execution Order

Execute tasks sequentially as each builds on the previous:

```
TASK-001 → TASK-002 → TASK-003 → TASK-004 → TASK-005 → TASK-006 → TASK-007
```

## Task Overview

### Phase 1: Foundation (P0 Priority - 45 minutes)

#### TASK-001: Add Request Model (15 min)
**File**: `pkg/handler/invoice/request/request.go`

Add request struct and validation:
```go
type GenerateSplitsRequest struct {
    LegacyNumber string `json:"legacy_number" binding:"required"`
}

func (r *GenerateSplitsRequest) Validate() error {
    // Validation logic
}
```

**Key Points**:
- Add after MarkPaidRequest struct
- Include Swagger @name annotation
- Validate non-empty legacy number
- Write unit tests

**Verification**: `go test ./pkg/handler/invoice/request/... -v`

---

#### TASK-002: Create Controller Method (30 min)
**File**: `pkg/controller/invoice/generate_splits.go` (new file)

Implement controller logic:
```go
func (c *controller) GenerateInvoiceSplitsByLegacyNumber(input GenerateSplitsInput) (*GenerateSplitsResult, error) {
    // 1. Query Notion by Legacy Number
    // 2. Validate invoice exists
    // 3. Enqueue worker job
    // 4. Return result
}
```

**Key Points**:
- Query Notion using QueryClientInvoiceByNumber
- Enqueue GenerateInvoiceSplitsMsg worker job
- Return structured result
- Write unit tests with mocks

**Verification**: `go test ./pkg/controller/invoice/... -run TestController_GenerateInvoiceSplitsByLegacyNumber -v`

---

### Phase 2: Integration (P1 Priority - 35 minutes)

#### TASK-003: Add Handler Interface (5 min)
**File**: `pkg/handler/invoice/interface.go`

Add method to interface:
```go
type IHandler interface {
    // ... existing methods ...
    GenerateSplits(c *gin.Context)
}
```

**Key Points**:
- Add after MarkPaid method
- Compiler will enforce implementation

**Verification**: `go build ./pkg/handler/invoice/...`

---

#### TASK-004: Implement Handler Method (30 min)
**File**: `pkg/handler/invoice/invoice.go`

Add response struct and handler:
```go
type GenerateSplitsResponse struct {
    LegacyNumber  string `json:"legacy_number"`
    InvoicePageID string `json:"invoice_page_id"`
    JobEnqueued   bool   `json:"job_enqueued"`
    Message       string `json:"message"`
}

func (h *handler) GenerateSplits(c *gin.Context) {
    // 1. Parse request
    // 2. Validate request
    // 3. Call controller
    // 4. Handle errors (404, 500)
    // 5. Return response
}
```

**Key Points**:
- Add complete Swagger annotations
- Handle 400, 404, 500 errors appropriately
- Follow view.CreateResponse pattern
- Write comprehensive unit tests

**Verification**: `go test ./pkg/handler/invoice/... -run TestHandler_GenerateSplits -v`

---

### Phase 3: Deployment (P2-P3 Priority - 35 minutes)

#### TASK-005: Register Route (10 min)
**File**: `pkg/routes/v1.go`

Register route in invoice group:
```go
invoiceGroup.POST("/generate-splits",
    conditionalAuthMW,
    conditionalPermMW(model.PermissionInvoiceEdit),
    h.Invoice.GenerateSplits)
```

**Key Points**:
- Add after mark-paid route
- Use PermissionInvoiceEdit permission
- Include auth and permission middleware

**Verification**:
```bash
make dev
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-TEST-001"}'
```

---

#### TASK-006: Update Swagger Documentation (5 min)
**Command**: `make gen-swagger`

Generate API documentation:
- Swagger will automatically parse handler annotations
- Updates docs/swagger.json and docs/swagger.yaml
- Visible at /swagger/index.html

**Key Points**:
- Run after all code is implemented
- Verify endpoint appears in Swagger UI
- Test "Try it out" functionality

**Verification**: Navigate to `http://localhost:8080/swagger/index.html`

---

#### TASK-007: Integration Testing (20 min)
**File**: `scripts/test/integration_generate_splits.sh` (new file)

End-to-end testing:
1. Test valid invoice
2. Test validation errors
3. Test not found errors
4. Test worker processing
5. Test idempotency

**Key Points**:
- Create integration test script
- Test all happy and error paths
- Verify worker actually processes jobs
- Check Notion data updates

**Verification**: `./scripts/test/integration_generate_splits.sh`

---

## Common Patterns to Follow

### 1. Error Handling
```go
if err != nil {
    l.Error(err, "descriptive message")
    if strings.Contains(err.Error(), "not found") {
        c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
        return
    }
    c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
    return
}
```

### 2. Logging
```go
l := h.logger.Fields(logger.Fields{
    "handler": "invoice",
    "method":  "GenerateSplits",
})
l.Debug("descriptive message")
l.Debugf("message with %s", variable)
l.Infof("success message: key=%s", value)
```

### 3. Response Pattern
```go
c.JSON(http.StatusOK, view.CreateResponse(responseData, nil, nil, request, ""))
```

### 4. Swagger Annotations
```go
// MethodName godoc
// @Summary Brief description
// @Description Detailed description
// @id uniqueOperationId
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body RequestType true "description"
// @Success 200 {object} ResponseType
// @Failure 400 {object} ErrorResponse
// @Router /path [method]
```

## Testing Strategy

### Unit Tests (Required for Each Task)

**Request Model**:
```bash
go test ./pkg/handler/invoice/request/... -v
```

**Controller**:
```bash
go test ./pkg/controller/invoice/... -run TestController_GenerateInvoiceSplitsByLegacyNumber -v
```

**Handler**:
```bash
go test ./pkg/handler/invoice/... -run TestHandler_GenerateSplits -v
```

### Integration Tests

**Manual Testing**:
```bash
# Start app
make dev

# Test success case
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}' | jq .

# Test validation error
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": ""}' | jq .

# Test not found
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-NONEXISTENT"}' | jq .
```

**Automated Testing**:
```bash
./scripts/test/integration_generate_splits.sh
```

## Verification Checklist

After completing all tasks, verify:

### Code Quality
- [ ] All files compile without errors: `go build ./...`
- [ ] All unit tests pass: `go test ./... -v`
- [ ] Code follows existing patterns
- [ ] No linter warnings: `make lint`

### Functionality
- [ ] Endpoint accessible at `/api/v1/invoices/generate-splits`
- [ ] Valid request returns 200 with correct response structure
- [ ] Empty legacy number returns 400
- [ ] Non-existent invoice returns 404
- [ ] Worker job is enqueued successfully
- [ ] Worker processes job and generates splits
- [ ] Idempotency works (no duplicate splits)

### Documentation
- [ ] Swagger docs generated: `make gen-swagger`
- [ ] Endpoint visible in Swagger UI
- [ ] All parameters documented
- [ ] Response schemas documented
- [ ] "Try it out" works in Swagger UI

### Logging
- [ ] Handler logs show request processing
- [ ] Controller logs show Notion query and job enqueue
- [ ] Worker logs show job processing
- [ ] Error cases are logged appropriately

## Troubleshooting

### Issue: Compilation Errors

**Symptom**: `go build` fails

**Solution**:
1. Check import statements
2. Verify all types are defined
3. Ensure interface is properly implemented
4. Run `go mod tidy`

### Issue: Tests Failing

**Symptom**: Unit tests fail

**Solution**:
1. Check mock expectations match actual calls
2. Verify test data matches expected format
3. Review error handling in test cases
4. Check table-driven test structure

### Issue: Endpoint Not Found

**Symptom**: 404 when calling endpoint

**Solution**:
1. Verify route registration in `pkg/routes/v1.go`
2. Check route path matches request URL
3. Ensure handler method is implemented
4. Restart application: `make dev`

### Issue: Worker Not Processing

**Symptom**: Job enqueued but not processed

**Solution**:
1. Check worker is running
2. Verify worker message type matches
3. Check worker logs for errors
4. Ensure Notion service is configured

### Issue: Swagger Not Updated

**Symptom**: Endpoint not in Swagger UI

**Solution**:
1. Ensure all Swagger annotations present
2. Run `make gen-swagger`
3. Check for generation errors
4. Restart application
5. Clear browser cache

## File Checklist

### New Files Created
- [ ] `pkg/controller/invoice/generate_splits.go`
- [ ] `pkg/controller/invoice/generate_splits_test.go`
- [ ] `scripts/test/integration_generate_splits.sh`

### Modified Files
- [ ] `pkg/handler/invoice/request/request.go`
- [ ] `pkg/handler/invoice/request/request_test.go` (or add tests)
- [ ] `pkg/handler/invoice/interface.go`
- [ ] `pkg/handler/invoice/invoice.go`
- [ ] `pkg/handler/invoice/invoice_test.go` (or add tests)
- [ ] `pkg/routes/v1.go`
- [ ] `docs/swagger.json` (auto-generated)
- [ ] `docs/swagger.yaml` (auto-generated)

### Files to Review (Reference)
- `pkg/handler/invoice/invoice.go` (MarkPaid method - line 527)
- `pkg/controller/invoice/mark_paid.go` (similar pattern)
- `pkg/worker/worker.go` (handleGenerateInvoiceSplits - line 104)
- `pkg/service/notion/invoice.go` (QueryClientInvoiceByNumber)

## Success Metrics

Implementation is complete when:

1. **All Tasks Completed**: TASK-001 through TASK-007 done
2. **All Tests Pass**: Unit and integration tests passing
3. **Endpoint Functional**: Returns correct responses for all cases
4. **Worker Processes Jobs**: Invoice splits are actually generated
5. **Documentation Updated**: Swagger docs include new endpoint
6. **Code Quality**: Passes linting and follows conventions
7. **Logs Clear**: All operations logged appropriately

## Timeline

Assuming focused work without interruptions:

- **Phase 1 (Foundation)**: 45 minutes
- **Phase 2 (Integration)**: 35 minutes
- **Phase 3 (Deployment)**: 35 minutes
- **Total**: ~2 hours

Add buffer for:
- Code review iterations
- Bug fixes
- Documentation review
- QA validation

**Realistic Total**: 3-4 hours including overhead

## Next Steps After Implementation

1. **Code Review**
   - Create PR with all changes
   - Request review from @huynguyenh @lmquang
   - Address feedback

2. **QA Testing**
   - Deploy to staging environment
   - Run full test suite
   - Manual testing with real data
   - Performance validation

3. **Documentation**
   - Update API documentation
   - Add usage examples
   - Document any limitations

4. **Deployment**
   - Merge to main branch
   - Deploy to production
   - Monitor logs and metrics
   - Verify functionality in production

## Support Resources

### Documentation
- ADR: `planning/ADRs/001-api-endpoint-design.md`
- Task Specs: `planning/specifications/TASK-*.md`
- Status: `planning/STATUS.md`

### Reference Code
- Similar endpoint: `pkg/handler/invoice/invoice.go:527`
- Worker: `pkg/worker/worker.go:104`
- Notion service: `pkg/service/notion/invoice.go`

### External Resources
- Notion API: Client Invoices DB `2bf64b29-b84c-80e2-8cc7-000bfe534203`
- Swagger UI: `http://localhost:8080/swagger/index.html`

---

**Implementation Guide Version**: 1.0
**Last Updated**: 2026-01-15
**Prepared By**: Claude Code (Project Manager Agent)
