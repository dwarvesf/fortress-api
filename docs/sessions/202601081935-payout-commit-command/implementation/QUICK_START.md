# Quick Start Guide: Payout Commit Implementation

## Overview

This guide provides a quick reference for developers implementing the `?payout commit` feature.

**Command**: `?payout commit <month> <batch>`
**Purpose**: Commit contractor payables from Pending to Paid status with cascade updates
**Repositories**: fortress-api, fortress-discord

---

## Architecture at a Glance

```
Discord User
    ↓
?payout commit 2025-01 15
    ↓
fortress-discord (Command)
    ↓
fortress-api (REST API)
    ↓
Notion API (4 databases)
```

### Data Flow

```
1. User issues command → Discord command validates args
2. Command calls API preview endpoint → Get pending payables
3. API filters by PayDay (from Service Rate) → Show confirmation
4. User clicks Confirm → Command calls API commit endpoint
5. API executes cascade updates:
   └─ For each Contractor Payable:
      ├─ For each Payout Item:
      │  ├─ Update Invoice Split (if exists) → Paid
      │  ├─ Update Refund Request (if exists) → Paid
      │  └─ Update Payout Item → Paid
      └─ Update Contractor Payable → Paid (with Payment Date)
6. API returns success/failure counts → Discord shows result
```

---

## Critical Implementation Details

### 1. Property Type Differences (MUST GET RIGHT)

**Invoice Split is DIFFERENT from all others:**

```go
// Invoice Split - uses Select type
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Select: &nt.SelectOptions{Name: "Paid"}, // ← Select
        },
    },
}

// All Others - use Status type
params := nt.UpdatePageParams{
    DatabasePageProperties: nt.DatabasePageProperties{
        "Status": nt.DatabasePageProperty{
            Status: &nt.SelectOptions{Name: "Paid"}, // ← Status
        },
    },
}
```

| Database | Property Name | Property Type |
|----------|--------------|---------------|
| Contractor Payables | Payment Status | **Status** |
| Contractor Payouts | Status | **Status** |
| Invoice Split | Status | **Select** ← DIFFERENT! |
| Refund Request | Status | **Status** |

**Getting this wrong will cause Notion API to reject your updates!**

---

### 2. Cascade Update Sequence

**Order matters for data consistency:**

```
1. Update Invoice Split → Paid
2. Update Refund Request → Paid
3. Update Payout Item → Paid
4. Update Contractor Payable → Paid (with Payment Date)
```

**Why this order?**
- Leaf nodes first (Invoice Split, Refund)
- Parent nodes last (Payout Item, Payable)
- Payment Date set last as final confirmation

---

### 3. Error Handling Strategy

**Best-Effort Approach:**

```go
// Continue on error, track failures
for _, payable := range payables {
    if err := commitSinglePayable(payable); err != nil {
        failCount++
        errors = append(errors, err)
        // Don't return - continue processing
    } else {
        successCount++
    }
}

return &CommitResponse{
    Updated: successCount,
    Failed:  failCount,
    Errors:  errors,
}
```

**Key Points:**
- Don't stop on first error
- Track all successes and failures
- Return detailed error information
- Log everything at DEBUG level

---

### 4. PayDay Filtering Logic

**How to filter by batch (PayDay):**

```go
// 1. Query ALL pending payables for period
payables := QueryPendingPayablesByPeriod(ctx, "2025-01-01")

// 2. For each payable, get contractor's PayDay
for _, payable := range payables {
    payDay := GetContractorPayDay(ctx, payable.ContractorPageID)

    // 3. Filter by matching batch
    if payDay == batch {
        filtered = append(filtered, payable)
    }
}
```

**Why?**
- Contractors have assigned PayDay (1 or 15) in Service Rate
- Batch parameter determines which bi-monthly group to process
- This ensures correct contractors are included

---

## File Organization

### fortress-api

```
pkg/
├── service/notion/
│   ├── contractor_payables.go   [MODIFY - 3 new methods]
│   ├── contractor_payouts.go     [MODIFY - 2 new methods]
│   ├── invoice_split.go          [MODIFY - 1 new method]
│   └── refund_requests.go        [MODIFY - 1 new method]
├── controller/contractorpayables/
│   ├── interface.go              [CREATE]
│   └── contractorpayables.go     [CREATE]
├── handler/contractorpayables/
│   ├── interface.go              [CREATE]
│   ├── request.go                [CREATE]
│   └── contractorpayables.go     [CREATE]
└── routes/
    └── v1.go                     [MODIFY - register routes]
```

### fortress-discord

```
pkg/
├── model/
│   └── payout.go                 [CREATE]
├── adapter/fortress/
│   └── payout.go                 [CREATE]
├── discord/
│   ├── service/payout/
│   │   ├── interface.go          [CREATE]
│   │   └── service.go            [CREATE]
│   ├── view/payout/
│   │   └── payout.go             [CREATE]
│   └── command/payout/
│       ├── command.go            [CREATE]
│       └── new.go                [CREATE]
```

---

## API Endpoints

### Preview Commit
```
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=15

Response 200 OK:
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "count": 3,
    "total_amount": 15000.00,
    "contractors": [
      {"name": "John Doe", "amount": 5000.00, "currency": "USD", "payable_id": "page-1"}
    ]
  }
}
```

### Execute Commit
```
POST /api/v1/contractor-payables/commit
Content-Type: application/json

{
  "month": "2025-01",
  "batch": 15
}

Response 200 OK (full success):
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "updated": 3,
    "failed": 0
  }
}

Response 207 Multi-Status (partial failure):
{
  "data": {
    "month": "2025-01",
    "batch": 15,
    "updated": 2,
    "failed": 1,
    "errors": [
      {"payable_id": "page-3", "error": "network timeout"}
    ]
  }
}
```

---

## Discord Command Flow

### User Experience

```
User: ?payout commit 2025-01 15

Bot: [Embed - Orange]
     Confirm Payout Commit

     Month: 2025-01
     Batch: 15
     Count: 3 payables
     Total: $15,000.00

     Contractors:
     • John Doe - $5,000.00
     • Jane Smith - $7,500.00
     • Bob Wilson - $2,500.00

     [Confirm] [Cancel]

User: [clicks Confirm]

Bot: [Embed - Green]
     Payout Commit Successful

     Successfully committed 3 payables for 2025-01 batch 15
```

---

## Testing Strategy

### Unit Tests Priority

1. **Property Type Tests** (CRITICAL)
   - Verify Invoice Split uses Select
   - Verify others use Status
   - Prevent Notion API rejections

2. **Pagination Tests**
   - Handle >100 payables
   - Cursor-based iteration
   - Prevent data loss

3. **Cascade Update Tests**
   - Correct update sequence
   - Handle missing relations
   - Track success/failure counts

4. **Error Handling Tests**
   - Continue on individual failures
   - Aggregate error details
   - Idempotency (re-running is safe)

### Manual Testing Checklist

- [ ] `?payout help` shows usage
- [ ] Invalid month format rejected
- [ ] Invalid batch (not 1 or 15) rejected
- [ ] No pending payables shows count=0
- [ ] Preview shows correct contractors and amounts
- [ ] Cancel button dismisses confirmation
- [ ] Confirm executes commit with cascade updates
- [ ] Partial failures show error details
- [ ] Non-admin users blocked by permission check
- [ ] Re-running commit is safe (idempotent)

---

## Common Pitfalls to Avoid

### 1. Wrong Property Type
❌ **WRONG**: Using Status for Invoice Split
```go
"Status": {Status: &nt.SelectOptions{Name: "Paid"}} // Will fail!
```

✅ **CORRECT**: Using Select for Invoice Split
```go
"Status": {Select: &nt.SelectOptions{Name: "Paid"}} // Works!
```

### 2. Stopping on First Error
❌ **WRONG**: Returning immediately on error
```go
if err := update(payable); err != nil {
    return err // Stops processing!
}
```

✅ **CORRECT**: Continue processing, track failures
```go
if err := update(payable); err != nil {
    failCount++
    errors = append(errors, err)
    // Continue to next payable
}
```

### 3. Missing Pagination
❌ **WRONG**: Assuming results fit in one page
```go
resp, _ := client.QueryDatabase(ctx, dbID, query)
return resp.Results // Missing payables if >100!
```

✅ **CORRECT**: Handle pagination with cursor
```go
for {
    resp, _ := client.QueryDatabase(ctx, dbID, query)
    results = append(results, resp.Results...)
    if !resp.HasMore {
        break
    }
    query.StartCursor = *resp.NextCursor
}
```

### 4. Forgetting to Filter by PayDay
❌ **WRONG**: Returning all pending payables
```go
return QueryPendingPayablesByPeriod(ctx, period)
```

✅ **CORRECT**: Filter by contractor's PayDay
```go
payables := QueryPendingPayablesByPeriod(ctx, period)
filtered := []Payable{}
for _, p := range payables {
    payDay := GetContractorPayDay(ctx, p.ContractorPageID)
    if payDay == batch {
        filtered = append(filtered, p)
    }
}
return filtered
```

---

## Development Workflow

### Recommended Order

1. **Start with Notion Services** (Phase 1)
   - Can test independently with mocked Notion client
   - Foundation for everything else
   - 7 methods to implement

2. **Build Controller/Handler** (Phase 2)
   - Use mocked Notion services for testing
   - Test cascade update logic thoroughly
   - 2 API endpoints to create

3. **Implement Discord Command** (Phase 3)
   - Can mock API responses during development
   - Test button interactions
   - Create embeds and views

4. **Integration Testing** (Phase 4)
   - Test full flow end-to-end
   - Use test Notion workspace (not production!)
   - Verify idempotency

### Parallel Work Opportunities

- Notion service methods can be developed in parallel
- Discord models/adapters can start while API is in progress
- Unit tests can be written alongside implementation

---

## Quick Reference: Key Methods

### Notion Services (fortress-api)

```go
// contractor_payables.go
QueryPendingPayablesByPeriod(ctx, period) []PendingPayable
UpdatePayableStatus(ctx, pageID, status, paymentDate) error
GetContractorPayDay(ctx, contractorPageID) int

// contractor_payouts.go
GetPayoutWithRelations(ctx, payoutPageID) *PayoutWithRelations
UpdatePayoutStatus(ctx, pageID, status) error

// invoice_split.go
UpdateInvoiceSplitStatus(ctx, pageID, status) error  // Uses Select!

// refund_requests.go
UpdateRefundRequestStatus(ctx, pageID, status) error
```

### Controller Methods (fortress-api)

```go
PreviewCommit(ctx, month, batch) *PreviewCommitResponse
CommitPayables(ctx, month, batch) *CommitResponse
```

### Discord Methods (fortress-discord)

```go
// Service
PreviewCommit(month, batch) *PayoutPreview
ExecuteCommit(month, batch) *PayoutCommitResult

// View
Help(message)
NoPayables(message, month, batch)
ShowConfirmation(message, preview)
ShowResult(message, result)

// Command
Execute(message)
ExecuteCommitConfirmation(message, month, batch)
```

---

## Environment Setup

### fortress-api (.env)
```bash
NOTION_DATABASE_CONTRACTOR_PAYABLES=2c264b29-b84c-8037-807c-000bf6d0792c
NOTION_DATABASE_CONTRACTOR_PAYOUTS=2c564b29-b84c-8045-80ee-000bee2e3669
NOTION_DATABASE_INVOICE_SPLIT=2c364b29-b84c-804f-9856-000b58702dea
NOTION_DATABASE_REFUND_REQUEST=2cc64b29-b84c-8066-adf2-cc56171cedf4
NOTION_DATABASE_SERVICE_RATE=2c464b29-b84c-80cf-bef6-000b42bce15e
```

### fortress-discord (.env)
```bash
FORTRESS_API_URL=http://localhost:8080  # Local development
FORTRESS_API_KEY=your-api-key
```

---

## Need Help?

- **Full Task Breakdown**: See [tasks.md](./tasks.md)
- **Specifications**: See [../planning/specifications/](../planning/specifications/)
- **Test Cases**: See [../test-cases/](../test-cases/)
- **ADR**: See [../planning/ADRs/ADR-001-cascade-status-update.md](../planning/ADRs/ADR-001-cascade-status-update.md)
- **Requirements**: See [../requirements/requirements.md](../requirements/requirements.md)

---

## Ready to Start?

1. Read [tasks.md](./tasks.md) for detailed task breakdown
2. Set up environment variables
3. Start with Phase 1 (Notion services)
4. Write tests alongside implementation
5. Update [STATUS.md](./STATUS.md) as you complete tasks

**Good luck! 🚀**
