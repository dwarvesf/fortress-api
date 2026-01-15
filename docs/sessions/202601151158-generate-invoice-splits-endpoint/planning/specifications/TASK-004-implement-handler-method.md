# TASK-004: Implement Handler Method

## Priority
P1 - Core Implementation

## Estimated Effort
30 minutes

## Description
Implement the `GenerateSplits` handler method that processes HTTP requests, validates input, calls the controller, and returns responses.

## Dependencies
- TASK-001 (Request model)
- TASK-002 (Controller method)
- TASK-003 (Handler interface)

## File to Modify
`/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/invoice/invoice.go`

## Implementation Details

### Add Response Struct

Add response struct after the `MarkPaidResponse` struct (around line 511):

```go
// GenerateSplitsResponse is the response for generate splits endpoint
type GenerateSplitsResponse struct {
	LegacyNumber  string `json:"legacy_number"`
	InvoicePageID string `json:"invoice_page_id"`
	JobEnqueued   bool   `json:"job_enqueued"`
	Message       string `json:"message"`
} // @name GenerateSplitsResponse
```

### Add Handler Method

Add the handler method at the end of the file (after the `MarkPaid` method around line 572):

```go
// GenerateSplits godoc
// @Summary Generate invoice splits by legacy number
// @Description Generate invoice splits by querying Notion with legacy number and enqueuing worker job
// @id generateInvoiceSplits
// @Tags Invoice
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body request.GenerateSplitsRequest true "Generate splits request"
// @Success 200 {object} GenerateSplitsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /invoices/generate-splits [post]
func (h *handler) GenerateSplits(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "invoice",
		"method":  "GenerateSplits",
	})

	l.Debug("handling generate invoice splits request")

	// 1. Parse request
	var req request.GenerateSplitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Error(err, "invalid request body")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	l.Debugf("received request: legacyNumber=%s", req.LegacyNumber)

	// 2. Validate request
	if err := req.Validate(); err != nil {
		l.Error(err, "request validation failed")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 3. Call controller
	result, err := h.controller.Invoice.GenerateInvoiceSplitsByLegacyNumber(invoiceCtrl.GenerateSplitsInput{
		LegacyNumber: req.LegacyNumber,
	})
	if err != nil {
		l.Error(err, "failed to generate invoice splits")
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, req, ""))
			return
		}
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, req, ""))
		return
	}

	// 4. Build response
	response := GenerateSplitsResponse{
		LegacyNumber:  result.LegacyNumber,
		InvoicePageID: result.InvoicePageID,
		JobEnqueued:   result.JobEnqueued,
		Message:       result.Message,
	}

	l.Infof("invoice splits generation initiated: legacyNumber=%s pageID=%s", result.LegacyNumber, result.InvoicePageID)
	c.JSON(http.StatusOK, view.CreateResponse(response, nil, nil, req, ""))
}
```

### Import Statements

Ensure the following imports are present at the top of the file:

```go
import (
	// ... existing imports ...
	invoiceCtrl "github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/handler/invoice/request"
	// ... existing imports ...
)
```

### Notes on Implementation

1. **Swagger Annotations**: Complete godoc with all parameters for API documentation
2. **Error Handling**: Distinguish between 404 (not found) and 500 (server error)
3. **Logging**: Use structured logging with context fields
4. **Response Pattern**: Follow `view.CreateResponse` pattern used throughout the codebase
5. **Validation**: Explicit validation call before controller invocation
6. **HTTP Status Codes**:
   - 200: Success
   - 400: Bad request (validation error)
   - 404: Invoice not found
   - 500: Internal server error

## Testing Requirements

### Handler Test File

Add to existing test file or create: `pkg/handler/invoice/invoice_test.go`

```go
func TestHandler_GenerateSplits(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockCtrlResult *invoiceCtrl.GenerateSplitsResult
		mockCtrlErr    error
		wantStatus     int
		wantContains   string
	}{
		{
			name:        "success",
			requestBody: `{"legacy_number": "INV-2024-001"}`,
			mockCtrlResult: &invoiceCtrl.GenerateSplitsResult{
				LegacyNumber:  "INV-2024-001",
				InvoicePageID: "test-page-id-123",
				JobEnqueued:   true,
				Message:       "Invoice splits generation job enqueued successfully",
			},
			mockCtrlErr:  nil,
			wantStatus:   http.StatusOK,
			wantContains: "INV-2024-001",
		},
		{
			name:         "invalid json",
			requestBody:  `{invalid}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "error",
		},
		{
			name:         "empty legacy number",
			requestBody:  `{"legacy_number": ""}`,
			wantStatus:   http.StatusBadRequest,
			wantContains: "error",
		},
		{
			name:         "invoice not found",
			requestBody:  `{"legacy_number": "INV-NOTFOUND"}`,
			mockCtrlErr:  errors.New("invoice with legacy number INV-NOTFOUND not found"),
			wantStatus:   http.StatusNotFound,
			wantContains: "not found",
		},
		{
			name:         "controller error",
			requestBody:  `{"legacy_number": "INV-2024-001"}`,
			mockCtrlErr:  errors.New("internal error"),
			wantStatus:   http.StatusInternalServerError,
			wantContains: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/v1/invoices/generate-splits", strings.NewReader(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Mock controller
			mockCtrl := new(MockInvoiceController)
			if tt.mockCtrlResult != nil || tt.mockCtrlErr != nil {
				mockCtrl.On("GenerateInvoiceSplitsByLegacyNumber", mock.Anything).
					Return(tt.mockCtrlResult, tt.mockCtrlErr)
			}

			mockController := &controller.Controller{
				Invoice: mockCtrl,
			}

			// Create handler
			handler := &handler{
				controller: mockController,
				logger:     testLogger,
			}

			// Execute
			handler.GenerateSplits(c)

			// Assert
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantContains)

			if tt.mockCtrlResult != nil || tt.mockCtrlErr != nil {
				mockCtrl.AssertExpectations(t)
			}
		})
	}
}
```

### Golden File Test (Optional)

Create golden file: `pkg/handler/invoice/testdata/generate_splits/success.golden`

## Acceptance Criteria

- [ ] `GenerateSplitsResponse` struct is defined with Swagger annotation
- [ ] `GenerateSplits` handler method is implemented
- [ ] Method has complete Swagger/godoc annotations
- [ ] Request parsing and validation is implemented
- [ ] Controller is called with correct input
- [ ] Error handling distinguishes between 404 and 500 errors
- [ ] Success response returns 200 with correct data structure
- [ ] Logging follows existing patterns
- [ ] Unit tests cover success and all error cases
- [ ] All tests pass: `go test ./pkg/handler/invoice/... -run TestHandler_GenerateSplits -v`
- [ ] Code builds: `go build ./pkg/handler/invoice/...`

## Verification Commands

```bash
# Run handler tests
go test ./pkg/handler/invoice/... -run TestHandler_GenerateSplits -v

# Build check
go build ./pkg/handler/invoice/...

# Run all invoice handler tests
go test ./pkg/handler/invoice/... -v

# Generate Swagger docs (after implementation)
make gen-swagger
```

## Reference Files
- Similar handler: `MarkPaid` method in same file (lines 527-572)
- Response pattern: `MarkPaidResponse` struct (lines 505-511)
- Controller import: See existing imports at top of file
