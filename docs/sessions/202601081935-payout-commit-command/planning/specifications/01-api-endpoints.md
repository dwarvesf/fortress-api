# Specification: API Endpoints for Payout Commit

## Overview

This specification defines the API endpoints, handlers, controllers, and routing for the payout commit functionality in fortress-api.

## Architecture

Following the established layered architecture pattern:

```
Routes → Handler → Controller → Service (Notion) → External API
```

## File Structure

```
fortress-api/
├── pkg/
│   ├── handler/
│   │   └── contractorpayables/
│   │       ├── interface.go           [CREATE]
│   │       ├── contractorpayables.go  [CREATE]
│   │       └── request.go             [CREATE]
│   ├── controller/
│   │   └── contractorpayables/
│   │       ├── contractorpayables.go  [CREATE]
│   │       └── interface.go           [CREATE]
│   └── routes/
│       └── v1.go                      [MODIFY]
```

## API Endpoints

### 1. Preview Commit

**Endpoint**: `GET /api/v1/contractor-payables/preview-commit`

**Purpose**: Preview payables that will be committed for a given month and batch without making changes.

**Query Parameters**:
- `month` (string, required): Month in YYYY-MM format (e.g., "2025-01")
- `batch` (int, required): PayDay batch, must be 1 or 15

**Request Example**:
```
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=15
```

**Response 200 OK**:
```json
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "count": 3,
    "total_amount": 15000.00,
    "contractors": [
      {
        "name": "John Doe",
        "amount": 5000.00,
        "currency": "USD",
        "payable_id": "page-id-1"
      },
      {
        "name": "Jane Smith",
        "amount": 7500.00,
        "currency": "USD",
        "payable_id": "page-id-2"
      },
      {
        "name": "Bob Wilson",
        "amount": 2500.00,
        "currency": "USD",
        "payable_id": "page-id-3"
      }
    ]
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid month format or batch value
- `404 Not Found`: No pending payables found for criteria
- `500 Internal Server Error`: Notion API error or service failure

### 2. Execute Commit

**Endpoint**: `POST /api/v1/contractor-payables/commit`

**Purpose**: Execute the payout commit, updating all related records.

**Request Body**:
```json
{
  "month": "2025-01",
  "batch": 15
}
```

**Response 200 OK**:
```json
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "updated": 3,
    "failed": 0
  }
}
```

**Response 207 Multi-Status** (partial failure):
```json
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "updated": 2,
    "failed": 1
  },
  "errors": [
    {
      "payable_id": "page-id-3",
      "error": "failed to update payout: network timeout"
    }
  ]
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or validation error
- `404 Not Found`: No pending payables found for criteria
- `500 Internal Server Error`: Complete failure (no updates succeeded)

## Handler Layer

### File: `pkg/handler/contractorpayables/interface.go`

```go
package contractorpayables

import "github.com/gin-gonic/gin"

type IHandler interface {
    PreviewCommit(c *gin.Context)
    Commit(c *gin.Context)
}
```

### File: `pkg/handler/contractorpayables/request.go`

```go
package contractorpayables

// PreviewCommitRequest contains query parameters for preview endpoint
type PreviewCommitRequest struct {
    Month string `form:"month" binding:"required"` // YYYY-MM format
    Batch int    `form:"batch" binding:"required,oneof=1 15"`
}

// CommitRequest contains the request body for commit endpoint
type CommitRequest struct {
    Month string `json:"month" binding:"required"` // YYYY-MM format
    Batch int    `json:"batch" binding:"required,oneof=1 15"`
}

// PreviewCommitResponse contains the preview data
type PreviewCommitResponse struct {
    Month       string               `json:"month"`
    Batch       int                  `json:"batch"`
    Count       int                  `json:"count"`
    TotalAmount float64              `json:"total_amount"`
    Contractors []ContractorPreview  `json:"contractors"`
}

// ContractorPreview contains preview data for a single contractor
type ContractorPreview struct {
    Name       string  `json:"name"`
    Amount     float64 `json:"amount"`
    Currency   string  `json:"currency"`
    PayableID  string  `json:"payable_id"`
}

// CommitResponse contains the result of commit operation
type CommitResponse struct {
    Month   string        `json:"month"`
    Batch   int           `json:"batch"`
    Updated int           `json:"updated"`
    Failed  int           `json:"failed"`
    Errors  []CommitError `json:"errors,omitempty"`
}

// CommitError contains error details for failed updates
type CommitError struct {
    PayableID string `json:"payable_id"`
    Error     string `json:"error"`
}
```

### File: `pkg/handler/contractorpayables/contractorpayables.go`

```go
package contractorpayables

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/controller"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
    controller controller.IController
    logger     logger.Logger
    config     *config.Config
}

func New(controller controller.IController, logger logger.Logger, config *config.Config) IHandler {
    return &handler{
        controller: controller,
        logger:     logger,
        config:     config,
    }
}

// PreviewCommit godoc
// @Summary Preview contractor payables commit
// @Description Preview which contractor payables would be committed for a given month and batch (PayDay)
// @Tags Contractor Payables
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format (e.g., 2025-01)"
// @Param batch query int true "PayDay batch (1 or 15)"
// @Success 200 {object} view.Response{data=PreviewCommitResponse}
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/contractor-payables/preview-commit [get]
func (h *handler) PreviewCommit(c *gin.Context) {
    l := h.logger.Fields(logger.Fields{
        "handler": "contractorpayables",
        "method":  "PreviewCommit",
    })

    var req PreviewCommitRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        l.Error(err, "failed to bind query parameters")
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    // Validate month format (YYYY-MM)
    if !isValidMonthFormat(req.Month) {
        l.Error(nil, "invalid month format")
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
            nil, nil, "Invalid month format. Use YYYY-MM (e.g., 2025-01)"))
        return
    }

    l.Debug("calling controller.ContractorPayables.PreviewCommit")

    result, err := h.controller.ContractorPayables.PreviewCommit(c.Request.Context(), req.Month, req.Batch)
    if err != nil {
        l.Error(err, "failed to preview commit")

        // Check if no payables found (not an error, return 200 with count=0)
        if strings.Contains(err.Error(), "no pending payables") {
            c.JSON(http.StatusOK, view.CreateResponse[any](
                PreviewCommitResponse{
                    Month:       req.Month,
                    Batch:       req.Batch,
                    Count:       0,
                    TotalAmount: 0,
                    Contractors: []ContractorPreview{},
                },
                nil, nil, nil, ""))
            return
        }

        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    c.JSON(http.StatusOK, view.CreateResponse[any](result, nil, nil, nil, ""))
}

// Commit godoc
// @Summary Commit contractor payables to Paid status
// @Description Execute the payout commit, updating all related records (Payables, Payouts, Invoice Splits, Refunds)
// @Tags Contractor Payables
// @Accept json
// @Produce json
// @Param request body CommitRequest true "Commit request"
// @Success 200 {object} view.Response{data=CommitResponse}
// @Success 207 {object} view.Response{data=CommitResponse} "Partial success (some updates failed)"
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /api/v1/contractor-payables/commit [post]
func (h *handler) Commit(c *gin.Context) {
    l := h.logger.Fields(logger.Fields{
        "handler": "contractorpayables",
        "method":  "Commit",
    })

    var req CommitRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        l.Error(err, "failed to bind request body")
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    // Validate month format (YYYY-MM)
    if !isValidMonthFormat(req.Month) {
        l.Error(nil, "invalid month format")
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
            nil, nil, "Invalid month format. Use YYYY-MM (e.g., 2025-01)"))
        return
    }

    l.Debug("calling controller.ContractorPayables.CommitPayables")

    result, err := h.controller.ContractorPayables.CommitPayables(c.Request.Context(), req.Month, req.Batch)
    if err != nil {
        l.Error(err, "failed to commit payables")

        // Check if no payables found
        if strings.Contains(err.Error(), "no pending payables") {
            c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
            return
        }

        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    // Return 207 Multi-Status if there were any failures
    statusCode := http.StatusOK
    if result.Failed > 0 {
        statusCode = http.StatusMultiStatus
    }

    c.JSON(statusCode, view.CreateResponse[any](result, nil, nil, nil, ""))
}

// isValidMonthFormat validates month is in YYYY-MM format
func isValidMonthFormat(month string) bool {
    // Check format: YYYY-MM (10 characters, hyphen at position 4)
    if len(month) != 7 {
        return false
    }
    if month[4] != '-' {
        return false
    }

    // Parse to validate it's a valid date
    parts := strings.Split(month, "-")
    if len(parts) != 2 {
        return false
    }

    // Year should be 4 digits
    if len(parts[0]) != 4 {
        return false
    }

    // Month should be 2 digits and between 01-12
    if len(parts[1]) != 2 {
        return false
    }

    // Could add more validation (numeric check, month range) if needed
    return true
}
```

## Controller Layer

### File: `pkg/controller/contractorpayables/interface.go`

```go
package contractorpayables

import (
    "context"

    "github.com/dwarvesf/fortress-api/pkg/handler/contractorpayables"
)

type IController interface {
    PreviewCommit(ctx context.Context, month string, batch int) (*contractorpayables.PreviewCommitResponse, error)
    CommitPayables(ctx context.Context, month string, batch int) (*contractorpayables.CommitResponse, error)
}
```

### File: `pkg/controller/contractorpayables/contractorpayables.go`

```go
package contractorpayables

import (
    "context"
    "fmt"
    "time"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/handler/contractorpayables"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/service"
)

type controller struct {
    config  *config.Config
    logger  logger.Logger
    service *service.Service
}

func New(config *config.Config, logger logger.Logger, service *service.Service) IController {
    return &controller{
        config:  config,
        logger:  logger,
        service: service,
    }
}

// PreviewCommit queries pending payables and returns preview data
func (c *controller) PreviewCommit(ctx context.Context, month string, batch int) (*contractorpayables.PreviewCommitResponse, error) {
    l := c.logger.Fields(logger.Fields{
        "controller": "contractorpayables",
        "method":     "PreviewCommit",
        "month":      month,
        "batch":      batch,
    })

    l.Debug("querying pending payables")

    // Convert month to period date (YYYY-MM-01)
    period := month + "-01"

    // Query all pending payables for the period
    payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, period)
    if err != nil {
        l.Error(err, "failed to query pending payables")
        return nil, fmt.Errorf("failed to query pending payables: %w", err)
    }

    l.Debug(fmt.Sprintf("found %d pending payables", len(payables)))

    // Filter by PayDay (batch)
    var filtered []contractorpayables.ContractorPreview
    var totalAmount float64

    for _, payable := range payables {
        // Get contractor's PayDay from Service Rate
        payDay, err := c.service.Notion.ContractorPayables.GetContractorPayDay(ctx, payable.ContractorPageID)
        if err != nil {
            l.Debug(fmt.Sprintf("failed to get PayDay for contractor %s: %v", payable.ContractorPageID, err))
            continue
        }

        // Filter by batch
        if payDay != batch {
            l.Debug(fmt.Sprintf("skipping payable %s: PayDay %d != batch %d", payable.PageID, payDay, batch))
            continue
        }

        preview := contractorpayables.ContractorPreview{
            Name:      payable.ContractorName,
            Amount:    payable.Total,
            Currency:  payable.Currency,
            PayableID: payable.PageID,
        }
        filtered = append(filtered, preview)
        totalAmount += payable.Total
    }

    l.Debug(fmt.Sprintf("filtered to %d payables for batch %d", len(filtered), batch))

    return &contractorpayables.PreviewCommitResponse{
        Month:       month,
        Batch:       batch,
        Count:       len(filtered),
        TotalAmount: totalAmount,
        Contractors: filtered,
    }, nil
}

// CommitPayables executes the cascade status update for all matching payables
func (c *controller) CommitPayables(ctx context.Context, month string, batch int) (*contractorpayables.CommitResponse, error) {
    l := c.logger.Fields(logger.Fields{
        "controller": "contractorpayables",
        "method":     "CommitPayables",
        "month":      month,
        "batch":      batch,
    })

    l.Debug("starting commit operation")

    // Convert month to period date (YYYY-MM-01)
    period := month + "-01"

    // Query all pending payables for the period
    payables, err := c.service.Notion.ContractorPayables.QueryPendingPayablesByPeriod(ctx, period)
    if err != nil {
        l.Error(err, "failed to query pending payables")
        return nil, fmt.Errorf("failed to query pending payables: %w", err)
    }

    if len(payables) == 0 {
        return nil, fmt.Errorf("no pending payables found for month %s", month)
    }

    // Filter by PayDay
    var toCommit []PayableToCommit
    for _, payable := range payables {
        payDay, err := c.service.Notion.ContractorPayables.GetContractorPayDay(ctx, payable.ContractorPageID)
        if err != nil {
            l.Debug(fmt.Sprintf("failed to get PayDay for contractor %s: %v", payable.ContractorPageID, err))
            continue
        }

        if payDay == batch {
            toCommit = append(toCommit, PayableToCommit{
                PageID:            payable.PageID,
                ContractorPageID:  payable.ContractorPageID,
                PayoutItemPageIDs: payable.PayoutItemPageIDs,
            })
        }
    }

    if len(toCommit) == 0 {
        return nil, fmt.Errorf("no pending payables found for month %s batch %d", month, batch)
    }

    l.Debug(fmt.Sprintf("committing %d payables", len(toCommit)))

    // Execute cascade updates for each payable
    var successCount, failCount int
    var errors []contractorpayables.CommitError

    for _, payable := range toCommit {
        if err := c.commitSinglePayable(ctx, payable); err != nil {
            l.Error(err, fmt.Sprintf("failed to commit payable %s", payable.PageID))
            failCount++
            errors = append(errors, contractorpayables.CommitError{
                PayableID: payable.PageID,
                Error:     err.Error(),
            })
        } else {
            successCount++
        }
    }

    l.Info(fmt.Sprintf("commit complete: %d succeeded, %d failed", successCount, failCount))

    return &contractorpayables.CommitResponse{
        Month:   month,
        Batch:   batch,
        Updated: successCount,
        Failed:  failCount,
        Errors:  errors,
    }, nil
}

// PayableToCommit contains the data needed to commit a single payable
type PayableToCommit struct {
    PageID            string
    ContractorPageID  string
    PayoutItemPageIDs []string
}

// commitSinglePayable performs the cascade update for a single payable
func (c *controller) commitSinglePayable(ctx context.Context, payable PayableToCommit) error {
    l := c.logger.Fields(logger.Fields{
        "controller": "contractorpayables",
        "method":     "commitSinglePayable",
        "payable_id": payable.PageID,
    })

    l.Debug("starting cascade update")

    // Step 1: Update each Payout Item and its related Invoice Split/Refund
    for _, payoutPageID := range payable.PayoutItemPageIDs {
        if err := c.commitPayoutItem(ctx, payoutPageID); err != nil {
            l.Error(err, fmt.Sprintf("failed to commit payout item %s", payoutPageID))
            // Continue with other payouts (best-effort)
        }
    }

    // Step 2: Update the Contractor Payable itself
    paymentDate := time.Now().Format("2006-01-02")
    if err := c.service.Notion.ContractorPayables.UpdatePayableStatus(ctx, payable.PageID, "Paid", paymentDate); err != nil {
        l.Error(err, "failed to update payable status")
        return fmt.Errorf("failed to update payable status: %w", err)
    }

    l.Debug("cascade update complete")
    return nil
}

// commitPayoutItem updates a payout item and its related records
func (c *controller) commitPayoutItem(ctx context.Context, payoutPageID string) error {
    l := c.logger.Fields(logger.Fields{
        "controller":    "contractorpayables",
        "method":        "commitPayoutItem",
        "payout_page_id": payoutPageID,
    })

    // Get payout with relations (Invoice Split, Refund)
    payout, err := c.service.Notion.ContractorPayouts.GetPayoutWithRelations(ctx, payoutPageID)
    if err != nil {
        l.Error(err, "failed to get payout with relations")
        return fmt.Errorf("failed to get payout with relations: %w", err)
    }

    // Update Invoice Split if exists
    if payout.InvoiceSplitID != "" {
        l.Debug(fmt.Sprintf("updating invoice split %s", payout.InvoiceSplitID))
        if err := c.service.Notion.InvoiceSplit.UpdateInvoiceSplitStatus(ctx, payout.InvoiceSplitID, "Paid"); err != nil {
            l.Error(err, "failed to update invoice split")
            // Continue (best-effort)
        }
    }

    // Update Refund Request if exists
    if payout.RefundRequestID != "" {
        l.Debug(fmt.Sprintf("updating refund request %s", payout.RefundRequestID))
        if err := c.service.Notion.RefundRequests.UpdateRefundRequestStatus(ctx, payout.RefundRequestID, "Paid"); err != nil {
            l.Error(err, "failed to update refund request")
            // Continue (best-effort)
        }
    }

    // Update Payout Item status
    l.Debug("updating payout status")
    if err := c.service.Notion.ContractorPayouts.UpdatePayoutStatus(ctx, payoutPageID, "Paid"); err != nil {
        l.Error(err, "failed to update payout status")
        return fmt.Errorf("failed to update payout status: %w", err)
    }

    return nil
}
```

## Routes Registration

### File: `pkg/routes/v1.go` (MODIFY)

Add the following to the routes registration:

```go
import (
    // ... existing imports
    contractorPayablesHandler "github.com/dwarvesf/fortress-api/pkg/handler/contractorpayables"
)

// Inside the router setup function, add:

// Contractor Payables routes
contractorPayablesGroup := v1.Group("/contractor-payables")
contractorPayablesGroup.Use(authMiddleware)
{
    // Preview commit requires PayrollsRead permission
    contractorPayablesGroup.GET("/preview-commit",
        middleware.RequirePermission(model.PermissionPayrollsRead),
        h.ContractorPayables.PreviewCommit,
    )

    // Commit requires PayrollsCreate permission
    contractorPayablesGroup.POST("/commit",
        middleware.RequirePermission(model.PermissionPayrollsCreate),
        h.ContractorPayables.Commit,
    )
}
```

## Controller Registration

In the main controller initialization (likely in `pkg/controller/controller.go`), add:

```go
import (
    contractorPayablesController "github.com/dwarvesf/fortress-api/pkg/controller/contractorpayables"
)

type Controller struct {
    // ... existing controllers
    ContractorPayables contractorPayablesController.IController
}

// In New() function:
ContractorPayables: contractorPayablesController.New(cfg, logger, service),
```

## Handler Registration

In the handler initialization (likely in `pkg/handler/handler.go`), add:

```go
import (
    contractorPayablesHandler "github.com/dwarvesf/fortress-api/pkg/handler/contractorpayables"
)

type Handler struct {
    // ... existing handlers
    ContractorPayables contractorPayablesHandler.IHandler
}

// In New() function:
ContractorPayables: contractorPayablesHandler.New(controller, logger, cfg),
```

## Testing Considerations

### Unit Tests

1. **Handler Tests**:
   - Test query parameter validation (month format, batch values)
   - Test request body validation
   - Test error response formats
   - Test success/partial failure responses

2. **Controller Tests**:
   - Mock Notion services
   - Test PayDay filtering logic
   - Test cascade update sequence
   - Test error handling and continue-on-error behavior
   - Test response aggregation (counts, errors)

### Integration Tests

1. Test full flow with test Notion databases
2. Test idempotency (re-running commit on same data)
3. Test partial failure scenarios
4. Test empty result handling

## Security Considerations

1. **Permissions**: Enforce PayrollsRead and PayrollsCreate permissions
2. **Validation**: Strict validation of month format and batch values
3. **Logging**: Log all operations with user context for audit
4. **Rate Limiting**: Consider rate limiting on commit endpoint

## Performance Considerations

1. **Batch Size**: Current design handles payables sequentially. For large batches (>50 payables), consider:
   - Adding progress feedback
   - Implementing pagination in preview
   - Adding timeouts

2. **Caching**: Consider caching PayDay values per contractor to reduce Service Rate queries

## Error Messages

User-friendly error messages:
- "Invalid month format. Use YYYY-MM (e.g., 2025-01)"
- "Batch must be 1 or 15"
- "No pending payables found for 2025-01 batch 15"
- "Successfully committed 3 payables for 2025-01 batch 15"
- "Committed 2/3 payables (1 failed - check logs)"

## References

- ADR-001: Cascade Status Update Strategy
- Spec: `/docs/specs/payout-commit-command.md`
- Existing Handler Pattern: `pkg/handler/payroll/mark_paid.go`
- Existing Controller Pattern: `pkg/controller/`
