# Implementation Tasks

## Overview

Implement Initialize Task Order Logs endpoint.

## Tasks

### Task 1: Add `CreateEmptyTimesheetLineItem` method ✅ COMPLETED

**File**: `pkg/service/notion/task_order_log.go` (lines 600-689)

**Description**: Add method to create empty Line Item for initialization

```go
func (s *TaskOrderLogService) CreateEmptyTimesheetLineItem(ctx context.Context, orderID, deploymentID string, month string) (string, error)
```

**Steps**:
1. ✅ Parse month to get target date (last day of month)
2. ✅ Create Timesheet page with:
   - Type: "Timesheet"
   - Status: "Pending Approval"
   - Date: Last day of month
   - Line Item Hours: 0
   - Deployment: Link to deploymentID
   - Timesheet: Empty relation
3. ✅ Call `addSubItemToOrder` to link Line Item to Order
4. ✅ Add DEBUG logging

---

### Task 2: Add `InitTaskOrderLogs` handler ✅ COMPLETED

**File**: `pkg/handler/notion/task_order_log.go` (lines 578-746)

**Description**: Add handler for initialization endpoint

**Steps**:
1. ✅ Parse and validate `month` query parameter
2. ✅ Call `QueryActiveDeploymentsByMonth` to get all active deployments
3. ✅ Group deployments by contractor using `groupDeploymentsByContractor`
4. ✅ For each contractor:
   - Check if Order exists via `CheckOrderExistsByContractor`
   - If not, create Order via `CreateOrder`
   - For each deployment:
     - Check if Line Item exists via `CheckLineItemExists`
     - If not, create via `CreateEmptyTimesheetLineItem`
5. ✅ Track counts: `ordersCreated`, `lineItemsCreated`, `deploymentsProcessed`, `skipped`
6. ✅ Return JSON response with counts and details
7. ✅ Add DEBUG logging throughout

---

### Task 3: Add route for endpoint ✅ COMPLETED

**File**: `pkg/routes/v1.go` (line 63)

**Description**: Register the new endpoint

**Steps**:
1. ✅ Add `POST /cronjobs/init-task-order-logs` route
2. ✅ Use existing cronjobs group with same middleware

**Additional**: Updated interface in `pkg/handler/notion/interface.go` (line 28)

---

### Task 4: Test compilation ✅ COMPLETED

**Description**: Verify code compiles

```bash
go build ./pkg/handler/notion/... ./pkg/service/notion/... ./pkg/routes/...
```

Result: Success - no errors

---

## Execution Order

1. ✅ Task 1 (service method)
2. ✅ Task 2 (handler)
3. ✅ Task 3 (route)
4. ✅ Task 4 (compilation test)

## Estimated Complexity

- Service layer: Low (simple method)
- Handler layer: Medium (iteration and tracking)
- Routes: Low (single line)
