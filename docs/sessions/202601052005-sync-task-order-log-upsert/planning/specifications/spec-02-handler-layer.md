# Specification: Handler Layer Changes

## File: `pkg/handler/notion/task_order_log.go`

### 1. Modify `SyncTaskOrderLogs` Handler

#### Remove first deployment search

**Current code** (lines 127-143):
```go
// Step 3b: Get first project's deployment for Order
var firstDeploymentID string
for projectID := range projectGroups {
    deploymentID, err := taskOrderLogService.GetDeploymentByContractorAndProject(...)
    ...
    firstDeploymentID = deploymentID
    break
}
```

**Change**: Remove this block entirely. Order creation no longer needs deployment.

#### Update Order creation call

**Current** (line 155):
```go
orderID, err = taskOrderLogService.CreateOrder(ctx, firstDeploymentID, month)
```

**New**:
```go
orderID, err = taskOrderLogService.CreateOrder(ctx, month)
```

#### Add upsert logic for line items

**Current flow** (lines 229-249):
```go
lineItemExists, lineItemID, err := taskOrderLogService.CheckLineItemExists(ctx, orderID, deploymentID)
if !lineItemExists {
    // Create new
} else {
    // Just log and skip
}
```

**New flow**:
```go
lineItemExists, lineItemID, err := taskOrderLogService.CheckLineItemExists(ctx, orderID, deploymentID)
if err != nil {
    // handle error
}

if !lineItemExists {
    // Create new line item (existing logic)
    lineItemsCreated++
} else {
    // Fetch existing details
    existingDetails, err := taskOrderLogService.GetLineItemDetails(ctx, lineItemID)
    if err != nil {
        // handle error, continue
    }

    // Compare: hours changed OR timesheet count changed
    hoursChanged := existingDetails.Hours != hours
    timesheetsChanged := !equalStringSlices(existingDetails.TimesheetIDs, timesheetIDs)

    if hoursChanged || timesheetsChanged {
        l.Debug(fmt.Sprintf("line item changed: hours %.1f->%.1f, timesheets %d->%d",
            existingDetails.Hours, hours,
            len(existingDetails.TimesheetIDs), len(timesheetIDs)))

        // Update line item
        err = taskOrderLogService.UpdateTimesheetLineItem(ctx, lineItemID, orderID, hours, summarizedPoW, timesheetIDs)
        if err != nil {
            // handle error
        }
        lineItemsUpdated++
    } else {
        l.Debug("line item unchanged, skipping update")
    }
}
```

### 2. Add helper function

```go
// equalStringSlices compares two string slices for equality (order-independent)
func equalStringSlices(a, b []string) bool {
    if len(a) != len(b) {
        return false
    }
    aMap := make(map[string]bool)
    for _, v := range a {
        aMap[v] = true
    }
    for _, v := range b {
        if !aMap[v] {
            return false
        }
    }
    return true
}
```

### 3. Update response

Add `line_items_updated` to response:
```go
c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
    "month":                  month,
    "orders_created":         ordersCreated,
    "line_items_created":     lineItemsCreated,
    "line_items_updated":     lineItemsUpdated,  // NEW
    "contractors_processed":  contractorsProcessed,
    "details":                details,
}, nil, nil, nil, "ok"))
```
