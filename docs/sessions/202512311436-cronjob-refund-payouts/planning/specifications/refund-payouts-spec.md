# Specification: Refund Payouts Processing

## Overview

Add support for `type=refund` to the existing `POST /cronjobs/contractor-payouts` endpoint.

## Data Flow

```
Refund Requests (Status=Approved)
         │
         ▼
  Query Approved Refunds
         │
         ▼
  For Each Refund:
    ├─ Check idempotency (payout exists?)
    │       │
    │       ├─ Yes → Skip
    │       │
    │       └─ No → Create Payout
    │                  │
    │                  └─ Type: "Refund"
    │                  └─ Direction: "Outgoing"
    │                  └─ Refund Request relation set
    │                  └─ Person from refund
    │
    └─ Continue to next
```

## Service Layer Changes

### 1. RefundRequestsService (NEW)

**File**: `pkg/service/notion/refund_requests.go`

```go
type RefundRequestsService struct {
    client *nt.Client
    cfg    *config.Config
    logger logger.Logger
}

type ApprovedRefundData struct {
    PageID           string
    RefundID         string  // Title
    Amount           float64
    Currency         string
    ContractorPageID string  // From Contractor relation
    ContractorName   string  // Rollup from Contractor
    Reason           string  // Select: Advance Return, etc.
    Description      string  // Rich text
    DateApproved     string  // Date
}

func (s *RefundRequestsService) QueryApprovedRefunds(ctx) ([]*ApprovedRefundData, error)
```

### 2. ContractorPayoutsService (EXTEND)

**File**: `pkg/service/notion/contractor_payouts.go`

```go
// New input struct for refund payouts
type CreateRefundPayoutInput struct {
    Name              string
    ContractorPageID  string
    RefundRequestID   string  // Refund Request relation
    Amount            float64
    Currency          string
    Date              string
}

func (s *ContractorPayoutsService) CheckPayoutExistsByRefundRequest(ctx, refundRequestPageID) (bool, string, error)

func (s *ContractorPayoutsService) CreateRefundPayout(ctx, input CreateRefundPayoutInput) (string, error)
```

## Handler Layer Changes

### processRefundPayouts (NEW)

**File**: `pkg/handler/notion/contractor_payouts.go`

Add new method `processRefundPayouts` called when `type=refund`:

```go
func (h *handler) processRefundPayouts(c *gin.Context, l logger.Logger, payoutType string) {
    // 1. Query approved refund requests
    // 2. For each refund:
    //    - Validate contractor
    //    - Check idempotency
    //    - Create payout with Direction=Outgoing
    // 3. Return response (do NOT update refund status)
}
```

## Property Mappings

### Refund Request → Payout

| Refund Request | Payout | Notes |
|----------------|--------|-------|
| PageID | Refund Request relation | Link to source |
| Amount | Amount | Direct copy |
| Currency | Currency | Direct copy |
| Contractor | Person | Relation |
| - | Type | "Refund" |
| - | Direction | "Outgoing" |
| - | Status | "Pending" |
| Date Approved | Date | Optional |

## API Response

Same format as contractor payroll:

```json
{
  "data": {
    "payouts_created": 2,
    "refunds_processed": 3,
    "refunds_skipped": 1,
    "errors": 0,
    "details": [...],
    "type": "Refund"
  }
}
```
