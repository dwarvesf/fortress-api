# Specification: Commission Payout Processing

## Overview

Process pending commission invoice splits and create corresponding payout entries.

## Components

### 1. InvoiceSplitService Extensions

**File:** `pkg/service/notion/invoice_split.go`

#### 1.1 PendingCommissionSplit struct

```go
type PendingCommissionSplit struct {
    PageID       string
    Name         string
    Amount       float64
    Currency     string
    Role         string
    PersonPageID string  // From Person relation
}
```

#### 1.2 QueryPendingCommissionSplits()

Query invoice splits with Status=Pending and Type=Commission.

**Input:** `ctx context.Context`
**Output:** `[]PendingCommissionSplit, error`

**Filter:**
```go
Filter: And(
    Status == "Pending",
    Type == "Commission"
)
```

**Extract:**
- Name (title)
- Amount (number)
- Currency (select)
- Role (select)
- Person (relation → first ID)

### 2. ContractorPayoutsService Extensions

**File:** `pkg/service/notion/contractor_payouts.go`

#### 2.1 CheckPayoutExistsByInvoiceSplit()

Check if payout already exists for given invoice split.

**Input:** `ctx context.Context, invoiceSplitPageID string`
**Output:** `(exists bool, existingPayoutID string, error)`

**Filter:**
```go
Filter: InvoiceSplit.Contains(invoiceSplitPageID)
```

#### 2.2 CreateCommissionPayoutInput struct

```go
type CreateCommissionPayoutInput struct {
    Name             string
    ContractorPageID string  // Person relation
    InvoiceSplitID   string  // Invoice Split relation
    Amount           float64
    Currency         string
}
```

#### 2.3 CreateCommissionPayout()

Create commission payout entry.

**Input:** `ctx context.Context, input CreateCommissionPayoutInput`
**Output:** `(pageID string, error)`

**Properties:**
- Name: input.Name
- Amount: input.Amount
- Currency: input.Currency (select)
- Person: input.ContractorPageID (relation)
- Invoice Split: input.InvoiceSplitID (relation)
- Type: "Commission" (select)
- Direction: "Outgoing" (select)
- Status: "Pending" (status)

### 3. Handler

**File:** `pkg/handler/notion/contractor_payouts.go`

#### 3.1 processCommissionPayouts()

```go
func (h *handler) processCommissionPayouts(c *gin.Context, l logger.Logger, payoutType string)
```

**Flow:**
1. Get InvoiceSplitService and ContractorPayoutsService
2. Query pending commission splits
3. For each split:
   - Validate Person relation exists
   - Check if payout already exists (idempotency)
   - Create payout if not exists
4. Return summary response

## Data Flow

```
POST /cronjobs/create-contractor-payouts?type=commission
       │
       ▼
processCommissionPayouts()
       │
       ├─► InvoiceSplitService.QueryPendingCommissionSplits()
       │
       ├─► For each split:
       │       │
       │       ├─► Validate PersonPageID != ""
       │       │
       │       ├─► ContractorPayoutsService.CheckPayoutExistsByInvoiceSplit()
       │       │
       │       └─► ContractorPayoutsService.CreateCommissionPayout()
       │
       └─► Return JSON response
```

## Response Format

```json
{
  "data": {
    "payouts_created": 5,
    "splits_processed": 7,
    "splits_skipped": 2,
    "errors": 0,
    "details": [
      {
        "split_page_id": "xxx",
        "split_name": "Sales - ABC - Jan 2025",
        "person_id": "yyy",
        "amount": 1000000,
        "payout_page_id": "zzz",
        "status": "created"
      }
    ],
    "type": "Commission"
  }
}
```

## Error Handling

- Log errors but continue processing other splits
- Track error count in response
- Include error reason in details array
