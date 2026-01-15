# TASK-002: Create Controller Method for Generate Splits

## Priority
P0 - Foundation

## Estimated Effort
30 minutes

## Description
Create a new controller method that queries Notion by Legacy Number and enqueues the worker job to generate invoice splits.

## Dependencies
- TASK-001 (Request model must exist)

## File to Create
`/Users/quang/workspace/dwarvesf/fortress-api/pkg/controller/invoice/generate_splits.go`

## Implementation Details

### Create New Controller File

Create file with the following structure:

```go
package invoice

import (
	"errors"
	"fmt"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/worker"
)

// GenerateSplitsInput contains the input for generating invoice splits
type GenerateSplitsInput struct {
	LegacyNumber string
}

// GenerateSplitsResult contains the result of the generate splits operation
type GenerateSplitsResult struct {
	LegacyNumber  string `json:"legacy_number"`
	InvoicePageID string `json:"invoice_page_id"`
	JobEnqueued   bool   `json:"job_enqueued"`
	Message       string `json:"message"`
}

// GenerateInvoiceSplitsByLegacyNumber queries Notion for an invoice by Legacy Number
// and enqueues a worker job to generate invoice splits.
func (c *controller) GenerateInvoiceSplitsByLegacyNumber(input GenerateSplitsInput) (*GenerateSplitsResult, error) {
	l := c.logger.Fields(logger.Fields{
		"controller":   "invoice",
		"method":       "GenerateInvoiceSplitsByLegacyNumber",
		"legacyNumber": input.LegacyNumber,
	})

	l.Debug("starting generate invoice splits by legacy number")

	// 1. Validate input
	if input.LegacyNumber == "" {
		l.Debug("empty legacy number provided")
		return nil, errors.New("legacy number is required")
	}

	// 2. Query Notion Client Invoices by Legacy Number
	l.Debug("querying Notion for invoice by legacy number")
	notionPage, err := c.service.Notion.QueryClientInvoiceByNumber(input.LegacyNumber)
	if err != nil {
		l.Errorf(err, "failed to query Notion for invoice")
		return nil, fmt.Errorf("failed to query invoice: %w", err)
	}

	if notionPage == nil {
		l.Debug("invoice not found in Notion")
		return nil, fmt.Errorf("invoice with legacy number %s not found", input.LegacyNumber)
	}

	l.Debugf("found invoice in Notion: pageID=%s", notionPage.ID)

	// 3. Enqueue worker job to generate splits
	l.Debug("enqueuing invoice splits generation job")
	c.worker.Enqueue(worker.GenerateInvoiceSplitsMsg, worker.GenerateInvoiceSplitsPayload{
		InvoicePageID: notionPage.ID,
	})
	l.Debug("invoice splits generation job enqueued successfully")

	// 4. Build result
	result := &GenerateSplitsResult{
		LegacyNumber:  input.LegacyNumber,
		InvoicePageID: notionPage.ID,
		JobEnqueued:   true,
		Message:       "Invoice splits generation job enqueued successfully",
	}

	l.Infof("invoice splits generation initiated: legacyNumber=%s pageID=%s", input.LegacyNumber, notionPage.ID)

	return result, nil
}
```

### Notes on Implementation

1. **Error Handling**: Return descriptive errors that can be converted to appropriate HTTP status codes
2. **Logging**: Follow existing logging patterns with structured fields
3. **Worker Integration**: Use the existing `GenerateInvoiceSplitsMsg` and `GenerateInvoiceSplitsPayload`
4. **Notion Service**: Reuse `QueryClientInvoiceByNumber` method (already exists)
5. **Result Struct**: Provide clear response structure for handler layer

### Reference Pattern

This controller follows the same pattern as `MarkInvoiceAsPaidByNumber` in `pkg/controller/invoice/mark_paid.go`:
- Query Notion by invoice number
- Validate existence
- Enqueue worker job
- Return result struct

## Testing Requirements

### Unit Test File

Create: `pkg/controller/invoice/generate_splits_test.go`

```go
package invoice

import (
	"errors"
	"testing"

	nt "github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dwarvesf/fortress-api/pkg/worker"
)

func TestController_GenerateInvoiceSplitsByLegacyNumber(t *testing.T) {
	tests := []struct {
		name          string
		input         GenerateSplitsInput
		mockNotionErr error
		mockNotionRes *nt.Page
		wantErr       bool
		wantResult    *GenerateSplitsResult
	}{
		{
			name: "success - invoice found and job enqueued",
			input: GenerateSplitsInput{
				LegacyNumber: "INV-2024-001",
			},
			mockNotionErr: nil,
			mockNotionRes: &nt.Page{
				ID: "test-page-id-123",
			},
			wantErr: false,
			wantResult: &GenerateSplitsResult{
				LegacyNumber:  "INV-2024-001",
				InvoicePageID: "test-page-id-123",
				JobEnqueued:   true,
				Message:       "Invoice splits generation job enqueued successfully",
			},
		},
		{
			name: "error - empty legacy number",
			input: GenerateSplitsInput{
				LegacyNumber: "",
			},
			wantErr: true,
		},
		{
			name: "error - invoice not found",
			input: GenerateSplitsInput{
				LegacyNumber: "INV-NOTFOUND",
			},
			mockNotionErr: nil,
			mockNotionRes: nil,
			wantErr:       true,
		},
		{
			name: "error - notion query error",
			input: GenerateSplitsInput{
				LegacyNumber: "INV-2024-001",
			},
			mockNotionErr: errors.New("notion api error"),
			mockNotionRes: nil,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := new(MockService)
			mockNotion := new(MockNotionService)
			mockWorker := new(MockWorker)
			mockLogger := new(MockLogger)

			if tt.input.LegacyNumber != "" {
				mockNotion.On("QueryClientInvoiceByNumber", tt.input.LegacyNumber).
					Return(tt.mockNotionRes, tt.mockNotionErr)
			}

			if tt.mockNotionRes != nil && tt.mockNotionErr == nil {
				mockWorker.On("Enqueue", worker.GenerateInvoiceSplitsMsg, mock.Anything).
					Return()
			}

			mockService.Notion = mockNotion
			mockLogger.On("Fields", mock.Anything).Return(mockLogger)
			mockLogger.On("Debug", mock.Anything)
			mockLogger.On("Debugf", mock.Anything, mock.Anything)
			mockLogger.On("Errorf", mock.Anything, mock.Anything)
			mockLogger.On("Infof", mock.Anything, mock.Anything, mock.Anything)

			ctrl := &controller{
				service: mockService,
				worker:  mockWorker,
				logger:  mockLogger,
			}

			// Execute
			result, err := ctrl.GenerateInvoiceSplitsByLegacyNumber(tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantResult.LegacyNumber, result.LegacyNumber)
				assert.Equal(t, tt.wantResult.InvoicePageID, result.InvoicePageID)
				assert.Equal(t, tt.wantResult.JobEnqueued, result.JobEnqueued)
			}

			mockNotion.AssertExpectations(t)
			mockWorker.AssertExpectations(t)
		})
	}
}
```

## Acceptance Criteria

- [ ] Controller file created at correct path
- [ ] `GenerateSplitsInput` struct is defined
- [ ] `GenerateSplitsResult` struct is defined with JSON tags
- [ ] `GenerateInvoiceSplitsByLegacyNumber` method is implemented
- [ ] Method queries Notion using `QueryClientInvoiceByNumber`
- [ ] Method enqueues worker job with correct message type and payload
- [ ] Error handling covers all edge cases (empty input, not found, query error)
- [ ] Logging follows existing patterns with structured fields
- [ ] Unit tests cover success and all error cases
- [ ] All tests pass: `go test ./pkg/controller/invoice/... -run TestController_GenerateInvoiceSplitsByLegacyNumber -v`
- [ ] Code builds without errors: `go build ./pkg/controller/invoice/...`

## Verification Commands

```bash
# Run tests
go test ./pkg/controller/invoice/... -run TestController_GenerateInvoiceSplitsByLegacyNumber -v

# Build check
go build ./pkg/controller/invoice/...

# Check test coverage
go test ./pkg/controller/invoice/... -cover
```

## Reference Files
- Similar controller: `pkg/controller/invoice/mark_paid.go` (MarkInvoiceAsPaidByNumber method)
- Worker types: `pkg/worker/message_types.go`
- Notion service interface: `pkg/service/notion/interface.go`
