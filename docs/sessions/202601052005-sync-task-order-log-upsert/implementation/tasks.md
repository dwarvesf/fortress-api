# Implementation Tasks

## Overview

Implement upsert logic for Sync Task Order Log endpoint.

## Tasks

### Task 1: Add `LineItemDetails` struct [DONE]

**File**: `pkg/service/notion/task_order_log.go`

**Description**: Add struct to hold line item data for comparison

```go
type LineItemDetails struct {
    PageID       string
    Hours        float64
    TimesheetIDs []string
    Status       string
}
```

---

### Task 2: Implement `GetLineItemDetails` method [DONE]

**File**: `pkg/service/notion/task_order_log.go`

**Description**: Fetch existing line item data from Notion

**Steps**:
1. Query Notion page by ID using `s.client.FindPageByID`
2. Extract `Line Item Hours` number property
3. Extract `Timesheet` relation IDs
4. Extract `Status` select property
5. Return `LineItemDetails` struct
6. Add DEBUG logging

---

### Task 3: Implement `UpdateTimesheetLineItem` method [DONE]

**File**: `pkg/service/notion/task_order_log.go`

**Description**: Update existing line item with new data

**Steps**:
1. Build update properties map
2. Set `Line Item Hours` to new value
3. Set `Proof of Works` to new summarized text
4. Set `Timesheet` relation to new IDs
5. Set `Status` to "Pending Approval"
6. Call Notion UpdatePage API
7. Call `UpdateOrderStatus` for parent Order
8. Add DEBUG logging

---

### Task 4: Modify `CreateOrder` to remove Deployment [DONE]

**File**: `pkg/service/notion/task_order_log.go`

**Description**: Remove Deployment parameter and relation

**Steps**:
1. Remove `deploymentID` parameter from function signature
2. Remove Deployment relation from properties map
3. Update function documentation

---

### Task 5: Add `equalStringSlices` helper [DONE]

**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Add helper function for slice comparison

```go
func equalStringSlices(a, b []string) bool
```

---

### Task 6: Update handler - Remove first deployment search [DONE]

**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Remove the block that searches for first deployment for Order

**Lines to remove**: 127-143 (approximately)

---

### Task 7: Update handler - Modify Order creation call [DONE]

**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Update CreateOrder call to remove deployment parameter

**Change**:
```go
// Before
orderID, err = taskOrderLogService.CreateOrder(ctx, firstDeploymentID, month)
// After
orderID, err = taskOrderLogService.CreateOrder(ctx, month)
```

---

### Task 8: Update handler - Add upsert logic [DONE]

**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Replace skip logic with compare/update logic

**Steps**:
1. When line item exists, fetch details via `GetLineItemDetails`
2. Compare hours and timesheet IDs
3. If changed, call `UpdateTimesheetLineItem`
4. Track `lineItemsUpdated` counter
5. Add DEBUG logging for change detection

---

### Task 9: Update handler - Add response field [DONE]

**File**: `pkg/handler/notion/task_order_log.go`

**Description**: Add `line_items_updated` to response

---

### Task 10: Update tests (if exist) [SKIPPED - no tests exist]

**File**: `pkg/handler/notion/*_test.go`

**Description**: Update any existing tests for the modified functions

---

## Execution Order

1. Task 1 (struct)
2. Task 2 (GetLineItemDetails)
3. Task 3 (UpdateTimesheetLineItem)
4. Task 4 (CreateOrder modification)
5. Task 5 (helper function)
6. Task 6-9 (handler changes - can be done together)
7. Task 10 (tests)

## Estimated Complexity

- Service layer: Medium (new methods + modification)
- Handler layer: Low-Medium (logic changes)
- Tests: Low (if tests exist)
