# Implementation Tasks: Generate Invoice Splits Endpoint

## Overview

Implement REST API endpoint `POST /api/v1/invoices/generate-splits` to generate invoice splits by Legacy Number.

**Total Tasks**: 7
**Estimated Time**: ~2 hours
**Execution Order**: Sequential (001 → 007)

---

## Task 1: Add Request Model

**File(s)**: `pkg/handler/invoice/request/request.go`

**Description**:
Add `GenerateSplitsRequest` struct with validation for accepting invoice Legacy Number.

**Implementation**:
```go
type GenerateSplitsRequest struct {
    LegacyNumber string `json:"legacy_number" binding:"required"`
}
```

**Acceptance**:
- Struct added with proper JSON tags
- Validation tag `binding:"required"` present
- Compiles without errors: `go build ./pkg/handler/invoice/request/...`

---

## Task 2: Create Controller Method

**File(s)**:
- `pkg/controller/invoice/generate_splits.go` (NEW)
- `pkg/controller/invoice/generate_splits_test.go` (NEW - optional but recommended)

**Description**:
Implement controller logic to:
1. Query Notion Client Invoices database by Legacy Number
2. Enqueue worker job for invoice splits generation
3. Return response data

**Implementation Pattern**:
```go
func (c *controller) GenerateInvoiceSplitsByLegacyNumber(legacyNumber string) (*model.GenerateSplitsResponse, error) {
    // 1. Query Notion
    notionPage, err := c.store.Notion.QueryClientInvoiceByNumber(legacyNumber)
    if err != nil {
        return nil, err
    }

    // 2. Enqueue worker job
    c.worker.Enqueue(worker.GenerateInvoiceSplitsMsg, worker.GenerateInvoiceSplitsPayload{
        InvoicePageID: notionPage.ID,
    })

    // 3. Return response
    return &model.GenerateSplitsResponse{
        LegacyNumber:   legacyNumber,
        InvoicePageID:  notionPage.ID,
        JobEnqueued:    true,
        Message:        "Invoice splits generation job enqueued successfully",
    }, nil
}
```

**Acceptance**:
- Controller method created in new file
- Uses existing `QueryClientInvoiceByNumber()` service method
- Enqueues worker job with correct message type
- Returns proper response model
- Handles errors appropriately
- Compiles without errors: `go build ./pkg/controller/invoice/...`

---

## Task 3: Add Handler Interface

**File(s)**: `pkg/handler/invoice/interface.go`

**Description**:
Add `GenerateSplits(c *gin.Context)` method signature to handler interface.

**Implementation**:
```go
type IHandler interface {
    // ... existing methods ...
    GenerateSplits(c *gin.Context)
}
```

**Acceptance**:
- Method signature added to interface
- Compiles without errors: `go build ./pkg/handler/invoice/...`

---

## Task 4: Implement Handler Method

**File(s)**:
- `pkg/handler/invoice/invoice.go`
- `pkg/handler/invoice/invoice_test.go` (optional test file)

**Description**:
Implement HTTP handler with:
1. Request parsing and validation
2. Controller method call
3. Response handling with proper HTTP status codes
4. Swagger annotations

**Implementation Pattern**:
```go
// GenerateSplits godoc
// @Summary Generate invoice splits by Legacy Number
// @Description Query Notion Client Invoices database and enqueue worker job to generate splits
// @id generateInvoiceSplits
// @Tags Invoice
// @Accept json
// @Produce json
// @Param body body request.GenerateSplitsRequest true "Generate Splits Request"
// @Success 200 {object} view.GenerateSplitsResponse
// @Failure 400 {object} view.ErrorResponse "Invalid request"
// @Failure 404 {object} view.ErrorResponse "Invoice not found"
// @Failure 500 {object} view.ErrorResponse "Internal server error"
// @Router /invoices/generate-splits [post]
func (h *handler) GenerateSplits(c *gin.Context) {
    var req request.GenerateSplitsRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
        return
    }

    l := h.logger.Fields(logger.Fields{
        "handler": "invoice",
        "method":  "GenerateSplits",
        "legacy_number": req.LegacyNumber,
    })

    resp, err := h.controller.GenerateInvoiceSplitsByLegacyNumber(req.LegacyNumber)
    if err != nil {
        l.Error(err, "failed to generate invoice splits")

        // Handle not found error
        if strings.Contains(err.Error(), "not found") {
            c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
            return
        }

        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
        return
    }

    l.Info("invoice splits generation job enqueued successfully")
    c.JSON(http.StatusOK, view.CreateResponse(resp, nil, nil, req, ""))
}
```

**Acceptance**:
- Handler method implemented with all error cases
- Request validation using `ShouldBindJSON`
- Proper logging with contextual fields
- HTTP status codes: 200 (success), 400 (bad request), 404 (not found), 500 (server error)
- Swagger annotations complete
- Compiles without errors: `go build ./pkg/handler/invoice/...`

---

## Task 5: Register Route

**File(s)**: `pkg/routes/v1.go`

**Description**:
Register the endpoint in the router with authentication and permission middleware.

**Implementation**:
```go
// In the private routes section (within authMiddleware)
private.POST("/invoices/generate-splits", permissionMiddleware(model.PermissionInvoiceEdit), h.Invoice.GenerateSplits)
```

**Acceptance**:
- Route registered under `/api/v1/invoices/generate-splits`
- Uses `authMiddleware` for authentication
- Uses `permissionMiddleware(model.PermissionInvoiceEdit)` for authorization
- Compiles without errors: `go build ./pkg/routes/...`
- Server starts successfully: `make dev`

---

## Task 6: Update Swagger Documentation

**File(s)**: Auto-generated files in `docs/swagger/`

**Description**:
Generate updated API documentation from Swagger annotations.

**Implementation**:
```bash
make gen-swagger
```

**Acceptance**:
- Swagger docs regenerated successfully
- New endpoint visible in Swagger UI at `http://localhost:8080/swagger/index.html`
- Request/response schemas properly documented
- All parameters and status codes documented

---

## Task 7: Integration Testing

**File(s)**:
- `scripts/test/integration_generate_splits.sh` (NEW - optional)
- Manual testing via curl/Postman

**Description**:
Test the complete flow end-to-end:
1. Valid request → 200 response with job enqueued
2. Invalid request → 400 validation error
3. Not found → 404 error
4. Worker processes the job successfully

**Test Cases**:

**Success Case**:
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}' | jq .
```
Expected: 200 with `job_enqueued: true`

**Validation Error**:
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": ""}' | jq .
```
Expected: 400 with validation error

**Not Found**:
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-NOTFOUND"}' | jq .
```
Expected: 404 with not found error

**Acceptance**:
- All test cases pass with expected responses
- Worker logs show job processing
- No errors in application logs
- Endpoint accessible and functional

---

## Execution Checklist

### Before Starting
- [ ] Development environment running (`make dev`)
- [ ] Notion API configured and accessible
- [ ] Worker infrastructure functional
- [ ] All dependencies installed

### During Implementation
- [ ] Task 1: Request model added
- [ ] Task 2: Controller method implemented
- [ ] Task 3: Handler interface updated
- [ ] Task 4: Handler method implemented
- [ ] Task 5: Route registered
- [ ] Task 6: Swagger docs generated
- [ ] Task 7: Integration testing completed

### After Completion
- [ ] All unit tests pass: `go test ./...`
- [ ] Linter clean: `make lint`
- [ ] Application builds: `make build`
- [ ] Manual testing successful
- [ ] Code follows project conventions
- [ ] Ready for code review

---

## Quick Reference

### Files to Create
1. `pkg/controller/invoice/generate_splits.go`

### Files to Modify
1. `pkg/handler/invoice/request/request.go`
2. `pkg/handler/invoice/interface.go`
3. `pkg/handler/invoice/invoice.go`
4. `pkg/routes/v1.go`

### Key Dependencies
- `service.Notion.QueryClientInvoiceByNumber()` - Already exists
- `worker.GenerateInvoiceSplitsMsg` - Already defined
- `worker.GenerateInvoiceSplitsPayload` - Already defined
- `model.PermissionInvoiceEdit` - Already exists

### Reference Implementation
- Similar handler: `pkg/handler/invoice/invoice.go:527` (MarkPaid)
- Similar controller: `pkg/controller/invoice/mark_paid.go`
- Worker handler: `pkg/worker/worker.go:104` (handleGenerateInvoiceSplits)

---

## Success Criteria

✅ All 7 tasks completed in order
✅ Endpoint returns correct responses for all cases
✅ Worker processes jobs successfully
✅ All tests pass
✅ Swagger documentation updated
✅ Code follows project conventions
✅ No linter warnings

---

**Ready to Implement**: Yes
**Next Step**: Use `/proceed` or `/dev:proceed` command to begin implementation
