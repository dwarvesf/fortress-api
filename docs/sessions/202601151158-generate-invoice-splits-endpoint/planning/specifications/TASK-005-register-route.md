# TASK-005: Register Route

## Priority
P2 - Integration

## Estimated Effort
10 minutes

## Description
Register the new endpoint in the route configuration with appropriate authentication and permission middleware.

## Dependencies
- TASK-001 (Request model)
- TASK-002 (Controller method)
- TASK-003 (Handler interface)
- TASK-004 (Handler implementation)

## File to Modify
`/Users/quang/workspace/dwarvesf/fortress-api/pkg/routes/v1.go`

## Implementation Details

### Add Route Registration

Add the route in the `invoiceGroup` block (around line 252-261). Add it after the `mark-paid` route:

**Current invoice routes (lines 252-261):**
```go
invoiceGroup := v1.Group("/invoices")
{
	invoiceGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.List)
	invoiceGroup.PUT("/:id/status", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.UpdateStatus)
	invoiceGroup.POST("/:id/calculate-commissions", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsCommissionRateEdit), h.Invoice.CalculateCommissions)
	invoiceGroup.GET("/template", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.GetTemplate)
	invoiceGroup.POST("/send", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.Send)
	invoiceGroup.POST("/contractor/generate", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceCreate), h.Invoice.GenerateContractorInvoice)
	invoiceGroup.POST("/mark-paid", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.MarkPaid)
}
```

**Updated invoice routes:**
```go
invoiceGroup := v1.Group("/invoices")
{
	invoiceGroup.GET("", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.List)
	invoiceGroup.PUT("/:id/status", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.UpdateStatus)
	invoiceGroup.POST("/:id/calculate-commissions", conditionalAuthMW, conditionalPermMW(model.PermissionProjectsCommissionRateEdit), h.Invoice.CalculateCommissions)
	invoiceGroup.GET("/template", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.GetTemplate)
	invoiceGroup.POST("/send", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceRead), h.Invoice.Send)
	invoiceGroup.POST("/contractor/generate", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceCreate), h.Invoice.GenerateContractorInvoice)
	invoiceGroup.POST("/mark-paid", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.MarkPaid)
	invoiceGroup.POST("/generate-splits", conditionalAuthMW, conditionalPermMW(model.PermissionInvoiceEdit), h.Invoice.GenerateSplits)  // Add this line
}
```

### Route Configuration Details

1. **HTTP Method**: `POST` - This endpoint modifies state by enqueuing a job
2. **Path**: `/generate-splits` - Clear and RESTful, follows existing pattern
3. **Authentication**: `conditionalAuthMW` - Requires authentication (bypassed in local env)
4. **Permission**: `model.PermissionInvoiceEdit` - Same as mark-paid, requires invoice edit permission
5. **Handler**: `h.Invoice.GenerateSplits` - Links to our handler method

### Full Endpoint URL

When deployed, the endpoint will be available at:
```
POST /api/v1/invoices/generate-splits
```

### Middleware Flow

1. **conditionalAuthMW**: Validates JWT token (or bypass in local env)
2. **conditionalPermMW(model.PermissionInvoiceEdit)**: Checks user has invoice edit permission
3. **h.Invoice.GenerateSplits**: Executes handler method

## Testing Requirements

### Manual Testing

After implementation, test the endpoint:

```bash
# Local environment (auth bypassed)
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}'

# With authentication
curl -X POST https://api.example.com/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"legacy_number": "INV-2024-001"}'

# Expected success response
{
  "data": {
    "legacy_number": "INV-2024-001",
    "invoice_page_id": "abc123...",
    "job_enqueued": true,
    "message": "Invoice splits generation job enqueued successfully"
  },
  "error": null
}
```

### Integration Test

Add to integration test suite (if exists):

```go
func TestGenerateSplitsEndpoint(t *testing.T) {
	// Start test server
	router := setupTestRouter()

	// Test valid request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/invoices/generate-splits",
		strings.NewReader(`{"legacy_number": "INV-2024-001"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Add more assertions...
}
```

## Acceptance Criteria

- [ ] Route is registered in the correct location (invoice group)
- [ ] Route uses `POST` method
- [ ] Route path is `/generate-splits`
- [ ] Route includes `conditionalAuthMW` middleware
- [ ] Route includes `conditionalPermMW(model.PermissionInvoiceEdit)` middleware
- [ ] Route links to `h.Invoice.GenerateSplits` handler
- [ ] Code compiles: `go build ./pkg/routes/...`
- [ ] Application starts successfully: `make dev`
- [ ] Endpoint is accessible at `/api/v1/invoices/generate-splits`
- [ ] Swagger documentation includes the new endpoint after running `make gen-swagger`

## Verification Commands

```bash
# Build check
go build ./pkg/routes/...

# Start application
make dev

# In another terminal, test endpoint
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-TEST-001"}'

# Generate Swagger docs
make gen-swagger

# Check Swagger UI
# Navigate to http://localhost:8080/swagger/index.html
# Look for POST /api/v1/invoices/generate-splits
```

## Notes

### Permission Model
The endpoint uses `model.PermissionInvoiceEdit` which restricts access to users with invoice editing permissions. This is appropriate because:
- Generating splits can affect invoice processing
- Similar to mark-paid functionality (uses same permission)
- Prevents unauthorized users from triggering worker jobs

### Local Environment
In local environment (`cfg.Env == "local"`), both authentication and permission middleware are bypassed for easier development. This is controlled by the conditional middleware setup at the top of the routes file.

## Reference Files
- Route group definition: `pkg/routes/v1.go` (lines 252-261)
- Similar route: `mark-paid` route in the same group
- Permission constants: `pkg/model/permission.go`
