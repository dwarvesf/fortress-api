# Implementation Tasks: Commission Payout Processing

## Task List

### Task 1: Add PendingCommissionSplit struct and QueryPendingCommissionSplits method
**File:** `pkg/service/notion/invoice_split.go`

**Changes:**
1. Add `PendingCommissionSplit` struct with fields: PageID, Name, Amount, Currency, Role, PersonPageID
2. Add `QueryPendingCommissionSplits(ctx context.Context) ([]PendingCommissionSplit, error)` method
3. Query filter: Status=Pending AND Type=Commission
4. Extract Person relation ID

**Dependencies:** None

---

### Task 2: Add CheckPayoutExistsByInvoiceSplit method
**File:** `pkg/service/notion/contractor_payouts.go`

**Changes:**
1. Add `CheckPayoutExistsByInvoiceSplit(ctx context.Context, invoiceSplitPageID string) (bool, string, error)` method
2. Query by "Invoice Split" relation contains invoiceSplitPageID

**Dependencies:** None

---

### Task 3: Add CreateCommissionPayoutInput struct and CreateCommissionPayout method
**File:** `pkg/service/notion/contractor_payouts.go`

**Changes:**
1. Add `CreateCommissionPayoutInput` struct
2. Add `CreateCommissionPayout(ctx context.Context, input CreateCommissionPayoutInput) (string, error)` method
3. Set properties: Name, Amount, Currency, Person, Invoice Split, Type=Commission, Direction=Outgoing, Status=Pending

**Dependencies:** None

---

### Task 4: Add processCommissionPayouts handler
**File:** `pkg/handler/notion/contractor_payouts.go`

**Changes:**
1. Add `processCommissionPayouts(c *gin.Context, l logger.Logger, payoutType string)` method
2. Replace `case "commission"` stub with call to `processCommissionPayouts`
3. Implement processing loop with idempotency check
4. Return JSON response matching existing format

**Dependencies:** Tasks 1, 2, 3

---

### Task 5: Build verification
**Command:** `go build ./...`

**Dependencies:** Tasks 1-4

---

## Execution Order

```
Task 1 ──┐
Task 2 ──┼──► Task 4 ──► Task 5
Task 3 ──┘
```

Tasks 1, 2, 3 can be done in parallel. Task 4 depends on all three. Task 5 is final verification.
