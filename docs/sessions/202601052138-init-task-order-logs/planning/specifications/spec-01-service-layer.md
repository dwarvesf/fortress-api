# Specification: Service Layer Changes

## Overview

Add method to create empty Line Items for initialization.

## Changes

### 1. Add `CreateEmptyTimesheetLineItem` method

**File**: `pkg/service/notion/task_order_log.go`

**Signature**:
```go
func (s *TaskOrderLogService) CreateEmptyTimesheetLineItem(ctx context.Context, orderID, deploymentID string, month string) (string, error)
```

**Description**: Creates a Timesheet Line Item with:
- Type: "Timesheet"
- Status: "Pending Approval"
- Date: Last day of month
- Deployment: Link to deployment
- Line Item Hours: 0
- Timesheet: Empty relation (no timesheets yet)
- Proof of Works: Empty

**Why new method**: The existing `CreateTimesheetLineItem` requires hours, proofOfWorks, and timesheetIDs. For initialization, we need a simpler version that creates an empty Line Item.

### 2. Existing Methods to Reuse

- `QueryActiveDeploymentsByMonth(ctx, month, contractorDiscord)` - Query active deployments
- `CheckOrderExistsByContractor(ctx, contractorID, month)` - Check if Order exists
- `CheckLineItemExists(ctx, orderID, deploymentID)` - Check if Line Item exists
- `CreateOrder(ctx, month)` - Create Order
- `addSubItemToOrder(ctx, orderID, lineItemID)` - Link Line Item to Order
