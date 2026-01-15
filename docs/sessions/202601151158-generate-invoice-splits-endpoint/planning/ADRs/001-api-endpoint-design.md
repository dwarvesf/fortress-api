# ADR 001: API Endpoint Design for Invoice Splits Generation

## Status
Accepted

## Context

We need to create an API endpoint that allows generating/regenerating invoice splits by invoice Legacy Number. This functionality is required to trigger the invoice splits generation process manually or programmatically without waiting for the invoice to be marked as paid.

### Current System Analysis

1. **Existing Worker Infrastructure**: The `handleGenerateInvoiceSplits` worker job already exists in `pkg/worker/worker.go:104` and processes invoice splits generation
2. **Worker Message Type**: `GenerateInvoiceSplitsMsg` constant and `GenerateInvoiceSplitsPayload` struct are defined in `pkg/worker/message_types.go`
3. **Notion Integration**:
   - Client Invoices database ID: `2bf64b29-b84c-80e2-8cc7-000bfe534203`
   - Legacy Number field exists and is used to query invoices
   - Splits Generated checkbox provides idempotency tracking
4. **Similar Pattern**: `POST /api/v1/invoices/mark-paid` endpoint follows the same pattern of accepting invoice number, querying Notion, and enqueuing worker job

### Business Requirements

- Accept invoice Legacy Number as input
- Query Notion Client Invoices database by Legacy Number
- Enqueue worker job to generate invoice splits
- Return success response immediately (async processing)
- Support both manual API calls and potential automation

## Decision

We will create a new REST endpoint following the existing codebase patterns:

### Endpoint Specification

```
POST /api/v1/invoices/generate-splits
```

**Request Body:**
```json
{
  "legacy_number": "INV-2024-001"
}
```

**Success Response (200):**
```json
{
  "data": {
    "legacy_number": "INV-2024-001",
    "invoice_page_id": "2bf64b29-b84c-80e2-8cc7-000bfe534203",
    "job_enqueued": true,
    "message": "Invoice splits generation job enqueued successfully"
  },
  "error": null
}
```

**Error Responses:**
- 400 Bad Request: Missing or invalid legacy_number
- 404 Not Found: Invoice not found in Notion
- 500 Internal Server Error: System error

### Technical Approach

1. **Handler Layer** (`pkg/handler/invoice/invoice.go`):
   - Add `GenerateSplits(c *gin.Context)` method
   - Parse and validate request body
   - Call controller method
   - Return standardized response

2. **Controller Layer** (new file `pkg/controller/invoice/generate_splits.go`):
   - Query Notion by Legacy Number using `service.Notion.QueryClientInvoiceByNumber()`
   - Extract invoice page ID
   - Enqueue worker job with `GenerateInvoiceSplitsPayload`
   - Return result struct

3. **Request Model** (`pkg/handler/invoice/request/request.go`):
   - Add `GenerateSplitsRequest` struct with validation

4. **Route Registration** (`pkg/routes/v1.go`):
   - Register under `/api/v1/invoices` group
   - Apply authentication and permission middleware

### File Structure

```
pkg/
├── routes/v1.go                                    [MODIFY]
├── handler/invoice/
│   ├── interface.go                                [MODIFY]
│   ├── invoice.go                                  [MODIFY]
│   └── request/request.go                          [MODIFY]
├── controller/invoice/
│   └── generate_splits.go                          [NEW]
```

## Consequences

### Positive

1. **Consistency**: Follows existing codebase patterns (similar to `mark-paid` endpoint)
2. **Separation of Concerns**: Clear layered architecture (handler → controller → service → worker)
3. **Async Processing**: Worker pattern allows non-blocking operation
4. **Idempotency**: Worker already checks `Splits Generated` checkbox to prevent duplicate processing
5. **Reusability**: Existing Notion query methods can be reused
6. **Testability**: Each layer can be unit tested independently

### Negative

1. **No Immediate Feedback**: Response returns before splits are actually generated (async nature)
2. **Error Visibility**: Worker errors are logged but not returned to API caller
3. **No Progress Tracking**: Caller cannot poll for job completion status

### Mitigation Strategies

1. For error visibility: Worker logs are centralized and can be monitored
2. For progress tracking: Future enhancement could add job status endpoint if needed
3. For immediate validation: Handler validates invoice existence before enqueuing

## Alternatives Considered

### Alternative 1: Synchronous Processing
**Rejected**: Would block HTTP request for potentially long-running operation. Worker pattern is already established in codebase.

### Alternative 2: Accept Invoice Page ID Instead of Legacy Number
**Rejected**: Legacy Number is more user-friendly and matches existing patterns (mark-paid endpoint). Page IDs are internal Notion identifiers.

### Alternative 3: Webhook-Only Approach
**Rejected**: API endpoint provides more flexibility for manual triggers, testing, and integration with other systems.

## Implementation Notes

### Dependencies
- Existing Notion service methods (`QueryClientInvoiceByNumber`)
- Existing worker infrastructure (`GenerateInvoiceSplitsMsg`, `handleGenerateInvoiceSplits`)
- Existing permission system (`model.PermissionInvoiceEdit`)

### Testing Strategy
- Unit tests for request validation
- Unit tests for controller logic (mock Notion service and worker)
- Integration tests for full endpoint flow
- Golden file comparison for handler responses

### Rollout Plan
1. Implement and test in local environment
2. Deploy to staging for QA validation
3. Document endpoint in Swagger
4. Production deployment with monitoring

## References

- Existing endpoint: `POST /api/v1/invoices/mark-paid` (pkg/handler/invoice/invoice.go:527)
- Worker implementation: `handleGenerateInvoiceSplits` (pkg/worker/worker.go:104)
- Notion service: `QueryClientInvoiceByNumber` (pkg/service/notion/invoice.go)
- Similar controller pattern: `pkg/controller/invoice/mark_paid.go`
